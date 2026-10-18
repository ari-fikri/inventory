package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"new-inventory/internal/adapter/http/middleware"
	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

type ValuationHandler struct {
	valuationService port.ValuationService
}

func NewValuationHandler(valuationService port.ValuationService) *ValuationHandler {
	return &ValuationHandler{valuationService: valuationService}
}

type ValuationSettingsRequest struct {
	ValuationMethod string `json:"valuation_method"`
}

type ValuationSettingsResponse struct {
	TenantID        string `json:"tenant_id"`
	ValuationMethod string `json:"valuation_method"`
}

func (h *ValuationHandler) GetItemValuation(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	sku := r.PathValue("sku")
	if sku == "" {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "SKU parameter is required"})
		return
	}

	valuation, err := h.valuationService.GetItemValuation(r.Context(), tenantID, sku)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to retrieve item valuation"})
		return
	}

	writeJSON(w, http.StatusOK, valuation)
}

func (h *ValuationHandler) GetValuationSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	method, err := h.valuationService.GetTenantValuationMethod(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to retrieve valuation settings"})
		return
	}

	writeJSON(w, http.StatusOK, ValuationSettingsResponse{
		TenantID:        tenantID,
		ValuationMethod: method,
	})
}

func (h *ValuationHandler) UpdateValuationSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	var req ValuationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	if err := h.valuationService.SetTenantValuationMethod(r.Context(), tenantID, req.ValuationMethod); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidValuationMethod), errors.Is(err, domain.ErrCannotChangeValuation):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to update valuation settings"})
		}
		return
	}

	writeJSON(w, http.StatusOK, ValuationSettingsResponse{
		TenantID:        tenantID,
		ValuationMethod: req.ValuationMethod,
	})
}
