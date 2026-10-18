package steps

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/cucumber/godog"
)

type CostLayerResponse struct {
	WarehouseID       string  `json:"warehouse_id"`
	Quantity          int     `json:"quantity"`
	RemainingQuantity int     `json:"remaining_quantity"`
	UnitCost          float64 `json:"unit_cost"`
	CreatedAt         string  `json:"created_at"`
}

type ItemValuationResponse struct {
	SKU             string              `json:"sku"`
	TenantID        string              `json:"tenant_id"`
	ValuationMethod string              `json:"valuation_method"`
	TotalQuantity   int                 `json:"total_quantity"`
	TotalValue      float64             `json:"total_value"`
	AverageUnitCost float64             `json:"average_unit_cost"`
	CostLayers      []CostLayerResponse `json:"cost_layers"`
}

func RegisterValuationSteps(sc *godog.ScenarioContext, ctx *TestContext) {
	sc.Step(`^I add (\d+) units of stock to item with SKU "([^"]*)" with unit cost ([0-9.]+) and reference "([^"]*)"$`, func(qty int, sku string, unitCost float64, ref string) error {
		ctx.LastSKU = sku
		payload := map[string]any{
			"quantity":  qty,
			"unit_cost": unitCost,
			"reference": ref,
		}
		return ctx.SendRequest("POST", "/api/v1/items/sku/"+sku+"/stock/in", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^I query inventory valuation for item with SKU "([^"]*)"$`, func(sku string) error {
		ctx.LastSKU = sku
		return ctx.SendRequest("GET", "/api/v1/items/sku/"+sku+"/valuation", ctx.CurrentTenantID, nil)
	})

	sc.Step(`^the valuation response should show total quantity (\d+) and total value ([0-9.]+)$`, func(expectedQty int, expectedValue float64) error {
		var valResp ItemValuationResponse
		if err := json.Unmarshal(ctx.LastBody, &valResp); err != nil {
			return fmt.Errorf("failed to unmarshal valuation response: %w, body: %s", err, string(ctx.LastBody))
		}

		if valResp.TotalQuantity != expectedQty {
			return fmt.Errorf("expected total quantity %d, got %d", expectedQty, valResp.TotalQuantity)
		}

		if math.Abs(valResp.TotalValue-expectedValue) > 0.001 {
			return fmt.Errorf("expected total value %.2f, got %.2f", expectedValue, valResp.TotalValue)
		}

		return nil
	})

	sc.Step(`^the response should contain total COGS ([0-9.]+)$`, func(expectedCOGS float64) error {
		var resp struct {
			TotalCOGS float64 `json:"total_cogs"`
		}
		if err := json.Unmarshal(ctx.LastBody, &resp); err != nil {
			return fmt.Errorf("failed to unmarshal stock response for COGS: %w, body: %s", err, string(ctx.LastBody))
		}
		if math.Abs(resp.TotalCOGS-expectedCOGS) > 0.001 {
			return fmt.Errorf("expected total COGS %.2f, got %.2f", expectedCOGS, resp.TotalCOGS)
		}
		return nil
	})
}
