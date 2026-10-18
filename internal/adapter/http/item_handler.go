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

func (h *ItemHandler) GetBySKU(w http.ResponseWriter, r *http.Request) {
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

	item, err := h.itemService.GetItemBySKU(r.Context(), tenantID, sku)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "internal server error"})
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

	writeJSON(w, http.StatusOK, resp)
}

type UpdateItemRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	UnitPrice    float64 `json:"unit_price"`
	MinThreshold int     `json:"min_threshold"`
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	cmd := port.UpdateItemCommand{
		TenantID:     tenantID,
		SKU:          sku,
		Name:         req.Name,
		Description:  req.Description,
		UnitPrice:    req.UnitPrice,
		MinThreshold: req.MinThreshold,
	}

	item, err := h.itemService.UpdateItem(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrItemNotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
		case errors.Is(err, domain.ErrEmptyName),
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

	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
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

	item, err := h.itemService.DeactivateItem(r.Context(), tenantID, sku)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "internal server error"})
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

	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemHandler) Activate(w http.ResponseWriter, r *http.Request) {
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

	item, err := h.itemService.ActivateItem(r.Context(), tenantID, sku)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "internal server error"})
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

	writeJSON(w, http.StatusOK, resp)
}


type LowStockItemResponseDTO struct {
	ItemResponseDTO
	CurrentQuantity int `json:"current_quantity"`
}

func (h *ItemHandler) ListLowStock(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	items, err := h.itemService.ListLowStock(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to list low stock items"})
		return
	}

	resps := make([]LowStockItemResponseDTO, 0, len(items))
	for _, it := range items {
		resps = append(resps, LowStockItemResponseDTO{
			ItemResponseDTO: ItemResponseDTO{
				ID:           it.Item.ID,
				TenantID:     it.Item.TenantID,
				SKU:          it.Item.SKU,
				Name:         it.Item.Name,
				Description:  it.Item.Description,
				UnitPrice:    it.Item.UnitPrice,
				MinThreshold: it.Item.MinThreshold,
				IsActive:     it.Item.IsActive,
				CreatedAt:    it.Item.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:    it.Item.UpdatedAt.Format(time.RFC3339Nano),
			},
			CurrentQuantity: it.CurrentQuantity,
		})
	}

	writeJSON(w, http.StatusOK, resps)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
