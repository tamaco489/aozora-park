# CLAUDE.md

このファイルは Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

架空テーマパーク **Aozora Park** の来園予約システム。Google Cloud のサーバレス構成 (Cloud Run / Firestore / Pub/Sub / Cloud Tasks / Cloud Scheduler / Identity Platform) を Terraform で構築し、Connect (connect-go / connect-web) で API を組む。環境は `dev` のみ。

- `.claude/rules/general/` — 応答、作業の進め方、justfile の規約
- `.claude/rules/github/` — コミットと PR の規約
- `.claude/rules/go/` — Go のコーディングとテストの規約
- `.claude/rules/terraform/` — Terraform のコーディング規約
- [docs/README.ja.md](docs/README.ja.md) — プロジェクトの構成・アーキテクチャ

## 制約事項

> [!IMPORTANT]
>
> - **即時性は求めない。時間をかけてでも根拠に基づく正確なアウトプットを行う**
> - **公式ドキュメントや関連資料の調査はメインコンテキストを汚さないよう、別途調査用エージェントに委譲する**

- **コード変更前に必ずファイルを Read ツールで読む**
- **変更は diff 形式で提示し、承認 (y) を得てから実行する**
- **git commit はユーザーの承認を得てから実行する**
- 応答は日本語・簡潔・直接的
- コメントは「なぜ」が自明でない場合のみ書く (「何をしているか」は書かない)
- コメントに句点 (。) を含めない
- AWS は使わない。クラウドは Google Cloud のみ

## 禁止事項

- `rm -rf` の使用禁止 — ファイル削除は `rm -f` を使う
- 明示的な指示なしの変更禁止
- Git フック・署名のスキップ禁止 (`--no-verify`, `--no-gpg-sign`)
- `main` ブランチへの直接 push 禁止
- 生成コードの手動編集禁止 — `backend/gen/` と `frontend/src/gen/` は `buf generate` で再生成する
- 機密情報のハードコーディング禁止 (API キー、fincode の認証情報、Slack Webhook URL、接続情報)
  - 秘匿値は Secret Manager に置く。Terraform はシークレットの入れ物と参照だけを定義し、値の投入はユーザーが行う
  - リポジトリと GitHub Secrets に長期クレデンシャルを置かない (デプロイは Workload Identity Federation)
- 絵文字の使用禁止 (明示的に求められた場合を除く)
- **インフラ適用・GCP リソース操作の禁止** — 以下はユーザーのみが実行する。Claude が実行してはならない:
  - `terraform apply` / `terraform destroy` (`terraform fmt` / `validate` / `plan` は可)
  - `gcloud run deploy` / `gcloud run jobs update` / `gcloud run jobs execute`
  - `firebase deploy`
  - Pub/Sub への publish、Cloud Tasks へのタスク投入、Firestore への書き込み
  - GCP の読み取り (`gcloud ... list` / `describe`) は確認目的で行ってよい
- **外部サービスの実呼び出し禁止** — fincode と Slack はユーザーの承認を得てから実行する。fincode は検証環境のみを使い、本番環境は使わない。テストはフェイクとエミュレータで行う

## フック

`git commit` のメッセージが `.claude/rules/github/` の規約に合わない場合、PreToolUse フック (`.claude/hooks/check-commit-message.py`) がコミットをブロックする。
