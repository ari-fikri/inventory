package http

import (
	"net/http"

	"new-inventory/internal/adapter/http/middleware"
)

type RouterConfig struct {
	// Future handlers will be injected here
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Base items endpoint placeholder
	mux.HandleFunc("GET /api/v1/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	// Wrap entire API with TenantMiddleware
	return middleware.TenantMiddleware(mux)
}
