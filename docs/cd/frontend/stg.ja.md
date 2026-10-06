# frontend の stg へのデプロイ

[English](./stg.md) | [日本語](./stg.ja.md)

[ドキュメント一覧](../../README.ja.md)に戻る。

手元から 1 コマンドで stg の frontend を更新します。構成の全体像は [frontend のデプロイの構成](./README.ja.md)にあります。
**ビルドもデプロイも手元で実行します。** backend と違い Cloud Build を経由しません。

## 前提

- firebase-tools が入っている。`.tool-versions` で管理している

```sh
asdf plugin add firebase
asdf install firebase 15.32.1
firebase --version
```

> [!NOTE]
> `.tool-versions` はリポジトリの直下にあります。**リポジトリの外では `firebase` コマンドが解決できません。**

- firebase CLI にログインしている。`gcloud` の ADC とは別の認証です

```sh
firebase login
firebase projects:list
```

`stg-aozora-park` が `(current)` 付きで一覧に出れば準備完了です。

- `stg-aozora-park` の Firebase Hosting に対する権限がある。プロジェクトのオーナーであれば満たします

## デプロイ

```sh
cd frontend
just deploy-stg
```

`build` に依存しているため、`npm run build` から自動で走ります。**古い `dist` がそのまま配信されることはありません。**

処理の流れは次のとおりです。

| 順  | 実行するもの | 内容                                                                  |
| --- | ------------ | --------------------------------------------------------------------- |
| 1   | 手元         | `tsc -b` で型を検査し、`vite build` で `frontend/dist` を作成する     |
| 2   | 手元         | `firebase deploy --only hosting --project stg-aozora-park` を実行する |
| 3   | Hosting      | ファイルをアップロードし、バージョンを確定してリリースする            |

**手元の作業ツリーがそのまま配信されます。** backend と違い GitHub からソースを取得しないため、コミットしていない変更も反映されます。

`--project` を明示しているのは、`.firebaserc` の既定が変わったときに `deploy-stg` という名前のまま別のプロジェクトへ配信されるのを防ぐためです。

## 疎通確認

画面を開きます。

```sh
open https://stg-aozora-park.web.app
```

> [!NOTE]
> `index.html` は `no-cache` で配信されるため、**ハードリロードをしなくても新しい画面が出ます。**
> 出ない場合はデプロイが完了していないか、`firebase.json` の `headers` が効いていません。

api への転送を確かめます。

```sh
curl -s -X POST https://stg-aozora-park.web.app/grpc.health.v1.Health/Check \
  -H "Content-Type: application/json" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"service":"aozorapark.park.v1.ParkService"}'
```

`{"status":"SERVING_STATUS_SERVING"}` が返ればよいです。
Firebase Hosting の `rewrites` が Cloud Run まで転送できていることの確認になります。

プリフライトが出ていないことは、開発者ツールの Network で確かめます。
**`OPTIONS` が 1 件も無い**のが正しい状態です。

配信中のファイルと `Cache-Control` を確かめます。

```sh
curl -sD - -o /dev/null https://stg-aozora-park.web.app/ | grep -i cache-control
```

`no-cache` が返ればよいです。

## ロールバック

Firebase Hosting はリリースの履歴を保持するため、前のバージョンに戻せます。
**手元で再ビルドする必要はありません。**

Firebase コンソールの Hosting の画面でリリースの一覧を開き、戻したいバージョンの「ロールバック」を選びます。

```sh
open https://console.firebase.google.com/project/stg-aozora-park/hosting/sites
```

`index.html` が `no-cache` のため、ロールバックも次のアクセスで反映されます。

## 仕組みと、backend との違い

| 項目             | frontend                     | backend                              |
| ---------------- | ---------------------------- | ------------------------------------ |
| ビルドの実行場所 | 手元 (CI からはランナー)     | Cloud Build (GCP の中)               |
| ソースの取得元   | 手元の作業ツリー             | GitHub (Developer Connect 経由)      |
| 未コミットの変更 | **反映される**               | 反映されない                         |
| デプロイ先       | Firebase Hosting             | Cloud Run                            |
| ロールバック     | コンソールからリリースを選ぶ | `gcloud run services update-traffic` |
| prd              | 未定                         | `api/v1.2.3` の形のタグの push       |

- ここに書いた手順は手元から起こす場合のものです。`main` に入った変更は `cd-frontend-stg` が自動で配信するため、通常は実行する必要がありません
- CI から配信する場合はランナーの上でビルドするため、**未コミットの変更は反映されません。** 手元から起こしたときだけ作業ツリーがそのまま配信されます
- `firebase.json` はエミュレータの設定と同居しています。`hosting` を足した影響で Hosting エミュレータが起動したため、`docker/firebase-emulator/Dockerfile` で `--only firestore` に限定しています
