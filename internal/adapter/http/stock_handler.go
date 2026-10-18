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

type StockHandler struct {
	stockService port.StockService
}

func NewStockHandler(stockService port.StockService) *StockHandler {
	return &StockHandler{stockService: stockService}
}

type StockInRequest struct {
	Quantity    int    `json:"quantity"`
	WarehouseID string `json:"warehouse_id"`
	Reference   string `json:"reference"`
	Notes       string `json:"notes"`
}

type StockInResponse struct {
	ItemID       string `json:"item_id"`
	TenantID     string `json:"tenant_id"`
	WarehouseID  string `json:"warehouse_id"`
	Quantity     int    `json:"quantity"`
	MovementType string `json:"movement_type"`
	MovementID   string `json:"movement_id"`
	Reference    string `json:"reference"`
	Notes        string `json:"notes"`
}

type StockLevelResponse struct {
	ItemID      string `json:"item_id"`
	TenantID    string `json:"tenant_id"`
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
	UpdatedAt   string `json:"updated_at"`
}

type StockMovementResponse struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	ItemID       string `json:"item_id"`
	WarehouseID  string `json:"warehouse_id"`
	MovementType string `json:"movement_type"`
	Quantity     int    `json:"quantity"`
	BalanceAfter int    `json:"balance_after"`
	Reference    string `json:"reference"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"created_at"`
}

func (h *StockHandler) StockIn(w http.ResponseWriter, r *http.Request) {
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

	var req StockInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	cmd := port.StockInCommand{
		TenantID:    tenantID,
		SKU:         sku,
		WarehouseID: req.WarehouseID,
		Quantity:    req.Quantity,
		Reference:   req.Reference,
		Notes:       req.Notes,
	}

	movement, stockLevel, err := h.stockService.StockIn(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrItemNotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
		case errors.Is(err, domain.ErrItemInactive):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuantity):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to process stock in"})
		}
		return
	}

	resp := StockInResponse{
		ItemID:       stockLevel.ItemID,
		TenantID:     stockLevel.TenantID,
		WarehouseID:  stockLevel.WarehouseID,
		Quantity:     stockLevel.Quantity,
		MovementType: movement.MovementType,
		MovementID:   movement.ID,
		Reference:    movement.Reference,
		Notes:        movement.Notes,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *StockHandler) StockOut(w http.ResponseWriter, r *http.Request) {
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

	var req StockInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	cmd := port.StockOutCommand{
		TenantID:    tenantID,
		SKU:         sku,
		WarehouseID: req.WarehouseID,
		Quantity:    req.Quantity,
		Reference:   req.Reference,
		Notes:       req.Notes,
	}

	movement, stockLevel, err := h.stockService.StockOut(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrItemNotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
		case errors.Is(err, domain.ErrItemInactive):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInsufficientStock):
			writeJSON(w, http.StatusUnprocessableEntity, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuantity):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to process stock out"})
		}
		return
	}

	resp := StockInResponse{
		ItemID:       stockLevel.ItemID,
		TenantID:     stockLevel.TenantID,
		WarehouseID:  stockLevel.WarehouseID,
		Quantity:     stockLevel.Quantity,
		MovementType: movement.MovementType,
		MovementID:   movement.ID,
		Reference:    movement.Reference,
		Notes:        movement.Notes,
	}

	writeJSON(w, http.StatusOK, resp)
}

type StockAdjustRequest struct {
	TargetQuantity int    `json:"target_quantity"`
	WarehouseID    string `json:"warehouse_id"`
	Reference      string `json:"reference"`
	Notes          string `json:"notes"`
}

func (h *StockHandler) StockAdjust(w http.ResponseWriter, r *http.Request) {
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

	var req StockAdjustRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	cmd := port.StockAdjustCommand{
		TenantID:       tenantID,
		SKU:            sku,
		WarehouseID:    req.WarehouseID,
		TargetQuantity: req.TargetQuantity,
		Reference:      req.Reference,
		Notes:          req.Notes,
	}

	movement, stockLevel, err := h.stockService.StockAdjust(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrItemNotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
		case errors.Is(err, domain.ErrItemInactive):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuantity):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to process stock adjust"})
		}
		return
	}

	resp := StockInResponse{
		ItemID:       stockLevel.ItemID,
		TenantID:     stockLevel.TenantID,
		WarehouseID:  stockLevel.WarehouseID,
		Quantity:     stockLevel.Quantity,
		MovementType: movement.MovementType,
		MovementID:   movement.ID,
		Reference:    movement.Reference,
		Notes:        movement.Notes,
	}

	writeJSON(w, http.StatusOK, resp)
}

type StockTransferRequest struct {
	SourceWarehouseID string `json:"source_warehouse_id"`
	TargetWarehouseID string `json:"target_warehouse_id"`
	Quantity          int    `json:"quantity"`
	Reference         string `json:"reference"`
	Notes             string `json:"notes"`
}

type StockTransferResponse struct {
	ItemID            string `json:"item_id"`
	TenantID          string `json:"tenant_id"`
	SourceWarehouseID string `json:"source_warehouse_id"`
	TargetWarehouseID string `json:"target_warehouse_id"`
	Quantity          int    `json:"quantity"`
	SourceBalance     int    `json:"source_balance"`
	TargetBalance     int    `json:"target_balance"`
	Reference         string `json:"reference"`
	Notes             string `json:"notes"`
}

func (h *StockHandler) StockTransfer(w http.ResponseWriter, r *http.Request) {
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

	var req StockTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	cmd := port.StockTransferCommand{
		TenantID:          tenantID,
		SKU:               sku,
		SourceWarehouseID: req.SourceWarehouseID,
		TargetWarehouseID: req.TargetWarehouseID,
		Quantity:          req.Quantity,
		Reference:         req.Reference,
		Notes:             req.Notes,
	}

	result, err := h.stockService.StockTransfer(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrItemNotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
		case errors.Is(err, domain.ErrItemInactive):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInsufficientStock):
			writeJSON(w, http.StatusUnprocessableEntity, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuantity), errors.Is(err, domain.ErrSameWarehouse):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to process stock transfer"})
		}
		return
	}

	resp := StockTransferResponse{
		ItemID:            result.ItemID,
		TenantID:          result.TenantID,
		SourceWarehouseID: result.SourceWarehouseID,
		TargetWarehouseID: result.TargetWarehouseID,
		Quantity:          result.Quantity,
		SourceBalance:     result.SourceBalance,
		TargetBalance:     result.TargetBalance,
		Reference:         req.Reference,
		Notes:             req.Notes,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *StockHandler) GetStock(w http.ResponseWriter, r *http.Request) {
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

	warehouseID := r.URL.Query().Get("warehouse_id")
	stock, err := h.stockService.GetStock(r.Context(), tenantID, sku, warehouseID)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to query stock"})
		return
	}

	resp := StockLevelResponse{
		ItemID:      stock.ItemID,
		TenantID:    stock.TenantID,
		WarehouseID: stock.WarehouseID,
		Quantity:    stock.Quantity,
		UpdatedAt:   stock.UpdatedAt.Format(time.RFC3339Nano),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *StockHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
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

	movements, err := h.stockService.ListMovements(r.Context(), tenantID, sku)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to query movements"})
		return
	}

	resps := make([]StockMovementResponse, 0, len(movements))
	for _, m := range movements {
		resps = append(resps, StockMovementResponse{
			ID:           m.ID,
			TenantID:     m.TenantID,
			ItemID:       m.ItemID,
			MovementType: m.MovementType,
			Quantity:     m.Quantity,
			BalanceAfter: m.BalanceAfter,
			Reference:    m.Reference,
			Notes:        m.Notes,
			CreatedAt:    m.CreatedAt.Format(time.RFC3339Nano),
		})
	}

	writeJSON(w, http.StatusOK, resps)
}
