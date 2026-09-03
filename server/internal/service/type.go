package service

import (
	"strings"

	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/model"
	"hazard-system/server/internal/repo"
)

// TypeService 隐患类型/分类业务逻辑（单表两级）。
type TypeService struct {
	types   *repo.HazardTypeRepo
	hazards *repo.HazardRepo
}

// NewTypeService 构造 TypeService。
func NewTypeService(types *repo.HazardTypeRepo, hazards *repo.HazardRepo) *TypeService {
	return &TypeService{types: types, hazards: hazards}
}

// List 类型/分类全量（扁平，前端组树）。
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

// Create 新增类型（parentId=0）或分类（parentId=大类id）。
func (s *TypeService) Create(req gen.HazardTypeCreateRequest) (*gen.HazardType, *Error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, Unprocessable("名称不能为空")
	}
	if req.ParentId < 0 {
		return nil, Unprocessable("非法的父级 id")
	}
	parentID := uint64(req.ParentId)
	if parentID > 0 {
		parent, err := s.types.GetByID(parentID)
		if err == repo.ErrNotFound {
			return nil, Unprocessable("父级（大类）不存在")
		}
		if err != nil {
			return nil, Internal(err)
		}
		if parent.ParentID != 0 {
			return nil, Unprocessable("父级必须是隐患类型（大类）")
		}
	}
	name := strings.TrimSpace(req.Name)
	dup, err := s.types.NameExists(parentID, name, 0)
	if err != nil {
		return nil, Internal(err)
	}
	if dup {
		return nil, Unprocessable("同层级下已存在同名「" + name + "」")
	}

	t := &model.HazardType{
		ParentID: parentID,
		Name:     name,
		Sort:     0,
		Status:   1,
	}
	if req.Sort != nil {
		t.Sort = *req.Sort
	}
	if req.Status != nil {
		t.Status = int(*req.Status)
	}
	if err := s.types.Create(t); err != nil {
		return nil, Internal(err)
	}
	out := toGenHazardType(t)
	return &out, nil
}

// Update 更新类型/分类。
func (s *TypeService) Update(id int64, req gen.HazardTypeUpdateRequest) (*gen.HazardType, *Error) {
	t, err := s.types.GetByID(uint64(id))
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, Internal(err)
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, Unprocessable("名称不能为空")
		}
		t.Name = strings.TrimSpace(*req.Name)
	}
	if req.Sort != nil {
		t.Sort = *req.Sort
	}
	if req.Status != nil {
		t.Status = int(*req.Status)
	}
	if req.ParentId != nil {
		newParent := uint64(*req.ParentId)
		if newParent == t.ID {
			return nil, Unprocessable("父级不能是自身")
		}
		if newParent > 0 {
			parent, pErr := s.types.GetByID(newParent)
			if pErr == repo.ErrNotFound {
				return nil, Unprocessable("父级（大类）不存在")
			}
			if pErr != nil {
				return nil, Internal(pErr)
			}
			if parent.ParentID != 0 {
				return nil, Unprocessable("父级必须是隐患类型（大类）")
			}
			// 已有子分类的大类不能降级为小类，否则破坏层级。
			children, cErr := s.types.CountChildren(t.ID)
			if cErr != nil {
				return nil, Internal(cErr)
			}
			if children > 0 {
				return nil, Unprocessable("该大类下存在分类，不能改为小类")
			}
		}
		t.ParentID = newParent
	}

	if req.Name != nil {
		dup, dupErr := s.types.NameExists(t.ParentID, t.Name, t.ID)
		if dupErr != nil {
			return nil, Internal(dupErr)
		}
		if dup {
			return nil, Unprocessable("同层级下已存在同名「" + t.Name + "」")
		}
	}

	if err := s.types.Update(t); err != nil {
		return nil, Internal(err)
	}
	out := toGenHazardType(t)
	return &out, nil
}

// Delete 删除类型/分类（存在子分类或被隐患引用时拒绝）。
func (s *TypeService) Delete(id int64) *Error {
	children, err := s.types.CountChildren(uint64(id))
	if err != nil {
		return Internal(err)
	}
	if children > 0 {
		return Conflict("该类型下存在分类，无法删除")
	}
	referenced, err := s.hazards.CountReferencedType(uint64(id))
	if err != nil {
		return Internal(err)
	}
	if referenced > 0 {
		return Conflict("该类型/分类已被隐患记录引用，无法删除")
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
		ParentId:  int64(t.ParentID),
		Name:      t.Name,
		Sort:      t.Sort,
		Status:    gen.HazardTypeStatus(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
