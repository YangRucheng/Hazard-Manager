# AGENTS.md — 面向 AI 编程智能体的项目约定

本文档约束在 `Hazard-Manager`（电气车间隐患闭环系统）仓库内工作的 AI 编程智能体。
与 `README.md`（面向人类开发者）互补，此处聚焦「如何改代码、如何提交、如何上线」。

## 项目概况

- 前端：Vue 3 + TypeScript + Vite + Naive UI（无 Pinia，鉴权为轻量组合式 store），位于 `web/`。
- 后端：Go + Gin + GORM + MySQL（本机 MariaDB），位于 `server/`。
- 契约：`api/openapi.yaml` 为前后端对齐的唯一依据；Go 端由 `oapi-codegen` 生成 `server/internal/gen/api.gen.go`，TS 端由 `openapi-typescript` 生成 `web/src/api/schema.d.ts`，**禁止手改生成文件**。
- 鉴权：JWT（HS256）；管理用户唯一 `admin`，密码由环境变量 `ADMIN_PASSWORD` 指定不入库；Claims 中 `user_type` 预留小程序用户（`mini`），管理端接口要求 `user_type == "admin"`。
- 图片：`images` 表只存 SHA-256 摘要与 uuid（不存文件名），其他表仅引用 uuid；相同摘要不重复保存；上传时生成最长边 320px JPEG 缩略图供列表预览。
- 部署：本地进程——后端 `:8090`，前端 Vite `:5173`（`/api` 代理到后端）。
- 默认工作目录：仓库根目录 `/workspace/隐患闭环系统`。远程仓库：`YangRucheng/Hazard-Manager`，默认分支 `master`。

## 必读先做

动手改代码前，先阅读并遵守：

- `api/openapi.yaml` — 接口契约；改后端接口时同步契约，并用两端 codegen 重新生成（见「契约同步」），否则两边编译对不上。
- `README.md` — 启动步骤、业务规则、默认值/联动规则、验收清单。
- `.github/workflows/` — CI 流水线（若有）。
- 后端分层 `config → auth → database → model → repo → service → handler → router`：handler 只做 HTTP 编解码，业务规则一律放 service。

## 验证命令（提交前必须通过）

后端在 `server/` 目录：

```bash
go build ./...
go vet ./...
gofmt -l ./internal ./cmd   # 应为空
go test ./...               # 单测（auth/model）；集成测试在设置 TEST_DB_DSN 时自动启用
# 含集成测试（需先建 hazard_system_test 库）：
TEST_DB_DSN='hazard:hazard_dev_password@tcp(127.0.0.1:3306)/hazard_system_test?charset=utf8mb4&parseTime=True&loc=Local' go test ./...
```

前端在 `web/` 目录：

```bash
pnpm typecheck   # vue-tsc --noEmit（strict，零 any）
pnpm build       # 生产构建
```

## 契约同步

修改任何接口 / 字段 / 枚举，只改 `api/openapi.yaml`，然后重新生成两端：

```bash
# Go 端
cd server && oapi-codegen -generate types,gin -package gen -o internal/gen/api.gen.go ../api/openapi.yaml
# TS 端
cd web && pnpm schema
```

`go build ./...` 与 `pnpm typecheck` 同时通过即视为契约一致；生成文件必须随本次改动一起提交。

## 代码与提交规范

- **提交信息**：中文 + Conventional Commits 前缀（`feat` / `fix` / `style` / `chore` / `ci` / `docs` / `refactor` / `test`），一行简洁标题 + 空行 + 详细说明（可选列点）。
- **分功能点提交**：一个逻辑改动（一个功能 / 一个修复）对应一个 commit；不要把无关改动混进同一 commit。
- **分支命名**：`<type>/<kebab-case-描述>`，如 `docs/agents-md`、`feat/hazard-export`。
- **Go 端类型安全**：所有 DB 模型与 DTO 均为类型化结构体，禁止 JSON 动态取值；枚举使用强类型（如 `model.HazardStatus` / `model.HazardLevel`）并经 `Valid()` 校验，DB 列为真实 MySQL ENUM。
- **前端类型安全**：TS strict，禁止 `any`（需要时用 `unknown` + 窄化）；接口类型只从 `src/api/schema.d.ts` 派生，禁止自行声明 `any` 结构。
- **图片约定**：新图经 `POST /api/v1/images` 上传换取 uuid（同图自动去重），库里只存逗号分隔的 uuid 串；`model.JoinImages` / `model.SplitImages` 负责拼接与拆分，勿手写字符串处理。
- **隐患类型模型约定**：`hazard_types` 每行 = 一个「大类(`major`)+小类(`minor`)」组合（无 `parent_id`、无排序、无启停，同一组合唯一）；`hazards` 只存 `type_id` 引用该组合行，**没有 `category_id`**。新增/编辑时大类用小类使用「下拉输入框」（可下拉复用已有大类，也允许输入新大类）。删除被隐患引用的类型由后端拒绝（409），只能修改。

## UI 规范（面向所有前端改动，必须遵守）

> 交互与观感对齐仓库内同作者的 `备件管理系统`（`/workspace/备件管理系统`，Web 同为 Vue3+NaiveUI）。

- **按钮必须是按钮**：一切可点击操作都用 `n-button`（及 `n-upload` 等原生控件）的实体按钮样式；**禁止 `text` 纯文字按钮 / “彩色文字当按钮”**。行内操作用 `size="small"` 的普通或 `secondary` 按钮，删除等危险操作用 `type="error" secondary`。
- **删除按钮放进编辑弹窗**：系统管理页（责任单位/隐患类型）列表“操作”列只放「编辑」；「删除」按钮放在**编辑弹窗底部左侧**，点击先二次确认再调 DELETE。被引用数据删除会返回 409，前端展示后端提示即可（不要据此臆造禁用逻辑）。附件管理无编辑弹窗，删除留在行内（二次确认后调 DELETE，被引用同样由后端 409 兜底）。
- **开关一律用 `n-switch` 组件**：状态启停直接渲染裸开关，**开关内部不放任何文字**（不用 checked/unchecked 插槽塞 启用/停用 字样），语义由列标题「状态」表达。
- **侧栏白底且可折叠**：左侧导航背景为白色（`Layout.siderColor: #ffffff`），桌面端支持折叠（`show-trigger="bar"`，折叠宽 64px）；移动端（`useIsMobile`，≤820px）改为顶部汉堡按钮 + `n-drawer` 抽屉导航。
- **移动端响应式**：新增页面/组件必须适配窄屏——工具条控件用 class 控宽并在媒体查询下占满整行；表格给 `:scroll-x` 横向滚动；弹窗宽度带 `max-width: 94vw`；分页移动端用 `simple`。页签（line 型 tabs）导航下方不放通栏分割线（主题中 `Tabs.tabBorderColor: 'transparent'` 已全局去除）。
- **禁用灰色小字说明**：不要用 `≤12px` + 灰色（如 `#6b7a90`/`#8a97ab`）的散落小字做功能说明/描述。确有需要说明时改用正式标题/`n-tag`/表单占位符/校验提示；总条数等次要信息用正文色，不用低对比灰色。
- **布局惯例**：页面用 `page-header`/工具条 + 卡片化 `n-card`，筛选区放顶部工具条；警示/成功等语义色只用于状态/等级标签与真实语义强调，不用于普通文字。
- 涉及 UI 的改动需在 `web/` 下 `pnpm typecheck && pnpm build` 通过。

## 标准开发与发布工作流（必须遵守）

> 目标：每个功能点独立成 PR，合并后不留残余分支。所有操作在仓库根目录执行。

### 1. 从最新 master 开功能分支

```bash
git checkout master && git pull
git checkout -b <type>/<描述>
```

### 2. 分功能点提交

```bash
git add <本次功能点涉及的文件>
git commit -m "type: 中文标题"
```

确认 `git status` 干净、无本功能点之外的改动（尤其确认 `.env` / `data/` 未混入）。

### 3. 推送并创建 PR

```bash
git push -u origin <分支名>
gh pr create --title "type: 中文标题" --body "<问题/原因/改动/验证>"
```

PR 标题与 commit 标题一致（squash 后作为最终 commit 标题）。

### 4. 检查通过后才合并

- 本地验证通过（见「验证命令」）。
- `gh pr checks <编号> --watch` 等 CI 全部通过（若有）。
- 自查 `gh pr diff <编号>` 确认改动只涉及本功能点。
- 通过后合并：本仓库只允许 squash：

```bash
gh pr merge <编号> --squash --delete-branch
```

`--delete-branch` 会同时删除本地和远端的功能分支。

### 5. 回到 master 并同步

```bash
git checkout master && git pull
```

### 6. 清理多余分支

- 合并后功能分支已由 `--delete-branch` 删除。
- 定期清理远端孤立 / 已关闭 PR 的分支：

```bash
git fetch --prune
gh pr list --state all   # 找出 CLOSED 且未合并的分支
git push origin --delete <多余分支名>
git branch -D <多余分支名>  # 如需删除本地对应分支
```

## 注意事项

- 不直接提交到 `master`：所有改动经功能分支 + PR 合并（squash）。
- 不手改生成文件（`api.gen.go`、`schema.d.ts`），统一通过脚本再生成。
- `server/.env` 含 `ADMIN_PASSWORD` / `JWT_SECRET`，已被 `.gitignore` 忽略，禁止提交。
- 涉及接口时同步更新 `api/openapi.yaml`，并在 PR 描述中说明改动与验证方式。
- PR 描述遵循「问题 / 原因 / 改动 / 验证」结构，便于 review 与回溯。