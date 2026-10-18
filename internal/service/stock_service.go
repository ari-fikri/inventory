package service

import (
	"context"
	"errors"
	"strings"
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

	wh := cmd.WarehouseID
	if wh == "" {
		wh = "default"
	}

	currentStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID, wh)
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
		WarehouseID:  wh,
		MovementType: domain.MovementTypeIn,
		Quantity:     cmd.Quantity,
		BalanceAfter: newBalance,
		Reference:    cmd.Reference,
		Notes:        cmd.Notes,
	})
	if err != nil {
		return nil, nil, err
	}

	var costLayer *domain.CostLayer
	if cmd.UnitCost > 0 {
		costLayer = domain.NewCostLayer(cmd.TenantID, item.ID, wh, cmd.Quantity, cmd.UnitCost)
	}

	if err := s.stockRepo.RecordMovementWithCostLayer(ctx, movement, newBalance, costLayer); err != nil {
		return nil, nil, err
	}

	updatedLevel := &domain.StockLevel{
		TenantID:    cmd.TenantID,
		ItemID:      item.ID,
		WarehouseID: wh,
		Quantity:    newBalance,
		UpdatedAt:   time.Now().UTC(),
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

	wh := cmd.WarehouseID
	if wh == "" {
		wh = "default"
	}

	currentStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID, wh)
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

	// Consume cost layers for FIFO COGS
	activeLayers, err := s.stockRepo.GetActiveCostLayers(ctx, cmd.TenantID, item.ID, wh)
	if err != nil {
		return nil, nil, err
	}

	remainingToFulfill := cmd.Quantity
	totalCOGS := 0.0
	var consumed []domain.ConsumedCostLayer

	for _, layer := range activeLayers {
		if remainingToFulfill <= 0 {
			break
		}
		take := remainingToFulfill
		if take > layer.RemainingQuantity {
			take = layer.RemainingQuantity
		}
		totalCOGS += float64(take) * layer.UnitCost
		consumed = append(consumed, domain.ConsumedCostLayer{
			LayerID:      layer.ID,
			ConsumedQty:  take,
			RemainingQty: layer.RemainingQuantity - take,
		})
		remainingToFulfill -= take
	}

	movement, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     cmd.TenantID,
		ItemID:       item.ID,
		WarehouseID:  wh,
		MovementType: domain.MovementTypeOut,
		Quantity:     cmd.Quantity,
		BalanceAfter: newBalance,
		TotalCOGS:    totalCOGS,
		Reference:    cmd.Reference,
		Notes:        cmd.Notes,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := s.stockRepo.RecordStockOutWithValuation(ctx, movement, newBalance, consumed); err != nil {
		return nil, nil, err
	}

	updatedLevel := &domain.StockLevel{
		TenantID:    cmd.TenantID,
		ItemID:      item.ID,
		WarehouseID: wh,
		Quantity:    newBalance,
		UpdatedAt:   time.Now().UTC(),
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

	wh := cmd.WarehouseID
	if wh == "" {
		wh = "default"
	}

	currentStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID, wh)
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
		WarehouseID:  wh,
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
		TenantID:    cmd.TenantID,
		ItemID:      item.ID,
		WarehouseID: wh,
		Quantity:    cmd.TargetQuantity,
		UpdatedAt:   time.Now().UTC(),
	}

	return movement, updatedLevel, nil
}

func (s *StockService) StockTransfer(ctx context.Context, cmd port.StockTransferCommand) (*domain.StockTransferResult, error) {
	if cmd.Quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}
	srcWH := strings.TrimSpace(cmd.SourceWarehouseID)
	if srcWH == "" {
		srcWH = "default"
	}
	destWH := strings.TrimSpace(cmd.TargetWarehouseID)
	if destWH == "" {
		destWH = "default"
	}
	if srcWH == destWH {
		return nil, domain.ErrSameWarehouse
	}

	item, err := s.itemRepo.GetBySKU(ctx, cmd.TenantID, cmd.SKU)
	if err != nil {
		return nil, err
	}
	if !item.IsActive {
		return nil, domain.ErrItemInactive
	}

	srcStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID, srcWH)
	srcQty := 0
	if err != nil && !errors.Is(err, domain.ErrStockNotFound) {
		return nil, err
	}
	if srcStock != nil {
		srcQty = srcStock.Quantity
	}

	if srcQty < cmd.Quantity {
		return nil, domain.ErrInsufficientTransferStock
	}

	destStock, err := s.stockRepo.GetStock(ctx, cmd.TenantID, item.ID, destWH)
	destQty := 0
	if err != nil && !errors.Is(err, domain.ErrStockNotFound) {
		return nil, err
	}
	if destStock != nil {
		destQty = destStock.Quantity
	}

	newSourceBal := srcQty - cmd.Quantity
	newDestBal := destQty + cmd.Quantity

	outMovement, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     cmd.TenantID,
		ItemID:       item.ID,
		WarehouseID:  srcWH,
		MovementType: domain.MovementTypeTransferOut,
		Quantity:     cmd.Quantity,
		BalanceAfter: newSourceBal,
		Reference:    cmd.Reference,
		Notes:        cmd.Notes,
	})
	if err != nil {
		return nil, err
	}

	inMovement, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     cmd.TenantID,
		ItemID:       item.ID,
		WarehouseID:  destWH,
		MovementType: domain.MovementTypeTransferIn,
		Quantity:     cmd.Quantity,
		BalanceAfter: newDestBal,
		Reference:    cmd.Reference,
		Notes:        cmd.Notes,
	})
	if err != nil {
		return nil, err
	}

	if err := s.stockRepo.RecordTransfer(ctx, outMovement, inMovement, newSourceBal, newDestBal); err != nil {
		return nil, err
	}

	return &domain.StockTransferResult{
		ItemID:            item.ID,
		TenantID:          cmd.TenantID,
		SourceWarehouseID: srcWH,
		TargetWarehouseID: destWH,
		Quantity:          cmd.Quantity,
		SourceBalance:     newSourceBal,
		TargetBalance:     newDestBal,
		OutMovement:       outMovement,
		InMovement:        inMovement,
	}, nil
}

func (s *StockService) GetStock(ctx context.Context, tenantID, sku, warehouseID string) (*domain.StockLevel, error) {
	item, err := s.itemRepo.GetBySKU(ctx, tenantID, sku)
	if err != nil {
		return nil, err
	}

	if warehouseID == "" {
		warehouseID = "default"
	}

	stock, err := s.stockRepo.GetStock(ctx, tenantID, item.ID, warehouseID)
	if err != nil {
		if errors.Is(err, domain.ErrStockNotFound) {
			return &domain.StockLevel{
				TenantID:    tenantID,
				ItemID:      item.ID,
				WarehouseID: warehouseID,
				Quantity:    0,
				UpdatedAt:   time.Now().UTC(),
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
