#!/usr/bin/env bash
# proto から Go と TypeScript を生成し、出力先のファイルを一覧する
# どこから呼ばれても proto/ を基準に動く

set -euo pipefail

cd "$(dirname "$0")/.."

buf generate

echo
echo "Generated files:"
find ../backend/gen ../frontend/src/gen -type f | LC_ALL=C sort | sed 's|^\.\./|  |'
