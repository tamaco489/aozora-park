package httpx

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

// preflightMaxAge はプリフライトの結果をブラウザが保持してよい秒数
//
// Chromium 系の上限が 2 時間のため、それを超える値を指定しても短く丸められる
const preflightMaxAge = 2 * 60 * 60

// WithCORS はブラウザからの呼び出しを許可するオリジンに限って通す
//
// 許可するメソッドとヘッダは connectrpc.com/cors から取る、connect のバージョンが上がったときに手で追随しなくて済む
// origins が空なら何も包まない、設定し忘れたまま全オリジンを通す状態を作らないため
func WithCORS(h http.Handler, origins []string) http.Handler {
	if len(origins) == 0 {
		return h
	}

	return cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         preflightMaxAge,
	}).Handler(h)
}
