// Package router 组装 Gin 引擎：中间件与路由注册。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hazard-system/server/internal/auth"
	"hazard-system/server/internal/config"
	"hazard-system/server/internal/gen"
	"hazard-system/server/internal/handler"
)

// HealthResponse 健康检查响应（类型化）。
type HealthResponse struct {
	Status string `json:"status"`
}

// NewRouter 构建 Gin 引擎。
// 鉴权策略：仅 POST /api/v1/auth/login 公开，其余 /api/v1/* 均需有效 JWT 且 user_type=admin。
func NewRouter(srv *handler.Server, am *auth.Manager, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(corsMiddleware(cfg.CORSOrigins))

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

// corsMiddleware 允许配置的跨域来源。
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
