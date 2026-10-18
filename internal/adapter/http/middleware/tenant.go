package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

type ErrorResponse struct {
	Error string `json:"error"`
}

// TenantMiddleware validates that X-Tenant-ID is present and attaches it to request context.
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
		if tenantID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: "X-Tenant-ID header is required",
			})
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID retrieves the tenant ID from context.
func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	return tenantID, ok && tenantID != ""
}
