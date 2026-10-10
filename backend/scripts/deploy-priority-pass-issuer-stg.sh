#!/usr/bin/env bash
# stg の Cloud Run の priority-pass-issuer をデプロイする
#
# 使い方: ./scripts/deploy-priority-pass-issuer-stg.sh <ref>
#   ref: ビルドするブランチ・タグ・SHA。どのコミットが stg に出るかを取り違えないよう省略できない
#
# 例:
#   ./scripts/deploy-priority-pass-issuer-stg.sh main
#   ./scripts/deploy-priority-pass-issuer-stg.sh feature/xxx
#
# 前提:
#   - gcloud にログインしている
#   - stg-aozora-park で cloudbuild.builds.create と、sa-deployer の actAs がある
#   - 指定した ref が GitHub に push 済みである (ソースは GitHub から取得する)
#
# サービスごとにスクリプトを分ける、デプロイのしかたが変わるのはサービス単位のため他のサービスに影響させない
# 同じ内容が .github/workflows/scripts/deploy-priority-pass-issuer-stg.sh にもある、片方を変更したらもう片方も直す
set -euo pipefail

cd "$(dirname "$0")/.."

readonly SERVICE=priority-pass-issuer
readonly PROJECT_ID=stg-aozora-park
readonly REGION=asia-northeast1
readonly REPOSITORY_LINK="projects/${PROJECT_ID}/locations/${REGION}/connections/github/gitRepositoryLinks/aozora-park"
readonly DEPLOYER="projects/${PROJECT_ID}/serviceAccounts/sa-deployer@${PROJECT_ID}.iam.gserviceaccount.com"

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <ref>" >&2
  exit 1
fi

readonly ref="$1"

# イメージのタグにブランチ名は使えない (スラッシュを含む) ため、リモートの ref を SHA に解決する
sha="$(git rev-parse "origin/${ref}" 2>/dev/null || git rev-parse "${ref}")"
readonly sha

echo "deploy ${SERVICE} to ${PROJECT_ID}: ref=${ref} sha=${sha}"

# ログを表示するため beta を使う (GA の submit は CLOUD_LOGGING_ONLY のログを出さない)
gcloud beta builds submit "${REPOSITORY_LINK}" \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --revision="${sha}" \
  --config=cloudbuild.yaml \
  --service-account="${DEPLOYER}" \
  --substitutions="_SERVICE=${SERVICE},_TAG=${sha}"
