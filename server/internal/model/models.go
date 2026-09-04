// Package model 定义 GORM 数据模型与强类型枚举/日期类型。
// 所有列均为类型化字段，禁止动态取值。
package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Date 表示 "2006-01-02" 的日期，映射 MySQL DATE 列，JSON 序列化为 YYYY-MM-DD。
type Date struct {
	time.Time
}

// NewDate 由 time.Time 构造 Date。
func NewDate(t time.Time) Date {
	y, m, d := t.Date()
	return Date{Time: time.Date(y, m, d, 0, 0, 0, 0, t.Location())}
}

// Today 返回当前日期。
func Today() Date { return NewDate(time.Now()) }

// AddDays 返回加上 n 天后的日期。
func (d Date) AddDays(n int) Date { return NewDate(d.Time.AddDate(0, 0, n)) }

// String 返回 YYYY-MM-DD。
func (d Date) String() string { return d.Time.Format("2006-01-02") }

// IsZero 判断是否为零值。
func (d Date) IsZero() bool { return d.Time.IsZero() }

// GormDataType 使 GORM 将其映射为 DATE 列。
func (Date) GormDataType() string { return "date" }

// Scan 实现 sql.Scanner。
func (d *Date) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		d.Time = time.Time{}
	case time.Time:
		d.Time = v
	case string:
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err != nil {
			return fmt.Errorf("无法解析日期 %q: %w", v, err)
		}
		d.Time = t
	case []byte:
		t, err := time.ParseInLocation("2006-01-02", string(v), time.Local)
		if err != nil {
			return fmt.Errorf("无法解析日期 %q: %w", string(v), err)
		}
		d.Time = t
	default:
		return fmt.Errorf("不支持的日期类型 %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer。
func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.String(), nil
}

// MarshalJSON 输出 "2006-01-02"。
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON 解析 "2006-01-02"。
func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		d.Time = time.Time{}
		return nil
	}
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("非法日期格式 %s", s)
	}
	t, err := time.ParseInLocation("2006-01-02", s[1:len(s)-1], time.Local)
	if err != nil {
		return fmt.Errorf("非法日期格式 %s", s)
	}
	d.Time = t
	return nil
}

// HazardStatus 整改状态，映射 MySQL ENUM。
type HazardStatus string

const (
	StatusPending HazardStatus = "待整改"
	StatusBlocked HazardStatus = "整改受阻"
	StatusDone    HazardStatus = "已整改"
)

// Valid 校验状态是否在枚举内。
func (s HazardStatus) Valid() bool {
	switch s {
	case StatusPending, StatusBlocked, StatusDone:
		return true
	}
	return false
}

// GormDataType 返回 MySQL ENUM 定义。
func (HazardStatus) GormDataType() string { return "enum('待整改','整改受阻','已整改')" }

// Scan 实现 sql.Scanner。
func (s *HazardStatus) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = ""
	case string:
		*s = HazardStatus(v)
	case []byte:
		*s = HazardStatus(string(v))
	default:
		return fmt.Errorf("不支持的整改状态类型 %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer。
func (s HazardStatus) Value() (driver.Value, error) { return string(s), nil }

// HazardLevel 隐患等级，映射 MySQL ENUM。
type HazardLevel string

const (
	LevelGeneral HazardLevel = "一般隐患"
	LevelMajor   HazardLevel = "重大隐患"
)

// Valid 校验等级是否在枚举内。
func (l HazardLevel) Valid() bool {
	switch l {
	case LevelGeneral, LevelMajor:
		return true
	}
	return false
}

// GormDataType 返回 MySQL ENUM 定义。
func (HazardLevel) GormDataType() string { return "enum('一般隐患','重大隐患')" }

// Scan 实现 sql.Scanner。
func (l *HazardLevel) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*l = ""
	case string:
		*l = HazardLevel(v)
	case []byte:
		*l = HazardLevel(string(v))
	default:
		return fmt.Errorf("不支持的隐患等级类型 %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer。
func (l HazardLevel) Value() (driver.Value, error) { return string(l), nil }

// Hazard 隐患记录主表。
type Hazard struct {
	ID             uint64         `gorm:"primaryKey" json:"id"`
	InspectionArea string         `gorm:"type:varchar(128);not null;default:''" json:"inspectionArea"`
	InspectionDate Date           `gorm:"type:date;not null" json:"inspectionDate"`
	Inspector      string         `gorm:"type:varchar(64);not null;default:''" json:"inspector"`
	Description    string         `gorm:"type:text;not null" json:"description"`
	Suggestion     *string        `gorm:"type:text" json:"suggestion"`
	UnitID         uint64         `gorm:"not null;index" json:"unitId"`
	Person         string         `gorm:"type:varchar(64);not null;default:''" json:"person"`
	DueDate        Date           `gorm:"type:date;not null" json:"dueDate"`
	RecheckPerson  *string        `gorm:"type:varchar(64)" json:"recheckPerson"`
	RectifyPerson  *string        `gorm:"type:varchar(64)" json:"rectifyPerson"`
	BeforeImages   string         `gorm:"type:varchar(2048);not null;default:''" json:"-"`
	Status         HazardStatus   `gorm:"type:enum('待整改','整改受阻','已整改');not null;default:'待整改'" json:"status"`
	AfterImages    string         `gorm:"type:varchar(2048);not null;default:''" json:"-"`
	TypeID         uint64         `gorm:"not null;index" json:"typeId"`
	Level          HazardLevel    `gorm:"type:enum('一般隐患','重大隐患');not null;default:'一般隐患'" json:"level"`
	Remark         *string        `gorm:"type:text" json:"remark"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	UnitName  string `gorm:"-" json:"unitName"`
	TypeMajor string `gorm:"-" json:"typeMajor"`
	TypeMinor string `gorm:"-" json:"typeMinor"`
}

// SplitImages 将逗号分隔的 uuid 串拆为数组；空串返回空数组。
func SplitImages(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if part := s[start:i]; part != "" {
				out = append(out, part)
			}
			start = i + 1
		}
	}
	return out
}

// JoinImages 将 uuid 数组拼接为逗号分隔串。
func JoinImages(ids []string) (string, error) {
	if len(ids) == 0 {
		return "", nil
	}
	if len(ids) > 20 {
		return "", errors.New("图片数量不能超过 20 张")
	}
	var b []byte
	for i, id := range ids {
		if len(id) != 32 {
			return "", fmt.Errorf("非法图片 uuid: %q", id)
		}
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, id...)
	}
	return string(b), nil
}

// ResponsibleUnit 责任单位枚举表。无排序字段，列表按创建顺序（id）返回。
type ResponsibleUnit struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(128);not null;uniqueIndex" json:"name"`
	Person    string         `gorm:"type:varchar(64);not null" json:"person"`
	Remark    *string        `gorm:"type:varchar(255)" json:"remark"`
	Status    int            `gorm:"not null;default:1" json:"status"` // 0=停用 1=启用
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// HazardType 隐患类型枚举表：每行一个「大类(major)+小类(minor)」组合，
// 无父级引用、无排序、无启停；同一组合唯一。隐患记录仅引用该行 id。
// 删除为物理删除（服务层保证未被引用），便于同名组合复用。
type HazardType struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Major     string    `gorm:"type:varchar(128);not null;uniqueIndex:uk_hazard_type_major_minor" json:"major"`
	Minor     string    `gorm:"type:varchar(128);not null;uniqueIndex:uk_hazard_type_major_minor" json:"minor"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Image 图片摘要表：只存 uuid 与摘要（SHA-256），不存文件名。
type Image struct {
	ID        string    `gorm:"type:char(32);primaryKey" json:"id"`
	Digest    string    `gorm:"type:char(64);not null;uniqueIndex" json:"digest"`
	MimeType  string    `gorm:"type:varchar(64);not null" json:"mimeType"`
	SizeBytes int64     `gorm:"not null" json:"sizeBytes"`
	Width     *int      `gorm:"" json:"width"`
	Height    *int      `gorm:"" json:"height"`
	CreatedAt time.Time `json:"createdAt"`
}

// HazardStats 隐患概览统计。
type HazardStats struct {
	Pending int64 `json:"pending"`
	Blocked int64 `json:"blocked"`
	Done    int64 `json:"done"`
	Overdue int64 `json:"overdue"`
}
