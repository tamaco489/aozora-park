#!/usr/bin/env bash
# stg の Cloud Run に backend のサービスをデプロイする
#
# 使い方: ./scripts/deploy-stg.sh <ref> [サービス名...]
#   ref        : ビルドするブランチ・タグ・SHA。どのコミットが stg に出るかを取り違えないよう省略できない
#   サービス名 : 省略すると全サービス。並べるとそれだけを対象にする
#
# 例:
#   ./scripts/deploy-stg.sh main                  # main の全サービス
#   ./scripts/deploy-stg.sh main api              # main の api だけ
#   ./scripts/deploy-stg.sh feature/xxx api       # 作業中のブランチの api だけ
#
# 前提:
#   - gcloud にログインしている
#   - stg-aozora-park で cloudbuild.builds.create と、sa-deployer の actAs がある
#   - 指定した ref が GitHub に push 済みである (ソースは GitHub から取得する)
#
# サービスごとに別のビルドを起こす、1 回のビルドで 2 サービスを作ると cloudbuild.yaml が digest を書く /workspace のパスが衝突する
# 変更の内容から対象を絞る判定は入れない、internal/platform や go.mod を触ると結局どちらも対象になるため既定は全サービスにする
# 片方のビルドが落ちてももう片方は投げる、どちらの結果も分かるよう set -e で打ち切らず最後にまとめて報告する
# 終了コードは 1 つでも落ちていれば 1、すべて成功したときだけ 0 にする
set -euo pipefail

cd "$(dirname "$0")/.."

readonly PROJECT_ID=stg-aozora-park
readonly REGION=asia-northeast1
readonly REPOSITORY_LINK="projects/${PROJECT_ID}/locations/${REGION}/connections/github/gitRepositoryLinks/aozora-park"
readonly DEPLOYER="projects/${PROJECT_ID}/serviceAccounts/sa-deployer@${PROJECT_ID}.iam.gserviceaccount.com"

# cloudbuild.yaml の _SERVICE に渡す値、Cloud Run のサービス名と cmd 配下のディレクトリ名に一致させる
readonly ALL_SERVICES=(api priority-pass-issuer)

if [[ $# -eq 0 ]]; then
  echo "usage: $0 <ref> [service...]  (service を省くと ${ALL_SERVICES[*]})" >&2
  exit 1
fi

ref="$1"
shift

services=("$@")
if [[ ${#services[@]} -eq 0 ]]; then
  services=("${ALL_SERVICES[@]}")
fi

# 打ち間違いをここで弾く、gcloud run deploy は知らない名前を渡されると新しい Cloud Run サービスを作成する
for service in "${services[@]}"; do
  case " ${ALL_SERVICES[*]} " in
    *" ${service} "*) ;;
    *)
      echo "unknown service: ${service}  (available: ${ALL_SERVICES[*]})" >&2
      exit 1
      ;;
  esac
done

# イメージのタグにブランチ名は使えない (スラッシュを含む) ため、リモートの ref を SHA に解決する
sha="$(git rev-parse "origin/${ref}" 2>/dev/null || git rev-parse "${ref}")"

echo "deploy to ${PROJECT_ID}: ref=${ref} sha=${sha} services=${services[*]}"

failed=()

for service in "${services[@]}"; do
  echo "--- submit build: ${service}"

  # ログを表示するため beta を使う (GA の submit は CLOUD_LOGGING_ONLY のログを出さない)
  # 引数と ALL_SERVICES の一覧は .github/workflows/cd-backend-stg.yaml と同じ、片方を変更したらもう片方も直す
  if ! gcloud beta builds submit "${REPOSITORY_LINK}" \
    --project="${PROJECT_ID}" \
    --region="${REGION}" \
    --revision="${sha}" \
    --config=cloudbuild.yaml \
    --service-account="${DEPLOYER}" \
    --substitutions="_SERVICE=${service},_TAG=${sha}"; then
    failed+=("${service}")
  fi
done

if [[ ${#failed[@]} -gt 0 ]]; then
  echo "failed: ${failed[*]}" >&2
  exit 1
fi

echo "deployed: ${services[*]}"
