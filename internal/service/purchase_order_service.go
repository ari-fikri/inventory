package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

var _ port.PurchaseOrderService = (*PurchaseOrderService)(nil)

type PurchaseOrderService struct {
	supplierRepo port.SupplierRepository
	poRepo       port.PurchaseOrderRepository
	itemRepo     port.ItemRepository
	stockService port.StockService
}

func NewPurchaseOrderService(
	supplierRepo port.SupplierRepository,
	poRepo port.PurchaseOrderRepository,
	itemRepo port.ItemRepository,
	stockService port.StockService,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		supplierRepo: supplierRepo,
		poRepo:       poRepo,
		itemRepo:     itemRepo,
		stockService: stockService,
	}
}

func (s *PurchaseOrderService) CreateSupplier(ctx context.Context, cmd port.CreateSupplierCommand) (*domain.Supplier, error) {
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, errors.New("supplier name cannot be empty")
	}

	supplier := &domain.Supplier{
		ID:          uuid.NewString(),
		TenantID:    cmd.TenantID,
		Name:        name,
		ContactName: strings.TrimSpace(cmd.ContactName),
		Email:       strings.TrimSpace(cmd.Email),
		Phone:       strings.TrimSpace(cmd.Phone),
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

func (s *PurchaseOrderService) ListSuppliers(ctx context.Context, tenantID string) ([]*domain.Supplier, error) {
	return s.supplierRepo.List(ctx, tenantID)
}

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, cmd port.CreatePOCommand) (*domain.PurchaseOrder, error) {
	if len(cmd.Items) == 0 {
		return nil, domain.ErrNoPOItems
	}

	var items []domain.PurchaseOrderItem
	for _, it := range cmd.Items {
		item, err := s.itemRepo.GetBySKU(ctx, cmd.TenantID, it.SKU)
		if err != nil {
			return nil, err
		}
		if it.Quantity <= 0 {
			return nil, domain.ErrInvalidQuantity
		}
		items = append(items, domain.PurchaseOrderItem{
			ItemID:          item.ID,
			SKU:             it.SKU,
			OrderedQuantity: it.Quantity,
			UnitCost:        it.UnitCost,
		})
	}

	po, err := domain.NewPurchaseOrder(cmd.TenantID, cmd.PONumber, cmd.SupplierID, cmd.WarehouseID, cmd.Notes, items)
	if err != nil {
		return nil, err
	}

	if err := s.poRepo.Create(ctx, po); err != nil {
		return nil, err
	}
	return po, nil
}

func (s *PurchaseOrderService) GetPurchaseOrderByNumber(ctx context.Context, tenantID, poNumber string) (*domain.PurchaseOrder, error) {
	return s.poRepo.GetByNumber(ctx, tenantID, poNumber)
}

func (s *PurchaseOrderService) SubmitPurchaseOrder(ctx context.Context, tenantID, poNumber string) (*domain.PurchaseOrder, error) {
	po, err := s.poRepo.GetByNumber(ctx, tenantID, poNumber)
	if err != nil {
		return nil, err
	}
	if po.Status != domain.POStatusDraft {
		return nil, domain.ErrInvalidPOStatus
	}

	po.Status = domain.POStatusOrdered
	po.UpdatedAt = time.Now().UTC()
	if err := s.poRepo.Update(ctx, po); err != nil {
		return nil, err
	}
	return po, nil
}
