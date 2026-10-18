package test

import (
	"context"
	"testing"

	inventoryHttp "new-inventory/internal/adapter/http"
	"new-inventory/test/steps"

	"github.com/cucumber/godog"
)

func InitializeScenario(sc *godog.ScenarioContext) {
	appRouter := inventoryHttp.NewRouter(inventoryHttp.RouterConfig{})

	testCtx := steps.NewTestContext(appRouter)
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		testCtx.Close()
		return ctx, nil
	})

	steps.RegisterTenantSteps(sc, testCtx)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
