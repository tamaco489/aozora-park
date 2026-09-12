package interceptor

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// Log は RPC の呼び出しを 1 件 1 行で記録する
//
// next は 1 つ内側の処理で、いちばん内側はハンドラそのもの
// next の前後に処理を挟んだ関数を返すことで、何段でも重ねられる
// UnaryInterceptorFunc は unary だけを包む型で、ストリーミングは素通しする
func Log(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			res, err := next(ctx, req)

			attrs := []slog.Attr{
				slog.String("procedure", req.Spec().Procedure),
				slog.Duration("elapsed", time.Since(start)),
				slog.String("code", code(err)),
			}

			level := slog.LevelInfo
			if err != nil {
				level = slog.LevelWarn
			}
			logger.LogAttrs(ctx, level, "rpc", attrs...)

			return res, err
		}
	}
}

// code は応答に載る connect のコードを文字列で返す
func code(err error) string {
	if err == nil {
		return "ok"
	}
	// Error が変換した後に呼ばれるので、ここで取れるコードが実際に応答に載る
	if connectErr, ok := errors.AsType[*connect.Error](err); ok {
		return connectErr.Code().String()
	}
	return connect.CodeInternal.String()
}
