package steps

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RegisterTenantSteps(sc *godog.ScenarioContext, ctx *TestContext) {
	sc.Step(`^I send a GET request to "([^"]*)" without a tenant ID$`, func(path string) error {
		return ctx.SendRequest("GET", path, "", nil)
	})

	sc.Step(`^the response status code should be (\d+)$`, func(expectedStatus int) error {
		if ctx.LastResponse == nil {
			return fmt.Errorf("no response received")
		}
		if ctx.LastResponse.StatusCode != expectedStatus {
			return fmt.Errorf("expected status %d, got %d with body: %s", expectedStatus, ctx.LastResponse.StatusCode, string(ctx.LastBody))
		}
		return nil
	})

	sc.Step(`^the response should contain error "([^"]*)"$`, func(expectedError string) error {
		var errResp ErrorResponse
		if err := json.Unmarshal(ctx.LastBody, &errResp); err != nil {
			return fmt.Errorf("failed to unmarshal error response: %w, body: %s", err, string(ctx.LastBody))
		}
		if errResp.Error != expectedError {
			return fmt.Errorf("expected error %q, got %q", expectedError, errResp.Error)
		}
		return nil
	})
}
