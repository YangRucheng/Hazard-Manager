package repo

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// UnitRepo 责任单位数据访问。
type UnitRepo struct{ db *gorm.DB }

// NewUnitRepo 构造 UnitRepo。
func NewUnitRepo(db *gorm.DB) *UnitRepo { return &UnitRepo{db: db} }

// List 单位列表（含停用项），可按关键词筛选，按创建顺序（id）返回。
func (r *UnitRepo) List(keyword *string) ([]model.ResponsibleUnit, error) {
	q := r.db.Model(&model.ResponsibleUnit{})
	if keyword != nil && *keyword != "" {
		kw := "%" + *keyword + "%"
		q = q.Where("name LIKE ? OR person LIKE ?", kw, kw)
	}
	items := make([]model.ResponsibleUnit, 0)
	err := q.Order("id ASC").Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("查询单位列表失败: %w", err)
	}
	return items, nil
}

// GetByID 按 id 查询单位。
func (r *UnitRepo) GetByID(id uint64) (*model.ResponsibleUnit, error) {
	var u model.ResponsibleUnit
	err := r.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询单位失败: %w", err)
	}
	return &u, nil
}

// GetByIDWithDeleted 含软删除查询（用于类型/单位联动校验时复用）。
func (r *UnitRepo) GetByIDWithDeleted(id uint64) (*model.ResponsibleUnit, error) {
	var u model.ResponsibleUnit
	err := r.db.Unscoped().First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询单位失败: %w", err)
	}
	return &u, nil
}

// Create 新增单位。
func (r *UnitRepo) Create(u *model.ResponsibleUnit) error {
	if err := r.db.Create(u).Error; err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("单位名称已存在")
		}
		return fmt.Errorf("新增单位失败: %w", err)
	}
	return nil
}

// Update 更新单位（Save 全字段更新，确保 status=0 停用等零值生效）。
func (r *UnitRepo) Update(u *model.ResponsibleUnit) error {
	if err := r.db.Save(u).Error; err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("单位名称已存在")
		}
		return fmt.Errorf("更新单位失败: %w", err)
	}
	return nil
}

// Delete 软删除单位。
func (r *UnitRepo) Delete(id uint64) error {
	res := r.db.Delete(&model.ResponsibleUnit{}, id)
	if res.Error != nil {
		return fmt.Errorf("删除单位失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func isDuplicateError(err error) bool {
	return err != nil && (contains(err.Error(), "Duplicate entry") || contains(err.Error(), "Error 1062"))
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
