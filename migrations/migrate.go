package migrations

import (
	"log"
	"strings"

	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/database"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/models"
)

func RunMigrations() {
	if database.DB == nil {
		log.Println("Database connection is not initialized, skipping migrations")
		return
	}

	err := database.DB.AutoMigrate(&models.Inventory{})
	if err != nil {
		// Handle generated column conflict (SQLSTATE 42611) when converting to managed column
		if strings.Contains(err.Error(), "42611") || strings.Contains(err.Error(), "generated column") {
			log.Println("Migration conflict detected: dropping incompatible 'available_stock' generated column and retrying...")
			// Drop the problematic column
			if execErr := database.DB.Exec("ALTER TABLE inventories DROP COLUMN IF EXISTS available_stock").Error; execErr != nil {
				log.Fatalf("Failed to drop conflict column: %v", execErr)
			}

			// Retry migration
			err = database.DB.AutoMigrate(&models.Inventory{})
			if err == nil {
				// Backfill data since we dropped the column
				log.Println("Backfilling 'available_stock' data...")
				database.DB.Exec("UPDATE inventories SET available_stock = total_stock - reserved_stock")
			}
		}
	}

	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
}
