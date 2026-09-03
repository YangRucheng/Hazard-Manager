package service

import (
	"os"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"hazard-system/server/internal/database"
	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/model"
	"hazard-system/server/internal/repo"
)

// setupIntegration 初始化集成测试环境。
// 用法：export TEST_DB_DSN='hazard:hazard_dev_password@tcp(127.0.0.1:3306)/hazard_system_test?charset=utf8mb4&parseTime=True&loc=Local'
// 需先创建测试库：CREATE DATABASE hazard_system_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
func setupIntegration(t *testing.T) *HazardService {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("未设置 TEST_DB_DSN，跳过集成测试")
	}
	db, err := database.Connect(dsn)
	require.NoError(t, err)
	resetTables(t, db)
	require.NoError(t, database.SeedIfEmpty(db))

	return NewHazardService(
		repo.NewHazardRepo(db),
		repo.NewUnitRepo(db),
		repo.NewHazardTypeRepo(db),
		repo.NewImageRepo(db),
	)
}

func resetTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	// 用 TRUNCATE 以重置自增计数器，保证 seed 数据 id 从 1 开始。
	require.NoError(t, db.Exec("TRUNCATE TABLE hazards").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE responsible_units").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE hazard_types").Error)
}

func TestCreate_DefaultsAndLinkage(t *testing.T) {
	svc := setupIntegration(t)

	h, appErr := svc.Create(gen.HazardCreateRequest{
		Description: "测试隐患：线路裸露",
		UnitId:      1, // seed 中 id=1 为电气车间/张三
		TypeId:      1,
		CategoryId:  3,
	})
	require.Nil(t, appErr)
	require.NotNil(t, h)

	assert.Equal(t, "华星现场", h.InspectionArea)
	assert.Equal(t, "电气自查", h.Inspector)
	require.NotNil(t, h.RecheckPerson)
	assert.Equal(t, "电气自查", *h.RecheckPerson)
	assert.Equal(t, gen.HazardStatus(model.StatusPending), h.Status)
	assert.Equal(t, gen.HazardLevel(model.LevelGeneral), h.Level)
	assert.Equal(t, "张三", h.Person) // 单位联动
	assert.Equal(t, "电气设备", h.TypeName)

	// 检查日期=今天，要求完成时间=检查日期+7 天。
	inspection := h.InspectionDate.Time
	due := h.DueDate.Time
	assert.Equal(t, inspection.AddDate(0, 0, 7), due)
}

func TestCreate_RespectsProvidedValues(t *testing.T) {
	svc := setupIntegration(t)
	date := types.Date{Time: parseTime("2026-08-20")}
	due := types.Date{Time: parseTime("2026-08-30")}
	recheck := "远程复核"

	h, appErr := svc.Create(gen.HazardCreateRequest{
		InspectionArea: strPtr("二号车间"),
		InspectionDate: &date,
		Inspector:      strPtr("巡检员乙"),
		DueDate:        &due,
		RecheckPerson:  &recheck,
		Status:         statusPtr("整改受阻"),
		Level:          levelPtr("重大隐患"),
		Description:    "测试隐患：给定值",
		UnitId:         2,
		TypeId:         1,
		CategoryId:     3,
	})
	require.Nil(t, appErr)
	assert.Equal(t, "二号车间", h.InspectionArea)
	assert.Equal(t, "2026-08-20", h.InspectionDate.Time.Format("2006-01-02"))
	assert.Equal(t, "巡检员乙", h.Inspector)
	assert.Equal(t, "2026-08-30", h.DueDate.Time.Format("2006-01-02"))
	require.NotNil(t, h.RecheckPerson)
	assert.Equal(t, "远程复核", *h.RecheckPerson)
	assert.Equal(t, gen.HazardStatus(model.StatusBlocked), h.Status)
	assert.Equal(t, gen.HazardLevel(model.LevelMajor), h.Level)
	assert.Equal(t, "李四", h.Person)
}

func TestCreate_CategoryMustBelongToType(t *testing.T) {
	svc := setupIntegration(t)
	_, appErr := svc.Create(gen.HazardCreateRequest{
		Description: "分类不匹配",
		UnitId:      1,
		TypeId:      1, // 电气设备
		CategoryId:  6, // 警示标识缺失 属于 安全防护(2)
	})
	require.NotNil(t, appErr)
	assert.Equal(t, 422, appErr.Status)
}

func TestCreate_UnitMustExist(t *testing.T) {
	svc := setupIntegration(t)
	_, appErr := svc.Create(gen.HazardCreateRequest{
		Description: "单位不存在",
		UnitId:      99999,
		TypeId:      1,
		CategoryId:  3,
	})
	require.NotNil(t, appErr)
	assert.Equal(t, 422, appErr.Status)
}

func TestCreate_InvalidStatusRejected(t *testing.T) {
	svc := setupIntegration(t)
	_, appErr := svc.Create(gen.HazardCreateRequest{
		Description: "非法状态",
		UnitId:      1,
		TypeId:      1,
		CategoryId:  3,
		Status:      statusPtr("已废弃"),
	})
	require.NotNil(t, appErr)
	assert.Equal(t, 422, appErr.Status)
}

func TestUpdate_UnitChangeRelinksPerson(t *testing.T) {
	svc := setupIntegration(t)
	created, appErr := svc.Create(gen.HazardCreateRequest{
		Description: "联动测试",
		UnitId:      1,
		TypeId:      1,
		CategoryId:  3,
	})
	require.Nil(t, appErr)

	newUnit := int64(2)
	updated, appErr := svc.Update(created.Id, gen.HazardUpdateRequest{UnitId: &newUnit})
	require.Nil(t, appErr)
	assert.Equal(t, "动力车间", updated.UnitName)
	assert.Equal(t, "李四", updated.Person)
}

func TestUpdate_NotFound(t *testing.T) {
	svc := setupIntegration(t)
	_, appErr := svc.Update(999999, gen.HazardUpdateRequest{})
	require.NotNil(t, appErr)
	assert.Equal(t, 404, appErr.Status)
}

func TestStats_AfterCreate(t *testing.T) {
	svc := setupIntegration(t)
	_, appErr := svc.Create(gen.HazardCreateRequest{
		Description: "统计测试", UnitId: 1, TypeId: 1, CategoryId: 3,
	})
	require.Nil(t, appErr)

	stats, appErr := svc.Stats()
	require.Nil(t, appErr)
	assert.Equal(t, 1, stats.Pending)
	assert.Equal(t, 0, stats.Blocked)
	assert.Equal(t, 0, stats.Done)
}

func parseTime(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func strPtr(s string) *string { return &s }

func statusPtr(s string) *gen.HazardStatus {
	v := gen.HazardStatus(s)
	return &v
}

func levelPtr(s string) *gen.HazardLevel {
	v := gen.HazardLevel(s)
	return &v
}
