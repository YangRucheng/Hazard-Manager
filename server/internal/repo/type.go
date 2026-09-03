package repo

import (
	stdErrors "errors"
	"fmt"

	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// HazardTypeRepo 隐患类型/分类数据访问（单表两级）。
type HazardTypeRepo struct{ db *gorm.DB }

// NewHazardTypeRepo 构造 HazardTypeRepo。
func NewHazardTypeRepo(db *gorm.DB) *HazardTypeRepo { return &HazardTypeRepo{db: db} }

// List 类型全量（含两类），按 parent_id、sort、id 排序。
func (r *HazardTypeRepo) List() ([]model.HazardType, error) {
	items := make([]model.HazardType, 0)
	err := r.db.Model(&model.HazardType{}).
		Order("parent_id ASC, sort ASC, id ASC").Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("查询类型列表失败: %w", err)
	}
	return items, nil
}

// GetByID 按 id 查询（含软删除，供校验历史引用）。
func (r *HazardTypeRepo) GetByID(id uint64) (*model.HazardType, error) {
	var t model.HazardType
	err := r.db.First(&t, id).Error
	if stdErrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询类型失败: %w", err)
	}
	return &t, nil
}

// GetByIDWithDeleted 含软删除查询。
func (r *HazardTypeRepo) GetByIDWithDeleted(id uint64) (*model.HazardType, error) {
	var t model.HazardType
	err := r.db.Unscoped().First(&t, id).Error
	if stdErrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询类型失败: %w", err)
	}
	return &t, nil
}

// CountChildren 统计某类型（大类）下的分类数量，用于删除保护。
func (r *HazardTypeRepo) CountChildren(parentID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.HazardType{}).Where("parent_id = ?", parentID).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("统计子分类失败: %w", err)
	}
	return n, nil
}

// NameExists 检查同父级下是否已存在同名（排除自身）。
func (r *HazardTypeRepo) NameExists(parentID uint64, name string, excludeID uint64) (bool, error) {
	var n int64
	q := r.db.Model(&model.HazardType{}).Where("parent_id = ? AND name = ?", parentID, name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&n).Error; err != nil {
		return false, fmt.Errorf("检查类型重名失败: %w", err)
	}
	return n > 0, nil
}

// Create 新增类型/分类。
func (r *HazardTypeRepo) Create(t *model.HazardType) error {
	if err := r.db.Create(t).Error; err != nil {
		return fmt.Errorf("新增类型失败: %w", err)
	}
	return nil
}

// Update 更新类型/分类（Save 全字段更新）。
func (r *HazardTypeRepo) Update(t *model.HazardType) error {
	if err := r.db.Save(t).Error; err != nil {
		return fmt.Errorf("更新类型失败: %w", err)
	}
	return nil
}

// Delete 软删除类型/分类。
func (r *HazardTypeRepo) Delete(id uint64) error {
	res := r.db.Delete(&model.HazardType{}, id)
	if res.Error != nil {
		return fmt.Errorf("删除类型失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
