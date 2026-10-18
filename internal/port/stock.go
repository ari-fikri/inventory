package port

import (
	"context"

	"new-inventory/internal/domain"
)

type StockInCommand struct {
	TenantID  string
	SKU       string
	Quantity  int
	Reference string
	Notes     string
}

type StockRepository interface {
	GetStock(ctx context.Context, tenantID, itemID string) (*domain.StockLevel, error)
	RecordMovement(ctx context.Context, movement *domain.StockMovement, newBalance int) error
	ListMovements(ctx context.Context, tenantID, itemID string) ([]*domain.StockMovement, error)
}

type StockService interface {
	StockIn(ctx context.Context, cmd StockInCommand) (*domain.StockMovement, *domain.StockLevel, error)
	GetStock(ctx context.Context, tenantID, sku string) (*domain.StockLevel, error)
	ListMovements(ctx context.Context, tenantID, sku string) ([]*domain.StockMovement, error)
}
