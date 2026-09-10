---
name: smart-commit
description: "git の差分を適切な粒度でグループ化し、コミットメッセージを生成してコミット・push・PR 作成まで行うスキル。「コミットして」「変更をコミットして push して」「PR 作成して」などのリクエストでトリガーする。"
---

# Smart Commit — 差分グループ化・自動コミットスキル

## 処理ステップ概要

| Step | 内容                   | 概要                                                             |
| ---- | ---------------------- | ---------------------------------------------------------------- |
| 1    | 差分の取得             | `git status` / `git diff` で変更ファイルを把握する               |
| 2    | ブランチ確認           | 現在のブランチから PR の base とテンプレートを決める             |
| 3    | グループ化             | 変更ファイルを機能・ディレクトリ・変更種別で適切にグループ化する |
| 4    | コミットメッセージ生成 | 規約に従い各グループのコミットメッセージを生成する               |
| 5    | コミット実行           | グループ単位で `git add` → `git commit` を順に実行する           |
| 6    | push                   | Bash の `git push` で現在のブランチを push する                  |
| 7    | PR 確認・作成 (任意)   | 既存 PR を確認し、なければテンプレートに従って作成する           |
| 8    | 結果報告               | コミット数・push 結果・PR URL をユーザーに報告する               |

## 制約事項

> [!IMPORTANT]
>
> - **コミットは必ず 1 グループずつ順番に実行する** (並列実行禁止)
> - `.env` やシークレットを含む可能性があるファイルはコミット前にユーザーに確認する
> - **`main` へ直接コミット・push しない** (`CLAUDE.md`)

- コミットメッセージは `.claude/rules/github/commit-types.md` と `.claude/rules/github/commit-subject.md` に従う
- PR のタイトル・本文は `.claude/rules/github/pr-description.md` に従い、本文は `.github/PULL_REQUEST_TEMPLATE/` のテンプレートを埋める
- push には GitHub MCP の `push_files` ではなく Bash の `git push` を使う
- PR 作成は GitHub MCP (`create_pull_request`) を使う
- **GitHub のラベルは付けない** (`.claude/rules/github/issue.md` の「ラベル」を参照)

---

## 各ステップの詳細

### Step 1: 差分の取得

以下のコマンドで変更状況を把握する。

```bash
git status
git diff --stat HEAD
```

- 未追跡ファイル・変更済みファイル・削除済みファイルをすべて把握する
- 変更がない場合はスキルを中断してユーザーに報告する

### Step 2: ブランチ確認

```bash
git branch --show-current
```

現在のブランチから、PR の base と使うテンプレートを決める。

| 現在のブランチ           | PR の base       | テンプレート |
| ------------------------ | ---------------- | ------------ |
| `release/main-issue-N/…` | `main`           | `release.md` |
| `feature/sub-issue-N/…`  | リリースブランチ | `sub.md`     |
| `chore/issue-N/…` など   | `main`           | `single.md`  |
| `main`                   | —                | —            |

- **`main` の場合**: 「現在 `main` ブランチです。作業ブランチを切りますか？」と確認する。直接コミットはしない
- サブブランチの場合、base にするリリースブランチを次で探し、複数あるか見つからない場合はユーザーに確認する

```bash
git branch -r --list 'origin/release/*'
```

### Step 3: グループ化

変更ファイルを以下の優先順位でグループ化する。

**グループ化の基準:**

| 優先度 | 基準                   | 例                                       |
| ------ | ---------------------- | ---------------------------------------- |
| 1      | 論理的なまとまり       | 新機能 1 つ、バグ修正 1 つ、など目的単位 |
| 2      | ディレクトリのまとまり | `.claude/rules/` 配下の変更をまとめる    |
| 3      | 変更種別のまとまり     | 設定ファイル群、ドキュメント群など       |

**グループ化のルール:**

- 1 コミットに含めるファイルは「同じ目的・同じ文脈」のものに限る
- 関係のないファイルを 1 つのコミットに混在させない
- 1 ファイルのみの変更でもそのままグループ単体でコミットしてよい

グループ化の結果をユーザーに提示し、問題がなければ Step 4 へ進む。

**提示フォーマット:**

```text
以下のグループでコミットします。よろしいですか？

[グループ 1] #12 feat: 優先パスの申込ハンドラを追加 (backend)
  - backend/internal/prioritypass/handler/connect.go

[グループ 2] #12 chore: commit rules を追加 (chore)
  - .claude/rules/github/commit-types.md
  - .claude/rules/github/commit-subject.md
```

### Step 4: コミットメッセージ生成

`.claude/rules/github/commit-subject.md` の形式でメッセージを生成する。

```text
#<Issue 番号> <type>: <subject> (<スコープ>)

<本文>
```

- **Issue 番号と本文は必須。** 対応する Issue が無い場合は先に Issue を作る
- 本文は「何が問題だったか → どう変えたか」の順で書く
- type は `.claude/rules/github/commit-types.md`、スコープは `.claude/rules/github/labels.md` から選ぶ

> [!NOTE]
>
> 規約に合わないメッセージは PreToolUse フック (`.claude/hooks/check-commit-message.py`) が拒否する。
> 拒否されたら書き直して再実行する。

### Step 5: コミット実行

グループ単位で順番にコミットする。

```bash
git add <対象ファイル...>
git commit -m "$(cat <<'MSG'
#<Issue 番号> <type>: <subject> (<スコープ>)

<本文>
MSG
)"
```

**注意事項:**

> [!IMPORTANT]
>
> - `git add -A` や `git add .` は使わず、ファイルを明示的に指定する
> - コミット前に `.env` やシークレットらしいファイルが含まれている場合はユーザーに確認する

### Step 6: push

全コミット完了後、Bash で push する。

```bash
git push origin <current-branch>
```

作業ブランチの初回 push では `-u` を付ける。

> [!NOTE]
>
> force push は行わない。

### Step 7: PR 確認・作成 (任意)

push 完了後、GitHub MCP (`list_pull_requests`) で現在のブランチの PR が既に存在するか確認する。

```text
既存 PR あり → PR の URL をユーザーに報告して終了 (新規作成しない)
既存 PR なし → 「PR を作成しますか？ (y/n)」とユーザーに確認する
```

**No** の場合はそのまま Step 8 へ進む。

**Yes** の場合は GitHub MCP (`create_pull_request`) で PR を作成する。

- base は Step 2 で決めたブランチにする
- タイトルは `.claude/rules/github/pr-description.md` の 3 形式から、ブランチに対応するものを選ぶ
- 本文は Step 2 で決めたテンプレートを読み、その見出し構成のまま埋める。埋められない項目とコメント (`<!-- -->`) は削除する
- PR の URL をユーザーに報告する

### Step 8: 結果報告

以下をユーザーに報告する。

- 作成したコミット数とその一覧 (hash・メッセージ)
- push 先のブランチ名
- PR URL (作成した場合のみ)
- エラーが発生した場合はその内容と原因
