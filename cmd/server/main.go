package main

import (
	"log"
	"net/http"
	"os"

	inventoryHttp "new-inventory/internal/adapter/http"
	reposql "new-inventory/internal/adapter/repository/sql"
	"new-inventory/internal/service"
)

func main() {
	dbDriver := os.Getenv("DATABASE_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_PATH")
		if dsn == "" {
			dsn = "inventory.db"
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbWrapper, err := reposql.Open(reposql.Config{
		Driver:        dbDriver,
		DSN:           dsn,
		RunMigrations: true,
	})
	if err != nil {
		log.Fatalf("failed to initialize database (%s): %v", dbDriver, err)
	}
	defer dbWrapper.Close()

	itemRepo := reposql.NewItemRepository(dbWrapper.DB, dbWrapper.Dialect)
	itemService := service.NewItemService(itemRepo)
	itemHandler := inventoryHttp.NewItemHandler(itemService)

	stockRepo := reposql.NewStockRepository(dbWrapper.DB, dbWrapper.Dialect)
	stockService := service.NewStockService(itemRepo, stockRepo)
	stockHandler := inventoryHttp.NewStockHandler(stockService)

	valRepo := reposql.NewValuationRepository(dbWrapper.DB, dbWrapper.Dialect)
	valService := service.NewValuationService(itemRepo, valRepo)
	valHandler := inventoryHttp.NewValuationHandler(valService)

	supplierRepo := reposql.NewSupplierRepository(dbWrapper.DB, dbWrapper.Dialect)
	poRepo := reposql.NewPurchaseOrderRepository(dbWrapper.DB, dbWrapper.Dialect)
	poService := service.NewPurchaseOrderService(supplierRepo, poRepo, itemRepo, stockService)
	poHandler := inventoryHttp.NewPurchaseOrderHandler(poService)

	router := inventoryHttp.NewRouter(inventoryHttp.RouterConfig{
		ItemHandler:          itemHandler,
		StockHandler:         stockHandler,
		ValuationHandler:     valHandler,
		PurchaseOrderHandler: poHandler,
	})

	log.Printf("Starting Multi-Tenant Inventory Web Service (%s) on port :%s", dbDriver, port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
