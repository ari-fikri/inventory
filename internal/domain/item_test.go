package domain_test

import (
	"testing"

	"new-inventory/internal/domain"
)

func TestNewItem_Valid(t *testing.T) {
	item, err := domain.NewItem(domain.NewItemParams{
		TenantID:     "tenant-1",
		SKU:          "SKU-100",
		Name:         "Gaming Mouse",
		Description:  "RGB Optical Mouse",
		UnitPrice:    49.99,
		MinThreshold: 10,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if item.ID == "" {
		t.Errorf("expected generated UUID, got empty")
	}
	if item.TenantID != "tenant-1" {
		t.Errorf("expected tenant-1, got %s", item.TenantID)
	}
	if item.SKU != "SKU-100" {
		t.Errorf("expected SKU-100, got %s", item.SKU)
	}
	if !item.IsActive {
		t.Errorf("expected item to be active by default")
	}
}

func TestNewItem_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		params      domain.NewItemParams
		expectedErr error
	}{
		{
			name: "missing tenant ID",
			params: domain.NewItemParams{
				TenantID: "",
				SKU:      "SKU-1",
				Name:     "Test",
			},
			expectedErr: domain.ErrEmptyTenantID,
		},
		{
			name: "missing SKU",
			params: domain.NewItemParams{
				TenantID: "tenant-1",
				SKU:      "",
				Name:     "Test",
			},
			expectedErr: domain.ErrEmptySKU,
		},
		{
			name: "missing name",
			params: domain.NewItemParams{
				TenantID: "tenant-1",
				SKU:      "SKU-1",
				Name:     "",
			},
			expectedErr: domain.ErrEmptyName,
		},
		{
			name: "negative unit price",
			params: domain.NewItemParams{
				TenantID:  "tenant-1",
				SKU:       "SKU-1",
				Name:      "Test",
				UnitPrice: -5.0,
			},
			expectedErr: domain.ErrNegativeUnitPrice,
		},
		{
			name: "negative min threshold",
			params: domain.NewItemParams{
				TenantID:     "tenant-1",
				SKU:          "SKU-1",
				Name:         "Test",
				MinThreshold: -1,
			},
			expectedErr: domain.ErrNegativeThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewItem(tt.params)
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
