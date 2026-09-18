# GitHub との接続は認可をブラウザで行う必要があり、Terraform では作成できないため gcloud で手作業で作成した
# 手作業で作成したリソースは state に記録されないため、定義だけを書くと Terraform は未作成とみなして新規に作成しようとし、同名の接続が既にあるため apply が失敗する
# この import で GCP 上の接続を cloud_build モジュールの定義に対応づけて state に記録し、作り直さずに Terraform の管理下に置く
# 作り直すと GitHub の認可が失われ、ブラウザでの認可をやり直す必要がある
# plan で取り込みと定義との差分を確かめてから apply し、取り込んだ後はこのブロックは何もしない
import {
  to = module.cloud_build.google_developer_connect_connection.github
  id = "projects/stg-aozora-park/locations/asia-northeast1/connections/github"
}
