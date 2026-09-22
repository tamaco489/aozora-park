#!/usr/bin/env bash
# stg の Cloud Run に api をデプロイする
#
# 使い方: ./scripts/deploy-stg.sh [ref]
#   ref: ビルドするブランチ・タグ・SHA (既定は main)
#
# 前提:
#   - gcloud にログインしている
#   - stg-aozora-park で cloudbuild.builds.create と、sa-deployer の actAs がある
#   - 指定した ref が GitHub に push 済みである (ソースは GitHub から取得する)
set -euo pipefail

cd "$(dirname "$0")/.."

readonly PROJECT_ID=stg-aozora-park
readonly REGION=asia-northeast1
readonly REPOSITORY_LINK="projects/${PROJECT_ID}/locations/${REGION}/connections/github/gitRepositoryLinks/aozora-park"
readonly DEPLOYER="projects/${PROJECT_ID}/serviceAccounts/sa-deployer@${PROJECT_ID}.iam.gserviceaccount.com"

ref="${1:-main}"

# イメージのタグにブランチ名は使えない (スラッシュを含む) ため、リモートの ref を SHA に解決する
sha="$(git rev-parse "origin/${ref}" 2>/dev/null || git rev-parse "${ref}")"

echo "deploy to ${PROJECT_ID}: ref=${ref} sha=${sha}"

# ログを表示するため beta を使う (GA の submit は CLOUD_LOGGING_ONLY のログを出さない)
gcloud beta builds submit "${REPOSITORY_LINK}" \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --revision="${sha}" \
  --config=cloudbuild.yaml \
  --service-account="${DEPLOYER}" \
  --substitutions="_TAG=${sha}"
