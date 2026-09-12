default:
    @just --list

# エミュレータを起動する、healthy になるまで待つ
up:
    docker compose up -d --wait

# エミュレータを停止する
down:
    docker compose down

# ログを追う
logs:
    docker compose logs -f

# 起動状態と健全性を見る
ps:
    docker compose ps

# イメージを組み立て直す
build:
    docker compose build
