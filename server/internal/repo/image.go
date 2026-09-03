package repo

import (
	"fmt"

	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// ImageRepo 图片摘要表数据访问。
type ImageRepo struct{ db *gorm.DB }

// NewImageRepo 构造 ImageRepo。
func NewImageRepo(db *gorm.DB) *ImageRepo { return &ImageRepo{db: db} }

// Exists 返回指定 uuid 集合中真实存在的子集。
func (r *ImageRepo) Exists(ids []string) (map[string]bool, error) {
	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		result[id] = false
	}
	if len(ids) == 0 {
		return result, nil
	}
	rows := make([]model.Image, 0, len(ids))
	if err := r.db.Model(&model.Image{}).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("校验图片引用失败: %w", err)
	}
	for _, img := range rows {
		result[img.ID] = true
	}
	return result, nil
}
