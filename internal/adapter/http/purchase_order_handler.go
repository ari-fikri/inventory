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

type PurchaseOrderHandler struct {
	poService port.PurchaseOrderService
}

func NewPurchaseOrderHandler(poService port.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{poService: poService}
}

type CreateSupplierRequest struct {
	Name        string `json:"name"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

type SupplierResponseDTO struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	CreatedAt   string `json:"created_at"`
}

type POLineItemRequestDTO struct {
	SKU      string  `json:"sku"`
	Quantity int     `json:"quantity"`
	UnitCost float64 `json:"unit_cost"`
}

type CreatePORequest struct {
	PONumber    string                 `json:"po_number"`
	SupplierID  string                 `json:"supplier_id"`
	WarehouseID string                 `json:"warehouse_id"`
	Notes       string                 `json:"notes"`
	Items       []POLineItemRequestDTO `json:"items"`
}

type POLineItemResponseDTO struct {
	ID               string  `json:"id"`
	SKU              string  `json:"sku"`
	OrderedQuantity  int     `json:"ordered_quantity"`
	ReceivedQuantity int     `json:"received_quantity"`
	UnitCost         float64 `json:"unit_cost"`
}

type PurchaseOrderResponseDTO struct {
	ID          string                  `json:"id"`
	TenantID    string                  `json:"tenant_id"`
	PONumber    string                  `json:"po_number"`
	SupplierID  string                  `json:"supplier_id"`
	WarehouseID string                  `json:"warehouse_id"`
	Status      string                  `json:"status"`
	Notes       string                  `json:"notes"`
	Items       []POLineItemResponseDTO `json:"items"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

func (h *PurchaseOrderHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	var req CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	cmd := port.CreateSupplierCommand{
		TenantID:    tenantID,
		Name:        req.Name,
		ContactName: req.ContactName,
		Email:       req.Email,
		Phone:       req.Phone,
	}

	supplier, err := h.poService.CreateSupplier(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateSupplierName) {
			writeJSON(w, http.StatusConflict, middleware.ErrorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		return
	}

	resp := SupplierResponseDTO{
		ID:          supplier.ID,
		TenantID:    supplier.TenantID,
		Name:        supplier.Name,
		ContactName: supplier.ContactName,
		Email:       supplier.Email,
		Phone:       supplier.Phone,
		CreatedAt:   supplier.CreatedAt.Format(time.RFC3339Nano),
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *PurchaseOrderHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	suppliers, err := h.poService.ListSuppliers(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to list suppliers"})
		return
	}

	dtos := make([]SupplierResponseDTO, 0, len(suppliers))
	for _, s := range suppliers {
		dtos = append(dtos, SupplierResponseDTO{
			ID:          s.ID,
			TenantID:    s.TenantID,
			Name:        s.Name,
			ContactName: s.ContactName,
			Email:       s.Email,
			Phone:       s.Phone,
			CreatedAt:   s.CreatedAt.Format(time.RFC3339Nano),
		})
	}

	writeJSON(w, http.StatusOK, dtos)
}

func (h *PurchaseOrderHandler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	var req CreatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	items := make([]port.POLineItemCommand, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, port.POLineItemCommand{
			SKU:      it.SKU,
			Quantity: it.Quantity,
			UnitCost: it.UnitCost,
		})
	}

	cmd := port.CreatePOCommand{
		TenantID:    tenantID,
		PONumber:    req.PONumber,
		SupplierID:  req.SupplierID,
		WarehouseID: req.WarehouseID,
		Notes:       req.Notes,
		Items:       items,
	}

	po, err := h.poService.CreatePurchaseOrder(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDuplicatePONumber):
			writeJSON(w, http.StatusConflict, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrItemNotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "item not found: " + err.Error()})
		case errors.Is(err, domain.ErrEmptyPONumber), errors.Is(err, domain.ErrNoPOItems), errors.Is(err, domain.ErrInvalidQuantity):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to create purchase order: " + err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusCreated, toPODTO(po))
}

func (h *PurchaseOrderHandler) GetPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	poNumber := r.PathValue("po_number")
	if poNumber == "" {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "po_number is required"})
		return
	}

	po, err := h.poService.GetPurchaseOrderByNumber(r.Context(), tenantID, poNumber)
	if err != nil {
		if errors.Is(err, domain.ErrPONotFound) {
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "purchase order not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to retrieve purchase order"})
		return
	}

	writeJSON(w, http.StatusOK, toPODTO(po))
}

func (h *PurchaseOrderHandler) SubmitPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	poNumber := r.PathValue("po_number")
	if poNumber == "" {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "po_number is required"})
		return
	}

	po, err := h.poService.SubmitPurchaseOrder(r.Context(), tenantID, poNumber)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPONotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "purchase order not found"})
		case errors.Is(err, domain.ErrInvalidPOStatus):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to submit purchase order"})
		}
		return
	}

	writeJSON(w, http.StatusOK, toPODTO(po))
}

type ReceivePOItemRequestDTO struct {
	SKU              string `json:"sku"`
	ReceivedQuantity int    `json:"received_quantity"`
}

type ReceivePORequest struct {
	Items []ReceivePOItemRequestDTO `json:"items"`
}

func (h *PurchaseOrderHandler) ReceivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "X-Tenant-ID header is required"})
		return
	}

	poNumber := r.PathValue("po_number")
	if poNumber == "" {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "po_number is required"})
		return
	}

	var req ReceivePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	items := make([]port.ReceivePOItemCommand, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, port.ReceivePOItemCommand{
			SKU:              it.SKU,
			ReceivedQuantity: it.ReceivedQuantity,
		})
	}

	cmd := port.ReceivePOCommand{
		TenantID: tenantID,
		PONumber: poNumber,
		Items:    items,
	}

	po, err := h.poService.ReceivePurchaseOrder(r.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPONotFound):
			writeJSON(w, http.StatusNotFound, middleware.ErrorResponse{Error: "purchase order not found"})
		case errors.Is(err, domain.ErrInvalidPOStatus):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrOverReceiving):
			writeJSON(w, http.StatusUnprocessableEntity, middleware.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuantity), errors.Is(err, domain.ErrItemNotFound):
			writeJSON(w, http.StatusBadRequest, middleware.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, middleware.ErrorResponse{Error: "failed to receive purchase order: " + err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, toPODTO(po))
}

func toPODTO(po *domain.PurchaseOrder) PurchaseOrderResponseDTO {
	itemsDTO := make([]POLineItemResponseDTO, 0, len(po.Items))
	for _, it := range po.Items {
		itemsDTO = append(itemsDTO, POLineItemResponseDTO{
			ID:               it.ID,
			SKU:              it.SKU,
			OrderedQuantity:  it.OrderedQuantity,
			ReceivedQuantity: it.ReceivedQuantity,
			UnitCost:         it.UnitCost,
		})
	}
	return PurchaseOrderResponseDTO{
		ID:          po.ID,
		TenantID:    po.TenantID,
		PONumber:    po.PONumber,
		SupplierID:  po.SupplierID,
		WarehouseID: po.WarehouseID,
		Status:      po.Status,
		Notes:       po.Notes,
		Items:       itemsDTO,
		CreatedAt:   po.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   po.UpdatedAt.Format(time.RFC3339Nano),
	}
}
