package migrate

import (
	"ai-review-system/services/shop/internal/model"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error { return db.AutoMigrate(&model.Shop{}) }
