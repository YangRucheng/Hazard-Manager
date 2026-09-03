package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestResolveAllowOrigin(t *testing.T) {
	assert.Equal(t, "*", resolveAllowOrigin("", ""))
	assert.Equal(t, "https://app.example.com", resolveAllowOrigin("https://app.example.com", ""))
	// Origin 优先于 Referer。
	assert.Equal(t, "https://a.com", resolveAllowOrigin("https://a.com", "https://b.com/path"))
	// 无 Origin 时从 Referer 提取协议+主机。
	assert.Equal(t, "https://app.example.com", resolveAllowOrigin("", "https://app.example.com/index.html"))
	assert.Equal(t, "http://localhost:5173", resolveAllowOrigin("", "http://localhost:5173/#/hazards"))
	// 非法协议忽略。
	assert.Equal(t, "", originFromHeader("javascript:alert(1)"))
	assert.Equal(t, "*", resolveAllowOrigin("", "not a valid referer"))
}

func TestCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(corsMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	// 1) 带 Origin 回显。
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://app.example.com")
	r.ServeHTTP(w, req)
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")

	// 2) 无 Origin 但带 Referer → 解析出来源。
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Referer", "https://app.example.com/hazards")
	r.ServeHTTP(w, req)
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))

	// 3) 无 Origin/Referer → 兜底 *。
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	// 4) OPTIONS 预检返回 204 并带允许头。
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}
