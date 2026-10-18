package test

import (
	"context"
	"testing"

	inventoryHttp "new-inventory/internal/adapter/http"
	"new-inventory/internal/adapter/repository/sqlite"
	"new-inventory/internal/service"
	"new-inventory/test/steps"

	"github.com/cucumber/godog"
)

func InitializeScenario(sc *godog.ScenarioContext) {
	db, err := sqlite.OpenDB(":memory:")
	if err != nil {
		panic("failed to open in-memory db: " + err.Error())
	}

	itemRepo := sqlite.NewItemRepository(db)
	itemService := service.NewItemService(itemRepo)
	itemHandler := inventoryHttp.NewItemHandler(itemService)

	appRouter := inventoryHttp.NewRouter(inventoryHttp.RouterConfig{
		ItemHandler: itemHandler,
	})

	testCtx := steps.NewTestContext(appRouter)
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		testCtx.Close()
		_ = db.Close()
		return ctx, nil
	})

	steps.RegisterTenantSteps(sc, testCtx)
	steps.RegisterItemSteps(sc, testCtx)
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
