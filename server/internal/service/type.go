package service

import (
	"strings"

	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/model"
	"hazard-system/server/internal/repo"
)

// TypeService 隐患类型业务逻辑（每行一个「大类+小类」组合）。
type TypeService struct {
	types   *repo.HazardTypeRepo
	hazards *repo.HazardRepo
}

// NewTypeService 构造 TypeService。
func NewTypeService(types *repo.HazardTypeRepo, hazards *repo.HazardRepo) *TypeService {
	return &TypeService{types: types, hazards: hazards}
}

// List 类型全量。
func (s *TypeService) List() ([]gen.HazardType, *Error) {
	items, err := s.types.List()
	if err != nil {
		return nil, Internal(err)
	}
	out := make([]gen.HazardType, 0, len(items))
	for i := range items {
		out = append(out, toGenHazardType(&items[i]))
	}
	return out, nil
}

// Create 新增隐患类型：需同时提供大类与小类，同一组合唯一。
func (s *TypeService) Create(req gen.HazardTypeCreateRequest) (*gen.HazardType, *Error) {
	major, minor := strings.TrimSpace(req.Major), strings.TrimSpace(req.Minor)
	if major == "" {
		return nil, Unprocessable("大类不能为空")
	}
	if minor == "" {
		return nil, Unprocessable("小类不能为空")
	}

	dup, err := s.types.ExistsMajorMinor(major, minor, 0)
	if err != nil {
		return nil, Internal(err)
	}
	if dup {
		return nil, Unprocessable("「" + major + " / " + minor + "」已存在，无需重复新增")
	}

	t := &model.HazardType{Major: major, Minor: minor}
	if err := s.types.Create(t); err != nil {
		if err == repo.ErrDuplicate {
			return nil, Unprocessable("「" + major + " / " + minor + "」已存在，无需重复新增")
		}
		return nil, Internal(err)
	}
	out := toGenHazardType(t)
	return &out, nil
}

// Update 更新隐患类型：被隐患引用的类型允许改名，但不允许删除。
func (s *TypeService) Update(id int64, req gen.HazardTypeUpdateRequest) (*gen.HazardType, *Error) {
	t, err := s.types.GetByID(uint64(id))
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, Internal(err)
	}

	if req.Major != nil {
		v := strings.TrimSpace(*req.Major)
		if v == "" {
			return nil, Unprocessable("大类不能为空")
		}
		t.Major = v
	}
	if req.Minor != nil {
		v := strings.TrimSpace(*req.Minor)
		if v == "" {
			return nil, Unprocessable("小类不能为空")
		}
		t.Minor = v
	}

	if req.Major != nil || req.Minor != nil {
		dup, dupErr := s.types.ExistsMajorMinor(t.Major, t.Minor, t.ID)
		if dupErr != nil {
			return nil, Internal(dupErr)
		}
		if dup {
			return nil, Unprocessable("「" + t.Major + " / " + t.Minor + "」与已有类型重复")
		}
	}

	if err := s.types.Update(t); err != nil {
		if err == repo.ErrDuplicate {
			return nil, Unprocessable("「" + t.Major + " / " + t.Minor + "」与已有类型重复")
		}
		return nil, Internal(err)
	}
	out := toGenHazardType(t)
	return &out, nil
}

// Delete 删除隐患类型：被隐患引用时拒绝（只允许修改）；
// 未被引用时物理删除（类型枚举无审计价值，且避免软删行占用唯一索引导致同名无法复用）。
func (s *TypeService) Delete(id int64) *Error {
	referenced, err := s.hazards.CountReferencedType(uint64(id))
	if err != nil {
		return Internal(err)
	}
	if referenced > 0 {
		return Conflict("该隐患类型已被隐患记录引用，只能修改，不能删除")
	}
	if err := s.types.Delete(uint64(id)); err == repo.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return Internal(err)
	}
	return nil
}

func toGenHazardType(t *model.HazardType) gen.HazardType {
	return gen.HazardType{
		Id:        int64(t.ID),
		Major:     t.Major,
		Minor:     t.Minor,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
