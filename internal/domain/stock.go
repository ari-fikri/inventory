package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
	ErrInsufficientStock = errors.New("insufficient stock for fulfillment")
	ErrStockNotFound     = errors.New("stock level not found")
)

const (
	MovementTypeIn     = "IN"
	MovementTypeOut    = "OUT"
	MovementTypeAdjust = "ADJUST"
)

type StockLevel struct {
	ID               string
	TenantID         string
	ItemID           string
	Quantity         int
	ReservedQuantity int
	UpdatedAt        time.Time
}

type StockMovement struct {
	ID           string
	TenantID     string
	ItemID       string
	MovementType string
	Quantity     int
	BalanceAfter int
	Reference    string
	Notes        string
	CreatedAt    time.Time
}

type NewStockMovementParams struct {
	TenantID     string
	ItemID       string
	MovementType string
	Quantity     int
	BalanceAfter int
	Reference    string
	Notes        string
}

func NewStockMovement(params NewStockMovementParams) (*StockMovement, error) {
	tenantID := strings.TrimSpace(params.TenantID)
	if tenantID == "" {
		return nil, ErrEmptyTenantID
	}
	if strings.TrimSpace(params.ItemID) == "" {
		return nil, errors.New("item ID cannot be empty")
	}
	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if params.BalanceAfter < 0 {
		return nil, ErrInsufficientStock
	}

	return &StockMovement{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
		ItemID:       params.ItemID,
		MovementType: params.MovementType,
		Quantity:     params.Quantity,
		BalanceAfter: params.BalanceAfter,
		Reference:    strings.TrimSpace(params.Reference),
		Notes:        strings.TrimSpace(params.Notes),
		CreatedAt:    time.Now().UTC(),
	}, nil
}
