package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	ValuationMethodFIFO = "FIFO"
	ValuationMethodAVCO = "AVCO"
)

var (
	ErrInvalidValuationMethod = errors.New("invalid valuation method; allowed methods are FIFO, AVCO")
	ErrCannotChangeValuation  = errors.New("cannot change valuation method when stock is greater than zero")
)

type CostLayer struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	ItemID            string    `json:"item_id"`
	WarehouseID       string    `json:"warehouse_id"`
	Quantity          int       `json:"quantity"`
	RemainingQuantity int       `json:"remaining_quantity"`
	UnitCost          float64   `json:"unit_cost"`
	CreatedAt         time.Time `json:"created_at"`
}

func NewCostLayer(tenantID, itemID, warehouseID string, qty int, unitCost float64) *CostLayer {
	wh := warehouseID
	if wh == "" {
		wh = "default"
	}
	return &CostLayer{
		ID:                uuid.NewString(),
		TenantID:          tenantID,
		ItemID:            itemID,
		WarehouseID:       wh,
		Quantity:          qty,
		RemainingQuantity: qty,
		UnitCost:          unitCost,
		CreatedAt:         time.Now().UTC(),
	}
}

type ItemValuation struct {
	SKU             string      `json:"sku"`
	TenantID        string      `json:"tenant_id"`
	ValuationMethod string      `json:"valuation_method"`
	TotalQuantity   int         `json:"total_quantity"`
	TotalValue      float64     `json:"total_value"`
	AverageUnitCost float64     `json:"average_unit_cost"`
	CostLayers      []CostLayer `json:"cost_layers"`
}

type ConsumedCostLayer struct {
	LayerID      string  `json:"layer_id"`
	ConsumedQty  int     `json:"consumed_qty"`
	RemainingQty int     `json:"remaining_qty"`
	UnitCost     float64 `json:"unit_cost"`
}
