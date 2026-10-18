package port

import (
	"context"

	"new-inventory/internal/domain"
)

type StockInCommand struct {
	TenantID    string
	SKU         string
	WarehouseID string
	Quantity    int
	Reference   string
	Notes       string
}

type StockOutCommand struct {
	TenantID    string
	SKU         string
	WarehouseID string
	Quantity    int
	Reference   string
	Notes       string
}

type StockAdjustCommand struct {
	TenantID       string
	SKU            string
	WarehouseID    string
	TargetQuantity int
	Reference      string
	Notes          string
}

type StockRepository interface {
	GetStock(ctx context.Context, tenantID, itemID, warehouseID string) (*domain.StockLevel, error)
	RecordMovement(ctx context.Context, movement *domain.StockMovement, newBalance int) error
	ListMovements(ctx context.Context, tenantID, itemID string) ([]*domain.StockMovement, error)
}

type StockService interface {
	StockIn(ctx context.Context, cmd StockInCommand) (*domain.StockMovement, *domain.StockLevel, error)
	StockOut(ctx context.Context, cmd StockOutCommand) (*domain.StockMovement, *domain.StockLevel, error)
	StockAdjust(ctx context.Context, cmd StockAdjustCommand) (*domain.StockMovement, *domain.StockLevel, error)
	GetStock(ctx context.Context, tenantID, sku, warehouseID string) (*domain.StockLevel, error)
	ListMovements(ctx context.Context, tenantID, sku string) ([]*domain.StockMovement, error)
}
