package database

import (
	"log"

	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// SeedIfEmpty 在枚举表为空时写入示例数据，保证界面可用（可替换为真实数据）。
func SeedIfEmpty(db *gorm.DB) error {
	var unitCount int64
	if err := db.Model(&model.ResponsibleUnit{}).Count(&unitCount).Error; err != nil {
		return err
	}
	if unitCount == 0 {
		units := []model.ResponsibleUnit{
			{Name: "电气车间", Person: "张三", Sort: 1, Status: 1},
			{Name: "动力车间", Person: "李四", Sort: 2, Status: 1},
			{Name: "自动化班组", Person: "王五", Sort: 3, Status: 1},
		}
		if err := db.Create(&units).Error; err != nil {
			return err
		}
		log.Printf("已写入示例责任单位 %d 条", len(units))
	}

	var typeCount int64
	if err := db.Model(&model.HazardType{}).Count(&typeCount).Error; err != nil {
		return err
	}
	if typeCount == 0 {
		types := []model.HazardType{
			{Major: "电气设备", Minor: "线路老化"},
			{Major: "电气设备", Minor: "接线不规范"},
			{Major: "电气设备", Minor: "绝缘破损"},
			{Major: "安全防护", Minor: "警示标识缺失"},
			{Major: "安全防护", Minor: "防护罩缺失"},
		}
		if err := db.Create(&types).Error; err != nil {
			return err
		}
		log.Printf("已写入示例隐患类型 %d 条", len(types))
	}
	return nil
}
