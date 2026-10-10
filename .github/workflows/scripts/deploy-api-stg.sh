#!/usr/bin/env bash
# stg の Cloud Run の api をデプロイする (cd-api-stg.yaml から呼ぶ)
#
# 使い方: .github/workflows/scripts/deploy-api-stg.sh <sha>
#   sha: ビルドするコミットの SHA。ワークフローは github.sha を渡す
#
# 前提:
#   - google-github-actions/auth で stg-aozora-park への認証が済んでいる
#   - gcloud に beta のコンポーネントが入っている
#
# サービスごとにスクリプトを分ける、デプロイのしかたが変わるのはサービス単位のため他のサービスに影響させない
# 同じ内容が backend/scripts/deploy-api-stg.sh にもある、片方を変更したらもう片方も直す
# 手元は ref を SHA に解決する必要があるが、ここは github.sha がそのまま渡るため解決しない
set -euo pipefail

# cloudbuild.yaml は手元から渡すため、リポジトリの backend を作業ディレクトリにする
cd "$(dirname "$0")/../../../backend"

readonly SERVICE=api
readonly PROJECT_ID=stg-aozora-park
readonly REGION=asia-northeast1
readonly REPOSITORY_LINK="projects/${PROJECT_ID}/locations/${REGION}/connections/github/gitRepositoryLinks/aozora-park"
readonly DEPLOYER="projects/${PROJECT_ID}/serviceAccounts/sa-deployer@${PROJECT_ID}.iam.gserviceaccount.com"

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <sha>" >&2
  exit 1
fi

readonly sha="$1"

echo "deploy ${SERVICE} to ${PROJECT_ID}: sha=${sha}"

# ログを表示するため beta を使う (GA の submit は CLOUD_LOGGING_ONLY のログを出さない)
gcloud beta builds submit "${REPOSITORY_LINK}" \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --revision="${sha}" \
  --config=cloudbuild.yaml \
  --service-account="${DEPLOYER}" \
  --substitutions="_SERVICE=${SERVICE},_TAG=${sha}"
