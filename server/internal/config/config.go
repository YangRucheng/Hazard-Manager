// Package config 负责从环境变量（可经 .env 加载）读取类型化配置。
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 应用配置，全部字段类型化，杜绝动态取值。
type Config struct {
	ServerPort     string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	JWTTokenTTL    time.Duration
	AdminPassword  string
	UploadDir      string
	MaxUploadBytes int64
}

// Load 读取配置。ADMIN_PASSWORD 与 JWT_SECRET 为必填项，缺失返回错误。
func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8090"),
		DBHost:         getEnv("DB_HOST", "127.0.0.1"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBUser:         getEnv("DB_USER", "hazard"),
		DBPassword:     getEnv("DB_PASSWORD", "hazard_dev_password"),
		DBName:         getEnv("DB_NAME", "hazard_system"),
		UploadDir:      getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxUploadBytes: int64(getEnvInt("MAX_UPLOAD_BYTES", 10*1024*1024)), // 默认 10MB
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("环境变量 JWT_SECRET 未设置，拒绝启动")
	}

	cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	if cfg.AdminPassword == "" {
		return nil, fmt.Errorf("环境变量 ADMIN_PASSWORD 未设置，拒绝启动")
	}

	ttlMinutes, err := strconv.Atoi(getEnv("JWT_TOKEN_TTL_MINUTES", "1440")) // 默认 24h
	if err != nil || ttlMinutes <= 0 {
		return nil, fmt.Errorf("JWT_TOKEN_TTL_MINUTES 必须为正整数")
	}
	cfg.JWTTokenTTL = time.Duration(ttlMinutes) * time.Minute

	return cfg, nil
}

// DSN 返回 MySQL 连接串。
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// ImageURLPrefix 生成图片相对路径前缀，前端以自己的 baseURL 拼接。
func (c *Config) ImageURLPrefix() string { return "/api/v1/images" }

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
