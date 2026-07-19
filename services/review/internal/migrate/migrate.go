package migrate

import (
	"ai-review-system/services/review/internal/model"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error { return db.AutoMigrate(&model.Review{}, &model.Like{}) }
