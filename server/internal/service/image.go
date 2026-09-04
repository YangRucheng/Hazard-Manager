package service

import (
	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/repo"
	"hazard-system/server/internal/upload"
)

// ImageService 附件（图片）管理业务逻辑。
type ImageService struct {
	images *repo.ImageRepo
	store  *upload.Store
}

// NewImageService 构造 ImageService。
func NewImageService(images *repo.ImageRepo, store *upload.Store) *ImageService {
	return &ImageService{images: images, store: store}
}

// List 附件分页列表（含引用计数，可只看未引用）。
func (s *ImageService) List(params gen.ListImagesParams) (*gen.ImageListResponse, *Error) {
	page := derefInt(params.Page, 1)
	pageSize := derefInt(params.PageSize, 20)
	page, pageSize = repo.DefaultPagination(page, pageSize)

	items, total, err := s.images.List(repo.ImageFilter{
		Page:       page,
		PageSize:   pageSize,
		OnlyUnused: params.OnlyUnused != nil && *params.OnlyUnused,
		Keyword:    params.Keyword,
	})
	if err != nil {
		return nil, Internal(err)
	}

	resp := &gen.ImageListResponse{
		Items: make([]gen.ImageSummary, 0, len(items)),
		Pagination: gen.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}
	for i := range items {
		img := &items[i].Image
		resp.Items = append(resp.Items, gen.ImageSummary{
			Id:           img.ID,
			MimeType:     img.MimeType,
			SizeBytes:    img.SizeBytes,
			Width:        img.Width,
			Height:       img.Height,
			RefCount:     items[i].RefCount,
			CreatedAt:    img.CreatedAt,
			Url:          "/api/v1/images/" + img.ID,
			ThumbnailUrl: "/api/v1/images/" + img.ID + "/thumbnail",
		})
	}
	return resp, nil
}

// Delete 删除未被引用的附件（先删文件再删记录；被引用返回 409）。
func (s *ImageService) Delete(id string) *Error {
	if _, err := s.images.GetByID(id); err != nil {
		if err == repo.ErrNotFound {
			return ErrNotFound
		}
		return Internal(err)
	}
	usage, err := s.images.UsageMap()
	if err != nil {
		return Internal(err)
	}
	if usage[id] > 0 {
		return Conflict("该附件正被隐患记录引用，无法删除")
	}
	if err := s.store.Remove(id); err != nil {
		return Internal(err)
	}
	if err := s.images.Delete(id); err != nil {
		if err == repo.ErrNotFound {
			return ErrNotFound
		}
		return Internal(err)
	}
	return nil
}
