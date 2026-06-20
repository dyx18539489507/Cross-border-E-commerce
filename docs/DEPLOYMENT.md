# 数字丝路部署指南

本文面向比赛演示服务器和后续测试环境。生产商用前仍需补充用户认证、权限隔离、密钥管理、审计日志、限流、监控和计费能力。

## 1. 环境要求

- 操作系统：推荐 Ubuntu 22.04/24.04 或同等级 Linux。
- Go：`1.23.x`，与 `go.mod` 保持一致。
- Node.js：`20.x`，npm `10.x`。
- FFmpeg：时间线合成、音频提取需要；Docker 镜像已安装。
- 数据库：默认 SQLite，无需单独服务；也支持 MySQL。
- 建议资源：2 核 CPU、4 GB 内存、20 GB 可用磁盘起步。真实视频生成和大量素材需要更多存储。

## 2. 配置准备

```bash
cp configs/config.example.yaml configs/config.yaml
cp .env.example .env
cp web/.env.example web/.env.local
```

真实密钥只写入 `.env` 或服务器密钥管理服务，不要写入 `configs/config.yaml`、前端源码或 Git。

关键默认值：

- 后端：`0.0.0.0:5678`，可用 `SERVER_HOST`、`SERVER_PORT` 覆盖。
- 前端开发服务器：`0.0.0.0:3012`。
- API 前缀：`VITE_API_BASE_URL=/api/v1`。
- Vite 开发代理：`VITE_DEV_PROXY_TARGET=http://127.0.0.1:5678`。
- SQLite：`./data/drama_generator.db`。
- 素材目录：`./data/storage`，浏览器默认通过 `/static/` 访问。

前后端同域部署时保持相对 API 前缀。前后端分域时，在构建前设置完整地址：

```bash
VITE_API_BASE_URL=https://api.example.com/api/v1 npm run build
```

## 3. 后端启动

```bash
go mod download
go build -o bin/digital-silk-road .
./bin/digital-silk-road
```

首次启动会自动创建 SQLite 目录、素材目录并执行 GORM AutoMigrate。健康检查：

```bash
curl http://127.0.0.1:5678/health
```

修改端口示例：

```bash
SERVER_PORT=5680 STORAGE_BASE_URL=http://localhost:5680/static ./bin/digital-silk-road
```

`SERVER_READ_TIMEOUT`、`SERVER_WRITE_TIMEOUT` 的单位为秒，默认 600，适配较慢的 AI 和媒体任务。

## 4. 前端启动与构建

```bash
cd web
npm ci
npm run dev -- --host 127.0.0.1
```

开发端口默认 `3012`。若被占用，Vite 会在终端显示实际端口，也可显式指定：

```bash
npm run dev -- --host 127.0.0.1 --port 3013
```

生产构建：

```bash
npm run build
npm run preview -- --host 127.0.0.1 --port 4173
```

后端会从 `web/dist` 提供前端静态文件，因此单进程部署可以直接访问 `http://服务器:5678/`。

## 5. 数据与持久化目录

| 路径 | 用途 | 是否持久化 |
|---|---|---|
| `data/drama_generator.db` | SQLite 主库 | 必须 |
| `data/drama_generator.db-wal`、`-shm` | SQLite WAL 文件 | 与主库一起管理 |
| `data/storage/` | 上传文件、图片、视频、音频、数字人输入输出、本地 Demo 素材 | 必须 |
| `/tmp/drama-video-merge/` | FFmpeg 下载、裁剪、合成临时文件 | 不需要，重启可清理 |
| 标准输出/标准错误 | 应用日志 | 建议由 systemd、Docker 或日志平台收集 |

本地存储会按业务动态创建 `characters`、`images`、`videos`、`audios`、`materials`、`digital-human`、`sfx` 等子目录，无需手工预建。

SQLite 备份建议先停止写入，再复制主库及 WAL 文件；安装了 sqlite3 时也可在线执行：

```bash
sqlite3 data/drama_generator.db ".backup 'backup/digital-silk-road-$(date +%F).db'"
```

## 6. 第三方服务边界

- Agent：`AGENT_API_KEY` / `DEEPSEEK_API_KEY`，可选视觉模型变量。
- 合规模型：`COMPLIANCE_API_KEY`；缺失时使用本地规则与保守兜底。
- 图片/视频：在“设置 > AI 服务配置”保存供应商 Base URL、Key、模型和 endpoint。
- 数字人/TTS：火山引擎 `VOLCENGINE_*` 变量。
- 音乐：搜索和代理依赖外部公开服务的可用性；音效生成依赖相应服务能力。
- 分发：`UPLOAD_POST_API_KEY`、`DISTRIBUTION_SECRET_KEY` 等。

未配置第三方 Key 时，接口会返回配置缺失或供应商错误；不会伪造真实图片、视频或数字人成功结果。Agent 调试模式可能返回明确标记的本地兜底方案。

## 7. Nginx 示例

前端静态目录假设为 `/srv/digital-silk-road/web/dist`，后端监听 `127.0.0.1:5678`：

```nginx
server {
    listen 80;
    server_name demo.example.com;
    root /srv/digital-silk-road/web/dist;
    index index.html;

    client_max_body_size 110m;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:5678;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;
        proxy_read_timeout 900s;
        proxy_send_timeout 900s;
    }

    location /static/ {
        proxy_pass http://127.0.0.1:5678;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900s;
    }
}
```

Agent SSE 流使用普通 HTTP stream，无单独 WebSocket 路由；`proxy_buffering off` 可避免事件被 Nginx 缓冲。

## 8. Docker Compose

```bash
cp configs/config.example.yaml configs/config.yaml
cp .env.example .env
docker compose build
docker compose up -d
docker compose ps
docker compose logs -f digital-silk-road
```

Compose 挂载 `./data:/app/data`，数据库和素材会保留在宿主机。远程服务器必须把 `STORAGE_BASE_URL` 改为外部可访问的 `/static` 地址。

## 9. 从克隆到上线

```bash
git clone <repository-url> digital-silk-road
cd digital-silk-road
cp configs/config.example.yaml configs/config.yaml
cp .env.example .env

go mod download
go build -o bin/digital-silk-road .

cd web
npm ci
npm run build
cd ..

APP_DEBUG=false ./bin/digital-silk-road
```

随后配置 Nginx、HTTPS、防火墙和进程守护，并执行：

```bash
API_BASE_URL=http://127.0.0.1:5678 ./scripts/smoke-test.sh
```

## 10. 常见问题

- 端口占用：用 `lsof -iTCP:5678 -sTCP:LISTEN` 检查，或设置 `SERVER_PORT`；前端用 `--port`。
- 前端请求失败：确认 `VITE_API_BASE_URL` 是完整 `/api/v1` 前缀，修改后必须重新构建。
- CORS：同域 Nginx 部署不需要额外 CORS；分域时设置 `SERVER_CORS_ORIGINS`，多个地址用逗号分隔。
- 上传失败：检查 `data/storage` 写权限、Nginx `client_max_body_size` 和磁盘空间。
- SQLite 无权限：运行用户必须能写 `data/`；不要把数据库放在只读镜像层。
- 生成失败：先检查“AI 服务配置”和对应环境变量，不要把供应商错误当成本地任务成功。
- npm audit：当前剩余风险来自 Vite 5 开发服务器，详见 `docs/FRONTEND_DEPENDENCY_RISKS.md`；生产环境只部署构建产物。
- Sass legacy API：构建警告，不影响当前产物；后续升级 Element Plus/Sass 调用链。
- chunk 过大：主包较大但可运行；后续按编辑器、图表和生成模块拆分 manual chunks。
