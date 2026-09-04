// 框架级异常统一 JSON 输出。
//
// 业务层 / handler / auth 中间件均已按契约输出 gen.Error（code + message）；
// 本文件补齐框架兜底路径，保证 404 / 405 / 500（panic）/ 参数绑定错误
// 也返回同构的 JSON，而非 gin 默认的纯文本或空响应体。
package router

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"hazard-system/server/internal/gen"
)

// errorResponse 以统一 JSON（契约 Error：code + message）写出错误响应。
func errorResponse(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gen.Error{Code: code, Message: message})
}

// recoveryMiddleware 捕获 handler panic，服务端记录完整堆栈，对客户端返回统一 JSON 500。
// 替代 gin.Recovery()：其默认仅输出空 500，不满足统一 JSON 约定。
func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] %s %s: %v\n%s", c.Request.Method, c.Request.URL.Path, rec, debug.Stack())
				errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
			}
		}()
		c.Next()
	}
}

// noRouteHandler 未知路由统一返回 JSON 404（替代 gin 默认纯文本 404）。
func noRouteHandler(c *gin.Context) {
	errorResponse(c, http.StatusNotFound, "NOT_FOUND", "接口不存在")
}

// noMethodHandler 路径存在但方法不允许时统一返回 JSON 405（并携带 Allow 头）。
func noMethodHandler(c *gin.Context) {
	errorResponse(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "请求方法不允许")
}

// oapiErrorHandler 将 oapi-codegen 生成的路径/查询参数解析错误转成统一 JSON。
// 生成代码目前固定以 400 调用本回调；此处按状态码做防御性映射，
// 并服务端记录原始错误便于排查。
func oapiErrorHandler(c *gin.Context, err error, status int) {
	code, message := "BAD_REQUEST", "请求参数格式错误"
	switch status {
	case http.StatusNotFound:
		code, message = "NOT_FOUND", "接口不存在"
	case http.StatusMethodNotAllowed:
		code, message = "METHOD_NOT_ALLOWED", "请求方法不允许"
	case http.StatusInternalServerError:
		code, message = "INTERNAL_ERROR", "服务器内部错误"
	}
	log.Printf("[oapi-error] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	errorResponse(c, status, code, message)
}
