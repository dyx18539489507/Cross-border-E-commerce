# Digital Silk Road - 越境EC AI Agent マーケティングエンジン

Digital Silk Road（数字絲路）は、越境ECの中小事業者、輸出工場、運営代行会社向けの AI Agent マーケティング基盤です。

商品理解、コンプライアンス支援、ローカライズ、マーケティング台本、画像・動画生成、デジタルヒューマン、タイムライン編集、プラットフォーム内分析を一つのワークフローに統合します。

## 主な機能

- Planning、Product、Compliance、Localization、Content、Critic による段階的 Multi-Agent ワークフロー
- 国・プラットフォーム・カテゴリ・キーワードに基づく軽量コンプライアンス RAG
- Agent 結果からマーケティングプロジェクトを作成する業務フロー
- 画像、動画、音楽、効果音、デジタルヒューマン、編集機能の再利用
- Agent Trace、Critic 評価、履歴、ワークベンチ、プラットフォーム内推定分析

## ローカル起動

```bash
cp configs/config.example.yaml configs/config.yaml
cp .env.example .env
go run .
```

```bash
cd web
npm ci
npm run dev -- --host 127.0.0.1
```

既定ポートはバックエンド `5678`、フロントエンド開発サーバー `3012` です。

## 配置・デモ・検証

- サーバー配置：`docs/DEPLOYMENT.md`
- 大会デモ手順：`docs/DEMO_SCRIPT.md`
- 受入テスト：`docs/ACCEPTANCE_TEST.md`
- 答弁 Q&A：`docs/PRESENTATION_QA.md`
- フロントエンド依存リスク：`docs/FRONTEND_DEPENDENCY_RISKS.md`

## 注意事項

- コンプライアンス結果は補助情報であり、法的助言ではありません。
- 分析画面は現在、システム内のプロジェクト・素材・生成・配信記録に基づく推定値です。Amazon、TikTok Shop、Shopee の実広告費、注文、売上ではありません。
- 画像、動画、デジタルヒューマンの実生成には各プロバイダーの API Key が必要です。未設定時に成功結果を偽装しません。
- 既存データとの互換性のため、一部内部モデル名とテーブル名は現段階で維持しています。

詳細は [README.md](README.md) または [README-CN.md](README-CN.md) を参照してください。
