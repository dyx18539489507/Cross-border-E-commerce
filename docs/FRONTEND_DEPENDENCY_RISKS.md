# 前端依赖与构建风险说明

审计日期：2026-06-19。

## 1. npm audit 结果

初始审计：13 项漏洞，其中 9 high、4 moderate。

执行不带 `--force` 的 `npm audit fix` 后：

- Axios 升级到 `1.18.0`。
- lodash-es/lodash 升级到 `4.18.1`。
- PostCSS 升级到 `8.5.15`。
- follow-redirects、form-data、immutable、minimatch、picomatch、rollup 等间接依赖升级到安全版本。
- 剩余 2 项：1 high、1 moderate。

剩余风险来自 Vite 5.4.21 依赖的 esbuild 0.21.5 和 Vite 开发服务器相关公告。npm 给出的完整修复路径是升级到 Vite 8.0.16，属于跨主版本升级，可能同时影响 `@vitejs/plugin-vue`、Node.js 版本要求、开发代理和构建配置，因此本轮未执行 `npm audit fix --force`。

## 2. 风险判断

- Vite/esbuild 仅用于开发与构建，不进入 Go 后端运行时。
- 生产环境只部署 `web/dist` 静态产物，不对公网暴露 `vite dev`。
- 开发服务器只应绑定可信网络；公开演示优先使用 Nginx 或 Go 后端提供静态产物。
- `web/package-lock.json` 已纳入交付，Docker 使用 `npm ci`，避免每次部署解析到不同依赖。

## 3. 后续升级建议

在独立升级分支中完成：

1. 升级 Node.js 到 Vite 8 要求的版本。
2. 升级 `vite` 与 `@vitejs/plugin-vue`。
3. 验证开发代理、SSE、媒体代理、动态导入和生产构建。
4. 运行 `npm audit`、`npm run build:check` 和核心页面回归。

## 4. Sass legacy API 警告

当前构建会输出 Dart Sass legacy JS API 弃用提示。它来自现有 Vue/Element Plus/Sass 构建链，不影响当前 CSS 产物。后续随 Vite、插件和组件库升级一起处理，不建议在比赛交付前单独替换 Sass 实现。

## 5. Chunk 体积警告

主包超过 Vite 默认 500 kB 警告阈值，主要来自 Element Plus、编辑器、FFmpeg、图表和多工作流页面。当前使用路由懒加载且功能可用。后续可：

- 配置 `build.rollupOptions.output.manualChunks`。
- 将 FFmpeg、专业制作台和图表模块按需加载。
- 精细化 Element Plus 组件导入。
- 用 bundle analyzer 定位重复依赖。

本轮不通过调高警告阈值掩盖问题，也不为减包删除核心功能。
