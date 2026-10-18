package sql_test

import (
	"context"
	"testing"
	"time"

	"new-inventory/internal/adapter/repository/dialect"
	reposql "new-inventory/internal/adapter/repository/sql"
	"new-inventory/internal/domain"
)

func setupTestDB(t *testing.T) *reposql.DB {
	t.Helper()
	db, err := reposql.Open(reposql.Config{
		Driver:        string(dialect.DriverSQLite),
		DSN:           ":memory:",
		RunMigrations: true,
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestSQLItemRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := reposql.NewItemRepository(db.DB, db.Dialect)
	ctx := context.Background()

	item, err := domain.NewItem(domain.NewItemParams{
		TenantID:     "tenant-sql",
		SKU:          "SQL-ITEM-01",
		Name:         "Database Book",
		Description:  "SQL in depth",
		UnitPrice:    45.50,
		MinThreshold: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Create
	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// 2. Duplicate SKU rejected
	if err := repo.Create(ctx, item); err != domain.ErrItemAlreadyExists {
		t.Fatalf("expected ErrItemAlreadyExists on duplicate create, got: %v", err)
	}

	// 3. GetBySKU
	retrieved, err := repo.GetBySKU(ctx, "tenant-sql", "SQL-ITEM-01")
	if err != nil {
		t.Fatalf("failed to retrieve item by SKU: %v", err)
	}
	if retrieved.ID != item.ID || retrieved.Name != item.Name {
		t.Errorf("item data mismatch: got %+v", retrieved)
	}

	// 4. Update
	_ = item.Update(domain.UpdateItemParams{
		Name:         "Updated Database Book",
		Description:  item.Description,
		UnitPrice:    55.00,
		MinThreshold: 10,
	})
	if err := repo.Update(ctx, item); err != nil {
		t.Fatalf("failed to update item: %v", err)
	}

	updated, _ := repo.GetBySKU(ctx, "tenant-sql", "SQL-ITEM-01")
	if updated.Name != "Updated Database Book" || updated.UnitPrice != 55.00 {
		t.Errorf("update mismatch: got %+v", updated)
	}
}

func TestSQLStockAndValuationRepository(t *testing.T) {
	db := setupTestDB(t)
	stockRepo := reposql.NewStockRepository(db.DB, db.Dialect)
	valRepo := reposql.NewValuationRepository(db.DB, db.Dialect)
	ctx := context.Background()

	tenantID := "tenant-stock-sql"
	itemID := "item-uuid-1"
	wh := "WH-TEST"

	// 1. Record stock movement IN with cost layer
	mv, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     tenantID,
		ItemID:       itemID,
		WarehouseID:  wh,
		MovementType: domain.MovementTypeIn,
		Quantity:     50,
		BalanceAfter: 50,
		Reference:    "PO-INIT",
		Notes:        "Initial batch",
	})
	if err != nil {
		t.Fatal(err)
	}

	costLayer := domain.NewCostLayer(tenantID, itemID, wh, 50, 20.0)
	if err := stockRepo.RecordMovementWithCostLayer(ctx, mv, 50, costLayer); err != nil {
		t.Fatalf("failed to record movement with cost layer: %v", err)
	}

	// 2. Query stock level
	stock, err := stockRepo.GetStock(ctx, tenantID, itemID, wh)
	if err != nil {
		t.Fatalf("failed to get stock: %v", err)
	}
	if stock.Quantity != 50 {
		t.Errorf("expected quantity 50, got %d", stock.Quantity)
	}

	// 3. Query active cost layers
	layers, err := valRepo.GetActiveCostLayers(ctx, tenantID, itemID, wh)
	if err != nil {
		t.Fatalf("failed to get active cost layers: %v", err)
	}
	if len(layers) != 1 || layers[0].RemainingQuantity != 50 || layers[0].UnitCost != 20.0 {
		t.Errorf("unexpected cost layers: %+v", layers)
	}

	// 4. Record stock out with valuation
	outMv, _ := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     tenantID,
		ItemID:       itemID,
		WarehouseID:  wh,
		MovementType: domain.MovementTypeOut,
		Quantity:     20,
		BalanceAfter: 30,
		TotalCOGS:    400.0,
		Reference:    "SO-01",
		Notes:        "Fulfillment",
	})
	consumed := []domain.ConsumedCostLayer{
		{
			LayerID:      layers[0].ID,
			ConsumedQty:  20,
			RemainingQty: 30,
			UnitCost:     20.0,
		},
	}
	if err := stockRepo.RecordStockOutWithValuation(ctx, outMv, 30, consumed); err != nil {
		t.Fatalf("failed to record stock out with valuation: %v", err)
	}

	stockAfter, _ := stockRepo.GetStock(ctx, tenantID, itemID, wh)
	if stockAfter.Quantity != 30 {
		t.Errorf("expected quantity 30 after dispatch, got %d", stockAfter.Quantity)
	}

	// 5. Total stock for tenant
	totalStock, err := valRepo.GetTenantTotalStock(ctx, tenantID)
	if err != nil || totalStock != 30 {
		t.Errorf("expected total stock 30, got %d (err: %v)", totalStock, err)
	}

	// 6. Tenant valuation method upsert
	if err := valRepo.SetTenantValuationMethod(ctx, tenantID, "FIFO"); err != nil {
		t.Fatalf("failed to set tenant valuation method: %v", err)
	}
	method, err := valRepo.GetTenantValuationMethod(ctx, tenantID)
	if err != nil || method != "FIFO" {
		t.Errorf("expected FIFO method, got %s (err: %v)", method, err)
	}
}

func TestSQLPurchaseOrderRepository(t *testing.T) {
	db := setupTestDB(t)
	supRepo := reposql.NewSupplierRepository(db.DB, db.Dialect)
	poRepo := reposql.NewPurchaseOrderRepository(db.DB, db.Dialect)
	ctx := context.Background()

	tenantID := "tenant-po-sql"

	// 1. Create Supplier
	supplier := &domain.Supplier{
		ID:          "sup-uuid-1",
		TenantID:    tenantID,
		Name:        "Global Tech Supply",
		ContactName: "Jane Doe",
		Email:       "jane@globaltech.com",
		Phone:       "123-456-7890",
		CreatedAt:   time.Now().UTC(),
	}
	if err := supRepo.Create(ctx, supplier); err != nil {
		t.Fatalf("failed to create supplier: %v", err)
	}

	// Duplicate supplier name rejected
	if err := supRepo.Create(ctx, supplier); err != domain.ErrDuplicateSupplierName {
		t.Fatalf("expected ErrDuplicateSupplierName, got: %v", err)
	}

	// 2. Create Purchase Order
	po, err := domain.NewPurchaseOrder(tenantID, "PO-SQL-100", supplier.ID, "WH-MAIN", "Test order", []domain.PurchaseOrderItem{
		{
			ItemID:          "item-uuid-1",
			SKU:             "SQL-SKU-1",
			OrderedQuantity: 100,
			UnitCost:        25.0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := poRepo.Create(ctx, po); err != nil {
		t.Fatalf("failed to create PO: %v", err)
	}

	// Duplicate PO rejected
	if err := poRepo.Create(ctx, po); err != domain.ErrDuplicatePONumber {
		t.Fatalf("expected ErrDuplicatePONumber, got: %v", err)
	}

	// Retrieve PO
	retrievedPO, err := poRepo.GetByNumber(ctx, tenantID, "PO-SQL-100")
	if err != nil {
		t.Fatalf("failed to get PO by number: %v", err)
	}
	if len(retrievedPO.Items) != 1 || retrievedPO.Items[0].SKU != "SQL-SKU-1" {
		t.Errorf("PO items mismatch: %+v", retrievedPO.Items)
	}

	// Update item received quantity
	if err := poRepo.UpdateItemReceivedQty(ctx, retrievedPO.Items[0].ID, 40); err != nil {
		t.Fatalf("failed to update item received qty: %v", err)
	}

	retrievedPO2, _ := poRepo.GetByNumber(ctx, tenantID, "PO-SQL-100")
	if retrievedPO2.Items[0].ReceivedQuantity != 40 {
		t.Errorf("expected received quantity 40, got %d", retrievedPO2.Items[0].ReceivedQuantity)
	}
}
