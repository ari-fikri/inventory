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
	ErrSameWarehouse     = errors.New("source and target warehouses cannot be the same")
)

const (
	MovementTypeIn          = "IN"
	MovementTypeOut         = "OUT"
	MovementTypeAdjust      = "ADJUST"
	MovementTypeTransferOut = "TRANSFER_OUT"
	MovementTypeTransferIn  = "TRANSFER_IN"
)

type StockLevel struct {
	ID               string
	TenantID         string
	ItemID           string
	WarehouseID      string
	Quantity         int
	ReservedQuantity int
	UpdatedAt        time.Time
}

type StockMovement struct {
	ID           string
	TenantID     string
	ItemID       string
	WarehouseID  string
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
	WarehouseID  string
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
	if params.Quantity <= 0 && params.MovementType != MovementTypeAdjust {
		return nil, ErrInvalidQuantity
	}
	if params.Quantity < 0 {
		return nil, ErrInvalidQuantity
	}
	if params.BalanceAfter < 0 {
		return nil, ErrInsufficientStock
	}

	wh := strings.TrimSpace(params.WarehouseID)
	if wh == "" {
		wh = "default"
	}

	return &StockMovement{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
		ItemID:       params.ItemID,
		WarehouseID:  wh,
		MovementType: params.MovementType,
		Quantity:     params.Quantity,
		BalanceAfter: params.BalanceAfter,
		Reference:    strings.TrimSpace(params.Reference),
		Notes:        strings.TrimSpace(params.Notes),
		CreatedAt:    time.Now().UTC(),
	}, nil
}

type StockTransferResult struct {
	ItemID            string
	TenantID          string
	SourceWarehouseID string
	TargetWarehouseID string
	Quantity          int
	SourceBalance     int
	TargetBalance     int
	OutMovement       *StockMovement
	InMovement        *StockMovement
}
