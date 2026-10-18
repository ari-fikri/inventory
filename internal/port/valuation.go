package port

import (
	"context"

	"new-inventory/internal/domain"
)

type ValuationRepository interface {
	GetCostLayers(ctx context.Context, tenantID, itemID string) ([]domain.CostLayer, error)
	GetActiveCostLayers(ctx context.Context, tenantID, itemID, warehouseID string) ([]domain.CostLayer, error)
	UpdateCostLayerRemaining(ctx context.Context, layerID string, remainingQty int) error
	GetTenantValuationMethod(ctx context.Context, tenantID string) (string, error)
	SetTenantValuationMethod(ctx context.Context, tenantID, method string) error
	GetTenantTotalStock(ctx context.Context, tenantID string) (int, error)
}

type ValuationService interface {
	GetItemValuation(ctx context.Context, tenantID, sku string) (*domain.ItemValuation, error)
	GetTenantValuationMethod(ctx context.Context, tenantID string) (string, error)
	SetTenantValuationMethod(ctx context.Context, tenantID, method string) error
}
