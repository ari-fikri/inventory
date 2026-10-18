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

	stockRepo := sqlite.NewStockRepository(db)
	stockService := service.NewStockService(itemRepo, stockRepo)
	stockHandler := inventoryHttp.NewStockHandler(stockService)

	valRepo := sqlite.NewValuationRepository(db)
	valService := service.NewValuationService(itemRepo, valRepo)
	valHandler := inventoryHttp.NewValuationHandler(valService)

	appRouter := inventoryHttp.NewRouter(inventoryHttp.RouterConfig{
		ItemHandler:      itemHandler,
		StockHandler:     stockHandler,
		ValuationHandler: valHandler,
	})

	testCtx := steps.NewTestContext(appRouter)
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		testCtx.Close()
		_ = db.Close()
		return ctx, nil
	})

	steps.RegisterTenantSteps(sc, testCtx)
	steps.RegisterItemSteps(sc, testCtx)
	steps.RegisterStockSteps(sc, testCtx)
	steps.RegisterValuationSteps(sc, testCtx)
	steps.RegisterPurchaseOrderSteps(sc, testCtx)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
