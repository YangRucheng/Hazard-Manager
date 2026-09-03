// Package auth 负责 JWT 签发/校验与 Gin 鉴权中间件。
// 用户体系：管理用户仅有 admin（密码来自环境变量，不入库）；
// 小程序用户为预留（user_type=mini），将来走同一签发/校验链路。
package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"hazard-system/server/internal/gen"
)

// UserType 用户类型。
type UserType string

const (
	// UserTypeAdmin 管理用户（网页端）。
	UserTypeAdmin UserType = "admin"
	// UserTypeMini 小程序用户（预留）。
	UserTypeMini UserType = "mini"
)

// AdminUsername 管理用户名固定为 admin。
const AdminUsername = "admin"

// Claims JWT 载荷，类型化字段。
type Claims struct {
	UserType UserType `json:"user_type"`
	jwt.RegisteredClaims
}

// Manager 负责 JWT 签发与校验。
type Manager struct {
	secret        []byte
	ttl           time.Duration
	adminPassword string
}

// NewManager 构造 Manager。
func NewManager(secret string, ttl time.Duration, adminPassword string) *Manager {
	return &Manager{
		secret:        []byte(secret),
		ttl:           ttl,
		adminPassword: adminPassword,
	}
}

// Sign 为指定用户签发 JWT（HS256）。为小程序用户预留同链路。
func (m *Manager) Sign(sub string, userType UserType) (string, error) {
	now := time.Now()
	claims := Claims{
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			Issuer:    "hazard-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("签发 JWT 失败: %w", err)
	}
	return signed, nil
}

// Verify 校验令牌签名与有效期，返回载荷。
func (m *Manager) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法 %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("令牌无效: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("令牌无效")
	}
	return claims, nil
}

// VerifyAdmin 校验管理用户账号密码（常量时间比较，避免时序侧信道）。
// 密码不与数据库比对——admin 密码由环境变量 ADMIN_PASSWORD 指定，不入库。
// 空密码一律拒绝（ConstantTimeCompare 对两个空串会返回相等，需显式拦截）。
func (m *Manager) VerifyAdmin(username, password string) bool {
	if password == "" || m.adminPassword == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(username), []byte(AdminUsername)) != 1 {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(password), []byte(m.adminPassword)) != 1 {
		return false
	}
	return true
}

const ctxClaimsKey = "auth.claims"

// JWTMiddleware 解析 Authorization: Bearer <JWT>，校验后将 Claims 写入上下文。
func JWTMiddleware(m *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeUnauthorized(c, "缺少认证令牌")
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if tokenString == "" {
			writeUnauthorized(c, "缺少认证令牌")
			return
		}
		claims, err := m.Verify(tokenString)
		if err != nil {
			writeUnauthorized(c, "认证令牌无效或已过期")
			return
		}
		c.Set(ctxClaimsKey, claims)
		c.Next()
	}
}

// RequireUserType 要求当前用户属于指定类型之一（管理端要求 admin）。
func RequireUserType(types ...UserType) gin.HandlerFunc {
	allowed := make(map[UserType]struct{}, len(types))
	for _, t := range types {
		allowed[t] = struct{}{}
	}
	return func(c *gin.Context) {
		claims, ok := c.Value(ctxClaimsKey).(*Claims)
		if !ok {
			writeUnauthorized(c, "缺少认证")
			return
		}
		if _, ok := allowed[claims.UserType]; !ok {
			writeForbidden(c, "无权限访问")
			return
		}
		c.Next()
	}
}

// ClaimsFrom 从上下文读取当前用户 Claims（中间件之后调用）。
func ClaimsFrom(c *gin.Context) *Claims {
	claims, _ := c.Value(ctxClaimsKey).(*Claims)
	return claims
}

func writeUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gen.Error{Code: "UNAUTHORIZED", Message: msg})
}

func writeForbidden(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gen.Error{Code: "FORBIDDEN", Message: msg})
}
