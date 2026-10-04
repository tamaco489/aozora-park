# frontend のデプロイの構成

[English](./overview.md) | [日本語](./overview.ja.md)

[ドキュメント一覧](../../README.ja.md)に戻る。

frontend を Firebase Hosting へ届ける経路をまとめます。手順は [frontend の stg へのデプロイ](./stg.ja.md)にあります。
backend を含めた全体の俯瞰は[デプロイの構成](../../deploy/overview.ja.md)にあります。
規約は `.claude/rules/ci/coding.md` の「CD」が持ちます。ここには現状と、その形を選んだ理由を置きます。

## デプロイの経路

![frontend のデプロイ経路](./images/flow.png)

**ビルドもデプロイも 1 か所で実行します。** backend と違い Cloud Build を経由しません。
配信するのは `frontend/dist` の静的ファイルだけで、サーバのプロセスはありません。

既定では `cd-frontend-stg` が GitHub Actions のランナーで `npm ci` からビルドと `firebase deploy` までを行います。
手元からの `just deploy-stg` も残しており、**実行する場所が違うだけで手順は同じ**です。

## api を同一オリジンにする

Firebase Hosting の `rewrites` が、RPC のパスを Cloud Run の `api` へ転送します。
ブラウザから見ると api は同じオリジンにあるため、**プリフライト (`OPTIONS`) が発生しません。**

| リクエストのパス   | 転送先                                                 |
| ------------------ | ------------------------------------------------------ |
| `/aozorapark.*/*`  | Cloud Run の `api`                                     |
| `/grpc.health.*/*` | Cloud Run の `api`                                     |
| 上記以外           | `frontend/dist` の静的ファイル。無ければ `/index.html` |

Connect は `Content-Type: application/json` と `Connect-Protocol-Version` を送るため、クロスオリジンでは必ずプリフライトを伴います。
RPC ごとに往復が 1 回増え、`min_instance_count = 0` の Cloud Run ではコールドスタートの原因にもなります。

`rewrites` の照合に `regex` を使っているのは、RPC のパスが `/aozorapark.park.v1.ParkService/GetPark` の形で、1 セグメント目にドットを含むためです。
glob の `**` はパスのセグメントを任意個マッチする記法で、セグメントの途中から使うと挙動が一意に決まりません。

## ローカルも同じ形にする

`frontend/vite.config.ts` の `server.proxy` が、同じ 2 つの接頭辞を `http://localhost:8080` へ転送します。
本番と同じく同一オリジンになるため、ローカルでもプリフライトが出ません。

接続先は `/` に固定しています。環境変数で切り替えません。
全環境で同じ値しか取らないものを設定として持つと、既定で除外されている `.env.*` をコミット対象へ戻すことになり、秘匿値を置ける場所を作ってしまうためです。

## キャッシュ

Firebase Hosting は既定で静的コンテンツに `Cache-Control: max-age=3600` を付けます。
`firebase.json` の `headers` で上書きしています。

| 対象   | 値                  | 理由                                                                                            |
| ------ | ------------------- | ----------------------------------------------------------------------------------------------- |
| すべて | `no-cache`          | デプロイを即座に反映させる。保存はするが使う前に必ず再検証するため、変更が無ければ `304` で済む |
| RPC    | `private, no-store` | CDN に利用者ごとのデータを残さない                                                              |

`index.html` を再検証させるのが要点です。
古い `index.html` が残ると、そこが参照するハッシュ付きのファイルは新しいリリースに存在しないため、画面が壊れます。

資産 (`/assets/**`) に長期キャッシュを付けていません。
`headers` は rewrite **前**のリクエスト URL で照合されるため、存在しない資産への要求が SPA のフォールバックで `index.html` を返すとき、その HTML まで長期間固定されてしまうためです。
入れる場合は、先にフォールバックから `/assets/` を除外する必要があります。

## stg と prd の違い

| 項目       | stg                                                              | prd                                                 |
| ---------- | ---------------------------------------------------------------- | --------------------------------------------------- |
| 自動の起点 | `main` への push (`frontend/**`、`firebase.json`、`.firebaserc`) | 未定。`spa/v1.2.3` の形のタグでの配信を想定している |
| 手で起こす | `workflow_dispatch`、または `just deploy-stg`                    | 未定                                                |
| 実行場所   | GitHub Actions のランナー。手で起こす場合は手元                  | 未定                                                |
| 認証       | Workload Identity Federation (`sa-cd-frontend`)                  | 未定                                                |
| 承認       | 無し                                                             | 未定                                                |
| 現在の状態 | 稼働中                                                           | GCP のプロジェクトが未作成                          |

backend のタグは `api/v1.2.3` の形で、Cloud Build のトリガが `^api/v[0-9]+\.[0-9]+\.[0-9]+$` で拾います。
接頭辞を付けたのは配信の対象が増えたときにトリガを分けるためなので、frontend は `spa/v1.2.3` の形にします。

ただし **Firebase Hosting には Cloud Build のトリガに相当する仕組みがありません。**
タグの push で配信するには GitHub Actions でタグを拾う形になります。stg の自動化でその土台はできましたが、
**prd には Workload Identity Federation を置いていない**ため、配信方法を決めるところからになります (理由は[デプロイの構成](../../deploy/overview.ja.md)にあります)。

## 登場するリソース

| リソース                  | 役割                                                                                                                                         |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Firebase Hosting のサイト | 既定のサイト `stg-aozora-park`。Firebase プロジェクト化で自動的に作成された (マイルストーン 3 で Identity Platform を有効化した際に連鎖した) |
| `firebase.json`           | 公開するディレクトリ、`rewrites`、`headers`。エミュレータの設定と同居する                                                                    |
| `.firebaserc`             | 既定のプロジェクト。`firebase` を直接実行したときの事故を防ぐ                                                                                |
| firebase-tools            | `.tool-versions` で管理する。デプロイにしか使わない                                                                                          |

### CD の SA に要る権限

`sa-cd-frontend` には Hosting の権限だけでは足りません。

| ロール                        | 付与する範囲         | 必要な理由                                                                                                  |
| ----------------------------- | -------------------- | ----------------------------------------------------------------------------------------------------------- |
| `roles/firebasehosting.admin` | プロジェクト         | Hosting はサイト単位の IAM を持たない                                                                       |
| `roles/run.viewer`            | `api` のサービスのみ | **`rewrites` の転送先が実在するかを Hosting がバージョンの確定時に検証する**ため、`run.services.get` が要る |

後者は firebase CLI ではなく **Hosting の API がサーバ側で行う検証**で要求されます。
手元から実行している間は実行者の権限で通っていたため、CD を自動化して初めて表面化しました (#89)。

既定のドメインは 2 つあり、どちらも同じサイトを指します。

| ドメイン                          | 用途                                              |
| --------------------------------- | ------------------------------------------------- |
| `stg-aozora-park.web.app`         | 配信                                              |
| `stg-aozora-park.firebaseapp.com` | Identity Platform が OAuth のリダイレクト先に使う |

## Terraform との分担

Terraform が持つのは API の有効化だけです。

| 対象                                            | 扱う場所                             | 理由                                              |
| ----------------------------------------------- | ------------------------------------ | ------------------------------------------------- |
| `firebase` と `firebasehosting` の API の有効化 | Terraform (`modules/project`)        | `google_project_service` で足りる                 |
| Firebase プロジェクト化                         | 手作業                               | `google_firebase_project` は `google-beta` が要る |
| Hosting の既定サイト                            | 手作業                               | 同上。Firebase プロジェクト化で自動的に作成される |
| Hosting の設定と配信                            | `firebase.json` と `firebase deploy` | 設定が頻繁に変わるため                            |

`google-beta` を入れる理由がこの 2 つだけのため、provider を増やさず手作業に寄せています。
Developer Connect の接続と同じ扱いです。

## 採らなかった案

### CORS を設定して別オリジンのまま使う

api に許可するオリジンを設定し、Hosting から `*.run.app` を直接呼ぶ案です。

- Connect のプリフライトのキャッシュは URL ごとに効くため、`Access-Control-Max-Age` を付けても RPC の種類だけ往復が増える
- 外部 LB と Cloud Armor で同一オリジンにする案は、固定費が月 18 ドル程度かかる

`rewrites` なら追加の費用がかからず、プリフライト自体が無くなります。

### `firebase login:ci` のトークンを使う

`firebase login:ci` が発行するトークンを GitHub Secrets に置けば CI から配信できますが、**採りません。**

- 公式が非推奨としており、`GOOGLE_APPLICATION_CREDENTIALS` (ADC) を使う形が推奨されている
- 失効しない長期クレデンシャルになり、リポジトリに長期クレデンシャルを置かない方針に反する

firebase CLI は google-auth-library の ADC をそのまま使い、**資格情報の型で分岐していません。**
そのため WIF が発行する `external_account` 型の資格情報でも認証が通り、`google-github-actions/auth` が書き出すものをそのまま読みます。

### `FirebaseExtended/action-hosting-deploy` を使う

サービスアカウントの JSON キーが必須で、WIF と ADC に対応していません。
長期クレデンシャルを置かない方針と相容れないため採りません。

### firebase-tools を frontend の devDependency にする

バージョンが `package-lock.json` で固定される利点がありますが、依存が 34 から 700 パッケージに増え、脆弱性が 12 件付き、CI の `npm ci` が遅くなります。
デプロイにしか使わないツールの代償として大きいため、`.tool-versions` での管理にしています。
