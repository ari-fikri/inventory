package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

var _ port.ValuationRepository = (*ValuationRepository)(nil)

type ValuationRepository struct {
	db *sql.DB
}

func NewValuationRepository(db *sql.DB) *ValuationRepository {
	return &ValuationRepository{db: db}
}

func (r *ValuationRepository) GetCostLayers(ctx context.Context, tenantID, itemID string) ([]domain.CostLayer, error) {
	query := `
		SELECT id, tenant_id, item_id, warehouse_id, quantity, remaining_quantity, unit_cost, created_at
		FROM stock_cost_layers
		WHERE tenant_id = ? AND item_id = ?
		ORDER BY rowid ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cost layers: %w", err)
	}
	defer rows.Close()

	var layers []domain.CostLayer
	for rows.Next() {
		var l domain.CostLayer
		var createdAtStr string
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ItemID, &l.WarehouseID, &l.Quantity, &l.RemainingQuantity, &l.UnitCost, &createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan cost layer: %w", err)
		}
		if t, err := time.Parse(time.RFC3339Nano, createdAtStr); err == nil {
			l.CreatedAt = t
		}
		layers = append(layers, l)
	}
	return layers, rows.Err()
}

func (r *ValuationRepository) GetActiveCostLayers(ctx context.Context, tenantID, itemID, warehouseID string) ([]domain.CostLayer, error) {
	query := `
		SELECT id, tenant_id, item_id, warehouse_id, quantity, remaining_quantity, unit_cost, created_at
		FROM stock_cost_layers
		WHERE tenant_id = ? AND item_id = ? AND warehouse_id = ? AND remaining_quantity > 0
		ORDER BY rowid ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, itemID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active cost layers: %w", err)
	}
	defer rows.Close()

	var layers []domain.CostLayer
	for rows.Next() {
		var l domain.CostLayer
		var createdAtStr string
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ItemID, &l.WarehouseID, &l.Quantity, &l.RemainingQuantity, &l.UnitCost, &createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan active cost layer: %w", err)
		}
		if t, err := time.Parse(time.RFC3339Nano, createdAtStr); err == nil {
			l.CreatedAt = t
		}
		layers = append(layers, l)
	}
	return layers, rows.Err()
}

func (r *ValuationRepository) UpdateCostLayerRemaining(ctx context.Context, layerID string, remainingQty int) error {
	query := `UPDATE stock_cost_layers SET remaining_quantity = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, remainingQty, layerID)
	if err != nil {
		return fmt.Errorf("failed to update cost layer remaining quantity: %w", err)
	}
	return nil
}

func (r *ValuationRepository) GetTenantValuationMethod(ctx context.Context, tenantID string) (string, error) {
	query := `SELECT valuation_method FROM tenant_settings WHERE tenant_id = ?`
	row := r.db.QueryRowContext(ctx, query, tenantID)
	var method string
	if err := row.Scan(&method); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ValuationMethodFIFO, nil
		}
		return "", fmt.Errorf("failed to query tenant valuation method: %w", err)
	}
	return method, nil
}

func (r *ValuationRepository) SetTenantValuationMethod(ctx context.Context, tenantID, method string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339Nano)
	query := `
		INSERT INTO tenant_settings (tenant_id, valuation_method, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(tenant_id) DO UPDATE SET
			valuation_method = excluded.valuation_method,
			updated_at = excluded.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, tenantID, method, nowStr)
	if err != nil {
		return fmt.Errorf("failed to upsert tenant valuation method: %w", err)
	}
	return nil
}

func (r *ValuationRepository) GetTenantTotalStock(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COALESCE(SUM(quantity), 0) FROM stock_levels WHERE tenant_id = ?`
	row := r.db.QueryRowContext(ctx, query, tenantID)
	var total int
	if err := row.Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to calculate tenant total stock: %w", err)
	}
	return total, nil
}
