package service

import (
	"context"
	"errors"

	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

type ItemService struct {
	repo port.ItemRepository
}

func NewItemService(repo port.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) CreateItem(ctx context.Context, cmd port.CreateItemCommand) (*domain.Item, error) {
	item, err := domain.NewItem(domain.NewItemParams{
		TenantID:     cmd.TenantID,
		SKU:          cmd.SKU,
		Name:         cmd.Name,
		Description:  cmd.Description,
		UnitPrice:    cmd.UnitPrice,
		MinThreshold: cmd.MinThreshold,
	})
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetBySKU(ctx, item.TenantID, item.SKU)
	if err != nil && !errors.Is(err, domain.ErrItemNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrItemAlreadyExists
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) GetItem(ctx context.Context, tenantID, id string) (*domain.Item, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *ItemService) GetItemBySKU(ctx context.Context, tenantID, sku string) (*domain.Item, error) {
	return s.repo.GetBySKU(ctx, tenantID, sku)
}

func (s *ItemService) ListItems(ctx context.Context, tenantID string) ([]*domain.Item, error) {
	return s.repo.List(ctx, tenantID)
}
