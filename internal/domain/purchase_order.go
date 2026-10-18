package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	POStatusDraft             = "DRAFT"
	POStatusOrdered           = "ORDERED"
	POStatusPartiallyReceived = "PARTIALLY_RECEIVED"
	POStatusReceived          = "RECEIVED"
	POStatusCancelled         = "CANCELLED"
)

var (
	ErrSupplierNotFound      = errors.New("supplier not found")
	ErrDuplicateSupplierName = errors.New("supplier name already exists")
	ErrPONotFound            = errors.New("purchase order not found")
	ErrDuplicatePONumber     = errors.New("purchase order number already exists")
	ErrInvalidPOStatus       = errors.New("invalid status transition for purchase order")
	ErrOverReceiving         = errors.New("received quantity exceeds ordered quantity")
	ErrEmptyPONumber         = errors.New("purchase order number cannot be empty")
	ErrNoPOItems             = errors.New("purchase order must contain at least one line item")
)

type Supplier struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	ContactName string    `json:"contact_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	CreatedAt   time.Time `json:"created_at"`
}

type PurchaseOrderItem struct {
	ID               string  `json:"id"`
	POID             string  `json:"po_id"`
	TenantID         string  `json:"tenant_id"`
	ItemID           string  `json:"item_id"`
	SKU              string  `json:"sku"`
	OrderedQuantity  int     `json:"ordered_quantity"`
	ReceivedQuantity int     `json:"received_quantity"`
	UnitCost         float64 `json:"unit_cost"`
}

type PurchaseOrder struct {
	ID          string              `json:"id"`
	TenantID    string              `json:"tenant_id"`
	PONumber    string              `json:"po_number"`
	SupplierID  string              `json:"supplier_id"`
	WarehouseID string              `json:"warehouse_id"`
	Status      string              `json:"status"`
	Notes       string              `json:"notes"`
	Items       []PurchaseOrderItem `json:"items"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func NewPurchaseOrder(tenantID, poNumber, supplierID, warehouseID, notes string, items []PurchaseOrderItem) (*PurchaseOrder, error) {
	poNumber = strings.TrimSpace(poNumber)
	if poNumber == "" {
		return nil, ErrEmptyPONumber
	}
	if len(items) == 0 {
		return nil, ErrNoPOItems
	}
	wh := strings.TrimSpace(warehouseID)
	if wh == "" {
		wh = "default"
	}
	now := time.Now().UTC()
	poID := uuid.NewString()

	for i := range items {
		items[i].ID = uuid.NewString()
		items[i].POID = poID
		items[i].TenantID = tenantID
	}

	return &PurchaseOrder{
		ID:          poID,
		TenantID:    tenantID,
		PONumber:    poNumber,
		SupplierID:  supplierID,
		WarehouseID: wh,
		Status:      POStatusDraft,
		Notes:       strings.TrimSpace(notes),
		Items:       items,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
