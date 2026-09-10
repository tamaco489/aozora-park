# CLAUDE.md

このファイルは Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

架空テーマパーク **Aozora Park** の来園予約システム。Google Cloud のサーバレス構成 (Cloud Run / Firestore / Pub/Sub / Cloud Tasks / Cloud Scheduler / Identity Platform) を Terraform で構築し、Connect (connect-go / connect-web) で API を組む。環境は `dev` のみ。

## 規約の場所

規約は `.claude/rules/` にあり、すべて読み込まれる。**このファイルには規約を再掲しない。**

| ディレクトリ             | 内容                               |
| ------------------------ | ---------------------------------- |
| `.claude/rules/general/` | 応答・表記、作業の進め方、justfile |
| `.claude/rules/github/`  | コミット、Issue、PR、ラベル        |
| `.claude/rules/go/`      | Go のコーディングとテスト          |
| `.claude/rules/infra/`   | Terraform のコーディング           |

- [docs/README.ja.md](docs/README.ja.md) — プロジェクトの構成・アーキテクチャ

## ソースコードの探索

> [!IMPORTANT]
>
> - **調査に即時性は一切求めない。** 時間をかけてでも正確に探索し、確信が持てるまで結論を出さない
> - **広く探すときは `Explore` サブエージェントに委譲する。** ファイルの中身をメインコンテキストに溜めない
> - **場所が分かっているファイルは委譲せず直接 Read する。** 委譲は「どこにあるか分からない」ときだけ
> - **探索範囲を先に絞る。** モノレポなので `proto/` `backend/` `frontend/` `infra/` のどこを見るかを決めてから探す
> - **生成コードを探索対象にしない。** `backend/gen/` と `frontend/src/gen/` は buf の出力で、定義元は `proto/` にある
> - **読まずに実装を語らない。** 推測で答えず、根拠にしたファイルと位置を示す

## このリポジトリの禁止事項

- **AWS を使わない。** クラウドは Google Cloud のみ
- `main` ブランチへの直接コミット・push の禁止
- Git フック・署名のスキップ禁止 (`--no-verify`, `--no-gpg-sign`)
- `rm -rf` の使用禁止 — ファイル削除は `rm -f` を使う
- 生成コードの手動編集禁止 — `backend/gen/` と `frontend/src/gen/` は `buf generate` で再生成する
- 機密情報のハードコーディング禁止 (API キー、fincode の認証情報、Slack Webhook URL、接続情報)
  - 秘匿値は Secret Manager に置く。Terraform はシークレットの入れ物と参照だけを定義し、値の投入はユーザーが行う
  - リポジトリと GitHub Secrets に長期クレデンシャルを置かない (デプロイは Workload Identity Federation)
- **インフラ適用・GCP リソース操作の禁止** — 以下はユーザーのみが実行する。Claude が実行してはならない:
  - `terraform apply` / `terraform destroy` (`terraform fmt` / `validate` / `plan` は可)
  - `gcloud run deploy` / `gcloud run jobs update` / `gcloud run jobs execute`
  - `firebase deploy`
  - Pub/Sub への publish、Cloud Tasks へのタスク投入、Firestore への書き込み
  - GCP の読み取り (`gcloud ... list` / `describe`) は確認目的で行ってよい
- **外部サービスの実呼び出し禁止** — fincode と Slack はユーザーの承認を得てから実行する。fincode は検証環境のみを使い、本番環境は使わない。テストはフェイクとエミュレータで行う

## フック

`.claude/rules/github/` の規約に合わない操作は PreToolUse フックがブロックする。

| フック                    | 対象                                       | 検査するもの                             |
| ------------------------- | ------------------------------------------ | ---------------------------------------- |
| `check-commit-message.py` | `git commit`                               | subject の形式・本文の有無               |
| `check-issue.py`          | `issue_write` (create) / `gh issue create` | タイトルの形式・本文の見出し             |
| `check-pr.py`             | `create_pull_request` / `gh pr create`     | タイトルの形式・ブランチ名・base・見出し |
