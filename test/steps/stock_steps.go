package steps

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

type StockLevelResponse struct {
	ItemID      string `json:"item_id"`
	TenantID    string `json:"tenant_id"`
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
	UpdatedAt   string `json:"updated_at"`
}

type StockMovementResponse struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	ItemID       string `json:"item_id"`
	WarehouseID  string `json:"warehouse_id"`
	MovementType string `json:"movement_type"`
	Quantity     int    `json:"quantity"`
	BalanceAfter int    `json:"balance_after"`
	Reference    string `json:"reference"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"created_at"`
}

func RegisterStockSteps(sc *godog.ScenarioContext, ctx *TestContext) {
	sc.Step(`^I add (\d+) units of stock to item with SKU "([^"]*)" with reference "([^"]*)" and note "([^"]*)"$`, func(qty int, sku, ref, note string) error {
		ctx.LastSKU = sku
		payload := map[string]any{
			"quantity":  qty,
			"reference": ref,
			"notes":     note,
		}
		return ctx.SendRequest("POST", "/api/v1/items/sku/"+sku+"/stock/in", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^I add (\d+) units of stock to item with SKU "([^"]*)" in warehouse "([^"]*)" with reference "([^"]*)" and note "([^"]*)"$`, func(qty int, sku, warehouseID, ref, note string) error {
		ctx.LastSKU = sku
		payload := map[string]any{
			"quantity":     qty,
			"warehouse_id": warehouseID,
			"reference":    ref,
			"notes":        note,
		}
		return ctx.SendRequest("POST", "/api/v1/items/sku/"+sku+"/stock/in", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^I adjust stock for item with SKU "([^"]*)" to target quantity (\d+) with reference "([^"]*)" and note "([^"]*)"$`, func(sku string, targetQty int, ref, note string) error {
		ctx.LastSKU = sku
		payload := map[string]any{
			"target_quantity": targetQty,
			"reference":       ref,
			"notes":           note,
		}
		return ctx.SendRequest("POST", "/api/v1/items/sku/"+sku+"/stock/adjust", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^I dispatch (\d+) units of stock from item with SKU "([^"]*)" with reference "([^"]*)" and note "([^"]*)"$`, func(qty int, sku, ref, note string) error {
		ctx.LastSKU = sku
		payload := map[string]any{
			"quantity":  qty,
			"reference": ref,
			"notes":     note,
		}
		return ctx.SendRequest("POST", "/api/v1/items/sku/"+sku+"/stock/out", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^I transfer (\d+) units of stock for item with SKU "([^"]*)" from warehouse "([^"]*)" to warehouse "([^"]*)" with reference "([^"]*)" and note "([^"]*)"$`, func(qty int, sku, srcWH, destWH, ref, note string) error {
		ctx.LastSKU = sku
		payload := map[string]any{
			"source_warehouse_id": srcWH,
			"target_warehouse_id": destWH,
			"quantity":            qty,
			"reference":           ref,
			"notes":               note,
		}
		return ctx.SendRequest("POST", "/api/v1/items/sku/"+sku+"/stock/transfer", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^the current stock for item with SKU "([^"]*)" should be (\d+)$`, func(sku string, expectedQty int) error {
		if err := ctx.SendRequest("GET", "/api/v1/items/sku/"+sku+"/stock", ctx.CurrentTenantID, nil); err != nil {
			return err
		}
		if ctx.LastResponse.StatusCode != 200 {
			return fmt.Errorf("expected 200 OK fetching stock, got %d with body: %s", ctx.LastResponse.StatusCode, string(ctx.LastBody))
		}

		var stock StockLevelResponse
		if err := json.Unmarshal(ctx.LastBody, &stock); err != nil {
			return fmt.Errorf("failed to unmarshal stock level: %w", err)
		}

		if stock.Quantity != expectedQty {
			return fmt.Errorf("expected stock quantity %d, got %d", expectedQty, stock.Quantity)
		}
		return nil
	})

	sc.Step(`^the stock for item with SKU "([^"]*)" in warehouse "([^"]*)" should be (\d+)$`, func(sku, warehouseID string, expectedQty int) error {
		if err := ctx.SendRequest("GET", "/api/v1/items/sku/"+sku+"/stock?warehouse_id="+warehouseID, ctx.CurrentTenantID, nil); err != nil {
			return err
		}
		if ctx.LastResponse.StatusCode != 200 {
			return fmt.Errorf("expected 200 OK fetching stock for warehouse %s, got %d with body: %s", warehouseID, ctx.LastResponse.StatusCode, string(ctx.LastBody))
		}

		var stock StockLevelResponse
		if err := json.Unmarshal(ctx.LastBody, &stock); err != nil {
			return fmt.Errorf("failed to unmarshal stock level: %w", err)
		}

		if stock.Quantity != expectedQty {
			return fmt.Errorf("expected stock quantity %d in warehouse %s, got %d", expectedQty, warehouseID, stock.Quantity)
		}
		return nil
	})

	sc.Step(`^a stock movement record of type "([^"]*)" with quantity (\d+) should be logged$`, func(expectedType string, expectedQty int) error {
		sku := ctx.LastSKU
		if err := ctx.SendRequest("GET", "/api/v1/items/sku/"+sku+"/movements", ctx.CurrentTenantID, nil); err != nil {
			return err
		}
		if ctx.LastResponse.StatusCode != 200 {
			return fmt.Errorf("expected 200 OK fetching movements, got %d with body: %s", ctx.LastResponse.StatusCode, string(ctx.LastBody))
		}

		var movements []StockMovementResponse
		if err := json.Unmarshal(ctx.LastBody, &movements); err != nil {
			return fmt.Errorf("failed to unmarshal movements: %w", err)
		}

		if len(movements) == 0 {
			return fmt.Errorf("expected at least one movement record, got none")
		}

		latest := movements[0]
		if latest.MovementType != expectedType {
			return fmt.Errorf("expected latest movement type %q, got %q", expectedType, latest.MovementType)
		}
		if latest.Quantity != expectedQty {
			return fmt.Errorf("expected latest movement quantity %d, got %d", expectedQty, latest.Quantity)
		}

		return nil
	})

	sc.Step(`^I query stock movements for item with SKU "([^"]*)"$`, func(sku string) error {
		ctx.LastSKU = sku
		return ctx.SendRequest("GET", "/api/v1/items/sku/"+sku+"/movements", ctx.CurrentTenantID, nil)
	})

	sc.Step(`^the stock movements history should contain (\d+) entries in reverse chronological order:$`, func(expectedCount int, table *godog.Table) error {
		var movements []StockMovementResponse
		if err := json.Unmarshal(ctx.LastBody, &movements); err != nil {
			return fmt.Errorf("failed to unmarshal movements: %w, body: %s", err, string(ctx.LastBody))
		}

		if len(movements) != expectedCount {
			return fmt.Errorf("expected %d movements, got %d with body: %s", expectedCount, len(movements), string(ctx.LastBody))
		}

		dataRows := table.Rows[1:]
		if len(dataRows) != len(movements) {
			return fmt.Errorf("table row count %d does not match movements count %d", len(dataRows), len(movements))
		}

		for i, row := range dataRows {
			expectedType := row.Cells[0].Value
			expectedQty, _ := strconv.Atoi(row.Cells[1].Value)
			expectedBalance, _ := strconv.Atoi(row.Cells[2].Value)
			expectedRef := row.Cells[3].Value

			actual := movements[i]
			if actual.MovementType != expectedType {
				return fmt.Errorf("entry %d: expected movement_type %q, got %q", i, expectedType, actual.MovementType)
			}
			if actual.Quantity != expectedQty {
				return fmt.Errorf("entry %d: expected quantity %d, got %d", i, expectedQty, actual.Quantity)
			}
			if actual.BalanceAfter != expectedBalance {
				return fmt.Errorf("entry %d: expected balance_after %d, got %d", i, expectedBalance, actual.BalanceAfter)
			}
			if actual.Reference != expectedRef {
				return fmt.Errorf("entry %d: expected reference %q, got %q", i, expectedRef, actual.Reference)
			}
		}

		return nil
	})
}
