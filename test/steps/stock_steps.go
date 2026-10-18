package steps

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

type StockLevelResponse struct {
	ItemID    string `json:"item_id"`
	TenantID  string `json:"tenant_id"`
	Quantity  int    `json:"quantity"`
	UpdatedAt string `json:"updated_at"`
}

type StockMovementResponse struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	ItemID       string `json:"item_id"`
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
}
