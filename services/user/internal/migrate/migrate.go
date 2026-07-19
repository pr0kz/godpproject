package migrate

import (
	"ai-review-system/services/user/internal/model"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error { return db.AutoMigrate(&model.User{}) }
