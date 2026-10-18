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
