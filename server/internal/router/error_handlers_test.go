// 框架级异常（404 / 405 / 500 panic / oapi 参数解析错误）统一 JSON 输出测试。
package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hazard-system/server/internal/auth"
	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/handler"
	"hazard-system/server/internal/upload"
)

// newTestRouter 用真实路由组装构建引擎（服务层依赖传 nil：
// 本测试只覆盖路由/框架层行为，不触及业务 handler 对服务层的调用）。
func newTestRouter(t *testing.T) (*gin.Engine, *auth.Manager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	am := auth.NewManager("test-secret", time.Hour, "pw")
	store, err := upload.NewStore(t.TempDir(), nil)
	require.NoError(t, err)
	srv := handler.NewServer(nil, nil, nil, nil, am, store, 0)
	return NewRouter(srv, am), am
}

// decodeErrorJSON 断言响应为 JSON 并解析为统一错误结构。
func decodeErrorJSON(t *testing.T, w *httptest.ResponseRecorder) gen.Error {
	t.Helper()
	ct := w.Header().Get("Content-Type")
	assert.True(t, strings.HasPrefix(ct, "application/json"), "期望 JSON 响应，实际 Content-Type=%q", ct)
	var e gen.Error
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e
}

// TestFrameworkErrorsUniformJSON 覆盖原本非 JSON 的兜底路径：
// 404（未知路由）、405（方法不允许）、500（handler panic）。
func TestFrameworkErrorsUniformJSON(t *testing.T) {
	engine, _ := newTestRouter(t)

	// 404：未知路由不再返回纯文本 404 page not found。
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/no-such-endpoint", nil))
	require.Equal(t, http.StatusNotFound, w.Code)
	e := decodeErrorJSON(t, w)
	assert.Equal(t, "NOT_FOUND", e.Code)
	assert.NotEmpty(t, e.Message)

	// 405：路径存在但方法不允许（PATCH 未注册于 /api/v1/hazards/1）。
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/api/v1/hazards/1", nil))
	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Header().Get("Allow"), "GET") // gin 按 RFC 7231 生成 Allow 头
	e = decodeErrorJSON(t, w)
	assert.Equal(t, "METHOD_NOT_ALLOWED", e.Code)
	assert.NotEmpty(t, e.Message)

	// 500：handler panic 由 recoveryMiddleware 兜底为统一 JSON（而非空 500）。
	engine.GET("/boom", func(c *gin.Context) { panic("boom") })
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
	require.Equal(t, http.StatusInternalServerError, w.Code)
	e = decodeErrorJSON(t, w)
	assert.Equal(t, "INTERNAL_ERROR", e.Code)
	assert.NotEmpty(t, e.Message)
}

// TestOapiBindErrorUniformJSON 路径参数格式错误走 oapi-codegen 生成的
// 参数解析路径，ErrorHandler 应输出统一 JSON 400（而非默认 {"msg": ...}）。
func TestOapiBindErrorUniformJSON(t *testing.T) {
	engine, am := newTestRouter(t)
	token, err := am.Sign(auth.AdminUsername, auth.UserTypeAdmin)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/hazards/not-an-int", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	e := decodeErrorJSON(t, w)
	assert.Equal(t, "BAD_REQUEST", e.Code)
	assert.NotEmpty(t, e.Message)
}

// TestAuthErrorsUniformJSON 鉴权失败与登录体错误本就统一，此处作回归守卫。
func TestAuthErrorsUniformJSON(t *testing.T) {
	engine, _ := newTestRouter(t)

	// 401：未带令牌访问受保护端点。
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/hazards/1", nil))
	require.Equal(t, http.StatusUnauthorized, w.Code)
	e := decodeErrorJSON(t, w)
	assert.Equal(t, "UNAUTHORIZED", e.Code)

	// 400：登录请求体非法（handler 层 writeBadRequest）。
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	e = decodeErrorJSON(t, w)
	assert.Equal(t, "BAD_REQUEST", e.Code)
}
