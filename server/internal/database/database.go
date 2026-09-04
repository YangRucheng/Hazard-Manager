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

// Connect 建立 GORM MySQL 连接，先兼容旧结构再开启 Automigrate。
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 旧版「单表两级」结构（hazard_types.parent_id + hazards.category_id）迁移到新版结构，
	// 必须先于 AutoMigrate 执行，否则旧表加 NOT NULL 新列会失败。
	if err := migrateLegacyHazardSchema(db); err != nil {
		return nil, fmt.Errorf("迁移旧版隐患类型结构失败: %w", err)
	}

	// 责任单位去掉排序功能：删除遗留 sort 列（幂等，列不存在则跳过）。
	if err := migrateUnitDropSort(db); err != nil {
		return nil, fmt.Errorf("迁移责任单位排序列失败: %w", err)
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

// migrateLegacyHazardSchema 将旧版「单表两级」的隐患类型结构迁移为新版「大类+小类两列」：
//   - 旧 hazard_types：parent_id=0 为大类行、>0 为小类行（name 列）。
//   - 旧 hazards：type_id 指向大类行、category_id 指向小类行。
//
// 新版：
//   - hazard_types 每行 = 一个「major(大类)+minor(小类)」组合（小类行的 id 保留不变）。
//   - hazards 只保留 type_id，指向转换后的小类组合行（category_id 删除）。
//
// 通过 information_schema 探测 legacy 列，仅在旧结构存在时执行一次。
func migrateLegacyHazardSchema(db *gorm.DB) error {
	var legacy int64
	err := db.Raw(`SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'hazard_types' AND column_name = 'parent_id'`).
		Scan(&legacy).Error
	if err != nil {
		return err
	}
	if legacy == 0 {
		return nil
	}
	log.Printf("检测到旧版隐患类型结构，正在迁移为「大类+小类」两列结构…")

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE hazard_types ADD COLUMN major varchar(128) NULL, ADD COLUMN minor varchar(128) NULL").Error; err != nil {
			return err
		}
		// 小类行补全 major/minor。
		if err := tx.Exec(`UPDATE hazard_types ht
			JOIN hazard_types p ON p.id = ht.parent_id
			SET ht.major = p.name, ht.minor = ht.name
			WHERE ht.parent_id <> 0`).Error; err != nil {
			return err
		}
		// 隐患记录的类型引用由「大类行」收敛为「小类组合行」。
		if err := tx.Exec("UPDATE hazards SET type_id = category_id WHERE category_id > 0").Error; err != nil {
			return err
		}
		// 大类行（parent_id=0）不再单独存在。
		if err := tx.Exec("DELETE FROM hazard_types WHERE parent_id = 0").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE hazard_types DROP COLUMN parent_id").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE hazard_types DROP COLUMN name").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE hazard_types DROP COLUMN sort").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE hazard_types DROP COLUMN status").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE hazards DROP COLUMN category_id").Error; err != nil {
			return err
		}
		// 迁移先行，AutoMigrate 不会回改既有列的可空性，这里显式补 NOT NULL。
		if err := tx.Exec("ALTER TABLE hazard_types MODIFY COLUMN major varchar(128) NOT NULL, MODIFY COLUMN minor varchar(128) NOT NULL").Error; err != nil {
			return err
		}
		log.Printf("旧版隐患类型结构迁移完成")
		return nil
	})
}

// migrateUnitDropSort 责任单位去排序功能：删除 responsible_units.sort 遗留列。
// 经 information_schema 探测，列存在才执行，幂等。
func migrateUnitDropSort(db *gorm.DB) error {
	var has int64
	err := db.Raw(`SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'responsible_units' AND column_name = 'sort'`).
		Scan(&has).Error
	if err != nil {
		return err
	}
	if has == 0 {
		return nil
	}
	if err := db.Exec("ALTER TABLE responsible_units DROP COLUMN sort").Error; err != nil {
		return err
	}
	log.Printf("已删除责任单位遗留排序列 sort（列表改为按创建顺序返回）")
	return nil
}
