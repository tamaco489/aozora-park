# デプロイの構成

[English](./README.md) | [日本語](./README.ja.md)

[ドキュメント一覧](../README.ja.md)に戻る。

このプロジェクトには、仕組みの異なる 2 つのデプロイの経路があります。ここでは両方を俯瞰します。
それぞれの詳細は [backend のデプロイの構成](./backend/README.ja.md)と [frontend のデプロイの構成](./frontend/README.ja.md)にあります。
GCP への認証に使う仕組みは [Workload Identity Federation](./wif/README.ja.md)にあります。

## 2 つの経路

|                  | backend (Cloud Run のサービスごと)                                             | frontend                                                         |
| ---------------- | ------------------------------------------------------------------------------ | ---------------------------------------------------------------- |
| 配信先           | Cloud Run                                                                      | Firebase Hosting                                                 |
| ワークフロー     | `cd-<サービス名>-stg.yaml` (サービスごとに 1 本)                               | `cd-frontend-stg.yaml`                                           |
| ビルドの実行場所 | **Cloud Build (GCP の中)**                                                     | **GitHub Actions のランナー**。手元から起こしたときは手元        |
| ソースの取得元   | GitHub (Developer Connect 経由)                                                | ワークフローの checkout。手元から起こしたときは作業ツリー        |
| 未コミットの変更 | 反映されない                                                                   | CI からは反映されない。手元からは**反映される**                  |
| 自動の起点 (stg) | `main` への push (`backend/**`)                                                | `main` への push (`frontend/**`、`firebase.json`、`.firebaserc`) |
| 手で起こす (stg) | `workflow_dispatch`、または `cd backend && just deploy-<サービス名>-stg <ref>` | `workflow_dispatch`、または `cd frontend && just deploy-stg`     |
| ロールバック     | `gcloud run services update-traffic`                                           | Firebase コンソールでリリースを選ぶ                              |
| prd              | `<サービス名>/v1.2.3` の形のタグの push。承認が必須                            | 未定。`spa/v1.2.3` の形のタグを想定している                      |
| 現在の状態       | stg は稼働中。prd は GCP のプロジェクトが未作成                                | stg は稼働中。prd は GCP のプロジェクトが未作成                  |

**backend は「GCP がソースを取りに行ってビルドする」、frontend は「ビルドした成果物を外から置く」**という違いです。
GitHub Actions が backend に対して行うのは Cloud Build の起動だけで、ビルドもデプロイも GCP の中で完結します。

backend を GCP の中でビルドするのは、イメージの作成に時間がかかり、実行環境を固定したいためです。
frontend は静的ファイルを作るだけで数秒で終わるため、ランナーの上で完結させています。

## 共通する方針

- **stg は `main` への push で自動的に配信されます。** backend はサービスごとの `cd-<サービス名>-stg`、frontend は `cd-frontend-stg` が起動します
- **長期のクレデンシャルをリポジトリと GitHub Secrets に置きません。** GCP への認証は Workload Identity Federation を使い、ワークフローは `id-token: write` で得た OIDC トークンから短命のアクセストークンを受け取ります
- **手元からも同じものを起こせます。** backend はサービスごとのレシピ、frontend は `just deploy-stg` で、CI が止まっていても配信できます
- **`workflow_dispatch` で任意のブランチから配信できます。** 作業中の内容を stg で確かめるためで、手元のレシピと同じ用途です
- **frontend から api への通信は Firebase Hosting の `rewrites` が Cloud Run へ転送します。** 同一オリジンになるため CORS を設定していません。Cloud Run は `allUsers` に公開したままです (Hosting からの転送は内部トラフィック扱いにならないため)。詳細は [frontend のデプロイの構成](./frontend/README.ja.md)にあります
- prd は GCP のプロジェクトが未作成のため、構成の定義だけを持ちます

### `prd` に WIF を置いていません

Workload Identity Federation のプールとプロバイダは `stg` にしかありません。
prd の配信方法が決まっておらず、**誰に何を許すかを決められないためです。**

stg のプロバイダは `assertion.repository` だけを見て、ブランチを問わずこのリポジトリからのトークンを受け入れます。
作業中のブランチから stg を更新できるようにするための意図的な緩さで、**prd に同じ条件を持ち込むことはできません。**
prd を自動化する時点で、タグなのかブランチなのか、承認を挟むのかを決め、それに合わせて `attribute_condition` を書きます。

## 詳細

| ドキュメント                                         | 内容                                                 |
| ---------------------------------------------------- | ---------------------------------------------------- |
| [backend のデプロイの構成](./backend/README.ja.md)   | Cloud Build と Developer Connect、Terraform との分担 |
| [backend の stg へのデプロイ](./backend/stg.ja.md)   | 手順、疎通確認、ロールバック                         |
| [frontend のデプロイの構成](./frontend/README.ja.md) | rewrites とキャッシュ、採らなかった案                |
| [frontend の stg へのデプロイ](./frontend/stg.ja.md) | 手順、疎通確認、ロールバック                         |

規約は `.claude/rules/cd/coding.md` が持ちます。ここには現状と、その形を選んだ理由を置きます。
