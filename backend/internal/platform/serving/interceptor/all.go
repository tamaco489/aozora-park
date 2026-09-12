package interceptor

import (
	"log/slog"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
)

// All はすべての RPC に共通で適用するインターセプタ (ハンドラの前後に挟む共通処理) を返す
//
// connect.WithInterceptors は渡した順に実行するため、下の並びがそのまま実行順になる
//
//   - 記録 を先頭に置くのは、検証で弾いた呼び出しもログに残すため
//   - 入力の検証 を末尾に置くのは、ハンドラの直前で入力を検査するため
//   - 検証が返すのは既に connect の形のエラーなので、変換はこれをそのまま通す
func All(logger *slog.Logger) connect.Option {
	return connect.WithInterceptors(
		Log(logger),
		Error(logger),
		validate.NewInterceptor(), // proto に書いた制約で入力を検査 (違反は InvalidArgument になる)
	)
}
