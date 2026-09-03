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
			{ParentID: 0, Name: "电气设备", Sort: 1, Status: 1},
			{ParentID: 0, Name: "安全防护", Sort: 2, Status: 1},
		}
		if err := db.Create(&types).Error; err != nil {
			return err
		}
		categories := []model.HazardType{
			{ParentID: types[0].ID, Name: "线路老化", Sort: 1, Status: 1},
			{ParentID: types[0].ID, Name: "接线不规范", Sort: 2, Status: 1},
			{ParentID: types[0].ID, Name: "绝缘破损", Sort: 3, Status: 1},
			{ParentID: types[1].ID, Name: "警示标识缺失", Sort: 1, Status: 1},
			{ParentID: types[1].ID, Name: "防护罩缺失", Sort: 2, Status: 1},
		}
		if err := db.Create(&categories).Error; err != nil {
			return err
		}
		log.Printf("已写入示例隐患类型 %d 条、分类 %d 条", len(types), len(categories))
	}
	return nil
}
