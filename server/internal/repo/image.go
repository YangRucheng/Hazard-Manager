package repo

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// ImageRepo 图片摘要表数据访问。
type ImageRepo struct{ db *gorm.DB }

// NewImageRepo 构造 ImageRepo。
func NewImageRepo(db *gorm.DB) *ImageRepo { return &ImageRepo{db: db} }

// ImageFilter 附件列表筛选条件（全类型化）。
type ImageFilter struct {
	Page       int
	PageSize   int
	OnlyUnused bool    // 仅返回未被任何隐患引用的附件
	Keyword    *string // 按 uuid / 摘要前缀匹配
}

// ImageWithRef 图片记录 + 隐患引用计数。
type ImageWithRef struct {
	model.Image
	RefCount int
}

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

// UsageMap 统计 uuid -> 被隐患引用次数（整改前/整改后图片每次出现计一次，
// 软删除隐患不计入，GORM 自动过滤）。
func (r *ImageRepo) UsageMap() (map[string]int, error) {
	var rows []model.Hazard
	if err := r.db.Model(&model.Hazard{}).Select("before_images", "after_images").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("统计图片引用失败: %w", err)
	}
	usage := make(map[string]int)
	for i := range rows {
		for _, id := range model.SplitImages(rows[i].BeforeImages) {
			usage[id]++
		}
		for _, id := range model.SplitImages(rows[i].AfterImages) {
			usage[id]++
		}
	}
	return usage, nil
}

// GetByID 按 uuid 查询图片记录。
func (r *ImageRepo) GetByID(id string) (*model.Image, error) {
	var img model.Image
	if err := r.db.Where("id = ?", id).First(&img).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("查询图片失败: %w", err)
	}
	return &img, nil
}

// Delete 物理删除图片记录（文件由 upload.Store 负责删除）。
func (r *ImageRepo) Delete(id string) error {
	res := r.db.Where("id = ?", id).Delete(&model.Image{})
	if res.Error != nil {
		return fmt.Errorf("删除图片记录失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List 分页查询附件并按引用数过滤。
// onlyUnused 需跨全表判定引用，故该模式下在内存中过滤后分页
// （附件按摘要去重、单厂规模有限，可接受）。
func (r *ImageRepo) List(f ImageFilter) ([]ImageWithRef, int64, error) {
	usage, err := r.UsageMap()
	if err != nil {
		return nil, 0, err
	}

	newQuery := func() *gorm.DB {
		q := r.db.Model(&model.Image{})
		if f.Keyword != nil && *f.Keyword != "" {
			prefix := *f.Keyword + "%"
			q = q.Where("id LIKE ? OR digest LIKE ?", prefix, prefix)
		}
		return q
	}

	if f.OnlyUnused {
		all := make([]model.Image, 0)
		if err := newQuery().Order("created_at DESC, id ASC").Find(&all).Error; err != nil {
			return nil, 0, fmt.Errorf("查询图片列表失败: %w", err)
		}
		unused := make([]model.Image, 0, len(all))
		for _, img := range all {
			if usage[img.ID] == 0 {
				unused = append(unused, img)
			}
		}
		total := int64(len(unused))
		start := (f.Page - 1) * f.PageSize
		if start > len(unused) {
			start = len(unused)
		}
		end := start + f.PageSize
		if end > len(unused) {
			end = len(unused)
		}
		out := make([]ImageWithRef, 0, end-start)
		for _, img := range unused[start:end] {
			out = append(out, ImageWithRef{Image: img, RefCount: 0})
		}
		return out, total, nil
	}

	var total int64
	if err := newQuery().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计图片数量失败: %w", err)
	}
	rows := make([]model.Image, 0, f.PageSize)
	if err := newQuery().
		Order("created_at DESC, id ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("查询图片列表失败: %w", err)
	}
	out := make([]ImageWithRef, 0, len(rows))
	for _, img := range rows {
		out = append(out, ImageWithRef{Image: img, RefCount: usage[img.ID]})
	}
	return out, total, nil
}
