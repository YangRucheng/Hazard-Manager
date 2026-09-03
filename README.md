# 电气车间隐患闭环系统

隐患记录闭环管理平台：隐患台账、责任单位/隐患类型枚举管理、图片上传（摘要去重 + 缩略图预览）、JWT 鉴权（管理端，架构预留小程序端）。

## 技术栈

| 端 | 技术 |
|---|---|
| 后端 | Go + Gin + GORM + MySQL（本机 MariaDB），JWT 鉴权（`golang-jwt/jwt/v5`） |
| 前端 | Vue3 + VueRouter + Vite + NaiveUI + TypeScript（strict，零 `any`） |
| 契约 | `api/openapi.yaml`（OpenAPI 3.0）为前后端对齐的唯一依据 |
| 代码生成 | Go：`oapi-codegen`（types + gin）；TS：`openapi-typescript` + `openapi-fetch` |

## 目录结构

```
├── api/openapi.yaml          # 唯一接口契约（含 JWT security scheme）
├── server/                   # Go + Gin 后端
│   ├── cmd/server/main.go    # 入口
│   ├── internal/
│   │   ├── config/           # 环境变量 -> 类型化配置
│   │   ├── auth/             # JWT 签发/校验、鉴权中间件
│   │   ├── database/         # MySQL 连接、Automigrate、示例数据
│   │   ├── model/            # GORM 模型 + MySQL ENUM 类型 + Date 类型
│   │   ├── repo/             # 数据访问层（类型化查询）
│   │   ├── service/          # 业务规则（默认值、单位->责任人联动、分类归属、图片校验）
│   │   ├── handler/          # 实现 oapi-codegen 生成的 ServerInterface
│   │   ├── router/           # 路由 + CORS + 鉴权中间件挂载
│   │   ├── upload/           # 图片存储：SHA-256 去重、UUID、缩略图
│   │   └── gen/              # 由 openapi.yaml 生成（勿手改）
│   └── db/                   # init.sql（建库建用户）/ seed.sql（示例数据）
└── web/                      # Vue3 + Vite 前端
    └── src/
        ├── api/              # client.ts（openapi-fetch + 鉴权）、schema.d.ts（生成）、upload.ts
        ├── theme/            # NaiveUI 蓝色主题
        ├── stores/auth.ts    # 令牌与用户名（localStorage）
        ├── router/           # 路由与登录守卫
        ├── layouts/          # 左侧菜单 + 顶部栏（admin 下拉退出登录）
        ├── views/            # 登录 / 工作台 / 隐患列表 / 隐患表单 / 枚举值管理（责任单位、隐患类型）
        └── components/       # ImageUpload / ImagePreview / StatusTag / LevelTag
```

## 快速启动

### 1. 初始化数据库

```bash
mysql -u root < server/db/init.sql
# 建库 hazard_system、建用户 hazard（默认密码 hazard_dev_password）
```

### 2. 配置后端

```bash
cd server
cp .env.example .env
# 修改 JWT_SECRET 与 ADMIN_PASSWORD（必填，缺失时服务拒绝启动）
```

### 3. 启动后端

```bash
cd server
go build -o bin/server ./cmd/server
./bin/server          # 监听 :8090，首次启动自动建表并写入示例枚举数据
```

### 4. 启动前端

```bash
cd web
pnpm install
pnpm dev             # http://127.0.0.1:5173（/api 自动代理到 8090）
```

访问 `http://127.0.0.1:5173`，用 `admin` / `.env` 中配置的 `ADMIN_PASSWORD` 登录。

> 后端换端口时同步修改 `web/vite.config.ts` 的 proxy target。

### 5. 生产构建与后端地址（HOST）

前端请求的 API base 由构建期环境变量 `HOST` 决定（该值即后端接口的地址），
`vite.config.ts` 将其注入为 `__API_HOST__`，`web/src/api/config.ts` 负责解析：

```bash
HOST=https://hazard-manager.qcloud.19890605.xyz pnpm build
```

- `HOST` 为裸域名（如 `https://host`）→ 前端请求 `${HOST}/api/v1/*`
- `HOST` 已含路径（如 `https://host/api/v1`）→ 直接使用
- 未传 `HOST` → 回退相对 `/api/v1`（开发经 Vite proxy；生产由 nginx 反代到后端）

镜像构建同样支持传参：`docker build --build-arg HOST=https://xxx .`（见 `web/Dockerfile`）。

### 6. 跨域（CORS）

后端默认放开跨域：按请求 `Origin` 回显，缺失时按 `Referer` 解析出来源，
两者皆无则回显 `*`。鉴权使用 `Authorization: Bearer`（非 Cookie），
无需精确来源白名单，前端部署在任何域名均可访问。

## 接口契约与代码生成

修改任何字段/端点，只改 `api/openapi.yaml`，然后重新生成两端：

```bash
# Go 端（生成 internal/gen/api.gen.go）
cd server
oapi-codegen -generate types,gin -package gen -o internal/gen/api.gen.go ../api/openapi.yaml

# TS 端（生成 src/api/schema.d.ts）
cd web
pnpm schema
```

验证对齐：`cd server && go build ./...`、`cd web && pnpm typecheck` 均通过即为契约一致。

## 业务规则

- **新增默认值**：检查区域 → 华星现场；检查日期 → 今天；检查人员 → 电气自查；要求完成整改时间 → 检查日期 + 7 天；复查人员留空 → 检查人员。
- **责任单位 → 责任人**：提交仅带 `unitId`，责任人由服务端从单位表带出并冗余快照（日后改单位责任人不回写历史记录）。
- **隐患类型/分类**：单表两级（`hazard_types.parent_id`，0=大类），业务上校验"分类必须属于所选类型"。
- **整改状态 / 隐患等级**：MySQL 真实 ENUM（中文值），Go 侧强类型枚举校验，可自由编辑不做强制状态机。
- **图片**：`images` 表只存 SHA-256 摘要与 uuid（不存文件名）；其他表仅存 uuid 逗号串；同摘要重复上传返回既有 uuid，不重复保存；上传时生成最长边 320px JPEG 缩略图供列表预览（原图缺失缩略图自动回退）。
- **删除保护**：单位/类型被隐患引用（含软删除历史）或存在子分类时返回 409；隐患删除为软删除。

## 鉴权设计

- 管理用户唯一为 `admin`，密码由环境变量 `ADMIN_PASSWORD` 指定，**不入数据库**；登录签发 HS256 JWT（`JWT_SECRET`、`JWT_TOKEN_TTL_MINUTES`，默认 1440 分钟）。
- 除 `POST /api/v1/auth/login` 外，全部接口要求 `Authorization: Bearer <JWT>` 且 `user_type == "admin"`。
- JWT Claims 含 `user_type`（`admin | mini`），为后续**小程序端**预留：小程序用户届时走同一签发/校验链路、另起路由组即可，无需改动管理端鉴权。
- 登出为无状态方案（前端清除令牌），不做黑名单/refresh token（记为本期取舍与扩展点）。
- 空密码恒定时间比较已显式拦截（防止 `ConstantTimeCompare` 空串误判）。

## 测试与质量

```bash
# 后端
cd server
go test ./...                       # 单测 + 集成测试（未设 TEST_DB_DSN 时集成测试跳过）
TEST_DB_DSN='hazard:hazard_dev_password@tcp(127.0.0.1:3306)/hazard_system_test?charset=utf8mb4&parseTime=True&loc=Local' go test ./...   # 含集成测试（先建 hazard_system_test 库）
go vet ./...

# 前端
cd web
pnpm typecheck
pnpm build
```

集成测试覆盖：默认值、检查日期+7 联动、单位→责任人联动、分类归属、非法状态拦截、单位不存在 404/422、统计。

## 验收清单

- [ ] 未登录访问受保护接口返回 401；错误密码与正确密码登录行为正确；`/auth/me` 返回 admin。
- [ ] 新增隐患：默认值（华星现场 / 今天 / 电气自查 / 检查日期+7 天 / 复查=检查人员）、选责任单位自动带出责任人、类型/分类级联、多图上传去重（同图同 uuid、磁盘无重复文件）、缩略图可预览。
- [ ] 列表：筛选（状态/等级/类型级联/区域/单位/关键字/日期范围）、分页、图片列缩略图、逾期标记、编辑/删除。
- [ ] 枚举管理：责任单位与隐患类型两个子 tab 增删改；被引用删除返回冲突提示。
- [ ] UI：左侧导航（工作台/隐患管理/枚举值管理）、顶部右上 admin 下拉退出登录、全站蓝色系、警示色仅用于状态/等级标签。