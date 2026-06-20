# 数字丝路——跨境电商 AI Agent 智能营销引擎

数字丝路是一个面向跨境电商中小卖家、外贸工厂和跨境代运营服务商的 AI Agent 智能营销平台。系统通过多智能体协同机制，将商品理解、合规分析、本地化策略、营销脚本生成、数字人表达和短视频成片串联为完整任务链，帮助用户降低跨境营销内容生产门槛与合规风险。

本项目基于现有 Go + Vue3 工程改造，不重写系统，不删除原有图片生成、视频生成、音乐、数字人、分镜、剪辑和分发等能力，而是将它们重新包装为“跨境营销内容生产能力”。

## 核心功能

- 多 Agent 工作流：PlanningAgent、ProductAgent、ComplianceAgent、LocalizationAgent、ContentAgent、CriticAgent 分阶段执行。
- 合规知识增强：内置 JSON 规则库，按目标国家、平台、类目和关键词进行轻量 RAG 检索。
- 跨境营销方案生成：输出商品卖点、合规边界、本地化标题、短视频脚本、数字人口播稿和投放建议。
- 一键进入工作台：Agent 方案可直接创建营销项目，带入商品、合规、脚本、数字人和投放信息。
- 营销素材生产：复用既有图片生成、视频生成、音乐/SFX、数字人、剪辑时间线和分发能力。
- 历史与实验：保存 Agent 历史、阶段 Trace、Critic 评分、命中规则和是否二次修订，支持论文实验对比。

## 数字丝路原生页面

- `/dashboard`：业务工作台总览。
- `/agent`、`/agent/result`、`/agent/history`：丝路 Agent 分析、结果与历史。
- `/projects`、`/projects/create`、`/projects/:id`：营销项目列表、创建与工作台。
- `/projects/:id/compliance`、`script`、`assets`、`editor`、`tasks`：项目合规、营销脚本与分镜、素材、剪辑和任务。
- `/compliance`、`/media/image`、`/media/video`、`/media/music`：合规中心与媒体生成能力。
- `/digital-human`、`/analytics`、`/settings`：数字人、数据分析与设置。

旧页面路径仅用于重定向兼容，不是菜单入口，也不会加载历史页面实现。

## 多 Agent 工作流

新增接口：

```text
POST /api/v1/agent/workflow
POST /api/v1/agent/generate-workflow
```

工作流阶段：

1. PlanningAgent：规划任务链、执行步骤、信息缺口和最终输出结构。
2. ProductAgent：理解商品类目、核心卖点、使用场景、目标用户、敏感点和结构化属性。
3. ComplianceAgent：检索合规知识库，输出风险等级、原因、依据、敏感表达和替代表达。
4. LocalizationAgent：生成目标市场语言风格、文化适配、视觉/口播建议和本地化卖点。
5. ContentAgent：生成营销标题、卖点文案、短视频脚本、数字人口播稿和投放建议。
6. CriticAgent：评审完整性、合规性、本地化程度和营销吸引力。

如果 CriticAgent 判断 `need_revise=true`，系统会把修改建议回传给 ContentAgent，最多执行一次二次修改，避免无限循环。最终结果包含 `traces`、`critic`、`revised`、`workflow_status` 和兼容原页面的 `SilkroadAgentResult`。

## 合规知识增强

规则库文件：

```text
data/compliance_rules.json
```

规则字段包括：

- `id`
- `country`
- `platform`
- `category`
- `risk_type`
- `rule_text`
- `forbidden_expressions`
- `safer_expressions`
- `source_url`
- `updated_at`

当前内置规则覆盖电子产品、美妆个护、母婴用品、服饰、家居用品，以及美国、英国、马来西亚、新加坡、沙特等市场和 Amazon、TikTok Shop、Shopee、通用平台等渠道。

合规输出固定包含提示：

```text
本结果仅用于跨境电商营销合规辅助，不构成法律意见；实际上架与投放前建议结合目标国家法规、平台政策和专业合规意见进行复核。
```

## 技术架构

- 后端：Go、Gin、GORM、SQLite/MySQL、OpenAI 兼容模型调用。
- 前端：Vue 3、TypeScript、Vite、Element Plus。
- Agent 服务：`application/services/silkroad_agent_service.go`、`application/services/silkroad_multi_agent_service.go`。
- 合规检索：`application/services/compliance_rule_service.go`、`data/compliance_rules.json`。
- HTTP 入口：`api/handlers/silkroad_agent.go`。
- 路由注册：`api/routes/routes.go`。
- 前端 Agent 封装：`web/src/api/agent.ts`。
- 结果页展示：`web/src/views/agent/AgentResultPage.vue`。

## 本地启动

后端：

```bash
cp configs/config.example.yaml configs/config.yaml
go mod download
go run .
```

前端：

```bash
cd web
npm ci
npm run dev -- --host 127.0.0.1
```

默认端口：后端 `5678`，前端开发服务器 `3012`。部署时可通过 `SERVER_PORT`、`VITE_API_BASE_URL` 和 `VITE_DEV_PROXY_TARGET` 覆盖。

构建与自检：

```bash
go test ./...
cd web
npm run build
```

## 环境变量

Agent 模型：

```bash
export AGENT_API_KEY="your-openai-compatible-key"
export AGENT_BASE_URL="https://api.deepseek.com"
export AGENT_TEXT_MODEL="deepseek-v4-pro"
export AGENT_VISION_API_KEY="optional-vision-key"
export AGENT_VISION_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"
export AGENT_VISION_MODEL="optional-vision-model"
```

合规模型：

```bash
export COMPLIANCE_API_KEY="your-compliance-model-key"
export COMPLIANCE_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"
export COMPLIANCE_MODEL="deepseek-v3-2-251201"
```

在开发模式且模型未配置时，系统会使用本地兜底逻辑，保证演示链路可继续运行。

## 主要 API

- `POST /api/v1/agent/extract`：轻量识别和字段归一化。
- `POST /api/v1/agent/generate`：保留兼容的原单次生成接口。
- `POST /api/v1/agent/workflow`：新增多 Agent 工作流接口。
- `POST /api/v1/agent/generate-workflow`：工作流别名接口。
- `POST /api/v1/agent/analyze`：过渡页 SSE 分析流。
- `POST /api/v1/agent/follow-up`：结果页追问 SSE。
- `GET /api/v1/agent/history`：Agent 历史列表。
- `GET /api/v1/agent/history/:id`：Agent 历史详情。
- `POST /api/v1/agent/:id/create-project`：从 Agent 历史创建营销项目。
- `POST /api/v1/agent/create-project`：直接从当前 Agent 结果创建营销项目。
- `GET /api/v1/workbench/summary`：基于项目、素材、任务和 Agent 历史的真实工作台统计。
- `GET /api/v1/analytics/summary`：基于平台内项目、生成素材、成片和分发记录的数据分析统计。
- `GET /api/v1/projects`：营销项目列表，新前端优先使用该入口。
- `POST /api/v1/projects/compliance-check`：商品合规预检。
- `POST /api/v1/projects`：复用现有兼容项目模型创建营销项目。
- `GET /api/v1/projects/:id/script`：读取营销脚本。
- `PUT /api/v1/projects/:id/script`：保存营销脚本与内容版本。
- `GET /api/v1/projects/:id/timeline`：读取时间线数据。
- `PUT /api/v1/projects/:id/timeline`：保存时间线数据。
- `GET /api/v1/projects/:id/assets`：读取项目素材。
- `GET /api/v1/projects/:id/tasks`：读取项目生成任务。
- `GET /api/v1/digital-humans`：数字人营销任务列表。
- `POST /api/v1/digital-humans`：创建数字人口播视频任务。
- `GET /api/v1/digital-humans/:id/status`：查询数字人任务状态。
- `GET /api/v1/digital-humans/:id/result`：查询数字人任务结果。
- `GET /api/v1/dramas` 等旧接口：保留兼容，不作为新前端主入口。

## 验收脚本

后端启动后可执行：

```bash
API_BASE_URL=http://127.0.0.1:5678 ./scripts/smoke-test.sh
```

脚本会检查健康检查、工作台统计、数据分析、Agent 历史、数字人任务、图片/视频任务、项目列表，以及有项目时的任务/素材子接口。比赛前可执行 `go run ./scripts/seed-demo.go` 写入明确标识的可重复 Demo 数据。

交付文档：

- `docs/DEPLOYMENT.md`：服务器、Docker、Nginx、持久化、备份与故障排查。
- `docs/DEMO_SCRIPT.md`：固定商品案例和 6-8 分钟演示路径。
- `docs/PRESENTATION_QA.md`：答辩问答与能力边界。
- `docs/ACCEPTANCE_TEST.md`：前后端和业务闭环验收。
- `docs/FRONTEND_DEPENDENCY_RISKS.md`：npm audit、Sass 和 chunk 风险说明。

## 旧能力的新定位

- AI 脚本生成：跨境营销脚本生成。
- 角色/数字人：数字人营销表达。
- 图片/视频生成：跨境营销素材生成。
- 音乐/剪辑：营销短视频成片。
- 分发/数据：营销投放与反馈分析。

## 论文实验支持

系统可记录：

- 每个 Agent 阶段的输入、输出、状态和耗时。
- 合规知识库命中规则。
- CriticAgent 评分和问题列表。
- 是否触发二次修改。
- 最终结构化营销方案。
- Agent 历史记录和项目沉淀结果。

这些数据可用于对比单 Prompt 生成、多 Agent 工作流、RAG 合规增强和 Critic 反馈修订对结果质量的影响。

## 注意事项

- 数据库表名、部分类型名和路由中仍保留 `dramas` 等历史命名，以保证兼容性。
- 图片、视频、音乐、数字人、剪辑等旧能力不会删除，而是继续服务跨境营销内容生产。
- 数据分析目前是平台内估算指标，不等同 TikTok Shop、Amazon、Shopee 的真实广告花费、转化率或店铺销售额。
- 合规结果仅供辅助，不构成法律意见；正式上架和投放前应结合目标国家法规、平台政策和专业意见复核。
