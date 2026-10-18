package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"new-inventory/internal/adapter/http/middleware"
	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

type ItemHandler struct {
	itemService port.ItemService
}

func NewItemHandler(itemService port.ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

type CreateItemRequest struct {
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	UnitPrice    float64 `json:"unit_price"`
	MinThreshold int     `json:"min_threshold"`
}

type ItemResponseDTO struct {
	ID           string  `json:"id"`
	TenantID     string  `json:"tenant_id"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	UnitPrice    float64 `json:"unit_price"`
	MinThreshold int     `json:"min_threshold"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	cmd := port.CreateItemCommand{
		TenantID:     tenantID,
		SKU:          req.SKU,
		Name:         req.Name,
		Description:  req.Description,
		UnitPrice:    req.UnitPrice,
		MinThreshold: req.MinThreshold,
	}

	item, err := h.itemService.CreateItem(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrItemAlreadyExists):
			writeJSON(w, http.StatusConflict, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrEmptySKU),
			errors.Is(err, domain.ErrEmptyName),
			errors.Is(err, domain.ErrEmptyTenantID),
			errors.Is(err, domain.ErrNegativeUnitPrice),
			errors.Is(err, domain.ErrNegativeThreshold):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	resp := ItemResponseDTO{
		ID:           item.ID,
		TenantID:     item.TenantID,
		SKU:          item.SKU,
		Name:         item.Name,
		Description:  item.Description,
		UnitPrice:    item.UnitPrice,
		MinThreshold: item.MinThreshold,
		IsActive:     item.IsActive,
		CreatedAt:    item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    item.UpdatedAt.Format(time.RFC3339Nano),
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	items, err := h.itemService.ListItems(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to list items"})
		return
	}

	resps := make([]ItemResponseDTO, 0, len(items))
	for _, item := range items {
		resps = append(resps, ItemResponseDTO{
			ID:           item.ID,
			TenantID:     item.TenantID,
			SKU:          item.SKU,
			Name:         item.Name,
			Description:  item.Description,
			UnitPrice:    item.UnitPrice,
			MinThreshold: item.MinThreshold,
			IsActive:     item.IsActive,
			CreatedAt:    item.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    item.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	writeJSON(w, http.StatusOK, resps)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
