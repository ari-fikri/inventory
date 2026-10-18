package service

import (
	"context"

	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

var _ port.ValuationService = (*ValuationService)(nil)

type ValuationService struct {
	itemRepo      port.ItemRepository
	valuationRepo port.ValuationRepository
}

func NewValuationService(itemRepo port.ItemRepository, valuationRepo port.ValuationRepository) *ValuationService {
	return &ValuationService{
		itemRepo:      itemRepo,
		valuationRepo: valuationRepo,
	}
}

func (s *ValuationService) GetItemValuation(ctx context.Context, tenantID, sku string) (*domain.ItemValuation, error) {
	item, err := s.itemRepo.GetBySKU(ctx, tenantID, sku)
	if err != nil {
		return nil, err
	}

	method, err := s.valuationRepo.GetTenantValuationMethod(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	layers, err := s.valuationRepo.GetCostLayers(ctx, tenantID, item.ID)
	if err != nil {
		return nil, err
	}

	totalQty := 0
	totalValue := 0.0
	for _, l := range layers {
		totalQty += l.RemainingQuantity
		totalValue += float64(l.RemainingQuantity) * l.UnitCost
	}

	avgUnitCost := 0.0
	if totalQty > 0 {
		avgUnitCost = totalValue / float64(totalQty)
	}

	return &domain.ItemValuation{
		SKU:             sku,
		TenantID:        tenantID,
		ValuationMethod: method,
		TotalQuantity:   totalQty,
		TotalValue:      totalValue,
		AverageUnitCost: avgUnitCost,
		CostLayers:      layers,
	}, nil
}

func (s *ValuationService) GetTenantValuationMethod(ctx context.Context, tenantID string) (string, error) {
	return s.valuationRepo.GetTenantValuationMethod(ctx, tenantID)
}

func (s *ValuationService) SetTenantValuationMethod(ctx context.Context, tenantID, method string) error {
	if method != domain.ValuationMethodFIFO && method != domain.ValuationMethodAVCO {
		return domain.ErrInvalidValuationMethod
	}

	totalStock, err := s.valuationRepo.GetTenantTotalStock(ctx, tenantID)
	if err != nil {
		return err
	}

	if totalStock > 0 {
		return domain.ErrCannotChangeValuation
	}

	return s.valuationRepo.SetTenantValuationMethod(ctx, tenantID, method)
}
