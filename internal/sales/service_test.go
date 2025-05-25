package sales

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest" // Para un logger de prueba
)

// Mock para la interfaz Storage
// Aunque ya tienes LocalStorage, es bueno entender cómo se haría un mock si LocalStorage no fuera suficiente.
// Para este caso, LocalStorage es perfecto como "fake" storage.

// TestNewService verifica la inicialización del servicio.
func TestNewService(t *testing.T) {
	mockStorage := NewLocalStorage() // Usamos tu LocalStorage como mock in-memory
	logger := zaptest.NewLogger(t)   // Logger para pruebas
	userAPIURL := "http://localhost:8080"

	svc := NewService(mockStorage, logger, userAPIURL)

	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.storage == nil {
		t.Error("Service storage was not initialized")
	}
	if svc.logger == nil {
		t.Error("Service logger was not initialized")
	}
	if svc.userAPIURL != userAPIURL {
		t.Errorf("Service userAPIURL mismatch: got %s, want %s", svc.userAPIURL, userAPIURL)
	}
}

// TestCreateSale_Success prueba la creación exitosa de una venta.
func TestCreateSale_Success(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)

	// Configurar un servidor de prueba para la API de usuarios
	// Este servidor mockeará la respuesta de la API de usuarios.
	mockUserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/test-user-id" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK) // Usuario encontrado
			return
		}
		w.WriteHeader(http.StatusNotFound) // Cualquier otra ruta o usuario no encontrado
	}))
	defer mockUserServer.Close() // Cierra el servidor mock al finalizar la prueba

	svc := NewService(mockStorage, logger, mockUserServer.URL) // Usamos la URL del servidor mock

	userID := "test-user-id"
	amount := 150.75

	sale, err := svc.CreateSale(userID, amount)
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}

	if sale == nil {
		t.Fatal("Created sale is nil")
	}
	if sale.ID == "" {
		t.Error("Sale ID is empty")
	}
	if sale.UserID != userID {
		t.Errorf("Sale UserID mismatch: got %s, want %s", sale.UserID, userID)
	}
	if sale.Amount != amount {
		t.Errorf("Sale Amount mismatch: got %f, want %f", sale.Amount, amount)
	}
	if sale.Status != "pending" && sale.Status != "approved" && sale.Status != "rejected" {
		t.Errorf("Sale Status is invalid: %s", sale.Status)
	}

	// Verificar que la venta se guardó en el storage
	storedSale, err := mockStorage.Read(sale.ID)
	if err != nil {
		t.Fatalf("Failed to read stored sale: %v", err)
	}
	if storedSale.ID != sale.ID {
		t.Errorf("Stored sale ID mismatch: got %s, want %s", storedSale.ID, sale.ID)
	}
}

// TestCreateSale_InvalidAmount prueba la creación con un monto inválido.
func TestCreateSale_InvalidAmount(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)
	// No necesitamos un servidor mock de usuarios para esta prueba ya que fallará antes.
	svc := NewService(mockStorage, logger, "http://dummyurl")

	userID := "test-user-id"
	amount := 0.0 // Monto inválido

	sale, err := svc.CreateSale(userID, amount)
	if err == nil {
		t.Fatal("CreateSale expected an error for invalid amount, got none")
	}
	if sale != nil {
		t.Error("CreateSale returned a sale for invalid amount, expected nil")
	}
	expectedErr := "amount must be greater than zero"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestCreateSale_UserNotFound prueba la creación cuando el usuario no existe.
func TestCreateSale_UserNotFound(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)

	mockUserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // Siempre responde 404
	}))
	defer mockUserServer.Close()

	svc := NewService(mockStorage, logger, mockUserServer.URL)

	userID := "non-existent-user"
	amount := 100.0

	sale, err := svc.CreateSale(userID, amount)
	if err == nil {
		t.Fatal("CreateSale expected an error for user not found, got none")
	}
	if sale != nil {
		t.Error("CreateSale returned a sale, expected nil")
	}
	expectedErr := "user with ID 'non-existent-user' not found"
	if err.Error() != expectedErr {
		t.Errorf("Expected error containing '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestSearchSale_Success prueba la búsqueda exitosa de ventas.
func TestSearchSale_Success(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)

	// Cargar algunas ventas de prueba en el storage
	s1 := &Sale{ID: "s1", UserID: "user1", Amount: 100, Status: "approved", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s2 := &Sale{ID: "s2", UserID: "user1", Amount: 200, Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s3 := &Sale{ID: "s3", UserID: "user2", Amount: 50, Status: "rejected", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s4 := &Sale{ID: "s4", UserID: "user1", Amount: 150, Status: "approved", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockStorage.Set(s1)
	mockStorage.Set(s2)
	mockStorage.Set(s3)
	mockStorage.Set(s4)

	mockUserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/user1" || r.URL.Path == "/users/user2" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockUserServer.Close()

	svc := NewService(mockStorage, logger, mockUserServer.URL)

	tests := []struct {
		name         string
		userID       string
		status       string
		expectedLen  int
		expectedMeta SalesMetadata
	}{
		{
			name:        "Search all for user1",
			userID:      "user1",
			status:      "",
			expectedLen: 3,
			expectedMeta: SalesMetadata{
				Quantity: 3, Approved: 2, Rejected: 0, Pending: 1, TotalAmount: 450,
			},
		},
		{
			name:        "Search pending for user1",
			userID:      "user1",
			status:      "pending",
			expectedLen: 1,
			expectedMeta: SalesMetadata{
				Quantity: 1, Approved: 0, Rejected: 0, Pending: 1, TotalAmount: 200,
			},
		},
		{
			name:        "Search all for user2",
			userID:      "user2",
			status:      "",
			expectedLen: 1,
			expectedMeta: SalesMetadata{
				Quantity: 1, Approved: 0, Rejected: 1, Pending: 0, TotalAmount: 50,
			},
		},
		{
			name:         "Search non-existent user",
			userID:       "user3",
			status:       "",
			expectedLen:  0, // No sales found, but user validation should fail
			expectedMeta: SalesMetadata{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sales, metadata, err := svc.SearchSale(tt.userID, tt.status)

			if tt.userID == "user3" { // Special case for user not found
				if err == nil {
					t.Fatalf("SearchSale expected error for non-existent user, got none")
				}
				expectedErr := "user with ID 'user3' not found"
				if err.Error() != expectedErr {
					t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
				}
				return // Skip further checks for this case
			}

			if err != nil {
				t.Fatalf("SearchSale failed: %v", err)
			}
			if len(sales) != tt.expectedLen {
				t.Errorf("SearchSale result count mismatch: got %d, want %d", len(sales), tt.expectedLen)
			}
			if metadata.Quantity != tt.expectedMeta.Quantity {
				t.Errorf("Metadata Quantity mismatch: got %d, want %d", metadata.Quantity, tt.expectedMeta.Quantity)
			}
			if metadata.Approved != tt.expectedMeta.Approved {
				t.Errorf("Metadata Approved mismatch: got %d, want %d", metadata.Approved, tt.expectedMeta.Approved)
			}
			if metadata.Rejected != tt.expectedMeta.Rejected {
				t.Errorf("Metadata Rejected mismatch: got %d, want %d", metadata.Rejected, tt.expectedMeta.Rejected)
			}
			if metadata.Pending != tt.expectedMeta.Pending {
				t.Errorf("Metadata Pending mismatch: got %d, want %d", metadata.Pending, tt.expectedMeta.Pending)
			}
			if metadata.TotalAmount != tt.expectedMeta.TotalAmount {
				t.Errorf("Metadata TotalAmount mismatch: got %f, want %f", metadata.TotalAmount, tt.expectedMeta.TotalAmount)
			}
		})
	}
}

// TestSearchSale_InvalidStatus prueba la búsqueda con un estado inválido.
func TestSearchSale_InvalidStatus(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)

	mockUserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Suponemos que el usuario existe para esta prueba
	}))
	defer mockUserServer.Close()

	svc := NewService(mockStorage, logger, mockUserServer.URL)

	userID := "user1"
	invalidStatus := "invalid"

	_, _, err := svc.SearchSale(userID, invalidStatus)
	if err == nil {
		t.Fatal("SearchSale expected an error for invalid status, got none")
	}
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("Expected ErrInvalidStatus, got %v", err)
	}
}

// TestUpdateSaleStatus_Success prueba la actualización exitosa del estado de una venta.
func TestUpdateSaleStatus_Success(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)

	// Pre-crear una venta en estado pendiente
	saleID := "test-sale-to-update"
	initialSale := &Sale{
		ID:        saleID,
		UserID:    "user1",
		Amount:    100.0,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}
	mockStorage.Set(initialSale)

	svc := NewService(mockStorage, logger, "http://dummyurl") // No user API needed for this test

	updatedSale, err := svc.UpdateSaleStatus(saleID, "approved")
	if err != nil {
		t.Fatalf("UpdateSaleStatus failed: %v", err)
	}

	if updatedSale == nil {
		t.Fatal("Updated sale is nil")
	}
	if updatedSale.Status != "approved" {
		t.Errorf("Sale status not updated: got %s, want %s", updatedSale.Status, "approved")
	}
	if updatedSale.Version != 2 {
		t.Errorf("Sale version not incremented: got %d, want %d", updatedSale.Version, 2)
	}

	// Verificar que el estado se actualizó en el storage
	storedSale, _ := mockStorage.Read(saleID)
	if storedSale.Status != "approved" {
		t.Errorf("Stored sale status not updated: got %s, want %s", storedSale.Status, "approved")
	}
}

// TestUpdateSaleStatus_NotFound prueba la actualización de una venta no existente.
func TestUpdateSaleStatus_NotFound(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)
	svc := NewService(mockStorage, logger, "http://dummyurl")

	_, err := svc.UpdateSaleStatus("non-existent-sale", "approved")
	if err == nil {
		t.Fatal("UpdateSaleStatus expected ErrNotFound, got none")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

// TestUpdateSaleStatus_InvalidNewStatus prueba la actualización con un estado nuevo inválido.
func TestUpdateSaleStatus_InvalidNewStatus(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)

	saleID := "test-sale-to-update"
	initialSale := &Sale{ID: saleID, UserID: "user1", Amount: 100, Status: "pending", Version: 1}
	mockStorage.Set(initialSale)

	svc := NewService(mockStorage, logger, "http://dummyurl")

	_, err := svc.UpdateSaleStatus(saleID, "invalid_status")
	if err == nil {
		t.Fatal("UpdateSaleStatus expected ErrInvalidStatus, got none")
	}
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("Expected ErrInvalidStatus, got %v", err)
	}
}

// TestUpdateSaleStatus_InvalidTransition prueba una transición de estado inválida.
func TestUpdateSaleStatus_InvalidTransition(t *testing.T) {
	mockStorage := NewLocalStorage()
	logger := zaptest.NewLogger(t)

	saleID := "test-sale-to-update"
	initialSale := &Sale{ID: saleID, UserID: "user1", Amount: 100, Status: "approved", Version: 1} // Already approved
	mockStorage.Set(initialSale)

	svc := NewService(mockStorage, logger, "http://dummyurl")

	_, err := svc.UpdateSaleStatus(saleID, "rejected") // Try to change from approved to rejected
	if err == nil {
		t.Fatal("UpdateSaleStatus expected ErrInvalidTransition, got none")
	}
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("Expected ErrInvalidTransition, got %v", err)
	}
}

// ----- Prueba para la función interna getRandomStatus -----
// Es una función no exportada, pero como TestRandomStatus está en el mismo paquete, podemos probarla.
func TestGetRandomStatus(t *testing.T) {
	// Ejecuta la función varias veces para asegurar que se generan los estados esperados
	statusCounts := make(map[string]int)
	numIterations := 1000

	for i := 0; i < numIterations; i++ {
		status := getRandomStatus() // Llamada a la función no exportada
		statusCounts[status]++
	}

	if statusCounts["pending"] == 0 || statusCounts["approved"] == 0 || statusCounts["rejected"] == 0 {
		t.Errorf("Not all statuses were generated. Counts: %v", statusCounts)
	}

	// Esto es una prueba de "probabilidad", no determinista.
	// Si bien no podemos asegurar que todas las veces generará los 3, después de 1000 iteraciones
	// debería haber generado al menos uno de cada.
	if len(statusCounts) != 3 {
		t.Errorf("Expected 3 unique statuses, got %d. Counts: %v", len(statusCounts), statusCounts)
	}
}

// ----- Prueba para la función interna validateUser -----
// Como es una función no exportada, solo puede ser probada dentro de este paquete.
func TestValidateUser(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name          string
		userID        string
		statusCode    int // El código de estado que simulará el servidor mock
		expectedValid bool
		expectedErr   bool
	}{
		{"User Exists", "user-exists", http.StatusOK, true, false},
		{"User Not Found", "user-not-found", http.StatusNotFound, false, false},
		{"Internal Server Error", "server-error", http.StatusInternalServerError, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configurar un servidor de prueba para la API de usuarios para cada caso
			mockUserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer mockUserServer.Close()

			svc := NewService(nil, logger, mockUserServer.URL) // Storage no es relevante aquí

			valid, err := svc.validateUser(tt.userID) // Llamada a la función no exportada

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
			if valid != tt.expectedValid {
				t.Errorf("Expected valid: %t, got: %t", tt.expectedValid, valid)
			}
		})
	}

	// Caso de error en la petición HTTP (ej. URL inválida o red)
	t.Run("HTTP Request Error", func(t *testing.T) {
		svc := NewService(nil, logger, "http://invalid-url-that-does-not-exist:12345")
		_, err := svc.validateUser("any-user")
		if err == nil {
			t.Fatal("Expected an error for HTTP request failure, got none")
		}
		// Podemos verificar el mensaje de error o el tipo de error si queremos ser más específicos
	})
}
