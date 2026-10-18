package port

import (
	"context"

	"new-inventory/internal/domain"
)

type CreateItemCommand struct {
	TenantID     string
	SKU          string
	Name         string
	Description  string
	UnitPrice    float64
	MinThreshold int
}

type ItemRepository interface {
	Create(ctx context.Context, item *domain.Item) error
	GetBySKU(ctx context.Context, tenantID, sku string) (*domain.Item, error)
	GetByID(ctx context.Context, tenantID, id string) (*domain.Item, error)
	List(ctx context.Context, tenantID string) ([]*domain.Item, error)
}

type ItemService interface {
	CreateItem(ctx context.Context, cmd CreateItemCommand) (*domain.Item, error)
	GetItem(ctx context.Context, tenantID, id string) (*domain.Item, error)
	ListItems(ctx context.Context, tenantID string) ([]*domain.Item, error)
}
