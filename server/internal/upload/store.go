// Package upload 负责图片落盘与摘要去重。
// 规则：数据库只记录 SHA-256 摘要与 uuid，不存文件名；其他表仅引用 uuid；
// 相同摘要的图片不重复保存，直接返回既有 uuid。
package upload

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"hazard-system/server/internal/model"
)

// AllowedMimeTypes 允许的图片类型（白名单）。
var AllowedMimeTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/gif":  "gif",
}

// MaxThumbSide 缩略图最长边像素。
const MaxThumbSide = 320

// Result 上传结果。
type Result struct {
	ID           string
	Digest       string
	MimeType     string
	SizeBytes    int64
	Width        *int
	Height       *int
	URL          string
	ThumbnailURL string
	Duplicate    bool
}

// Store 图片存储：目录 + 数据库摘要表。
type Store struct {
	dir string
	db  *gorm.DB
}

// NewStore 创建存储目录并返回 Store。
func NewStore(dir string, db *gorm.DB) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("创建图片目录失败: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "thumbs"), 0o755); err != nil {
		return nil, fmt.Errorf("创建缩略图目录失败: %w", err)
	}
	return &Store{dir: dir, db: db}, nil
}

// Save 处理上传文件：摘要去重 -> 无则落盘 + 生成缩略图 + 入库。
// maxBytes 为大小上限。
func (s *Store) Save(fileHeader *multipart.FileHeader, maxBytes int64) (*Result, error) {
	if maxBytes > 0 && fileHeader.Size > maxBytes {
		return nil, fmt.Errorf("图片大小超过上限（%d 字节）", maxBytes)
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("读取上传内容失败: %w", err)
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("图片大小超过上限（%d 字节）", maxBytes)
	}
	if len(data) == 0 {
		return nil, errors.New("上传文件为空")
	}

	digest := sha256.Sum256(data)
	digestHex := hex.EncodeToString(digest[:])

	// 摘要去重：命中既有记录直接返回，不重复保存。
	var existing model.Image
	err = s.db.Where("digest = ?", digestHex).First(&existing).Error
	if err == nil {
		return s.toResult(existing, true), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询图片摘要失败: %w", err)
	}

	mimeType := http.DetectContentType(data)
	ext, ok := AllowedMimeTypes[mimeType]
	if !ok {
		return nil, fmt.Errorf("不支持的图片格式 %q，仅支持 jpeg/png/webp/gif", mimeType)
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %w", err)
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	id := newUUID()
	originalPath := filepath.Join(s.dir, id+"."+ext)
	if err := os.WriteFile(originalPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("保存原图失败: %w", err)
	}

	// 生成缩略图（最长边 320px 的 JPEG）。
	thumbPath := filepath.Join(s.dir, "thumbs", id+".jpg")
	if err := s.writeThumbnail(img, thumbPath); err != nil {
		return nil, fmt.Errorf("生成缩略图失败: %w", err)
	}

	imageModel := model.Image{
		ID:        id,
		Digest:    digestHex,
		MimeType:  mimeType,
		SizeBytes: int64(len(data)),
		Width:     intPtr(width),
		Height:    intPtr(height),
	}
	if err := s.db.Create(&imageModel).Error; err != nil {
		// digest 唯一索引兜底并发：撞键则重查返回既有记录。
		if isDuplicateKey(err) {
			var dup model.Image
			if err2 := s.db.Where("digest = ?", digestHex).First(&dup).Error; err2 == nil {
				return s.toResult(dup, true), nil
			}
		}
		return nil, fmt.Errorf("保存图片记录失败: %w", err)
	}

	return s.toResult(imageModel, false), nil
}

// Resolve 根据 uuid 返回原图路径与 MIME。
func (s *Store) Resolve(id string) (path string, mimeType string, ok bool) {
	var img model.Image
	if err := s.db.Where("id = ?", id).First(&img).Error; err != nil {
		return "", "", false
	}
	ext, ok := AllowedMimeTypes[img.MimeType]
	if !ok {
		return "", "", false
	}
	return filepath.Join(s.dir, id+"."+ext), img.MimeType, true
}

// ResolveThumbnail 根据 uuid 返回缩略图路径；缩略图缺失时回退原图。
func (s *Store) ResolveThumbnail(id string) (path string, mimeType string, ok bool) {
	thumb := filepath.Join(s.dir, "thumbs", id+".jpg")
	if _, err := os.Stat(thumb); err == nil {
		return thumb, "image/jpeg", true
	}
	return s.Resolve(id)
}

// Remove 删除 uuid 对应的落盘原图与缩略图（文件不存在视为成功）。
func (s *Store) Remove(id string) error {
	var img model.Image
	err := s.db.Where("id = ?", id).First(&img).Error
	ext := ""
	switch {
	case err == nil:
		if e, ok := AllowedMimeTypes[img.MimeType]; ok {
			ext = e
		}
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("查询图片记录失败: %w", err)
	}
	targets := []string{filepath.Join(s.dir, "thumbs", id+".jpg")}
	if ext != "" {
		targets = append(targets, filepath.Join(s.dir, id+"."+ext))
	}
	for _, p := range targets {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除图片文件失败: %w", err)
		}
	}
	return nil
}

// toResult 将图片模型转为上传结果（URL 为相对路径）。
func (s *Store) toResult(img model.Image, duplicate bool) *Result {
	return &Result{
		ID:           img.ID,
		Digest:       img.Digest,
		MimeType:     img.MimeType,
		SizeBytes:    img.SizeBytes,
		Width:        img.Width,
		Height:       img.Height,
		URL:          "/api/v1/images/" + img.ID,
		ThumbnailURL: "/api/v1/images/" + img.ID + "/thumbnail",
		Duplicate:    duplicate,
	}
}

func (s *Store) writeThumbnail(img image.Image, path string) error {
	thumb := imaging.Fit(img, MaxThumbSide, MaxThumbSide, imaging.Lanczos)
	return imaging.Save(thumb, path, imaging.JPEGQuality(80))
}

func newUUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func intPtr(n int) *int { return &n }

func isDuplicateKey(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "Error 1062"))
}
