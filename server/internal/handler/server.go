// Package handler 实现 oapi-codegen 生成的 ServerInterface（HTTP 编解码层）。
package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hazard-system/server/internal/auth"
	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/service"
	"hazard-system/server/internal/upload"
	"hazard-system/server/internal/version"
)

// Server 实现全部 API 端点。
type Server struct {
	hazards  *service.HazardService
	units    *service.UnitService
	types    *service.TypeService
	images   *service.ImageService
	auth     *auth.Manager
	store    *upload.Store
	maxBytes int64
}

// NewServer 构造 Server。
func NewServer(
	hazards *service.HazardService,
	units *service.UnitService,
	types *service.TypeService,
	images *service.ImageService,
	am *auth.Manager,
	store *upload.Store,
	maxBytes int64,
) *Server {
	return &Server{hazards: hazards, units: units, types: types, images: images, auth: am, store: store, maxBytes: maxBytes}
}

// ---- 鉴权 ----

// Login 管理员登录。
func (s *Server) Login(c *gin.Context) {
	var req gen.LoginJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "请求体格式错误")
		return
	}
	if !s.auth.VerifyAdmin(req.Username, req.Password) {
		c.JSON(http.StatusUnauthorized, gen.Error{Code: "INVALID_CREDENTIALS", Message: "用户名或密码错误"})
		return
	}
	token, err := s.auth.Sign(auth.AdminUsername, auth.UserTypeAdmin)
	if err != nil {
		writeServiceError(c, service.Internal(err))
		return
	}
	c.JSON(http.StatusOK, gen.LoginResponse{
		Token: token,
		User: gen.UserInfo{
			Username: auth.AdminUsername,
			UserType: gen.Admin,
		},
	})
}

// GetCurrentUser 当前用户信息。
func (s *Server) GetCurrentUser(c *gin.Context) {
	claims := auth.ClaimsFrom(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gen.Error{Code: "UNAUTHORIZED", Message: "缺少认证"})
		return
	}
	c.JSON(http.StatusOK, gen.UserInfo{
		Username: claims.Subject,
		UserType: gen.UserInfoUserType(claims.UserType),
	})
}

// ---- 系统信息 ----

// GetSystemInfo 系统信息：返回后端编译时间与启动时间。
func (s *Server) GetSystemInfo(c *gin.Context) {
	buildTime := version.BuildTime
	if buildTime == "" {
		buildTime = "unknown"
	}
	c.JSON(http.StatusOK, gen.SystemInfo{
		BuildTime: buildTime,
		StartTime: version.StartTime,
	})
}

// ---- 隐患记录 ----

// ListHazards 隐患分页列表。
func (s *Server) ListHazards(c *gin.Context, params gen.ListHazardsParams) {
	resp, appErr := s.hazards.List(params)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateHazard 新增隐患。
func (s *Server) CreateHazard(c *gin.Context) {
	var req gen.CreateHazardJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("绑定新增隐患请求失败: %v", err)
		writeBadRequest(c, "请求体格式错误")
		return
	}
	h, appErr := s.hazards.Create(req)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, h)
}

// GetHazard 隐患详情。
func (s *Server) GetHazard(c *gin.Context, id int64) {
	h, appErr := s.hazards.Get(id)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, h)
}

// UpdateHazard 更新隐患。
func (s *Server) UpdateHazard(c *gin.Context, id int64) {
	var req gen.UpdateHazardJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "请求体格式错误")
		return
	}
	h, appErr := s.hazards.Update(id, req)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, h)
}

// DeleteHazard 删除隐患（软删除）。
func (s *Server) DeleteHazard(c *gin.Context, id int64) {
	if appErr := s.hazards.Delete(id); appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetHazardStats 隐患概览统计。
func (s *Server) GetHazardStats(c *gin.Context) {
	stats, appErr := s.hazards.Stats()
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ---- 责任单位 ----

// ListUnits 责任单位列表。
func (s *Server) ListUnits(c *gin.Context, params gen.ListUnitsParams) {
	items, appErr := s.units.List(params.Keyword)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateUnit 新增责任单位。
func (s *Server) CreateUnit(c *gin.Context) {
	var req gen.CreateUnitJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "请求体格式错误")
		return
	}
	u, appErr := s.units.Create(req)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, u)
}

// UpdateUnit 更新责任单位。
func (s *Server) UpdateUnit(c *gin.Context, id int64) {
	var req gen.UpdateUnitJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "请求体格式错误")
		return
	}
	u, appErr := s.units.Update(id, req)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, u)
}

// DeleteUnit 删除责任单位。
func (s *Server) DeleteUnit(c *gin.Context, id int64) {
	if appErr := s.units.Delete(id); appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- 隐患类型/分类 ----

// ListHazardTypes 类型/分类全量。
func (s *Server) ListHazardTypes(c *gin.Context) {
	items, appErr := s.types.List()
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateHazardType 新增类型/分类。
func (s *Server) CreateHazardType(c *gin.Context) {
	var req gen.CreateHazardTypeJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "请求体格式错误")
		return
	}
	t, appErr := s.types.Create(req)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, t)
}

// UpdateHazardType 更新类型/分类。
func (s *Server) UpdateHazardType(c *gin.Context, id int64) {
	var req gen.UpdateHazardTypeJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "请求体格式错误")
		return
	}
	t, appErr := s.types.Update(id, req)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, t)
}

// DeleteHazardType 删除类型/分类。
func (s *Server) DeleteHazardType(c *gin.Context, id int64) {
	if appErr := s.types.Delete(id); appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- 图片 ----

// ListImages 附件分页列表（含引用计数）。
func (s *Server) ListImages(c *gin.Context, params gen.ListImagesParams) {
	resp, appErr := s.images.List(params)
	if appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteImage 删除未被引用的附件。
func (s *Server) DeleteImage(c *gin.Context, id string) {
	if appErr := s.images.Delete(id); appErr != nil {
		writeServiceError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadImage 上传图片（摘要去重）。
func (s *Server) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		writeBadRequest(c, "缺少文件字段 file")
		return
	}
	res, err := s.store.Save(fileHeader, s.maxBytes)
	if err != nil {
		log.Printf("图片上传失败: %v", err)
		writeBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gen.ImageInfo{
		Id:           res.ID,
		Digest:       res.Digest,
		MimeType:     res.MimeType,
		SizeBytes:    res.SizeBytes,
		Width:        res.Width,
		Height:       res.Height,
		Url:          res.URL,
		ThumbnailUrl: res.ThumbnailURL,
		Duplicate:    &res.Duplicate,
	})
}

// GetImage 获取原图。
func (s *Server) GetImage(c *gin.Context, id string) {
	path, mime, ok := s.store.Resolve(id)
	if !ok {
		writeNotFound(c, "图片不存在")
		return
	}
	c.Header("Content-Type", mime)
	c.File(path)
}

// GetImageThumbnail 获取缩略图（缺失回退原图）。
func (s *Server) GetImageThumbnail(c *gin.Context, id string) {
	path, mime, ok := s.store.ResolveThumbnail(id)
	if !ok {
		writeNotFound(c, "图片不存在")
		return
	}
	c.Header("Content-Type", mime)
	c.File(path)
}

// ---- 错误输出 ----

func writeServiceError(c *gin.Context, appErr *service.Error) {
	c.JSON(appErr.Status, gen.Error{Code: appErr.Code, Message: appErr.Message})
}

func writeBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gen.Error{Code: "BAD_REQUEST", Message: msg})
}

func writeNotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gen.Error{Code: "NOT_FOUND", Message: msg})
}
