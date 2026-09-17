#!/usr/bin/env bash
# docs/api/openapi.yaml から Redoc の HTML を docs/api/redoc.html に生成する
# proto を変更したら just generate の後に流し、生成物と一緒にコミットする
# 前提: Node (npx) が使えること

set -euo pipefail

cd "$(dirname "$0")/.."

out=../docs/api/redoc.html

# 既定では実行のたびに Redocly へ利用状況を送るため止める
REDOCLY_TELEMETRY=off npx -y @redocly/cli@2.53.2 build-docs ../docs/api/openapi.yaml -o "$out"

# Redoc 本体の CSS にある、ブラウザが無視する記述を削除する (エディタの警告が生成物に出続けるため)
#   - 接頭辞のない font-smoothing と tap-highlight-color は存在しないプロパティ
#   - ドットが 2 つ続くセレクタは不正で、規則ごと適用されない
# CLI は CSS を 1 行 1 規則に圧縮して出すため、その形に合わせて削除する
sed -i.bak -E \
  -e 's/([;{])(font-smoothing|tap-highlight-color):[^;}]*;/\1/g' \
  -e '/^[^{}]*\.\.[A-Za-z0-9_-]+[^{}]*\{[^{}]*\}(\/\*!sc\*\/)?[[:space:]]*$/d' \
  "$out"
rm -f "$out.bak"
