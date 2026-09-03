// Package database 负责 MySQL 连接与表结构迁移。
package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"hazard-system/server/internal/model"
)

// Connect 建立 GORM MySQL 连接并开启 Automigrate。
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 建表顺序：枚举表先于主表。
	if err := db.AutoMigrate(
		&model.ResponsibleUnit{},
		&model.HazardType{},
		&model.Hazard{},
		&model.Image{},
	); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	log.Printf("数据库连接与迁移完成")
	return db, nil
}
