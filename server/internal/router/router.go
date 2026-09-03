// Package router 组装 Gin 引擎：中间件与路由注册。
package router

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"hazard-system/server/internal/auth"
	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/handler"
)

// HealthResponse 健康检查响应（类型化）。
type HealthResponse struct {
	Status string `json:"status"`
}

// NewRouter 构建 Gin 引擎。
// 鉴权策略：仅 POST /api/v1/auth/login 公开，其余 /api/v1/* 均需有效 JWT 且 user_type=admin。
func NewRouter(srv *handler.Server, am *auth.Manager) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(corsMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
	})

	api := r.Group("/api/v1")
	api.Use(apiAuthMiddleware(am))

	// 全部端点注册（含 login，login 在 apiAuthMiddleware 中放行）。
	gen.RegisterHandlersWithOptions(api, srv, gen.GinServerOptions{BaseURL: ""})

	return r
}

// apiAuthMiddleware 校验 JWT：login 端点放行，其余要求 admin 用户。
func apiAuthMiddleware(am *auth.Manager) gin.HandlerFunc {
	jwtMW := auth.JWTMiddleware(am)
	adminMW := auth.RequireUserType(auth.UserTypeAdmin)
	return func(c *gin.Context) {
		if c.FullPath() == "/api/v1/auth/login" || c.FullPath() == "/healthz" {
			c.Next()
			return
		}
		jwtMW(c)
		if c.IsAborted() {
			return
		}
		adminMW(c)
	}
}

// resolveAllowOrigin 解析放行的跨域来源：
//  1. 优先取 Origin 请求头；
//  2. 否则从 Referer 解析出协议 + 主机（前端页面跨域请求通常携带 Referer）；
//  3. 兜底使用 *。
//
// 鉴权采用 Authorization: Bearer（非 Cookie），不启用 Allow-Credentials，
// 因此按来源回显或 * 均不会造成凭据泄露。
func resolveAllowOrigin(origin, referer string) string {
	if o := originFromHeader(origin); o != "" {
		return o
	}
	if r := originFromReferer(referer); r != "" {
		return r
	}
	return "*"
}

// originFromHeader 校验 Origin 是否为合法 http(s) 来源。
func originFromHeader(origin string) string {
	if origin == "" {
		return ""
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// originFromReferer 从 Referer 提取协议 + 主机（去掉路径）。
func originFromReferer(referer string) string {
	if referer == "" {
		return ""
	}
	u, err := url.Parse(referer)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// corsMiddleware 允许跨域：来源按 Origin/Referer 回显，兜底 *。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowOrigin := resolveAllowOrigin(c.GetHeader("Origin"), c.GetHeader("Referer"))
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
