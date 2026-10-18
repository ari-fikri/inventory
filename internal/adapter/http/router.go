package http

import (
	"net/http"

	"new-inventory/internal/adapter/http/middleware"
)

type RouterConfig struct {
	ItemHandler *ItemHandler
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	if cfg.ItemHandler != nil {
		mux.HandleFunc("POST /api/v1/items", cfg.ItemHandler.Create)
		mux.HandleFunc("GET /api/v1/items", cfg.ItemHandler.List)
	}

	// Wrap entire API with TenantMiddleware
	return middleware.TenantMiddleware(mux)
}
