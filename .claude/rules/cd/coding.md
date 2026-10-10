# CD のコーディング規約

ディレクトリは `.github/workflows/`。GitHub Actions で stg への配信を行う。

**検査 (`ci-*`) の規約は `.claude/rules/ci/coding.md` が持つ。** ワークフローの書き方のうち CD に固有でないもの
(バージョンの固定、`${{ }}` を `env` 経由で渡すこと、`working-directory` で対象のディレクトリに入ること、ステップの `name` を英語にすること) は、そちらに従う。

経路と、その形を選んだ理由は `docs/cd/README.ja.md` が持つ。認証の仕組みは `docs/cd/wif/README.ja.md` が持つ。

## 方針

**stg は `main` への push で GitHub Actions から配信する。** 手元からの `just deploy-stg` も残し、どちらからでも同じものが入る形にする。

- **GCP への認証は Workload Identity Federation で行う。** 長期クレデンシャルを GitHub Secrets に置かない
- **`prd` には WIF を置いていない。** 配信方法が未定で、`attribute_condition` の絞り方を決められないため

## 構成

### ファイルの分け方

- デプロイは `cd-<対象>-<環境>.yaml` にする。`labels.md` のラベルと同じ語を使う
- **ファイル名に環境を入れる。** 配信先が 1 つに決め打ちで、環境が増えたときに分かれる単位がファイルになるため
- **対象は配信する単位にする。** Cloud Run は service 1 つが 1 単位のため `cd-api-stg.yaml` `cd-priority-pass-issuer-stg.yaml` のようにサービス名を使い、1 つにまとまっている frontend は `cd-frontend-stg.yaml` にする
- **1 ファイル 1 対象**にする。1 つのワークフローで複数のサービスを配信しない。片方の失敗が他方を巻き込まず、手で起こすときに対象を選べる
- `name` はファイル名から拡張子を外したものにする
- **ジョブ名は `<対象>-deploy-<環境>` にする** (`api-deploy-stg`)。検査のジョブと同じ名前空間にあるため、重複させない
- **デプロイのコマンドはワークフローに直接書かず、`.github/workflows/scripts/<ワークフロー名から cd- を外した名前>.sh` に置いて呼ぶ。** 引数が長く、対象ごとに同じ形が並ぶため

## トリガ

```yaml
on:
  push:
    branches: [main]
    paths:
      - "<対象のディレクトリ>/**"
      - ".github/workflows/<このファイル>"
      - ".github/workflows/scripts/<呼ぶスクリプト>"
  workflow_dispatch:
```

- `main` への push で自動的に配信する。作業ブランチへの push では動かさない
- `paths` には対象のディレクトリと**そのワークフロー自身**、そこから**呼ぶスクリプト**を入れる。配信に影響する設定ファイルが対象の外にある場合は、それも入れる
- **サービスごとにワークフローを分けても `paths` は絞らない。** 変更の内容から対象のサービスを判定すると、共通のコードを触ったときに片方が古いまま残る
- **`workflow_dispatch` を併記し、任意のブランチから手でも起こせるようにする。** 作業中のブランチの内容を stg で確かめるため
- **`workflow_dispatch` はファイルが既定ブランチに無いと選択肢に出ない。** 追加した直後は `main` へ入るまで使えない

## 同時実行

```yaml
concurrency:
  group: <ワークフロー名>
  cancel-in-progress: false
```

**打ち切らずに直列化する。** 検査と違い、ジョブを止めても GCP 側の処理は止まらない。
打ち切ると古い成果物が後から反映されうるため、順序を保つほうを取る。

## 権限

```yaml
permissions:
  contents: read
  id-token: write
```

- **この 2 つだけにする。** 既定に任せない
- **`id-token: write` は GitHub の OIDC トークンを受け取るためのもので、GCP の権限ではない。** 誰を通すかは GCP 側のプロバイダと SA が決める
- **WIF の provider 名と SA のメールアドレスは秘匿値ではない。** `env` に直書きし、GitHub Secrets に入れない
- サードパーティのアクションを追加するときは PR に理由を書く

## backend (Cloud Run)

- Cloud Build が GitHub のソースを取得してビルドし、Cloud Run を更新する。**GitHub Actions が行うのは Cloud Build の起動だけ**で、ビルドとデプロイは GCP の中で完結する
- stg はサービスごとのワークフロー (`cd-api-stg.yaml` など) が、同名のスクリプトを呼んで `gcloud beta builds submit` を実行する。ビルドする ref は `github.sha` を使う
- 手元からは `just deploy-<サービス名>-stg <ref>` で同じことができる。Developer Connect のリポジトリは手動のトリガを作れないため、どちらも `gcloud builds submit` を使う
- **`<ref>` に既定値を置かない。** どのコミットが stg に出るかを取り違えないよう、手元から起こすときは必ず書かせる
- **1 サービス 1 スクリプトにする。** デプロイのしかたが変わるのはサービス単位 (service と job で使うコマンドが違う) のため、他のサービスを巻き込まずに変更できる形にする
- **同じ内容のスクリプトが `.github/workflows/scripts/` と `backend/scripts/` の 2 か所にある。片方を変更したらもう片方も直す** (両方にコメントを残している)。CI 側から `backend/scripts/` を呼ばないのは、手元は ref を SHA に解決するのに対し CI は `github.sha` をそのまま使うため
- **複数のサービスをまとめて配信するレシピとワークフローを作らない。** `main` へのマージでは、サービスごとのワークフローがそれぞれ起動して結果的に全サービスが入る
- prd は `<サービス名>/v1.2.3` の形のタグの push で起動するトリガから実行する。トリガはサービスごとに作成し、承認を必須にする
- ビルド定義は対象のディレクトリに置く (`backend/cloudbuild.yaml`)。イメージのタグにはコミットの SHA を使う
- ビルドは `sa-deployer` で走らせる。ユーザー指定の SA ではログの保存先を選べないため `logging: CLOUD_LOGGING_ONLY` を指定する

## frontend (Firebase Hosting)

- **ビルドもデプロイも Cloud Build を経由しない。** `npm ci` からビルドと `firebase deploy` までを 1 か所で実行する
- stg は `cd-frontend-stg.yaml` が実行する。`paths` には `frontend/**` に加えて `firebase.json` と `.firebaserc` を入れる。`rewrites` と `headers` の変更は frontend のファイルを触らずに起きるため
- 手元からは `just deploy-stg` で同じことができる。`build` に依存させ、古い `dist` を配信しない
- **`firebase deploy` の認証は ADC で行う。** 非推奨の `firebase login:ci` のトークンを使わない。firebase CLI は google-auth-library の ADC をそのまま使うため、WIF の資格情報でも通る
- `firebase deploy` には `--project` を明示する。`.firebaserc` の既定が変わっても配信先が動かないようにする
- firebase-tools は `.tool-versions` で管理する。frontend の devDependency にしない
- prd のタグは `spa/v1.2.3` の形にする。backend の `api/v1.2.3` と接頭辞を揃える
- **prd の配信方法はまだ決めていない**

## api を同一オリジンにする

- **CORS を設定せず、Firebase Hosting の `rewrites` で Cloud Run へ転送する**
- `rewrites` の照合には `regex` を使う。glob の `**` は RPC のパスで挙動が一意に決まらない
- ローカルも `frontend/vite.config.ts` の `server.proxy` で同一オリジンにする
- **`firebase.json` の `headers` で `Cache-Control` を指定する。** 既定のままではデプロイが反映されない

## 実行

- **ワークフローの手動実行と、起動条件を広げる変更の push は Claude が行わない** (`CLAUDE.md`)。実行する主体が GitHub Actions でも stg への実デプロイになる
- **`firebase.json` `backend/cloudbuild.yaml` デプロイのレシピを変更したら、`docs/cd/` 配下を更新する。** 手順書が古いまま実行されると事故につながる
