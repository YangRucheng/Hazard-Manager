package repo

import (
	stdErrors "errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// ErrDuplicate 唯一键冲突（如并发新增同「大类+小类」组合）。
var ErrDuplicate = stdErrors.New("类型组合已存在")

// isDuplicateKeyError 判断是否为 MySQL 唯一键冲突（1062）。
func isDuplicateKeyError(err error) bool {
	var me *mysql.MySQLError
	return stdErrors.As(err, &me) && me.Number == 1062
}

// HazardTypeRepo 隐患类型数据访问（每行一个「大类+小类」组合）。
type HazardTypeRepo struct{ db *gorm.DB }

// NewHazardTypeRepo 构造 HazardTypeRepo。
func NewHazardTypeRepo(db *gorm.DB) *HazardTypeRepo { return &HazardTypeRepo{db: db} }

// List 类型全量，按大类、小类、id 排序。
func (r *HazardTypeRepo) List() ([]model.HazardType, error) {
	items := make([]model.HazardType, 0)
	err := r.db.Model(&model.HazardType{}).
		Order("major ASC, minor ASC, id ASC").Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("查询类型列表失败: %w", err)
	}
	return items, nil
}

// GetByID 按 id 查询。
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

// ExistsMajorMinor 检查「大类+小类」组合是否已存在（排除自身）。
func (r *HazardTypeRepo) ExistsMajorMinor(major, minor string, excludeID uint64) (bool, error) {
	var n int64
	q := r.db.Model(&model.HazardType{}).Where("major = ? AND minor = ?", major, minor)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&n).Error; err != nil {
		return false, fmt.Errorf("检查类型组合重复失败: %w", err)
	}
	return n > 0, nil
}

// Create 新增类型。
func (r *HazardTypeRepo) Create(t *model.HazardType) error {
	if err := r.db.Create(t).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("新增类型失败: %w", err)
	}
	return nil
}

// Update 更新类型（Save 全字段更新）。
func (r *HazardTypeRepo) Update(t *model.HazardType) error {
	if err := r.db.Save(t).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("更新类型失败: %w", err)
	}
	return nil
}

// Delete 物理删除类型（调用方需先确认未被引用）。
func (r *HazardTypeRepo) Delete(id uint64) error {
	res := r.db.Unscoped().Delete(&model.HazardType{}, id)
	if res.Error != nil {
		return fmt.Errorf("删除类型失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
