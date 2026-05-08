# ccLoad 与本项目差异及可迁移能力评估

## 结论摘要

`ccLoad/` 不是本项目的同构分支，而是另一个架构方向的 AI API 代理项目。它的核心特点是：后端根目录单体 Go 服务、所有 `/v1/*` 与 `/v1beta/*` 走统一透明代理入口、渠道/Key/日志/令牌/设置统一落 SQL 存储，并提供静态 HTML 管理页。

本项目当前是：`backend-go/` 按协议拆分 handler/provider/converter/scheduler，前端是 Vue 3 + Vite + Vuetify，配置主数据仍在 `.config/config.json`，指标单独落 SQLite。它对 Claude Messages、OpenAI Chat、Codex Responses、Gemini、Images 分协议建模，协议边界更清晰，前端交互更完整。

因此不建议把 `ccLoad` 整体覆盖进来。更合理的方向是：保留本项目分协议架构和 Vue 前端，把 `ccLoad` 中更成熟的运行态能力按模块迁入。

## 顶层结构差异

| 维度 | 本项目 | ccLoad |
| --- | --- | --- |
| 后端位置 | `backend-go/` 独立 Go module | 根目录 Go module |
| 前端形态 | `frontend/` Vue 3 + Vite + Vuetify，构建后 embed | `web/` 多个静态 HTML/CSS/JS 页面，直接 embed |
| 后端模块 | `config`、`handlers/<protocol>`、`providers`、`converters`、`scheduler`、`metrics` 等分层 | 主要集中在 `internal/app`，辅以 `storage`、`protocol`、`cooldown`、`model`、`util` |
| 配置主存储 | `.config/config.json`，metrics 单独 SQLite | SQLite/MySQL/混合存储，渠道、Key、日志、Token、设置统一在数据库 |
| 代理路由 | 显式注册 `/v1/messages`、`/v1/responses`、`/v1/chat/completions`、Gemini、Images | `/v1/*`、`/v1beta/*` 统一捕获后识别协议和请求族 |
| 管理 API | `/api/<kind>/...`，按 messages/responses/chat/gemini/images 分族 | `/admin/...` 单套渠道模型，靠 `channel_type` 与暴露协议区分 |
| 鉴权 | `PROXY_ACCESS_KEY` / `ADMIN_ACCESS_KEY` 静态密钥 | 管理登录 session + 数据库 API tokens，支持令牌限额/模型限制 |
| 测试规模 | 后端约 199 个 Go 文件、101 个 `_test.go` | 约 282 个 Go 文件、154 个 `_test.go` |

## 功能差异

### 本项目已经具备，没必要重复迁移

- 多协议代理：Messages、Responses、Chat、Gemini、Images 都已有独立 handler。
- 协议转换：已有 `backend-go/internal/converters/` 与 `providers/`，并有 Responses/Chat/Gemini/Claude 相关测试。
- 多 BaseURL：`UpstreamConfig.BaseURLs`、`GetAllBaseURLs()`、metrics 多 URL 聚合已存在。
- 渠道级自定义请求头：`CustomHeaders` 已存在，且上游认证头有覆盖保护。
- 路由前缀：本项目支持 `/:routePrefix/...`。
- Claude failover rules、key 黑名单、模型健康检查、promotion、circuit breaker、capability test、channel logs 等核心调度和管理能力已有。

### ccLoad 明显更强的能力

- 数据库存储体系：`storage.Store` 覆盖渠道、API keys、cooldown、logs、debug logs、metrics、auth tokens、settings、admin sessions；支持 SQLite、MySQL、MySQL+SQLite 混合模式。
- API token 管理：令牌从 Web 管理，支持 per-token 成本限额、模型限制、使用统计、TTFB。
- 成本核算：日志中记录 cost、cache token、service_tier、cost_multiplier，并支持渠道每日成本限制。
- 透明代理入口：统一接管 `/v1/*`，支持更多 OpenAI 兼容路径的自然扩展。
- 协议暴露模型：一个渠道可以配置 `protocol_transforms`，即同一上游渠道向外暴露多个客户端协议。
- URL 运行态管理：多 URL EWMA 延迟权重、URL 级禁用/恢复、URL 级统计。
- 健康分排序：基于成功率、样本置信度动态调整有效优先级。
- Debug logs：保存上游请求/响应原始数据并脱敏，适合排查第三方渠道协议不一致。
- Active requests：管理端可查看正在进行中的请求。
- CSV 导入/导出与批量操作：渠道批量导入、批量启停、批量优先级、批量删除。
- 系统设置热管理：`system_settings` 表驱动管理运行配置。
- Zstd 管理端压缩、可信代理配置、全局并发信号量等生产化细节。

### ccLoad 相对弱或不适合直接迁移的部分

- 前端是静态 HTML 页面，和本项目 Vue/Vuetify 架构不匹配，不建议直接搬 UI 文件。
- 后端 `internal/app` 职责较集中；本项目现有分协议模块更清晰，直接迁入会破坏当前模块边界。
- 统一 `/v1/*` 透明代理很灵活，但会削弱本项目当前对协议差异、指标隔离和测试矩阵的显式控制。
- ccLoad 的 SQL 存储改造会替换当前 `.config/config.json` 主配置模型，属于高风险架构迁移。
- ccLoad 文档和部分注释编码显示异常，需要迁移前统一检查编码与文本质量。

## 建议迁移清单

### 优先级 P0：最值得迁，收益高且边界清楚

1. **API token 管理与限额**
   - 迁移目标：引入可管理的代理访问 token，支持启停、模型限制、成本限额、使用统计。
   - 适配方式：后端新增独立 token store 和 `/api/tokens`，前端用 Vue 新建 token 管理页。
   - 不建议照搬 `/admin/auth-tokens` 路由和静态 HTML。

2. **Debug logs**
   - 迁移目标：可按 channel log 查看脱敏后的上游 request/response，用于排查协议兼容问题。
   - 适配方式：接入本项目 `channel log` 生命周期，单独 SQLite 表或扩展 metrics store。
   - 注意：默认关闭，避免记录大 body 或敏感信息。

3. **Active requests**
   - 迁移目标：前端展示当前流式/非流式请求正在用哪个渠道、模型、BaseURL、耗时、状态。
   - 适配方式：在 `common.TryUpstreamWithAllKeys` 或各 handler 入口注册运行态 request。
   - 这是纯运行态能力，不必依赖 ccLoad 的 SQL 存储。

4. **CSV/批量渠道操作**
   - 迁移目标：批量导入/导出、批量启停、批量优先级调整、批量删除。
   - 适配方式：落在当前 `ConfigManager` 与 Vue 渠道页上，复用现有 `UpstreamConfig` 格式。
   - 本项目明确不需要旧格式兼容，可以设计新的导入格式。

### 优先级 P1：值得迁，但需要与现有调度融合

1. **URL 级 EWMA 选择和禁用**
   - 本项目已有多 BaseURL 和 `warmup.URLManager`，可吸收 ccLoad 的 URL 级统计、手动禁用、延迟加权思想。
   - 不建议另起一套 `URLSelector`；应扩展现有 `warmup` / `scheduler` 语义。

2. **健康分动态排序**
   - ccLoad 的 `effective_priority = base_priority - failure_rate * weight * confidence` 思路可迁入。
   - 本项目已有 priority、promotion、circuit breaker，迁入时要明确排序优先级：promotion > disabled/suspended > circuit/cooldown > health score > configured priority。

3. **成本统计与每日成本限制**
   - 适合引入到 channel metrics 和前端 dashboard。
   - 需要先统一 token usage、cache token、service_tier、cost multiplier 的数据合同。

4. **系统设置热管理**
   - 可把部分 env/config.json 设置迁到 UI 管理，例如日志保留、健康分窗口、检查间隔。
   - 但不要一次性替换所有环境变量；安全相关配置仍应保留环境变量。

5. **Zstd 压缩中间件**
   - 对管理 API 和静态资源有收益。
   - 本项目已有 gzip middleware，迁移前要评估浏览器兼容、代理兼容和 double-compress 风险。

### 优先级 P2：可借鉴设计，不建议近期迁

1. **统一 SQL 主存储 / MySQL / 混合存储**
   - 长期看有价值，尤其是多实例、HuggingFace/容器重启不丢配置场景。
   - 但它会重构 `ConfigManager`、配置加载、前端 API、迁移/备份策略，风险最大。
   - 如果做，建议先把 metrics/logs/tokens 放 SQL，再决定是否把 channels 从 config.json 迁出。

2. **统一透明代理 `/v1/*`**
   - 适合 OpenAI 兼容路径扩展。
   - 本项目目前的显式 handler 更利于协议测试和隔离，不建议整体替换；可以只为低风险 OpenAI 兼容路径做 fallback。

3. **协议 transform registry**
   - ccLoad 的 registry 抽象优雅，但本项目已有 converters/providers。
   - 可以借鉴“TransformPlan + registry”的元数据模型，避免直接替换现有转换器。

4. **管理登录 session**
   - 当前项目的静态 admin key 简单可靠。
   - 若要做多用户/审计，再考虑引入 session；否则优先做 API token 管理即可。

## 不建议迁移的内容

- `ccLoad/web/` 静态页面：只参考页面能力，不复制实现。
- `ccLoad/internal/app` 整体 server/proxy handler：会和当前分协议 handler 冲突。
- `ccLoad` 的路由命名：本项目已有 `/api/<kind>/...` 合同，除非做破坏式 API 重设计，否则不应迁 `/admin/...`。
- `ccLoad` 的完整 Store interface 一次性迁入：接口过大，和本项目当前 package ownership 不匹配。
- 旧格式兼容迁移逻辑：本项目已明确“不保留 backward compatibility”，新格式可以直接设计。

## 推荐落地顺序

1. 先做 `active requests` + `debug logs`，两者对排障收益最大，且不要求改变配置存储。
2. 再做 API token 管理和 token 维度统计，为多用户/限额/审计打基础。
3. 然后做 CSV/批量操作和 URL 级禁用，提升管理效率。
4. 最后评估健康分排序、成本限制、SQL 主存储这类会影响调度或数据模型的改造。

## 迁移时的代码位置建议

- Token 管理：`backend-go/internal/auth` 或 `backend-go/internal/tokens`，前端新增 `frontend/src/views/TokensView.vue` 与 `src/services/api.ts` 方法。
- Debug logs：扩展 `backend-go/internal/metrics` 或新建 `backend-go/internal/debuglog`，不要塞进 protocol handler。
- Active requests：新建 `backend-go/internal/runtime` 或 `backend-go/internal/handlers/common/active_requests.go`，由 common failover/stream 路径写入状态。
- URL 级状态：优先扩展 `backend-go/internal/warmup` 和 `backend-go/internal/scheduler`。
- 批量渠道操作：按现有协议族落在 `internal/handlers/<protocol>/channels.go`，前端复用 `ChannelOrchestration.vue`。
- 成本统计：先扩展 `metrics.ChannelLog` / SQLite schema，再接 UI。

## 总体判断

`ccLoad` 值得长期吸收的是运行态治理能力：可管理 token、调试日志、请求可见性、URL 级健康、成本/限额、SQL 化运营数据。本项目不应迁它的静态前端和单体代理结构，而应把这些能力重写进现有分协议后端和 Vue 管理台。

如果只选一个最有价值的迁移方向，建议从“API token 管理 + debug logs + active requests”开始。这三项能显著提升生产可用性，且不会先把项目拖进配置存储大迁移。
