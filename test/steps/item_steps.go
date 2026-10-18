package steps

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

type ItemResponse struct {
	ID           string  `json:"id"`
	TenantID     string  `json:"tenant_id"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	UnitPrice    float64 `json:"unit_price"`
	MinThreshold int     `json:"min_threshold"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func RegisterItemSteps(sc *godog.ScenarioContext, ctx *TestContext) {
	sc.Step(`^I am authenticated as tenant "([^"]*)"$`, func(tenantID string) error {
		ctx.CurrentTenantID = tenantID
		return nil
	})

	sc.Step(`^an item exists with SKU "([^"]*)" and name "([^"]*)"$`, func(sku, name string) error {
		payload := map[string]any{
			"sku":           sku,
			"name":          name,
			"description":   "Pre-existing item for testing",
			"unit_price":    25.0,
			"min_threshold": 1,
		}
		if err := ctx.SendRequest("POST", "/api/v1/items", ctx.CurrentTenantID, payload); err != nil {
			return err
		}
		if ctx.LastResponse.StatusCode != 201 {
			return fmt.Errorf("failed to setup pre-existing item, status %d: %s", ctx.LastResponse.StatusCode, string(ctx.LastBody))
		}
		return nil
	})

	sc.Step(`^I get item by SKU "([^"]*)" as tenant "([^"]*)"$`, func(sku, tenantID string) error {
		ctx.CurrentTenantID = tenantID
		return ctx.SendRequest("GET", "/api/v1/items/sku/"+sku, tenantID, nil)
	})

	sc.Step(`^I create an item with the following details:$`, func(table *godog.Table) error {
		payload := make(map[string]any)

		for _, row := range table.Rows {
			if len(row.Cells) < 2 {
				continue
			}
			key := row.Cells[0].Value
			val := row.Cells[1].Value

			switch key {
			case "unit_price":
				f, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return fmt.Errorf("invalid unit_price: %w", err)
				}
				payload[key] = f
			case "min_threshold":
				i, err := strconv.Atoi(val)
				if err != nil {
					return fmt.Errorf("invalid min_threshold: %w", err)
				}
				payload[key] = i
			default:
				payload[key] = val
			}
		}

		return ctx.SendRequest("POST", "/api/v1/items", ctx.CurrentTenantID, payload)
	})

	sc.Step(`^the response should contain item details:$`, func(table *godog.Table) error {
		var item ItemResponse
		if err := json.Unmarshal(ctx.LastBody, &item); err != nil {
			return fmt.Errorf("failed to unmarshal item response: %w, body: %s", err, string(ctx.LastBody))
		}

		for _, row := range table.Rows {
			if len(row.Cells) < 2 {
				continue
			}
			key := row.Cells[0].Value
			expected := row.Cells[1].Value

			switch key {
			case "sku":
				if item.SKU != expected {
					return fmt.Errorf("expected sku %q, got %q", expected, item.SKU)
				}
			case "name":
				if item.Name != expected {
					return fmt.Errorf("expected name %q, got %q", expected, item.Name)
				}
			case "tenant_id":
				if item.TenantID != expected {
					return fmt.Errorf("expected tenant_id %q, got %q", expected, item.TenantID)
				}
			case "is_active":
				expBool, _ := strconv.ParseBool(expected)
				if item.IsActive != expBool {
					return fmt.Errorf("expected is_active %v, got %v", expBool, item.IsActive)
				}
			}
		}
		return nil
	})
}
