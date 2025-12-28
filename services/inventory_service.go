package services

import (
	"context"
	"errors"

	"github.com/savanrajyaguru/ecommerce-go-inventory-service/config"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/cache"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/models"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/repository"
)

type InventoryService struct {
	repo *repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) CreateInventory(ctx context.Context, inv *models.Inventory) error {
	// Check if already exists
	existing, _ := s.repo.GetByProductID(ctx, inv.ProductID)
	if existing != nil {
		return errors.New("inventory already exists for this product")
	}

	// Default calcs
	inv.ReservedStock = 0
	// inv.AvailableStock = inv.TotalStock // If we were manually managing it

	return s.repo.Create(ctx, inv)
}

func (s *InventoryService) UpdateTotalStock(ctx context.Context, productID uint, newTotal int) error {
	// We might want to check if newTotal >= ReservedStock
	// For simplicity, let's just update for now or check current.
	// Best approach: Transaction

	// Simple update via Repo
	err := s.repo.UpdateStock(ctx, productID, newTotal)
	if err == nil {
		s.invalidateCache(ctx, productID)
	}
	return err
}

func (s *InventoryService) GetInventory(ctx context.Context, productID uint) (*models.Inventory, error) {
	// Cache Look aside
	// TODO: Implement Cache Get/Set similar to Product Service
	// Leaving minimal for now to focus on core logic

	inv, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Fill Computed Field if not using Generated Column in structs or validation
	inv.AvailableStock = inv.TotalStock - inv.ReservedStock

	return inv, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, productID uint, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}

	// Repo handles atomic update
	err := s.repo.ReserveStock(ctx, productID, qty)
	if err == nil {
		s.invalidateCache(ctx, productID)
	}
	return err
}

func (s *InventoryService) ReleaseStock(ctx context.Context, productID uint, qty int) error {
	if qty <= 0 {
		return nil
	}

	err := s.repo.ReleaseStock(ctx, productID, qty)
	if err == nil {
		s.invalidateCache(ctx, productID)
	}
	return err
}

func (s *InventoryService) ListByCategory(ctx context.Context, category string) ([]models.Inventory, error) {
	// Validate enum?
	return s.repo.ListByCategory(ctx, models.ProductCategory(category))
}

func (s *InventoryService) invalidateCache(ctx context.Context, productID uint) {
	if config.AppConfig.Redis.Host != "" && cache.RDB != nil {
		// key := fmt.Sprintf("inventory:%d", productID)
		// cache.RDB.Del(ctx, key)
	}
}
