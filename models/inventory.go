package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductCategory string

const (
	MobilesAndTabs ProductCategory = "MobilesAndTabs"
	Cloths         ProductCategory = "Cloths"
	Electronics    ProductCategory = "Electronics"
	Grocery        ProductCategory = "Grocery"
	Beauty         ProductCategory = "Beauty"
)

type Inventory struct {
	ID            uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID     uint            `gorm:"not null;uniqueIndex" json:"product_id"`
	Category      ProductCategory `gorm:"type:varchar(50);not null;index" json:"category"`
	TotalStock    int             `gorm:"not null;default:0" json:"total_stock"`
	ReservedStock int             `gorm:"not null;default:0" json:"reserved_stock"`
	// Re-defining AvailableStock to be a managed field (Application Layer synchronization)
	AvailableStock int `gorm:"not null;default:0" json:"available_stock"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeSave hook to ensure consistency
func (i *Inventory) BeforeSave(tx *gorm.DB) (err error) {
	i.AvailableStock = i.TotalStock - i.ReservedStock
	return
}

func (Inventory) TableName() string {
	return "inventories"
}
