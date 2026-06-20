# 数字丝路验收说明

本文件用于快速验收“数字丝路——跨境电商 AI Agent 智能营销引擎”的核心闭环，避免只通过页面观感判断功能是否真实接入。

## 1. 环境要求

- Go 1.23+
- Node.js 20+
- npm 10+
- 默认 SQLite；也可配置 MySQL
- 后端默认端口：`5678`
- 前端默认端口：`3012`
- 可选 AI/生成服务环境变量：`AGENT_API_KEY`、`DEEPSEEK_API_KEY`、图片/视频/数字人供应商 Key

合规结果和 Agent 结果仅用于跨境电商营销辅助，不构成法律意见；实际上架与投放前仍需结合目标国家法规、平台政策和专业合规意见复核。

## 2. 后端验收

```bash
go mod tidy
go build ./...
go test ./...
go run .
```

服务启动后执行：

```bash
API_BASE_URL=http://127.0.0.1:5678 ./scripts/smoke-test.sh
```

脚本会检查：

- `/health`
- `/api/v1/workbench/summary`
- `/api/v1/analytics/summary`
- `/api/v1/agent/history`
- `/api/v1/digital-humans`
- `/api/v1/images`
- `/api/v1/videos`
- `/api/v1/projects`
- 有项目数据时检查 `/api/v1/projects/:id/tasks` 和 `/api/v1/projects/:id/assets`

## 3. 前端验收

```bash
cd web
npm ci
npm run build
npm run dev -- --host 127.0.0.1
```

浏览器访问：

- `http://127.0.0.1:3012/`
- `http://127.0.0.1:3012/dashboard`
- `http://127.0.0.1:3012/agent`
- `http://127.0.0.1:3012/projects`
- `http://127.0.0.1:3012/projects/create`
- `http://127.0.0.1:3012/digital-human`
- `http://127.0.0.1:3012/analytics`

页面验收重点：

- 工作台统计来自 `/api/v1/workbench/summary`，具备 loading、error、empty 状态。
- 数据分析来自 `/api/v1/analytics/summary`，指标标注为平台内估算，不冒充真实广告或店铺数据。
- Agent 页面通过 `web/src/api/agent.ts` 调用后端，支持 workflow、history、follow-up 和一键创建营销项目。
- 数字人页面或工作流弹窗可创建任务、查看历史、查询状态、查看结果。
- 营销视频剪辑页的音乐、音效、上传、媒体代理走 `web/src/api/`，视图层不直接请求后端 URL。

## 4. 可重复 Demo 数据

比赛演示可先执行：

```bash
DEMO_DEVICE_ID=demo-device-digital-silk-road go run ./scripts/seed-demo.go
```

前端 `web/.env.local` 同时设置：

```env
VITE_DEMO_DEVICE_ID=demo-device-digital-silk-road
```

Seed 明确标识为 Demo，不调用第三方 AI，不生成虚假广告、订单、销售额或数字人成功结果。

## 5. 业务闭环验收

1. 进入商品录入页，填写商品名称、描述、目标市场和平台，上传商品图片。
2. 进入丝路 Agent，执行多 Agent 工作流并查看执行链、合规规则命中、Critic 评分。
3. 点击“一键创建营销项目”，确认跳转到项目或脚本工作台。
4. 在项目详情查看商品信息、目标市场、合规结论、营销内容和数字人建议。
5. 进入合规分析，确认风险等级、风险原因和修改建议来自接口。
6. 进入脚本/内容生成，保存脚本后刷新页面确认数据仍存在。
7. 创建图片、视频或数字人任务，确认任务进入历史列表并可查询状态。
8. 进入 `/projects/:id/editor` 营销视频剪辑页，验证音乐/音效搜索、素材预览和时间线保存。
9. 回到工作台和数据分析页，确认项目数、任务数、素材统计发生真实变化。

## 6. 当前边界

- 数据分析基于平台内项目、素材生成、成片和分发记录估算，不等同 TikTok Shop、Amazon、Shopee 的真实广告消耗、订单或转化数据。
- 图片、视频、数字人能力依赖外部供应商 Key；未配置时可能只能创建本地记录或显示供应商错误。
- 后端仍保留 `dramas`、`episodes`、`storyboards`、`characters` 等旧表与兼容接口，新对外入口优先使用 `/projects`。
