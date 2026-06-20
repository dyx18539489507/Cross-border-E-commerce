# Digital Silk Road - Cross-border E-commerce AI Agent Marketing Engine

Digital Silk Road is an AI Agent marketing platform for cross-border e-commerce sellers, export factories, and agency operators. It connects product understanding, compliance-assisted analysis, localization strategy, marketing script generation, digital-human presentation, media production, editing, and distribution into one practical workflow.

The project is built on the existing Go + Vue3 codebase. Historical image generation, video generation, music, digital human, shot planning, timeline editing, and distribution capabilities are retained as cross-border marketing production tools.

## Core Capabilities

- Multi-Agent workflow: PlanningAgent, ProductAgent, ComplianceAgent, LocalizationAgent, ContentAgent, and CriticAgent run in stages.
- Compliance knowledge enhancement: local JSON rule base plus lightweight keyword retrieval for target markets, platforms, and categories.
- Marketing content production: product selling points, localized titles, short-video scripts, digital-human voiceover, visual style, and campaign advice.
- Workbench loop: Agent results can create a marketing project and enter script, content generation, compliance, or editing workspaces.
- Reused media stack: image/video generation, music/SFX, digital human, marketing shot planning, timeline editing, and distribution support marketing video production.
- History and experiment support: Agent sessions, workflow traces, critic scores, and matched rules are saved for replay and paper experiments.

## Native Application Routes

- `/dashboard`, `/agent`, `/agent/result`, `/agent/history`
- `/projects`, `/projects/create`, `/projects/:id`
- `/projects/:id/compliance`, `/script`, `/assets`, `/editor`, `/tasks`
- `/compliance`, `/media/image`, `/media/video`, `/media/music`
- `/digital-human`, `/analytics`, `/settings`

Legacy browser paths only redirect for compatibility. They are not menu entries and do not load historical page implementations.

## Multi-Agent Workflow

`POST /api/v1/agent/workflow` executes a real staged workflow:

1. PlanningAgent builds the task chain, missing information list, and output plan.
2. ProductAgent structures category, selling points, scenarios, users, attributes, and sensitive points.
3. ComplianceAgent retrieves matching compliance rules and generates a risk boundary with disclaimer.
4. LocalizationAgent adapts language style, culture, visuals, voiceover, and local selling points.
5. ContentAgent generates marketing titles, scripts, digital-human copy, and promotion advice.
6. CriticAgent scores completeness, compliance, localization, and marketing appeal.

If the CriticAgent sets `need_revise=true`, ContentAgent receives revision advice and regenerates content once. The workflow response includes `traces`, `critic`, `revised`, `workflow_status`, and the final `SilkroadAgentResult`.

## Compliance RAG

The lightweight compliance knowledge base lives at:

```text
data/compliance_rules.json
```

Each rule includes country, platform, category, risk type, forbidden expressions, safer expressions, source URL, and update date. Current seed rules cover electronics, beauty/personal care, baby products, fashion, and home goods across the US, UK, Malaysia, Singapore, Saudi Arabia, Amazon, TikTok Shop, Shopee, and general platforms.

Compliance output is only an assistant signal:

> This result is for cross-border e-commerce marketing compliance assistance only and does not constitute legal advice. Before listing or advertising, review target-country regulations, platform policies, and professional compliance opinions.

## Architecture

- Backend: Go, Gin, GORM, SQLite/MySQL, OpenAI-compatible LLM calls.
- Frontend: Vue 3, TypeScript, Vite, Element Plus.
- Agent service: `application/services/silkroad_agent_service.go` and `silkroad_multi_agent_service.go`.
- Compliance retrieval: `application/services/compliance_rule_service.go` and `data/compliance_rules.json`.
- Agent APIs: `api/handlers/silkroad_agent.go`, registered in `api/routes/routes.go`.
- Workbench persistence: Agent results are adapted to the existing compatibility data model behind the project API.

## Local Setup

Backend:

```bash
cp configs/config.example.yaml configs/config.yaml
go mod download
go run .
```

Frontend:

```bash
cd web
npm ci
npm run dev -- --host 127.0.0.1
```

Default ports: backend `5678`, frontend dev server `3012`. Use `SERVER_PORT`, `VITE_API_BASE_URL`, and `VITE_DEV_PROXY_TARGET` to override deployment endpoints.

Build checks:

```bash
go test ./...
cd web
npm run build
```

## Environment Variables

Common Agent variables:

```bash
export AGENT_API_KEY="your-openai-compatible-key"
export AGENT_BASE_URL="https://api.deepseek.com"
export AGENT_TEXT_MODEL="deepseek-v4-pro"
export AGENT_VISION_API_KEY="optional-vision-key"
export AGENT_VISION_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"
export AGENT_VISION_MODEL="optional-vision-model"
```

Compliance service variables:

```bash
export COMPLIANCE_API_KEY="your-compliance-model-key"
export COMPLIANCE_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"
export COMPLIANCE_MODEL="deepseek-v3-2-251201"
```

When LLM credentials are absent in debug mode, the workflow uses local fallback logic so the staged showcase remains available.

## Main APIs

- `POST /api/v1/agent/extract` - normalize Agent input.
- `POST /api/v1/agent/generate` - legacy-compatible single-result generation.
- `POST /api/v1/agent/workflow` - staged multi-Agent workflow.
- `POST /api/v1/agent/generate-workflow` - alias for workflow generation.
- `POST /api/v1/agent/analyze` - transition-page SSE analysis.
- `POST /api/v1/agent/follow-up` - result-page follow-up SSE.
- `GET /api/v1/agent/history` - list Agent histories.
- `GET /api/v1/agent/history/:id` - get one history.
- `POST /api/v1/agent/:id/create-project` - create a marketing project from Agent history.
- `POST /api/v1/agent/create-project` - create a marketing project from posted Agent result.
- `GET /api/v1/workbench/summary` - real dashboard summary based on projects, assets, tasks, and Agent histories.
- `GET /api/v1/analytics/summary` - in-platform analytics summary based on projects, generated assets, videos, and distribution records.
- `GET /api/v1/projects` - list marketing projects; this is the preferred public alias over legacy project routes.
- `POST /api/v1/projects/compliance-check` - product compliance precheck.
- `POST /api/v1/projects` - create a marketing project using the existing compatible project model.
- `GET /api/v1/projects/:id/script` - read saved marketing scripts.
- `PUT /api/v1/projects/:id/script` - save marketing scripts and content versions.
- `GET /api/v1/projects/:id/timeline` - read saved timeline data.
- `PUT /api/v1/projects/:id/timeline` - save editable timeline data.
- `GET /api/v1/projects/:id/assets` - list project assets.
- `GET /api/v1/projects/:id/tasks` - list project generation tasks.
- `GET /api/v1/digital-humans` - list digital-human marketing tasks.
- `POST /api/v1/digital-humans` - create a digital-human speaking-video task.
- `GET /api/v1/digital-humans/:id/status` - get task status.
- `GET /api/v1/digital-humans/:id/result` - get task result.
- `GET /api/v1/dramas` and related legacy routes - retained for backward compatibility.

## Acceptance Checks

Run `API_BASE_URL=http://127.0.0.1:5678 ./scripts/smoke-test.sh` after starting the backend. For a stable competition dataset, run `go run ./scripts/seed-demo.go` with the documented Demo device ID.

Delivery documents:

- `docs/DEPLOYMENT.md` - server, Docker, Nginx, persistence, backup, and troubleshooting.
- `docs/DEMO_SCRIPT.md` - fixed product case and 6-8 minute presentation route.
- `docs/PRESENTATION_QA.md` - competition defense questions and boundaries.
- `docs/ACCEPTANCE_TEST.md` - front-end, back-end, and business-flow checks.
- `docs/FRONTEND_DEPENDENCY_RISKS.md` - npm audit, Sass, and chunk warnings.

## Paper Experiment Support

The system records intermediate Agent traces, matched compliance rules, CriticAgent scores, revision status, final structured output, and session history. These fields support ablation comparison between single-prompt generation, staged workflow generation, RAG-enhanced compliance, and Critic-guided revision.

## Notes

- Existing database/model names such as `dramas` are retained for compatibility, but the product experience is positioned as cross-border marketing projects.
- Existing image, video, music, digital-human, shot planning, editing, and distribution capabilities remain available as marketing content production modules.
- Analytics currently reflect in-platform estimated metrics, not live ad spend, conversion, or store sales from TikTok Shop, Amazon, or Shopee.
- Compliance results are auxiliary and conservative; they are not legal opinions.
