# stg の Firestore への接続

[English](./stg.md) | [日本語](./stg.ja.md)

[ドキュメント一覧](../../README.ja.md)に戻る。

ローカルの api を、エミュレータではなく GCP の `stg-aozora-park` の Firestore に向けて起動します。
Firestore の `(default)` データベースは Terraform (`infra/envs/stg`) で作成済みです。

> [!WARNING]
> 実際のデータベースに書き込みます。動作を確かめるために作成したドキュメントは、確認後に削除してください。

## 前提

- `stg-aozora-park` で Firestore を読み書きできる権限がある (プロジェクトのオーナーなど)
- ADC (Application Default Credentials) を用意している

```sh
gcloud auth application-default print-access-token > /dev/null && echo ADC OK
```

`ADC OK` と表示されなければ、ログインします。

```sh
gcloud auth application-default login
```

## 起動

```sh
cd backend
just run-api-stg
```

`just run-api` との違いは次のとおりです。

| 項目                         | `just run-api`     | `just run-api-stg` |
| ---------------------------- | ------------------ | ------------------ |
| 接続先                       | エミュレータ       | stg の Firestore   |
| `GOOGLE_CLOUD_PROJECT`       | `demo-aozora-park` | `stg-aozora-park`  |
| `FIRESTORE_EMULATOR_HOST`    | `localhost:18080`  | 設定しない         |
| `GOOGLE_CLOUD_QUOTA_PROJECT` | 設定しない         | `stg-aozora-park`  |

- `FIRESTORE_EMULATOR_HOST` がシェルに残っていると、SDK はエミュレータに繋ぎます。レシピはこの変数を外してから起動します。
- API の課金先 (quota project) は、ADC に記録された値ではなく `stg-aozora-park` に固定します。
- 待ち受けるポートと、許可するオリジン (`http://localhost:5173`) は `just run-api` と同じです。そのため、frontend の `just dev` からそのまま呼べます。

## 確認

api を起動した状態で、画面か [`backend/tools/http/park.http`](../../../backend/tools/http/park.http) から Park を登録・参照・更新します。
`.http` は VS Code の拡張機能 REST Client (`humao.rest-client`) で開き、リクエストの上に出る「Send Request」で実行します。
取得と更新は、直前の登録で返ったパークを対象にするため、先に登録を実行します。

書き込んだドキュメントは、[Firestore Studio](https://console.cloud.google.com/firestore/databases/-default-/data/panel?project=stg-aozora-park) の `parks` コレクションで見られます。
