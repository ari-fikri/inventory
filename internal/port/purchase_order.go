package port

import (
	"context"

	"new-inventory/internal/domain"
)

type CreateSupplierCommand struct {
	TenantID    string
	Name        string
	ContactName string
	Email       string
	Phone       string
}

type POLineItemCommand struct {
	SKU      string
	Quantity int
	UnitCost float64
}

type CreatePOCommand struct {
	TenantID    string
	PONumber    string
	SupplierID  string
	WarehouseID string
	Notes       string
	Items       []POLineItemCommand
}

type ReceivePOItemCommand struct {
	SKU              string
	ReceivedQuantity int
}

type ReceivePOCommand struct {
	TenantID string
	PONumber string
	Items    []ReceivePOItemCommand
}

type SupplierRepository interface {
	Create(ctx context.Context, supplier *domain.Supplier) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Supplier, error)
	GetByName(ctx context.Context, tenantID, name string) (*domain.Supplier, error)
	List(ctx context.Context, tenantID string) ([]*domain.Supplier, error)
}

type PurchaseOrderRepository interface {
	Create(ctx context.Context, po *domain.PurchaseOrder) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.PurchaseOrder, error)
	GetByNumber(ctx context.Context, tenantID, poNumber string) (*domain.PurchaseOrder, error)
	Update(ctx context.Context, po *domain.PurchaseOrder) error
	UpdateItemReceivedQty(ctx context.Context, itemID string, receivedQty int) error
}

type PurchaseOrderService interface {
	CreateSupplier(ctx context.Context, cmd CreateSupplierCommand) (*domain.Supplier, error)
	ListSuppliers(ctx context.Context, tenantID string) ([]*domain.Supplier, error)
	CreatePurchaseOrder(ctx context.Context, cmd CreatePOCommand) (*domain.PurchaseOrder, error)
	GetPurchaseOrderByNumber(ctx context.Context, tenantID, poNumber string) (*domain.PurchaseOrder, error)
	SubmitPurchaseOrder(ctx context.Context, tenantID, poNumber string) (*domain.PurchaseOrder, error)
	ReceivePurchaseOrder(ctx context.Context, cmd ReceivePOCommand) (*domain.PurchaseOrder, error)
}
