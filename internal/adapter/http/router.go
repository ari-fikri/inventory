package http

import (
	"net/http"

	"new-inventory/internal/adapter/http/middleware"
)

type RouterConfig struct {
	ItemHandler      *ItemHandler
	StockHandler     *StockHandler
	ValuationHandler *ValuationHandler
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	if cfg.ItemHandler != nil {
		mux.HandleFunc("POST /api/v1/items", cfg.ItemHandler.Create)
		mux.HandleFunc("GET /api/v1/items", cfg.ItemHandler.List)
		mux.HandleFunc("GET /api/v1/items/low-stock", cfg.ItemHandler.ListLowStock)
		mux.HandleFunc("GET /api/v1/items/sku/{sku}", cfg.ItemHandler.GetBySKU)
		mux.HandleFunc("PUT /api/v1/items/sku/{sku}", cfg.ItemHandler.Update)
		mux.HandleFunc("PATCH /api/v1/items/sku/{sku}/deactivate", cfg.ItemHandler.Deactivate)
		mux.HandleFunc("PATCH /api/v1/items/sku/{sku}/activate", cfg.ItemHandler.Activate)
	}

	if cfg.StockHandler != nil {
		mux.HandleFunc("POST /api/v1/items/sku/{sku}/stock/in", cfg.StockHandler.StockIn)
		mux.HandleFunc("POST /api/v1/items/sku/{sku}/stock/out", cfg.StockHandler.StockOut)
		mux.HandleFunc("POST /api/v1/items/sku/{sku}/stock/adjust", cfg.StockHandler.StockAdjust)
		mux.HandleFunc("POST /api/v1/items/sku/{sku}/stock/transfer", cfg.StockHandler.StockTransfer)
		mux.HandleFunc("GET /api/v1/items/sku/{sku}/stock", cfg.StockHandler.GetStock)
		mux.HandleFunc("GET /api/v1/items/sku/{sku}/movements", cfg.StockHandler.ListMovements)
	}

	if cfg.ValuationHandler != nil {
		mux.HandleFunc("GET /api/v1/items/sku/{sku}/valuation", cfg.ValuationHandler.GetItemValuation)
		mux.HandleFunc("GET /api/v1/valuation/settings", cfg.ValuationHandler.GetValuationSettings)
		mux.HandleFunc("PUT /api/v1/valuation/settings", cfg.ValuationHandler.UpdateValuationSettings)
	}

	// Wrap entire API with TenantMiddleware
	return middleware.TenantMiddleware(mux)
}
