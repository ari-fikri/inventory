package main

import (
	"log"
	"net/http"
	"os"

	inventoryHttp "new-inventory/internal/adapter/http"
	"new-inventory/internal/adapter/repository/sqlite"
	"new-inventory/internal/service"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "inventory.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sqlite.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize sqlite: %v", err)
	}
	defer db.Close()

	itemRepo := sqlite.NewItemRepository(db)
	itemService := service.NewItemService(itemRepo)
	itemHandler := inventoryHttp.NewItemHandler(itemService)

	stockRepo := sqlite.NewStockRepository(db)
	stockService := service.NewStockService(itemRepo, stockRepo)
	stockHandler := inventoryHttp.NewStockHandler(stockService)

	valRepo := sqlite.NewValuationRepository(db)
	valService := service.NewValuationService(itemRepo, valRepo)
	valHandler := inventoryHttp.NewValuationHandler(valService)

	supplierRepo := sqlite.NewSupplierRepository(db)
	poRepo := sqlite.NewPurchaseOrderRepository(db)
	poService := service.NewPurchaseOrderService(supplierRepo, poRepo, itemRepo, stockService)
	poHandler := inventoryHttp.NewPurchaseOrderHandler(poService)

	router := inventoryHttp.NewRouter(inventoryHttp.RouterConfig{
		ItemHandler:          itemHandler,
		StockHandler:         stockHandler,
		ValuationHandler:     valHandler,
		PurchaseOrderHandler: poHandler,
	})

	log.Printf("Starting Multi-Tenant Inventory Web Service on port :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
