// Package interceptor は connect の共通処理を持つ
package interceptor

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

// codeHeader は機械可読なエラーコードを載せる応答ヘッダ
const codeHeader = "Aozora-Error-Code"

// errInternal は想定外の失敗で応答に載せる文言、原因はログにだけ残す
var errInternal = errors.New("internal error")

// Error は apperr のエラーを connect のエラーに変換する
//
// 変換を行うのはここだけにする。ハンドラは connect.NewError を組み立てない
func Error(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, req)
			if err == nil {
				return res, nil
			}

			// 既に connect のエラーはそのまま返す、protovalidate の検証結果がこれにあたる
			if _, ok := errors.AsType[*connect.Error](err); ok {
				return nil, err
			}

			if appErr, ok := errors.AsType[*apperr.Error](err); ok {
				// 包んだ文脈がクライアントに出ないよう、応答にはセンチネルのメッセージだけを載せる
				converted := connect.NewError(appErr.Kind.ConnectCode(), appErr)

				// クライアントが文字列を解析せずに分岐できるようにコードをヘッダで返す
				converted.Meta().Set(codeHeader, appErr.Code)

				// 応答から落とした文脈はここで残す、これが無いとどの ID で失敗したか追えない
				logger.InfoContext(ctx, "業務エラー",
					slog.String("procedure", req.Spec().Procedure),
					slog.String("code", appErr.Code),
					slog.Any("error", err),
				)

				return nil, converted
			}

			logger.ErrorContext(ctx, "未分類のエラー",
				slog.String("procedure", req.Spec().Procedure),
				slog.Any("error", err),
			)

			return nil, connect.NewError(connect.CodeInternal, errInternal)
		}
	}
}
