package interceptor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

// discardLogger は検証の対象がログではないため出力を捨てる
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestError(t *testing.T) {
	sentinel := apperr.New(apperr.KindConflict, "PARK_SOLD_OUT", "在庫が不足している")

	tests := map[string]struct {
		err         error
		wantCode    connect.Code
		wantMeta    string
		wantMessage string
	}{
		"正常系_apperrの場合_Kindに対応するコードになること": {
			err:      sentinel,
			wantCode: connect.CodeFailedPrecondition,
			wantMeta: "PARK_SOLD_OUT",
		},
		"正常系_apperrをラップした場合_包んだ文脈が応答に出ないこと": {
			err:         fmt.Errorf("create park %s: %w", "park-1", sentinel),
			wantCode:    connect.CodeFailedPrecondition,
			wantMeta:    "PARK_SOLD_OUT",
			wantMessage: sentinel.Error(),
		},
		"正常系_connectのエラーの場合_そのまま返ること": {
			err:      connect.NewError(connect.CodeInvalidArgument, errors.New("validation")),
			wantCode: connect.CodeInvalidArgument,
		},
		"異常系_分類できないエラーの場合_Internalに畳まれること": {
			err:      errors.New("なにかの失敗"),
			wantCode: connect.CodeInternal,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			next := func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
				return nil, tt.err
			}

			_, err := Error(discardLogger())(next)(context.Background(), connect.NewRequest(&parkv1.GetParkRequest{}))

			got, ok := errors.AsType[*connect.Error](err)
			if !ok {
				t.Fatalf("Error()(next)(...) = %v, want *connect.Error", err)
			}
			if got.Code() != tt.wantCode {
				t.Errorf("Error()(next)(%v) のコード = %v, want %v", tt.err, got.Code(), tt.wantCode)
			}
			if tt.wantMessage != "" && got.Message() != tt.wantMessage {
				t.Errorf("Error()(next)(%v) のメッセージ = %q, want %q", tt.err, got.Message(), tt.wantMessage)
			}
			if tt.wantMeta != "" && got.Meta().Get(codeHeader) != tt.wantMeta {
				t.Errorf("Error()(next)(%v) の %s = %q, want %q", tt.err, codeHeader, got.Meta().Get(codeHeader), tt.wantMeta)
			}
		})
	}
}

func TestErrorPassesThroughSuccess(t *testing.T) {
	want := connect.NewResponse(&parkv1.GetParkResponse{})
	next := func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		return want, nil
	}

	got, err := Error(discardLogger())(next)(context.Background(), connect.NewRequest(&parkv1.GetParkRequest{}))
	if err != nil {
		t.Fatalf("Error()(next)(...) = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Error()(next)(...) = %v, want %v", got, want)
	}
}

func TestCode(t *testing.T) {
	tests := map[string]struct {
		err  error
		want string
	}{
		"正常系_エラーが無い場合_okになること": {
			err:  nil,
			want: "ok",
		},
		"正常系_connectのエラーの場合_そのコードになること": {
			err:  connect.NewError(connect.CodeNotFound, errors.New("見つからない")),
			want: "not_found",
		},
		"異常系_connect以外のエラーの場合_internalになること": {
			err:  errors.New("なにかの失敗"),
			want: "internal",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := code(tt.err); got != tt.want {
				t.Errorf("code(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}
