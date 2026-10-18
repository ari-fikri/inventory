package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

var _ port.SupplierRepository = (*SupplierRepository)(nil)
var _ port.PurchaseOrderRepository = (*PurchaseOrderRepository)(nil)

type SupplierRepository struct {
	db *sql.DB
}

func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) Create(ctx context.Context, s *domain.Supplier) error {
	query := `
		INSERT INTO suppliers (id, tenant_id, name, contact_name, email, phone, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		s.ID,
		s.TenantID,
		s.Name,
		s.ContactName,
		s.Email,
		s.Phone,
		s.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrDuplicateSupplierName
		}
		return fmt.Errorf("failed to insert supplier: %w", err)
	}
	return nil
}

func (r *SupplierRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Supplier, error) {
	query := `
		SELECT id, tenant_id, name, contact_name, email, phone, created_at
		FROM suppliers
		WHERE tenant_id = ? AND id = ?
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)
	return scanSupplier(row)
}

func (r *SupplierRepository) GetByName(ctx context.Context, tenantID, name string) (*domain.Supplier, error) {
	query := `
		SELECT id, tenant_id, name, contact_name, email, phone, created_at
		FROM suppliers
		WHERE tenant_id = ? AND name = ?
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, name)
	return scanSupplier(row)
}

func (r *SupplierRepository) List(ctx context.Context, tenantID string) ([]*domain.Supplier, error) {
	query := `
		SELECT id, tenant_id, name, contact_name, email, phone, created_at
		FROM suppliers
		WHERE tenant_id = ?
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []*domain.Supplier
	for rows.Next() {
		var s domain.Supplier
		var createdAtStr string
		if err := rows.Scan(&s.ID, &s.TenantID, &s.Name, &s.ContactName, &s.Email, &s.Phone, &createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		if t, err := time.Parse(time.RFC3339Nano, createdAtStr); err == nil {
			s.CreatedAt = t
		}
		suppliers = append(suppliers, &s)
	}
	return suppliers, rows.Err()
}

func scanSupplier(row *sql.Row) (*domain.Supplier, error) {
	var s domain.Supplier
	var createdAtStr string
	err := row.Scan(&s.ID, &s.TenantID, &s.Name, &s.ContactName, &s.Email, &s.Phone, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to scan supplier: %w", err)
	}
	if t, err := time.Parse(time.RFC3339Nano, createdAtStr); err == nil {
		s.CreatedAt = t
	}
	return &s, nil
}

type PurchaseOrderRepository struct {
	db *sql.DB
}

func NewPurchaseOrderRepository(db *sql.DB) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

func (r *PurchaseOrderRepository) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertPO := `
		INSERT INTO purchase_orders (id, tenant_id, po_number, supplier_id, warehouse_id, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(
		ctx,
		insertPO,
		po.ID,
		po.TenantID,
		po.PONumber,
		po.SupplierID,
		po.WarehouseID,
		po.Status,
		po.Notes,
		po.CreatedAt.Format(time.RFC3339Nano),
		po.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrDuplicatePONumber
		}
		return fmt.Errorf("failed to insert purchase order: %w", err)
	}

	insertItem := `
		INSERT INTO purchase_order_items (id, po_id, tenant_id, item_id, sku, ordered_quantity, received_quantity, unit_cost)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	for _, it := range po.Items {
		_, err = tx.ExecContext(
			ctx,
			insertItem,
			it.ID,
			po.ID,
			po.TenantID,
			it.ItemID,
			it.SKU,
			it.OrderedQuantity,
			it.ReceivedQuantity,
			it.UnitCost,
		)
		if err != nil {
			return fmt.Errorf("failed to insert po item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit purchase order transaction: %w", err)
	}
	return nil
}

func (r *PurchaseOrderRepository) GetByNumber(ctx context.Context, tenantID, poNumber string) (*domain.PurchaseOrder, error) {
	query := `
		SELECT id, tenant_id, po_number, supplier_id, warehouse_id, status, notes, created_at, updated_at
		FROM purchase_orders
		WHERE tenant_id = ? AND po_number = ?
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, poNumber)
	po, err := scanPO(row)
	if err != nil {
		return nil, err
	}

	items, err := r.getItemsForPO(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	po.Items = items
	return po, nil
}

func (r *PurchaseOrderRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.PurchaseOrder, error) {
	query := `
		SELECT id, tenant_id, po_number, supplier_id, warehouse_id, status, notes, created_at, updated_at
		FROM purchase_orders
		WHERE tenant_id = ? AND id = ?
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)
	po, err := scanPO(row)
	if err != nil {
		return nil, err
	}

	items, err := r.getItemsForPO(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	po.Items = items
	return po, nil
}

func (r *PurchaseOrderRepository) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	query := `
		UPDATE purchase_orders
		SET status = ?, notes = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	nowStr := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.ExecContext(ctx, query, po.Status, po.Notes, nowStr, po.TenantID, po.ID)
	if err != nil {
		return fmt.Errorf("failed to update purchase order: %w", err)
	}
	return nil
}

func (r *PurchaseOrderRepository) UpdateItemReceivedQty(ctx context.Context, itemID string, receivedQty int) error {
	query := `UPDATE purchase_order_items SET received_quantity = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, receivedQty, itemID)
	if err != nil {
		return fmt.Errorf("failed to update po item received quantity: %w", err)
	}
	return nil
}

func (r *PurchaseOrderRepository) getItemsForPO(ctx context.Context, poID string) ([]domain.PurchaseOrderItem, error) {
	query := `
		SELECT id, po_id, tenant_id, item_id, sku, ordered_quantity, received_quantity, unit_cost
		FROM purchase_order_items
		WHERE po_id = ?
		ORDER BY rowid ASC
	`
	rows, err := r.db.QueryContext(ctx, query, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to query po items: %w", err)
	}
	defer rows.Close()

	var items []domain.PurchaseOrderItem
	for rows.Next() {
		var it domain.PurchaseOrderItem
		if err := rows.Scan(&it.ID, &it.POID, &it.TenantID, &it.ItemID, &it.SKU, &it.OrderedQuantity, &it.ReceivedQuantity, &it.UnitCost); err != nil {
			return nil, fmt.Errorf("failed to scan po item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func scanPO(row *sql.Row) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	var createdAtStr, updatedAtStr string
	err := row.Scan(&po.ID, &po.TenantID, &po.PONumber, &po.SupplierID, &po.WarehouseID, &po.Status, &po.Notes, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPONotFound
		}
		return nil, fmt.Errorf("failed to scan purchase order: %w", err)
	}
	if t, err := time.Parse(time.RFC3339Nano, createdAtStr); err == nil {
		po.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339Nano, updatedAtStr); err == nil {
		po.UpdatedAt = t
	}
	return &po, nil
}
