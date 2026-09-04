package service

import (
	"github.com/oapi-codegen/runtime/types"
	"strings"

	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/model"
	"hazard-system/server/internal/repo"
)

// 默认值常量。
const (
	DefaultArea    = "华星现场"
	DefaultPerson  = "电气自查"
	DefaultDueDays = 7
)

// HazardService 隐患记录业务逻辑。
type HazardService struct {
	hazards *repo.HazardRepo
	units   *repo.UnitRepo
	types   *repo.HazardTypeRepo
	images  *repo.ImageRepo
}

// NewHazardService 构造 HazardService。
func NewHazardService(hazards *repo.HazardRepo, units *repo.UnitRepo, types *repo.HazardTypeRepo, images *repo.ImageRepo) *HazardService {
	return &HazardService{hazards: hazards, units: units, types: types, images: images}
}

// List 分页查询并转换为响应 DTO。
func (s *HazardService) List(params gen.ListHazardsParams) (*gen.HazardListResponse, *Error) {
	filter := repo.HazardFilter{
		Status:   optionalString(params.Status),
		Level:    optionalString(params.Level),
		TypeID:   optionalUint64(params.TypeId),
		UnitID:   optionalUint64(params.UnitId),
		Area:     params.Area,
		Keyword:  params.Keyword,
		Page:     derefInt(params.Page, 1),
		PageSize: derefInt(params.PageSize, 20),
	}
	if params.DateFrom != nil {
		v := (*params.DateFrom).Time.Format("2006-01-02")
		filter.DateFrom = &v
	}
	if params.DateTo != nil {
		v := (*params.DateTo).Time.Format("2006-01-02")
		filter.DateTo = &v
	}

	items, total, err := s.hazards.List(filter)
	if err != nil {
		return nil, Internal(err)
	}

	resp := &gen.HazardListResponse{
		Items: make([]gen.Hazard, 0, len(items)),
		Pagination: gen.Pagination{
			Page:     filter.Page,
			PageSize: filter.PageSize,
			Total:    total,
		},
	}
	for i := range items {
		resp.Items = append(resp.Items, toGenHazard(&items[i]))
	}
	return resp, nil
}

// Get 查询详情。
func (s *HazardService) Get(id int64) (*gen.Hazard, *Error) {
	h, err := s.hazards.GetByID(uint64(id))
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, Internal(err)
	}
	out := toGenHazard(h)
	return &out, nil
}

// Create 新增隐患：默认值兜底 + 单位->责任人联动 + 分类归属校验 + 图片校验。
func (s *HazardService) Create(req gen.HazardCreateRequest) (*gen.Hazard, *Error) {
	if strings.TrimSpace(req.Description) == "" {
		return nil, Unprocessable("隐患描述不能为空")
	}
	if req.UnitId <= 0 {
		return nil, Unprocessable("请选择责任单位")
	}
	if req.TypeId <= 0 {
		return nil, Unprocessable("请选择隐患类型")
	}

	h := &model.Hazard{
		Description: strings.TrimSpace(req.Description),
		Suggestion:  req.Suggestion,
		Remark:      req.Remark,
	}

	// 默认值。
	h.InspectionArea = DefaultArea
	if req.InspectionArea != nil {
		if v := strings.TrimSpace(*req.InspectionArea); v != "" {
			h.InspectionArea = v
		}
	}
	if req.InspectionDate != nil && !(*req.InspectionDate).Time.IsZero() {
		h.InspectionDate = model.Date{Time: (*req.InspectionDate).Time}
	} else {
		h.InspectionDate = model.Today()
	}
	h.Inspector = DefaultPerson
	if req.Inspector != nil {
		if v := strings.TrimSpace(*req.Inspector); v != "" {
			h.Inspector = v
		}
	}
	if req.DueDate != nil && !(*req.DueDate).Time.IsZero() {
		h.DueDate = model.Date{Time: (*req.DueDate).Time}
	} else {
		h.DueDate = h.InspectionDate.AddDays(DefaultDueDays)
	}

	// 复查人员：新增为空则填检查人员。
	if req.RecheckPerson == nil || strings.TrimSpace(*req.RecheckPerson) == "" {
		h.RecheckPerson = &h.Inspector
	} else {
		v := strings.TrimSpace(*req.RecheckPerson)
		h.RecheckPerson = &v
	}

	// 状态/等级校验与默认。
	status := model.StatusPending
	if req.Status != nil {
		status = model.HazardStatus(*req.Status)
	}
	if !status.Valid() {
		return nil, Unprocessable("非法的整改状态：" + string(status))
	}
	h.Status = status

	level := model.LevelGeneral
	if req.Level != nil {
		level = model.HazardLevel(*req.Level)
	}
	if !level.Valid() {
		return nil, Unprocessable("非法的隐患等级：" + string(level))
	}
	h.Level = level

	// 责任单位 -> 责任人联动（一一对应）。
	unit, err := s.units.GetByID(uint64(req.UnitId))
	if err == repo.ErrNotFound {
		return nil, Unprocessable("责任单位不存在")
	}
	if err != nil {
		return nil, Internal(err)
	}
	h.UnitID = unit.ID
	h.Person = strings.TrimSpace(unit.Person)
	if h.Person == "" {
		return nil, Unprocessable("责任单位「" + unit.Name + "」未配置责任人")
	}

	// 隐患类型：须为已存在的类型组合行。
	typeRow, typeErr := s.types.GetByID(uint64(req.TypeId))
	if typeErr == repo.ErrNotFound {
		return nil, Unprocessable("隐患类型不存在")
	}
	if typeErr != nil {
		return nil, Internal(typeErr)
	}
	h.TypeID = typeRow.ID

	// 图片 uuid 校验。
	before, after, imgErr := s.validateImages(req.BeforeImageIds, req.AfterImageIds)
	if imgErr != nil {
		return nil, imgErr
	}
	h.BeforeImages = before
	h.AfterImages = after

	if err := s.hazards.Create(h); err != nil {
		return nil, Internal(err)
	}
	// 重新查询以带出名称快照。
	created, getErr := s.hazards.GetByID(h.ID)
	if getErr != nil {
		return nil, Internal(getErr)
	}
	out := toGenHazard(created)
	return &out, nil
}

// Update 更新隐患：仅应用提供的字段，联动规则与创建一致。
func (s *HazardService) Update(id int64, req gen.HazardUpdateRequest) (*gen.Hazard, *Error) {
	h, err := s.hazards.GetByID(uint64(id))
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, Internal(err)
	}

	if req.InspectionArea != nil {
		h.InspectionArea = strings.TrimSpace(*req.InspectionArea)
		if h.InspectionArea == "" {
			h.InspectionArea = DefaultArea
		}
	}
	if req.InspectionDate != nil && !(*req.InspectionDate).Time.IsZero() {
		h.InspectionDate = model.Date{Time: (*req.InspectionDate).Time}
	}
	if req.Inspector != nil {
		h.Inspector = strings.TrimSpace(*req.Inspector)
		if h.Inspector == "" {
			h.Inspector = DefaultPerson
		}
	}
	if req.Description != nil {
		if strings.TrimSpace(*req.Description) == "" {
			return nil, Unprocessable("隐患描述不能为空")
		}
		h.Description = strings.TrimSpace(*req.Description)
	}
	if req.Suggestion != nil {
		h.Suggestion = req.Suggestion
	}
	if req.Remark != nil {
		h.Remark = req.Remark
	}
	if req.DueDate != nil && !(*req.DueDate).Time.IsZero() {
		h.DueDate = model.Date{Time: (*req.DueDate).Time}
	}
	// 复查人员：显式提供时更新（含清空），未提供保留原值。
	if req.RecheckPerson != nil {
		v := strings.TrimSpace(*req.RecheckPerson)
		h.RecheckPerson = &v
	}
	if req.Status != nil {
		status := model.HazardStatus(*req.Status)
		if !status.Valid() {
			return nil, Unprocessable("非法的整改状态：" + string(status))
		}
		h.Status = status
	}
	if req.Level != nil {
		level := model.HazardLevel(*req.Level)
		if !level.Valid() {
			return nil, Unprocessable("非法的隐患等级：" + string(level))
		}
		h.Level = level
	}

	// 单位变更 -> 重新联动责任人。
	if req.UnitId != nil && *req.UnitId > 0 && uint64(*req.UnitId) != h.UnitID {
		unit, unitErr := s.units.GetByID(uint64(*req.UnitId))
		if unitErr == repo.ErrNotFound {
			return nil, Unprocessable("责任单位不存在")
		}
		if unitErr != nil {
			return nil, Internal(unitErr)
		}
		if strings.TrimSpace(unit.Person) == "" {
			return nil, Unprocessable("责任单位「" + unit.Name + "」未配置责任人")
		}
		h.UnitID = unit.ID
		h.Person = strings.TrimSpace(unit.Person)
	}

	// 隐患类型变更 -> 校验存在性。
	if req.TypeId != nil {
		newTypeID := uint64(*req.TypeId)
		if newTypeID <= 0 {
			return nil, Unprocessable("请选择隐患类型")
		}
		if validateErr := s.validateType(newTypeID); validateErr != nil {
			return nil, validateErr
		}
		h.TypeID = newTypeID
	}

	// 图片变更。
	if req.BeforeImageIds != nil || req.AfterImageIds != nil {
		before := h.BeforeImages
		after := h.AfterImages
		var imgErr *Error
		if req.BeforeImageIds != nil {
			joined, e := model.JoinImages(*req.BeforeImageIds)
			if e != nil {
				return nil, Unprocessable(e.Error())
			}
			before = joined
		}
		if req.AfterImageIds != nil {
			joined, e := model.JoinImages(*req.AfterImageIds)
			if e != nil {
				return nil, Unprocessable(e.Error())
			}
			after = joined
		}
		before, after, imgErr = s.validateImageStrings(before, after)
		if imgErr != nil {
			return nil, imgErr
		}
		h.BeforeImages = before
		h.AfterImages = after
	}

	if err := s.hazards.Update(h); err != nil {
		return nil, Internal(err)
	}
	updated, getErr := s.hazards.GetByID(h.ID)
	if getErr != nil {
		return nil, Internal(getErr)
	}
	out := toGenHazard(updated)
	return &out, nil
}

// Delete 删除隐患（软删除）。
func (s *HazardService) Delete(id int64) *Error {
	if err := s.hazards.Delete(uint64(id)); err == repo.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return Internal(err)
	}
	return nil
}

// Stats 隐患概览统计。
func (s *HazardService) Stats() (*gen.HazardStats, *Error) {
	stats, err := s.hazards.Stats()
	if err != nil {
		return nil, Internal(err)
	}
	return &gen.HazardStats{
		Pending: int(stats.Pending),
		Blocked: int(stats.Blocked),
		Done:    int(stats.Done),
		Overdue: int(stats.Overdue),
	}, nil
}

// validateImages 校验并拼接创建时的图片 uuid。
func (s *HazardService) validateImages(before, after *[]string) (string, string, *Error) {
	beforeStr := ""
	afterStr := ""
	if before != nil {
		joined, err := model.JoinImages(*before)
		if err != nil {
			return "", "", Unprocessable(err.Error())
		}
		beforeStr = joined
	}
	if after != nil {
		joined, err := model.JoinImages(*after)
		if err != nil {
			return "", "", Unprocessable(err.Error())
		}
		afterStr = joined
	}
	return s.validateImageStrings(beforeStr, afterStr)
}

// validateImageStrings 校验图片 uuid 均存在于摘要表。
func (s *HazardService) validateImageStrings(before, after string) (string, string, *Error) {
	ids := append(model.SplitImages(before), model.SplitImages(after)...)
	if len(ids) == 0 {
		return before, after, nil
	}
	exists, err := s.images.Exists(ids)
	if err != nil {
		return "", "", Internal(err)
	}
	for _, id := range ids {
		if !exists[id] {
			return "", "", Unprocessable("图片不存在：" + id)
		}
	}
	return before, after, nil
}

// validateType 校验隐患类型组合行存在。
func (s *HazardService) validateType(typeID uint64) *Error {
	if _, err := s.types.GetByID(typeID); err == repo.ErrNotFound {
		return Unprocessable("隐患类型不存在")
	} else if err != nil {
		return Internal(err)
	}
	return nil
}

// toGenHazard 模型 -> 响应 DTO（图片逗号串拆为 uuid 数组）。
func toGenHazard(h *model.Hazard) gen.Hazard {
	out := gen.Hazard{
		Id:             int64(h.ID),
		InspectionArea: h.InspectionArea,
		InspectionDate: types.Date{Time: h.InspectionDate.Time},
		Inspector:      h.Inspector,
		Description:    h.Description,
		Suggestion:     h.Suggestion,
		UnitId:         int64(h.UnitID),
		UnitName:       h.UnitName,
		Person:         h.Person,
		DueDate:        types.Date{Time: h.DueDate.Time},
		RecheckPerson:  h.RecheckPerson,
		Status:         gen.HazardStatus(h.Status),
		TypeId:         int64(h.TypeID),
		TypeMajor:      h.TypeMajor,
		TypeMinor:      h.TypeMinor,
		Level:          gen.HazardLevel(h.Level),
		Remark:         h.Remark,
		CreatedAt:      h.CreatedAt,
		UpdatedAt:      h.UpdatedAt,
		BeforeImageIds: toIDList(model.SplitImages(h.BeforeImages)),
		AfterImageIds:  toIDList(model.SplitImages(h.AfterImages)),
	}
	return out
}

func toIDList(ids []string) *[]string {
	if len(ids) == 0 {
		return nil
	}
	return &ids
}

func optionalString[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func optionalUint64(v *int64) *uint64 {
	if v == nil {
		return nil
	}
	u := uint64(*v)
	return &u
}

func derefInt(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}
