// Package version 承载编译期注入的后端构建信息。
package version

import "time"

// BuildTime 后端编译时间。经 -ldflags -X hazard-system/server/internal/version.BuildTime=<RFC3339> 注入；
// 本地 go run/build 未注入时为 "unknown"。
var BuildTime = "unknown"

// StartTime 服务进程启动时刻（包加载即进程启动时）。
var StartTime = time.Now()
