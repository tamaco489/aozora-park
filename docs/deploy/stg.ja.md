# stg へのデプロイ

[English](./stg.md) | [日本語](./stg.ja.md)

[ドキュメント一覧](../README.ja.md)に戻る。

`main` へマージした後に、手元から 1 コマンドで stg の api を更新します。
ビルドとデプロイは Cloud Build が GCP の中で実行し、ソースは GitHub から Developer Connect 経由で取得します。

## 前提

- `gcloud` にログインしている

```sh
gcloud auth login
```

- `stg-aozora-park` で次の 2 つができる。プロジェクトのオーナーであれば両方を満たす
  - ビルドの作成 (`cloudbuild.builds.create`)
  - `sa-deployer` を使うこと (`sa-deployer` への `iam.serviceAccountUser`)
- デプロイしたいコミットを GitHub に push 済みである

**手元の変更は反映されません。** ソースは GitHub から取得するため、コミットして push していないものはビルドに含まれません。

## デプロイ

```sh
cd backend
just deploy-stg              # main をデプロイする
just deploy-stg <ref>        # ブランチ・タグ・SHA を指定する
```

`<ref>` は `origin/<ref>` として解決し、無ければそのまま SHA として扱います。
解決した SHA をイメージのタグに使うため、どのコミットが動いているかをイメージから追えます。

処理の流れは次のとおりです。

| 順  | 実行するもの | 内容                                                                        |
| --- | ------------ | --------------------------------------------------------------------------- |
| 1   | 手元         | ref をコミットの SHA に解決し、`gcloud beta builds submit` でビルドを投げる |
| 2   | Cloud Build  | Developer Connect のリンクから、その SHA のソースを取得する                 |
| 3   | Cloud Build  | `backend/Dockerfile` でイメージを組み立て、Artifact Registry に push する   |
| 4   | Cloud Build  | push したイメージをダイジェストで指定し、Cloud Run の `api` にデプロイする  |

ビルドは `sa-deployer` で走ります。ログはコマンドの出力に流れます。
履歴は [Cloud Build の一覧](https://console.cloud.google.com/cloud-build/builds?project=stg-aozora-park)で見られます。リージョンは `asia-northeast1` を選びます。

> [!NOTE]
> ビルド定義 (`backend/cloudbuild.yaml`) は手元のファイルを読みます。
> GitHub から取得するのはソースだけです。`cloudbuild.yaml` を変更したら、その内容でビルドされます。

## 疎通確認

URL は Terraform の出力で確かめます。

```sh
cd infra
terraform -chdir=envs/stg output -raw api_uri
```

ヘルスチェックを呼びます。

```sh
buf curl -d '{"service":""}' <api_uri>/grpc.health.v1.Health/Check
```

`{"status":"SERVING"}` が返ればよいです。

> [!NOTE]
> `buf curl` は `--schema` を渡さない場合、サーバのリフレクションで定義を取得します。
> リフレクションは gRPC のため HTTP/2 が要ります。
> Cloud Run は、クライアントとは HTTP/2 で通信しても、コンテナへ転送するときは既定でリクエストを HTTP/1.1 に変換します。
> そのため Terraform でコンテナのポートに `h2c` と名前を付け、コンテナまで HTTP/2 のまま通しています。

パークの登録と取得を確かめます。**stg の Firestore に書き込みます。** 確認で作成したドキュメントは後で削除してください。

```sh
buf curl -d '{"name":"デプロイ確認","defaultDailyCapacity":100,"inventoryDays":7}' \
  <api_uri>/aozorapark.park.v1.ParkService/CreatePark

buf curl -d '{"parkId":"<作成した park_id>"}' \
  <api_uri>/aozorapark.park.v1.ParkService/GetPark
```

動いているリビジョンとイメージは、次で確かめます。

```sh
gcloud run services describe api --region=asia-northeast1 --project=stg-aozora-park \
  --format='value(status.latestReadyRevisionName,spec.template.spec.containers[0].image)'
```

## ロールバック

前のリビジョンに戻します。Cloud Run はリビジョンを残すため、イメージを作り直す必要はありません。

```sh
gcloud run revisions list --service=api --region=asia-northeast1 --project=stg-aozora-park

gcloud run services update-traffic api \
  --region=asia-northeast1 --project=stg-aozora-park \
  --to-revisions=<戻す先のリビジョン>=100
```

戻した後に、もう一度デプロイすると最新のリビジョンに切り替わります。

```sh
gcloud run services update-traffic api \
  --region=asia-northeast1 --project=stg-aozora-park --to-latest
```

## 仕組みと、prd との違い

| 項目           | stg                                                                  | prd                                     |
| -------------- | -------------------------------------------------------------------- | --------------------------------------- |
| 起こし方       | 手元からの `just deploy-stg`                                         | `api/v1.2.3` の形のタグの push          |
| トリガ         | 作らない。Developer Connect のリポジトリは手動のトリガを作れないため | Cloud Build のトリガ (Terraform で定義) |
| 承認           | 無し                                                                 | 必須。タグの push だけでは走らない      |
| イメージのタグ | コミットの SHA                                                       | コミットの SHA                          |

- Terraform は Cloud Run の `image` を `ignore_changes` で無視します。デプロイでの差し替えを drift にしないためです。
- GitHub Actions は検査だけを行い、デプロイには関わりません。長期のクレデンシャルをリポジトリに置かない構成です。
- 接続 (Developer Connect) の作成手順は、設計ドキュメントの「5. Developer Connect の接続」にあります。
