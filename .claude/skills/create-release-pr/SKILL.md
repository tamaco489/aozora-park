---
name: create-release-pr
description: "リリースブランチを main へマージするリリース PR を作成するスキル。サブ Issue の消化状況を確認し、release.md テンプレートを埋めて PR を出す。「リリース PR を作って」「リリースブランチを main にマージする PR を作成して」などのリクエストでトリガーする。"
---

# Create Release PR — リリース PR 作成スキル

リリースブランチ 1 本を `main` へマージする PR を作る。メイン Issue に対応する。

**サブ PR (`feature/sub-issue-N/...` → リリースブランチ) はこのスキルの対象外。** `smart-commit` が作る。

## 処理ステップ概要

| Step | 内容           | 概要                                                        |
| ---- | -------------- | ----------------------------------------------------------- |
| 1    | 前提の確認     | リリースブランチ上にいることとメイン Issue を確定する       |
| 2    | 消化状況の確認 | サブ Issue とサブ PR の状態を集め、未完了があれば確認を取る |
| 3    | 差分の把握     | `main` との差分から、実際に入る変更を確かめる               |
| 4    | 本文の作成     | `release.md` を埋め、提示して承認を得る                     |
| 5    | PR 作成        | GitHub MCP で作成する                                       |
| 6    | 結果報告       | PR の URL と残作業を報告する                                |
| 7    | 振り返り       | 今回の進め方を振り返り、改善案があれば提案して反映する      |

## 制約事項

- **PR の作成は Step 4 の承認を得てから行う**
- **マージはしない。** `main` へのマージはユーザーが行う
- 本文の型は `.github/PULL_REQUEST_TEMPLATE/release.md` を正とし、このファイルにテンプレートを持たない
- PR 作成は GitHub MCP (`create_pull_request`) を使う
- **リリース PR とメイン Issue には、サブ Issue のラベルをまとめて付ける。** `create_pull_request` はラベルを受け取らないため、作成後に `gh pr edit` で付ける

---

## 各ステップの詳細

### Step 1: 前提の確認

```bash
git branch --show-current
```

- ブランチ名が `release/main-issue-<番号>/<名前>` でなければ、**このスキルの対象外であることを伝えて中断する**
- ブランチ名からメイン Issue の番号を取り出し、`issue_read` で内容 (目的・完了条件・識別子) を読む
- リリースブランチが push 済みで、`main` より進んでいることを確かめる

### Step 2: 消化状況の確認

メイン Issue に紐づくサブ Issue と、リリースブランチへの PR を集める。

**リリース PR はリリースブランチを作った直後に出してよい。** サブ Issue が全件 open でも構わない。
先に出しておくとサブ PR の集約先とリリースの全体像が 1 か所で追える。
未完了があることは中断の理由にならず、下の確認は「未完了があると承知のうえで作るか」を尋ねるためのもの。

- `issue_read` (method: `get_sub_issues`) でサブ Issue の一覧と open / closed を取る
- `list_pull_requests` で base がリリースブランチの PR を取り、マージ済みかを見る

未完了があれば一覧にして、**続行するかを確認する。**

```text
未完了のサブ Issue が 2 件あります。このままリリース PR を作りますか？

  #725 [sub] [feat] 記事一覧画面の刷新 — open (PR なし)
  #728 [sub] [fix] 残数の減算を修正 — PR #741 レビュー中
```

### Step 3: 差分の把握

```bash
git log main..HEAD --oneline
git diff --stat main...HEAD
```

- サブ Issue の一覧と実際のコミットが食い違っていないかを見る
- 想定外のコミットが混じっていれば報告する

### Step 4: 本文の作成

`.github/PULL_REQUEST_TEMPLATE/release.md` を読み、その見出し構成のまま埋める。

- タイトルは `[Release] [識別子] <メイン Issue のタイトル>`。識別子はメイン Issue と揃える
- 「目的」はメイン Issue の目的を要約し、背景の詳細はメイン Issue へのリンクで済ませる
- 「サブイシュー」の表は Step 2 で集めた Issue 番号・内容・状態 (マージ済みの PR 番号) で埋める
- 「マージ前の確認事項」はメイン Issue の完了条件から作る
- 「関連 Issue」には `Closes #<メイン Issue の番号>` を書く
- 個々の変更の説明はサブ PR にあるので繰り返さない。**統合して初めて確かめられることだけを書く**

タイトルと本文を提示して承認を得る。

### Step 5: PR 作成とラベル

```text
create_pull_request(owner, repo, title, body, head: <リリースブランチ>, base: "main")
```

作成できたら、Step 2 で集めたサブ Issue のラベルを重複を除いてまとめ、PR とメイン Issue の両方に付ける。

```bash
gh pr edit <PR 番号> --add-label "<ラベル>,<ラベル>,..."
gh issue edit <メイン Issue の番号> --add-label "<ラベル>,<ラベル>,..."
```

- ラベルはサブ Issue に付いているものを使う。**PR の差分から推測しない**
- サブ Issue にラベルが付いていないものがあれば、**そのまま付けずに報告する**
- **GitHub 上に無いラベルを渡すとエラーになる。** その場合は `sync-labels` スキルで同期してからやり直す

### Step 6: 結果報告

- PR の URL
- PR とメイン Issue に付けたラベルの一覧
- 未完了のサブ Issue が残っている場合はその一覧
- 次にやること (レビュー後、ユーザーが `main` へマージし、メイン Issue が閉じる)

### Step 7: 振り返り

**このスキルを実行した過程を振り返り、次に同じ作業をするときに効率が上がる点があれば提案する。**

観点と進め方は `create-release-issues` の Step 9 と同じ。

- 改善案が無ければ「なし」と伝えて終わる
- 提案は多くても 3 件。「観察した事実 → 何が困るか → どのファイルをどう直すか」の形で書く
- ユーザーが採用したものだけ、diff を提示して承認を得てから修正する
- 修正対象は `SKILL.md`、`.claude/rules/`、`.github/PULL_REQUEST_TEMPLATE/` に限る
