package domain_test

import (
	"testing"

	"new-inventory/internal/domain"
)

func TestNewStockMovement_Valid(t *testing.T) {
	movement, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     "tenant-1",
		ItemID:       "item-uuid-1",
		MovementType: domain.MovementTypeIn,
		Quantity:     15,
		BalanceAfter: 30,
		Reference:    "PO-999",
		Notes:        "Restocked from supplier",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if movement.ID == "" {
		t.Errorf("expected non-empty movement ID")
	}
	if movement.MovementType != domain.MovementTypeIn {
		t.Errorf("expected IN, got %s", movement.MovementType)
	}
	if movement.BalanceAfter != 30 {
		t.Errorf("expected balance after 30, got %d", movement.BalanceAfter)
	}
}

func TestNewStockMovement_ValidationErrors(t *testing.T) {
	_, err := domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     "tenant-1",
		ItemID:       "item-uuid-1",
		MovementType: domain.MovementTypeIn,
		Quantity:     0,
		BalanceAfter: 0,
	})
	if err != domain.ErrInvalidQuantity {
		t.Fatalf("expected ErrInvalidQuantity for 0 quantity, got %v", err)
	}

	_, err = domain.NewStockMovement(domain.NewStockMovementParams{
		TenantID:     "tenant-1",
		ItemID:       "item-uuid-1",
		MovementType: domain.MovementTypeOut,
		Quantity:     10,
		BalanceAfter: -5,
	})
	if err != domain.ErrInsufficientStock {
		t.Fatalf("expected ErrInsufficientStock for negative balance, got %v", err)
	}
}
