package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"new-inventory/internal/adapter/http/middleware"
)

func TestTenantMiddleware_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	rec := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.TenantMiddleware(nextHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp middleware.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMsg := "X-Tenant-ID header is required"
	if errResp.Error != expectedMsg {
		t.Fatalf("expected error %q, got %q", expectedMsg, errResp.Error)
	}
}

func TestTenantMiddleware_ValidHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("X-Tenant-ID", "tenant-test-123")
	rec := httptest.NewRecorder()

	var extractedTenant string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := middleware.GetTenantID(r.Context())
		if !ok {
			t.Errorf("expected tenant ID in context, but not found")
		}
		extractedTenant = tenantID
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.TenantMiddleware(nextHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if extractedTenant != "tenant-test-123" {
		t.Fatalf("expected tenant %q, got %q", "tenant-test-123", extractedTenant)
	}
}
