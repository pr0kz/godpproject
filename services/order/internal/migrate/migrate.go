package migrate

import (
	"ai-review-system/services/order/internal/model"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error { return db.AutoMigrate(&model.Coupon{}, &model.Order{}) }
