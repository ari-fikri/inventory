package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyTenantID     = errors.New("tenant ID cannot be empty")
	ErrEmptySKU          = errors.New("SKU cannot be empty")
	ErrEmptyName         = errors.New("name cannot be empty")
	ErrNegativeUnitPrice = errors.New("unit price cannot be negative")
	ErrNegativeThreshold = errors.New("min threshold cannot be negative")
	ErrItemAlreadyExists = errors.New("item with this SKU already exists for this tenant")
	ErrItemNotFound      = errors.New("item not found")
)

type Item struct {
	ID           string
	TenantID     string
	SKU          string
	Name         string
	Description  string
	UnitPrice    float64
	MinThreshold int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LowStockItem struct {
	Item            *Item
	CurrentQuantity int
}

type NewItemParams struct {
	TenantID     string
	SKU          string
	Name         string
	Description  string
	UnitPrice    float64
	MinThreshold int
}

func NewItem(params NewItemParams) (*Item, error) {
	tenantID := strings.TrimSpace(params.TenantID)
	if tenantID == "" {
		return nil, ErrEmptyTenantID
	}

	sku := strings.TrimSpace(params.SKU)
	if sku == "" {
		return nil, ErrEmptySKU
	}

	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, ErrEmptyName
	}

	if params.UnitPrice < 0 {
		return nil, ErrNegativeUnitPrice
	}

	if params.MinThreshold < 0 {
		return nil, ErrNegativeThreshold
	}

	now := time.Now().UTC()
	return &Item{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
		SKU:          sku,
		Name:         name,
		Description:  strings.TrimSpace(params.Description),
		UnitPrice:    params.UnitPrice,
		MinThreshold: params.MinThreshold,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
