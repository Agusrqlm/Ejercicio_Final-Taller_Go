package sales

import "errors"

// ErrNotFound is returned when a sale with the given ID is not found.
var ErrNotFound = errors.New("sale not found")

// ErrEmptyID is returned when trying to store a sale with an empty ID.
var ErrEmptyID = errors.New("empty sale ID")

// Storage is the main interface for our sales storage layer.
type Storage interface {
	Set(sale *Sale) error
	Read(id string) (*Sale, error) // Aunque no se pide explícitamente ahora, puede ser útil
	GetAll() ([]*Sale, error)
	// Update(sale *Sale) error     // Podríamos necesitar esto en el futuro
	// Delete(id string) error     // Podríamos necesitar esto en el futuro
}

// LocalStorage provides an in-memory implementation for storing sales.
type LocalStorage struct {
	m map[string]*Sale
}

// NewLocalStorage instantiates a new LocalStorage for sales with an empty map.
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		m: map[string]*Sale{},
	}
}

// Set stores a sale in the local storage.
// Returns ErrEmptyID if the sale has an empty ID.
func (l *LocalStorage) Set(sale *Sale) error {
	if sale.ID == "" {
		return ErrEmptyID
	}
	l.m[sale.ID] = sale
	return nil
}

// Read retrieves a sale from the local storage by ID.
// Returns ErrNotFound if the sale is not found.
func (l *LocalStorage) Read(id string) (*Sale, error) {
	s, ok := l.m[id]
	if !ok {
		return nil, ErrNotFound
	}
	return s, nil
}

// GetAll retrieves all sales from the local storage. <-- ¡NUEVA IMPLEMENTACIÓN!
func (l *LocalStorage) GetAll() ([]*Sale, error) {
	sales := make([]*Sale, 0, len(l.m))
	for _, s := range l.m {
		sales = append(sales, s)
	}
	return sales, nil
}

// // Update updates a sale in the local storage.
// // Returns ErrNotFound if the sale does not exist.
// func (l *LocalStorage) Update(sale *Sale) error {
// 	if _, ok := l.m[sale.ID]; !ok {
// 		return ErrNotFound
// 	}
// 	l.m[sale.ID] = sale
// 	return nil
// }

// // Delete removes a sale from the local storage by ID.
// // Returns ErrNotFound if the sale does not exist.
// func (l *LocalStorage) Delete(id string) error {
// 	if _, ok := l.m[id]; !ok {
// 		return ErrNotFound
// 	}
// 	delete(l.m, id)
// 	return nil
// }
