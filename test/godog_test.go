package test

import (
	"context"
	"net/http"
	"testing"

	"new-inventory/test/steps"

	"github.com/cucumber/godog"
)

func InitializeScenario(sc *godog.ScenarioContext) {
	// Dummy handler for RED phase
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	testCtx := steps.NewTestContext(dummyHandler)
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
