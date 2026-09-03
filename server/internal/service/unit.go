package service

import (
	"strings"

	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/model"
	"hazard-system/server/internal/repo"
)

// UnitService 责任单位业务逻辑。
type UnitService struct {
	units   *repo.UnitRepo
	hazards *repo.HazardRepo
}

// NewUnitService 构造 UnitService。
func NewUnitService(units *repo.UnitRepo, hazards *repo.HazardRepo) *UnitService {
	return &UnitService{units: units, hazards: hazards}
}

// List 单位列表（含停用项，供下拉与管理页）。
func (s *UnitService) List(keyword *string) ([]gen.ResponsibleUnit, *Error) {
	items, err := s.units.List(keyword)
	if err != nil {
		return nil, Internal(err)
	}
	out := make([]gen.ResponsibleUnit, 0, len(items))
	for i := range items {
		out = append(out, toGenUnit(&items[i]))
	}
	return out, nil
}

// Create 新增单位。
func (s *UnitService) Create(req gen.UnitCreateRequest) (*gen.ResponsibleUnit, *Error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, Unprocessable("单位名称不能为空")
	}
	if strings.TrimSpace(req.Person) == "" {
		return nil, Unprocessable("责任人不能为空（单位与责任人一一对应）")
	}
	u := &model.ResponsibleUnit{
		Name:   strings.TrimSpace(req.Name),
		Person: strings.TrimSpace(req.Person),
		Remark: req.Remark,
		Sort:   0,
		Status: 1,
	}
	if req.Sort != nil {
		u.Sort = *req.Sort
	}
	if req.Status != nil {
		u.Status = int(*req.Status)
	}
	if err := s.units.Create(u); err != nil {
		if strings.Contains(err.Error(), "已存在") {
			return nil, Unprocessable(err.Error())
		}
		return nil, Internal(err)
	}
	out := toGenUnit(u)
	return &out, nil
}

// Update 更新单位。
func (s *UnitService) Update(id int64, req gen.UnitUpdateRequest) (*gen.ResponsibleUnit, *Error) {
	u, err := s.units.GetByID(uint64(id))
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, Internal(err)
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, Unprocessable("单位名称不能为空")
		}
		u.Name = strings.TrimSpace(*req.Name)
	}
	if req.Person != nil {
		if strings.TrimSpace(*req.Person) == "" {
			return nil, Unprocessable("责任人不能为空（单位与责任人一一对应）")
		}
		u.Person = strings.TrimSpace(*req.Person)
	}
	if req.Remark != nil {
		u.Remark = req.Remark
	}
	if req.Sort != nil {
		u.Sort = *req.Sort
	}
	if req.Status != nil {
		u.Status = int(*req.Status)
	}
	if err := s.units.Update(u); err != nil {
		if strings.Contains(err.Error(), "已存在") {
			return nil, Unprocessable(err.Error())
		}
		return nil, Internal(err)
	}
	out := toGenUnit(u)
	return &out, nil
}

// Delete 删除单位（被隐患引用时拒绝）。
func (s *UnitService) Delete(id int64) *Error {
	n, err := s.hazards.CountReferencedUnit(uint64(id))
	if err != nil {
		return Internal(err)
	}
	if n > 0 {
		return Conflict("该责任单位已被隐患记录引用，无法删除")
	}
	if err := s.units.Delete(uint64(id)); err == repo.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return Internal(err)
	}
	return nil
}

func toGenUnit(u *model.ResponsibleUnit) gen.ResponsibleUnit {
	return gen.ResponsibleUnit{
		Id:        int64(u.ID),
		Name:      u.Name,
		Person:    u.Person,
		Remark:    u.Remark,
		Sort:      u.Sort,
		Status:    gen.ResponsibleUnitStatus(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
