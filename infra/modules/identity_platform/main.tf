# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/identity_platform_config
# 運営スタッフが管理画面にログインするための認証基盤として、プロジェクトの Identity Platform を有効化し、ログイン方法を設定する
# WARN: 一度有効化すると削除できず、destroy しても state から外れるだけになる
resource "google_identity_platform_config" "operator_login" {
  project = var.project_id

  sign_in {
    email {
      # メールアドレスでのログインを有効にする
      enabled = true
      # パスワードを必須にする (false にするとパスワードなしのメールリンクでのログインになる)
      password_required = true
    }

    # 有効化すると GCP が無効の値を返すため、書かないと plan のたびに差分が出る
    phone_number {
      enabled = false
    }
  }

  # 運営アカウントは管理者が発行するため、クライアントからの作成と削除を無効にする
  client {
    permissions {
      # 利用者がログイン画面から自分でアカウントを作成できないようにする
      disabled_user_signup = true
      # 利用者が自分のアカウントを削除できないようにする
      disabled_user_deletion = true
    }
  }
}
