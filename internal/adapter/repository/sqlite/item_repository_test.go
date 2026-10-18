package sqlite_test

import (
	"context"
	"testing"

	"new-inventory/internal/adapter/repository/sqlite"
	"new-inventory/internal/domain"
)

func setupTestDB(t *testing.T) *sqlite.ItemRepository {
	db, err := sqlite.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return sqlite.NewItemRepository(db)
}

func TestItemRepository_CreateAndGet(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	item, err := domain.NewItem(domain.NewItemParams{
		TenantID:     "tenant-alpha",
		SKU:          "KEY-01",
		Name:         "Mechanical Keyboard",
		Description:  "Clicky switches",
		UnitPrice:    89.99,
		MinThreshold: 2,
	})
	if err != nil {
		t.Fatalf("failed to create domain item: %v", err)
	}

	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("failed to save item: %v", err)
	}

	retrieved, err := repo.GetBySKU(ctx, "tenant-alpha", "KEY-01")
	if err != nil {
		t.Fatalf("failed to get item by SKU: %v", err)
	}

	if retrieved.ID != item.ID {
		t.Errorf("expected ID %s, got %s", item.ID, retrieved.ID)
	}
	if retrieved.Name != item.Name {
		t.Errorf("expected Name %s, got %s", item.Name, retrieved.Name)
	}

	// Should not find for another tenant
	_, err = repo.GetBySKU(ctx, "tenant-beta", "KEY-01")
	if err != domain.ErrItemNotFound {
		t.Errorf("expected ErrItemNotFound for tenant-beta, got %v", err)
	}
}

func TestItemRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	item, err := domain.NewItem(domain.NewItemParams{
		TenantID:     "tenant-alpha",
		SKU:          "KEY-UPDATE",
		Name:         "Old Keyboard",
		Description:  "Old Switches",
		UnitPrice:    50.0,
		MinThreshold: 1,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("failed to insert item: %v", err)
	}

	_ = item.Update(domain.UpdateItemParams{
		Name:         "New Keyboard",
		Description:  "New Switches",
		UnitPrice:    75.0,
		MinThreshold: 5,
	})

	if err := repo.Update(ctx, item); err != nil {
		t.Fatalf("failed to update item: %v", err)
	}

	updated, err := repo.GetBySKU(ctx, "tenant-alpha", "KEY-UPDATE")
	if err != nil {
		t.Fatalf("failed to get updated item: %v", err)
	}
	if updated.Name != "New Keyboard" {
		t.Errorf("expected New Keyboard, got %s", updated.Name)
	}
	if updated.UnitPrice != 75.0 {
		t.Errorf("expected 75.0, got %f", updated.UnitPrice)
	}
	if updated.MinThreshold != 5 {
		t.Errorf("expected 5, got %d", updated.MinThreshold)
	}
}

