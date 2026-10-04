# ドキュメント一覧

[English](./README.md) | [日本語](./README.ja.md)

このディレクトリにはプロジェクトのドキュメントが含まれています。

| ドキュメント                                                  | 内容                                                       |
| ------------------------------------------------------------- | ---------------------------------------------------------- |
| [backend のパッケージ構成](./backend/packages/overview.ja.md) | 層と依存の向き、判断の記録                                 |
| [Firestore エミュレータ](./backend/firestore/emulator.ja.md)  | ローカルでの起動と接続                                     |
| [stg の Firestore への接続](./backend/firestore/stg.ja.md)    | ローカルの api を stg の Firestore に向ける                |
| [API 仕様 (OpenAPI)](./api/openapi.yaml)                      | proto から生成した仕様、手で編集しない                     |
| [API 仕様 (Redoc)](./api/redoc.html)                          | 上の仕様の HTML 形式、手元のブラウザで開いて見る           |
| [デプロイの構成](./cd/overview.ja.md)                         | backend と frontend の 2 つの経路の俯瞰と対比              |
| [Workload Identity Federation](./cd/wif/overview.ja.md)       | 鍵を置かずに GitHub Actions から Google Cloud へ入る仕組み |
| [backend のデプロイの構成](./backend/deploy/overview.ja.md)   | Cloud Build と Developer Connect、Terraform との分担       |
| [backend の stg へのデプロイ](./backend/deploy/stg.ja.md)     | 手元から stg の api を更新する、疎通確認とロールバック     |
| [frontend のデプロイの構成](./frontend/deploy/overview.ja.md) | rewrites による同一オリジン化とキャッシュ、採らなかった案  |
| [frontend の stg へのデプロイ](./frontend/deploy/stg.ja.md)   | 手元から stg の画面を更新する、疎通確認とロールバック      |
