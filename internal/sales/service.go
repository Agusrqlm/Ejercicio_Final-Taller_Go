package sales

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Error para transiciones inválidas
var ErrInvalidTransition = errors.New("invalid status transition")

// Error para estados inválidos
var ErrInvalidStatus = errors.New("invalid status value")

// Service provides high-level sales management operations on a Storage backend.
type Service struct {
	storage    Storage
	logger     *zap.Logger
	userAPIURL string // URL base de la API de usuarios
}

// NewService creates a new Sales Service.
func NewService(storage Storage, logger *zap.Logger, userAPIURL string) *Service {
	if logger == nil {
		logger, _ = zap.NewProduction()
		defer logger.Sync()
	}
	return &Service{
		storage:    storage,
		logger:     logger,
		userAPIURL: userAPIURL,
	}
}

// CreateSale handles the creation of a new sale.
func (s *Service) CreateSale(userID string, amount float64) (*Sale, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	// Validar que el usuario existe llamando a la API de usuarios
	userExists, err := s.validateUser(userID)
	if err != nil {
		s.logger.Error("error validating user", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("error validating user: %w", err)
	}
	if !userExists {
		return nil, fmt.Errorf("user with ID '%s' not found", userID)
	}

	sale := &Sale{
		ID:        uuid.NewString(),
		UserID:    userID,
		Amount:    amount,
		Status:    getRandomStatus(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	if err := s.storage.Set(sale); err != nil {
		s.logger.Error("failed to save sale", zap.String("sale_id", sale.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to save sale: %w", err)
	}

	s.logger.Info("sale created", zap.String("sale_id", sale.ID), zap.Any("sale", sale))
	return sale, nil
}

func (s *Service) validateUser(userID string) (bool, error) {
	url := fmt.Sprintf("%s/users/%s", s.userAPIURL, userID)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("error making request to user API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return false, fmt.Errorf("user API returned unexpected status: %d", resp.StatusCode)
	}
}

func getRandomStatus() string {
	statuses := []string{"pending", "approved", "rejected"}
	randomIndex := rand.Intn(len(statuses))
	return statuses[randomIndex]
}

// Modificar el estado de una venta
func (s *Service) UpdateSaleStatus(saleID, newStatus string) (*Sale, error) {
	sale, err := s.storage.Read(saleID)
	if err != nil {
		return nil, ErrNotFound
	}

	if newStatus != "approved" && newStatus != "rejected" {
		return nil, ErrInvalidStatus

	}

	if sale.Status != "pending" {
		return nil, ErrInvalidTransition
	}

	sale.Status = newStatus
	sale.UpdatedAt = time.Now()
	sale.Version++

	if err := s.storage.Set(sale); err != nil {
		s.logger.Error("failed to update sale", zap.String("sale_id", sale.ID), zap.Error(err))
		return nil, err
	}

	return sale, nil
}
