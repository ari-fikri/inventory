package steps

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

type SupplierResponse struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
}

type POLineItemResponse struct {
	ID               string  `json:"id"`
	SKU              string  `json:"sku"`
	OrderedQuantity  int     `json:"ordered_quantity"`
	ReceivedQuantity int     `json:"received_quantity"`
	UnitCost         float64 `json:"unit_cost"`
}

type PurchaseOrderResponse struct {
	ID          string               `json:"id"`
	TenantID    string               `json:"tenant_id"`
	PONumber    string               `json:"po_number"`
	SupplierID  string               `json:"supplier_id"`
	WarehouseID string               `json:"warehouse_id"`
	Status      string               `json:"status"`
	Notes       string               `json:"notes"`
	Items       []POLineItemResponse `json:"items"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

func RegisterPurchaseOrderSteps(sc *godog.ScenarioContext, ctx *TestContext) {
	supplierMap := make(map[string]string) // name -> id

	sc.Step(`^I create a supplier with name "([^"]*)", email "([^"]*)", and phone "([^"]*)"$`, func(name, email, phone string) error {
		payload := map[string]any{
			"name":  name,
			"email": email,
			"phone": phone,
		}
		if err := ctx.SendRequest("POST", "/api/v1/suppliers", ctx.CurrentTenantID, payload); err != nil {
			return err
		}
		if ctx.LastResponse.StatusCode == 201 {
			var sup SupplierResponse
			if err := json.Unmarshal(ctx.LastBody, &sup); err == nil {
				supplierMap[name] = sup.ID
			}
		}
		return nil
	})

	sc.Step(`^I create a purchase order with PO number "([^"]*)", supplier "([^"]*)", destination warehouse "([^"]*)", and line items:$`, func(poNum, supplierName, wh string, table *godog.Table) error {
		supplierID := supplierMap[supplierName]

		type poItemPayload struct {
			SKU      string  `json:"sku"`
			Quantity int     `json:"quantity"`
			UnitCost float64 `json:"unit_cost"`
		}

		var items []poItemPayload
		for _, row := range table.Rows[1:] {
			sku := row.Cells[0].Value
			qty, _ := strconv.Atoi(row.Cells[1].Value)
			cost, _ := strconv.ParseFloat(row.Cells[2].Value, 64)
			items = append(items, poItemPayload{
				SKU:      sku,
				Quantity: qty,
				UnitCost: cost,
			})
		}

		payload := map[string]any{
			"po_number":    poNum,
			"supplier_id":  supplierID,
			"warehouse_id": wh,
			"items":        items,
		}

		return ctx.SendRequest("POST", "/api/v1/purchase-orders", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^the purchase order status should be "([^"]*)"$`, func(expectedStatus string) error {
		var po PurchaseOrderResponse
		if err := json.Unmarshal(ctx.LastBody, &po); err != nil {
			return fmt.Errorf("failed to unmarshal purchase order response: %w, body: %s", err, string(ctx.LastBody))
		}
		if po.Status != expectedStatus {
			return fmt.Errorf("expected PO status %q, got %q", expectedStatus, po.Status)
		}
		return nil
	})

	sc.Step(`^a supplier exists with name "([^"]*)", email "([^"]*)", and phone "([^"]*)"$`, func(name, email, phone string) error {
		payload := map[string]any{
			"name":  name,
			"email": email,
			"phone": phone,
		}
		if err := ctx.SendRequest("POST", "/api/v1/suppliers", ctx.CurrentTenantID, payload); err != nil {
			return err
		}
		var sup SupplierResponse
		if err := json.Unmarshal(ctx.LastBody, &sup); err == nil {
			supplierMap[name] = sup.ID
		}
		return nil
	})

	sc.Step(`^a draft purchase order "([^"]*)" exists for supplier "([^"]*)" in warehouse "([^"]*)" with items:$`, func(poNum, supplierName, wh string, table *godog.Table) error {
		supplierID := supplierMap[supplierName]

		type poItemPayload struct {
			SKU      string  `json:"sku"`
			Quantity int     `json:"quantity"`
			UnitCost float64 `json:"unit_cost"`
		}

		var items []poItemPayload
		for _, row := range table.Rows[1:] {
			sku := row.Cells[0].Value
			qty, _ := strconv.Atoi(row.Cells[1].Value)
			cost, _ := strconv.ParseFloat(row.Cells[2].Value, 64)
			items = append(items, poItemPayload{
				SKU:      sku,
				Quantity: qty,
				UnitCost: cost,
			})
		}

		payload := map[string]any{
			"po_number":    poNum,
			"supplier_id":  supplierID,
			"warehouse_id": wh,
			"items":        items,
		}

		return ctx.SendRequest("POST", "/api/v1/purchase-orders", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^I submit purchase order "([^"]*)"$`, func(poNum string) error {
		return ctx.SendRequest("PATCH", "/api/v1/purchase-orders/"+poNum+"/submit", ctx.CurrentTenantID, nil)
	})

	sc.Step(`^I receive delivery for purchase order "([^"]*)" with items:$`, func(poNum string, table *godog.Table) error {
		type receiveItemPayload struct {
			SKU              string `json:"sku"`
			ReceivedQuantity int    `json:"received_quantity"`
		}

		var items []receiveItemPayload
		for _, row := range table.Rows[1:] {
			sku := row.Cells[0].Value
			qty, _ := strconv.Atoi(row.Cells[1].Value)
			items = append(items, receiveItemPayload{
				SKU:              sku,
				ReceivedQuantity: qty,
			})
		}

		payload := map[string]any{
			"items": items,
		}

		return ctx.SendRequest("POST", "/api/v1/purchase-orders/"+poNum+"/receive", ctx.CurrentTenantID, payload)
	})
}
