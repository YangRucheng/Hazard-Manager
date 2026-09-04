// Package repo 提供类型化的数据访问层（GORM）。
package repo

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// ErrNotFound 资源不存在。
var ErrNotFound = errors.New("资源不存在")

// HazardFilter 隐患列表筛选条件（全类型化）。
type HazardFilter struct {
	Status   *string
	Level    *string
	TypeID   *uint64
	UnitID   *uint64
	Area     *string
	Keyword  *string
	DateFrom *string // YYYY-MM-DD
	DateTo   *string // YYYY-MM-DD
	Page     int
	PageSize int
}

// DefaultPagination 返回默认分页。
func DefaultPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// HazardRepo 隐患记录数据访问。
type HazardRepo struct{ db *gorm.DB }

// NewHazardRepo 构造 HazardRepo。
func NewHazardRepo(db *gorm.DB) *HazardRepo { return &HazardRepo{db: db} }

func (r *HazardRepo) applyFilter(q *gorm.DB, f HazardFilter) *gorm.DB {
	if f.Status != nil && *f.Status != "" {
		q = q.Where("hazards.status = ?", *f.Status)
	}
	if f.Level != nil && *f.Level != "" {
		q = q.Where("hazards.level = ?", *f.Level)
	}
	if f.TypeID != nil && *f.TypeID > 0 {
		q = q.Where("hazards.type_id = ?", *f.TypeID)
	}
	if f.UnitID != nil && *f.UnitID > 0 {
		q = q.Where("hazards.unit_id = ?", *f.UnitID)
	}
	if f.Area != nil && *f.Area != "" {
		q = q.Where("hazards.inspection_area LIKE ?", "%"+*f.Area+"%")
	}
	if f.Keyword != nil && *f.Keyword != "" {
		kw := "%" + *f.Keyword + "%"
		q = q.Where(`(hazards.description LIKE ? OR hazards.suggestion LIKE ? OR hazards.remark LIKE ?
			OR hazards.inspection_area LIKE ? OR hazards.inspector LIKE ? OR hazards.person LIKE ?)`,
			kw, kw, kw, kw, kw, kw)
	}
	if f.DateFrom != nil && *f.DateFrom != "" {
		q = q.Where("hazards.inspection_date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil && *f.DateTo != "" {
		q = q.Where("hazards.inspection_date <= ?", *f.DateTo)
	}
	return q
}

// hazardRow 查询结果：内嵌表模型 + JOIN 出的名称字段。
// 说明：model.Hazard 中的名称字段标记 gorm:"-"，GORM 不写入表也不扫描，
// 因此 JOIN 查询用本类型承载，再转回 model.Hazard。
type hazardRow struct {
	model.Hazard
	UnitName  string `gorm:"column:unit_name"`
	TypeMajor string `gorm:"column:type_major"`
	TypeMinor string `gorm:"column:type_minor"`
}

func fromRows(rows []hazardRow) []model.Hazard {
	out := make([]model.Hazard, len(rows))
	for i := range rows {
		out[i] = rows[i].Hazard
		out[i].UnitName = rows[i].UnitName
		out[i].TypeMajor = rows[i].TypeMajor
		out[i].TypeMinor = rows[i].TypeMinor
	}
	return out
}

func fromRow(row hazardRow) model.Hazard {
	rows := fromRows([]hazardRow{row})
	return rows[0]
}

// List 分页查询隐患（JOIN 返回单位/类型名称）并按创建时间倒序。
func (r *HazardRepo) List(f HazardFilter) ([]model.Hazard, int64, error) {
	page, pageSize := DefaultPagination(f.Page, f.PageSize)

	base := r.db.Model(&model.Hazard{}).
		Select(`hazards.id, hazards.inspection_area, hazards.inspection_date, hazards.inspector,
			hazards.description, hazards.suggestion, hazards.unit_id, hazards.person, hazards.due_date,
			hazards.recheck_person, hazards.before_images, hazards.status, hazards.after_images,
			hazards.type_id, hazards.level, hazards.remark,
			hazards.created_at, hazards.updated_at,
			COALESCE(u.name, '') AS unit_name,
			COALESCE(t.major, '') AS type_major, COALESCE(t.minor, '') AS type_minor`).
		Joins("LEFT JOIN responsible_units u ON u.id = hazards.unit_id").
		Joins("LEFT JOIN hazard_types t ON t.id = hazards.type_id")

	base = r.applyFilter(base, f)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计隐患失败: %w", err)
	}

	rows := make([]hazardRow, 0)
	err := base.Order("hazards.created_at DESC, hazards.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询隐患列表失败: %w", err)
	}
	return fromRows(rows), total, nil
}

// GetByID 按 id 查询隐患详情。
func (r *HazardRepo) GetByID(id uint64) (*model.Hazard, error) {
	var row hazardRow
	err := r.db.Model(&model.Hazard{}).
		Select(`hazards.id, hazards.inspection_area, hazards.inspection_date, hazards.inspector,
			hazards.description, hazards.suggestion, hazards.unit_id, hazards.person, hazards.due_date,
			hazards.recheck_person, hazards.before_images, hazards.status, hazards.after_images,
			hazards.type_id, hazards.level, hazards.remark,
			hazards.created_at, hazards.updated_at,
			COALESCE(u.name, '') AS unit_name,
			COALESCE(t.major, '') AS type_major, COALESCE(t.minor, '') AS type_minor`).
		Joins("LEFT JOIN responsible_units u ON u.id = hazards.unit_id").
		Joins("LEFT JOIN hazard_types t ON t.id = hazards.type_id").
		Where("hazards.id = ?", id).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询隐患详情失败: %w", err)
	}
	h := fromRow(row)
	return &h, nil
}

// Create 新增隐患。
func (r *HazardRepo) Create(h *model.Hazard) error {
	if err := r.db.Create(h).Error; err != nil {
		return fmt.Errorf("新增隐患失败: %w", err)
	}
	return nil
}

// Update 更新隐患（Save 全字段更新，调用方需传入完整模型）。
func (r *HazardRepo) Update(h *model.Hazard) error {
	if err := r.db.Save(h).Error; err != nil {
		return fmt.Errorf("更新隐患失败: %w", err)
	}
	return nil
}

// Delete 软删除隐患。
func (r *HazardRepo) Delete(id uint64) error {
	res := r.db.Delete(&model.Hazard{}, id)
	if res.Error != nil {
		return fmt.Errorf("删除隐患失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Stats 返回状态计数与逾期未整改数。
func (r *HazardRepo) Stats() (*model.HazardStats, error) {
	stats := &model.HazardStats{}
	rows, err := r.db.Model(&model.Hazard{}).
		Select("status, COUNT(*) AS cnt").
		Group("status").Rows()
	if err != nil {
		return nil, fmt.Errorf("统计隐患状态失败: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status model.HazardStatus
		var cnt int64
		if err := rows.Scan(&status, &cnt); err != nil {
			return nil, fmt.Errorf("解析统计结果失败: %w", err)
		}
		switch status {
		case model.StatusPending:
			stats.Pending = cnt
		case model.StatusBlocked:
			stats.Blocked = cnt
		case model.StatusDone:
			stats.Done = cnt
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历统计结果失败: %w", err)
	}

	today := model.Today().String()
	if err := r.db.Model(&model.Hazard{}).
		Where("due_date < ? AND status <> ?", today, model.StatusDone).
		Count(&stats.Overdue).Error; err != nil {
		return nil, fmt.Errorf("统计逾期隐患失败: %w", err)
	}
	return stats, nil
}

// CountReferenced 统计某单位被（含软删除的）隐患引用的数量，用于删除保护。
func (r *HazardRepo) CountReferencedUnit(unitID uint64) (int64, error) {
	var n int64
	err := r.db.Unscoped().Model(&model.Hazard{}).Where("unit_id = ?", unitID).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("检查单位引用失败: %w", err)
	}
	return n, nil
}

// CountReferencedType 统计某隐患类型被（含软删除的）隐患引用的数量，用于删除保护。
func (r *HazardRepo) CountReferencedType(typeID uint64) (int64, error) {
	var n int64
	err := r.db.Unscoped().Model(&model.Hazard{}).
		Where("type_id = ?", typeID).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("检查类型引用失败: %w", err)
	}
	return n, nil
}
