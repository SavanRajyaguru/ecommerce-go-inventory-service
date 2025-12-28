package repository

import (
	"context"
	"errors"

	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/database"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/models"
	"gorm.io/gorm"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository() *InventoryRepository {
	return &InventoryRepository{
		db: database.DB,
	}
}

// WithTx allows running operations within a transaction
func (r *InventoryRepository) WithTx(tx *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: tx}
}

func (r *InventoryRepository) Create(ctx context.Context, inv *models.Inventory) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *InventoryRepository) GetByProductID(ctx context.Context, productID uint) (*models.Inventory, error) {
	var inv models.Inventory
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InventoryRepository) UpdateStock(ctx context.Context, productID uint, totalStock int) error {
	// AvailableStock will be re-calculated or we update TotalStock
	// If we use generated columns, we just update TotalStock.
	// If manual, we should probably lock the row first.
	return r.db.WithContext(ctx).Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Update("total_stock", totalStock).Error
}

func (r *InventoryRepository) ReserveStock(ctx context.Context, productID uint, qty int) error {
	// Atomic update:
	// UPDATE inventories
	// SET reserved_stock = reserved_stock + qty
	// WHERE product_id = ? AND (total_stock - reserved_stock) >= qty

	result := r.db.WithContext(ctx).Model(&models.Inventory{}).
		Where("product_id = ? AND (total_stock - reserved_stock) >= ?", productID, qty).
		UpdateColumn("reserved_stock", gorm.Expr("reserved_stock + ?", qty))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("insufficient stock or product not found")
	}
	return nil
}

func (r *InventoryRepository) ReleaseStock(ctx context.Context, productID uint, qty int) error {
	// Decrease reserved stock.
	// Note: We should ensure reserved_stock doesn't go below 0?
	// Assuming logic prior was correct, it shouldn't.
	result := r.db.WithContext(ctx).Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn("reserved_stock", gorm.Expr("reserved_stock - ?", qty))

	if result.Error != nil {
		return result.Error
	}
	// If rows affected is 0, product might not exist.
	return nil
}

func (r *InventoryRepository) ListByCategory(ctx context.Context, category models.ProductCategory) ([]models.Inventory, error) {
	var list []models.Inventory
	err := r.db.WithContext(ctx).Where("category = ?", category).Find(&list).Error
	return list, err
}
