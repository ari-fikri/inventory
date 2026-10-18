package service

import (
	"context"
	"errors"
	"time"

	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

var _ port.StockService = (*StockService)(nil)

type StockService struct {
	itemRepo  port.ItemRepository
	stockRepo port.StockRepository
}

func NewStockService(itemRepo port.ItemRepository, stockRepo port.StockRepository) *StockService {
	return &StockService{
		itemRepo:  itemRepo,
		stockRepo: stockRepo,
	}
}

func (s *StockService) StockIn(ctx context.Context, cmd port.StockInCommand) (*domain.StockMovement, *domain.StockLevel, error) {
	if cmd.Quantity <= 0 {
		return nil, nil, domain.ErrInvalidQuantity
	}

	item, err := s.itemRepo.GetBySKU(ctx, cmd.TenantID, cmd.SKU)
	if err != nil {
		return nil, nil, err
	}
	if !item.IsActive {
		return nil, nil, domain.ErrItemInactive
	}

	currentStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID)
	currentQty := 0
	if err != nil && !errors.Is(err, domain.ErrStockNotFound) {
		return nil, nil, err
	}
	if currentStock != nil {
		currentQty = currentStock.Quantity
	}

	newBalance := currentQty + cmd.Quantity

	movement, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     cmd.TenantID,
		ItemID:       item.ID,
		MovementType: domain.MovementTypeIn,
		Quantity:     cmd.Quantity,
		BalanceAfter: newBalance,
		Reference:    cmd.Reference,
		Notes:        cmd.Notes,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := s.stockRepo.RecordMovement(ctx, movement, newBalance); err != nil {
		return nil, nil, err
	}

	updatedLevel := &domain.StockLevel{
		TenantID:  cmd.TenantID,
		ItemID:    item.ID,
		Quantity:  newBalance,
		UpdatedAt: time.Now().UTC(),
	}

	return movement, updatedLevel, nil
}

func (s *StockService) StockOut(ctx context.Context, cmd port.StockOutCommand) (*domain.StockMovement, *domain.StockLevel, error) {
	if cmd.Quantity <= 0 {
		return nil, nil, domain.ErrInvalidQuantity
	}

	item, err := s.itemRepo.GetBySKU(ctx, cmd.TenantID, cmd.SKU)
	if err != nil {
		return nil, nil, err
	}
	if !item.IsActive {
		return nil, nil, domain.ErrItemInactive
	}

	currentStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID)
	currentQty := 0
	if err != nil && !errors.Is(err, domain.ErrStockNotFound) {
		return nil, nil, err
	}
	if currentStock != nil {
		currentQty = currentStock.Quantity
	}

	if currentQty < cmd.Quantity {
		return nil, nil, domain.ErrInsufficientStock
	}

	newBalance := currentQty - cmd.Quantity

	movement, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     cmd.TenantID,
		ItemID:       item.ID,
		MovementType: domain.MovementTypeOut,
		Quantity:     cmd.Quantity,
		BalanceAfter: newBalance,
		Reference:    cmd.Reference,
		Notes:        cmd.Notes,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := s.stockRepo.RecordMovement(ctx, movement, newBalance); err != nil {
		return nil, nil, err
	}

	updatedLevel := &domain.StockLevel{
		TenantID:  cmd.TenantID,
		ItemID:    item.ID,
		Quantity:  newBalance,
		UpdatedAt: time.Now().UTC(),
	}

	return movement, updatedLevel, nil
}

func (s *StockService) StockAdjust(ctx context.Context, cmd port.StockAdjustCommand) (*domain.StockMovement, *domain.StockLevel, error) {
	if cmd.TargetQuantity < 0 {
		return nil, nil, domain.ErrInvalidQuantity
	}

	item, err := s.itemRepo.GetBySKU(ctx, cmd.TenantID, cmd.SKU)
	if err != nil {
		return nil, nil, err
	}
	if !item.IsActive {
		return nil, nil, domain.ErrItemInactive
	}

	currentStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID)
	currentQty := 0
	if err != nil && !errors.Is(err, domain.ErrStockNotFound) {
		return nil, nil, err
	}
	if currentStock != nil {
		currentQty = currentStock.Quantity
	}

	delta := cmd.TargetQuantity - currentQty
	absDelta := delta
	if absDelta < 0 {
		absDelta = -absDelta
	}

	movement, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     cmd.TenantID,
		ItemID:       item.ID,
		MovementType: domain.MovementTypeAdjust,
		Quantity:     absDelta,
		BalanceAfter: cmd.TargetQuantity,
		Reference:    cmd.Reference,
		Notes:        cmd.Notes,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := s.stockRepo.RecordMovement(ctx, movement, cmd.TargetQuantity); err != nil {
		return nil, nil, err
	}

	updatedLevel := &domain.StockLevel{
		TenantID:  cmd.TenantID,
		ItemID:    item.ID,
		Quantity:  cmd.TargetQuantity,
		UpdatedAt: time.Now().UTC(),
	}

	return movement, updatedLevel, nil
}

func (s *StockService) GetStock(ctx context.Context, tenantID, sku string) (*domain.StockLevel, error) {
	item, err := s.itemRepo.GetBySKU(ctx, tenantID, sku)
	if err != nil {
		return nil, err
	}

	stock, err := s.stockRepo.GetStock(ctx, tenantID, item.ID)
	if err != nil {
		if errors.Is(err, domain.ErrStockNotFound) {
			return &domain.StockLevel{
				TenantID:  tenantID,
				ItemID:    item.ID,
				Quantity:  0,
				UpdatedAt: time.Now().UTC(),
			}, nil
		}
		return nil, err
	}

	return stock, nil
}

func (s *StockService) ListMovements(ctx context.Context, tenantID, sku string) ([]*domain.StockMovement, error) {
	item, err := s.itemRepo.GetBySKU(ctx, tenantID, sku)
	if err != nil {
		return nil, err
	}

	return s.stockRepo.ListMovements(ctx, tenantID, item.ID)
}
