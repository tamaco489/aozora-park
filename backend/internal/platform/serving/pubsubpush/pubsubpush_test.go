package pubsubpush

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

// publishedAt は publishTime の検証に使う固定の時刻
var publishedAt = time.Date(
	2026, 4, 1,
	10, 30, 0, 0,
	time.UTC,
)

func TestNewHandlerDecodesEnvelope(t *testing.T) {
	tests := map[string]struct {
		body string
		want *Message
	}{
		"正常系_DLQを経由していない場合_deliveryAttemptが0になること": {
			body: envelopeHelper(`"messageId": "1518", "data": "eyJwYXNzSWQiOiJwLTEifQ==", "attributes": {"passId": "p-1"}, "publishTime": "2026-04-01T10:30:00Z"`, ""),
			want: &Message{
				ID:              "1518",
				Data:            []byte(`{"passId":"p-1"}`),
				Attributes:      map[string]string{"passId": "p-1"},
				PublishTime:     publishedAt,
				DeliveryAttempt: 0,
				Subscription:    "projects/demo/subscriptions/allocate",
			},
		},
		"正常系_DLQを経由した場合_deliveryAttemptが入ること": {
			body: envelopeHelper(`"messageId": "1519", "data": "eyJwYXNzSWQiOiJwLTIifQ=="`, `"deliveryAttempt": 3,`),
			want: &Message{
				ID:              "1519",
				Data:            []byte(`{"passId":"p-2"}`),
				DeliveryAttempt: 3,
				Subscription:    "projects/demo/subscriptions/allocate",
			},
		},
		"境界値_本文が空の場合_エンベロープとして受け付けること": {
			body: envelopeHelper(`"messageId": "1520"`, ""),
			want: &Message{
				ID:           "1520",
				Subscription: "projects/demo/subscriptions/allocate",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var got *Message
			handler := NewHandler(discardLoggerHelper(), func(_ context.Context, msg *Message) error {
				got = msg

				return nil
			})

			res := serveHelper(
				t,
				handler,
				http.MethodPost,
				tt.body,
			)
			if res.Code != http.StatusNoContent {
				t.Errorf("NewHandler(...).ServeHTTP(%s) のコード = %d, want %d",
					tt.body,
					res.Code,
					http.StatusNoContent,
				)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewHandler(...).ServeHTTP(%s) が渡した Message の差分 (-want +got):\n%s",
					tt.body,
					diff,
				)
			}
		})
	}
}

func TestNewHandlerAcksWhatCannotBeRetried(t *testing.T) {
	tests := map[string]struct {
		body    string
		handler Handler
		called  bool // called はハンドラまで届くべきかどうか
	}{
		"異常系_JSONが壊れている場合_ハンドラを呼ばずackすること": {
			body:    `{"message": {`,
			handler: func(context.Context, *Message) error { return nil },
			called:  false,
		},
		"異常系_messageが無い場合_ハンドラを呼ばずackすること": {
			body:    `{"subscription": "projects/demo/subscriptions/allocate"}`,
			handler: func(context.Context, *Message) error { return nil },
			called:  false,
		},
		"異常系_dataが壊れたbase64の場合_ハンドラを呼ばずackすること": {
			body:    envelopeHelper(`"messageId": "1521", "data": "!!!notbase64!!!"`, ""),
			handler: func(context.Context, *Message) error { return nil },
			called:  false,
		},
		"異常系_再実行で直らない失敗の場合_ackすること": {
			body: envelopeHelper(`"messageId": "1522"`, ""),
			handler: func(context.Context, *Message) error {
				return apperr.New(
					apperr.KindConflict,
					"TEST_CONFLICT",
					"やり直しても同じ",
				)
			},
			called: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			called := false
			handler := NewHandler(discardLoggerHelper(), func(ctx context.Context, msg *Message) error {
				called = true

				return tt.handler(ctx, msg)
			})

			res := serveHelper(
				t,
				handler,
				http.MethodPost,
				tt.body,
			)
			if res.Code != http.StatusNoContent {
				t.Errorf("NewHandler(...).ServeHTTP(%s) のコード = %d, want %d",
					tt.body,
					res.Code,
					http.StatusNoContent,
				)
			}

			if called != tt.called {
				t.Errorf("NewHandler(...).ServeHTTP(%s) のハンドラの呼び出し = %t, want %t",
					tt.body,
					called,
					tt.called,
				)
			}
		})
	}
}

func TestNewHandlerRetriesWhatCanSucceedLater(t *testing.T) {
	tests := map[string]struct {
		err error
	}{
		"正常系_再実行で直りうるセンチネルの場合_再配信させること": {
			err: apperr.NewRetryable(
				apperr.KindNotFound,
				"TEST_NOT_READY",
				"まだ揃っていない",
			),
		},
		"正常系_分類していないエラーの場合_再配信させること": {
			err: errors.New("firestore unavailable"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(discardLoggerHelper(), func(context.Context, *Message) error {
				return tt.err
			})

			body := envelopeHelper(`"messageId": "1523"`, "")

			res := serveHelper(
				t,
				handler,
				http.MethodPost,
				body,
			)
			if res.Code != http.StatusInternalServerError {
				t.Errorf("NewHandler(...).ServeHTTP(%v) のコード = %d, want %d",
					tt.err,
					res.Code,
					http.StatusInternalServerError,
				)
			}
		})
	}
}

// TestNewHandlerRejectsOtherMethods は push 以外からの呼び出しを弾くことを確かめる
func TestNewHandlerRejectsOtherMethods(t *testing.T) {
	called := false
	handler := NewHandler(discardLoggerHelper(), func(context.Context, *Message) error {
		called = true

		return nil
	})

	res := serveHelper(
		t,
		handler,
		http.MethodGet,
		"",
	)
	if res.Code != http.StatusMethodNotAllowed {
		t.Errorf("NewHandler(...).ServeHTTP(GET) のコード = %d, want %d",
			res.Code,
			http.StatusMethodNotAllowed,
		)
	}

	if called {
		t.Errorf("NewHandler(...).ServeHTTP(GET) のハンドラの呼び出し = %t, want %t",
			called,
			false,
		)
	}
}

// FuzzDecode は外から任意の本文を渡されても落ちないことを確かめる
//
// push の受け口はプロセスの外から任意のバイト列を受け取る境界になる
func FuzzDecode(f *testing.F) {
	f.Add(envelopeHelper(`"messageId": "1518", "data": "eyJhIjoxfQ=="`, ""))
	f.Add(`{"message": {`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(envelopeHelper(`"messageId": "1518", "data": "!!!"`, ""))
	f.Add(`{"message": {"messageId": 1518}}`)

	f.Fuzz(func(t *testing.T, body string) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/pubsub/push",
			strings.NewReader(body),
		)

		msg, err := decode(req)
		if err != nil {
			return
		}

		// エラーを返さなかったものは機能側へ渡るため、識別子が空のまま通していないことを確かめる
		if msg.ID == "" {
			t.Errorf("decode(%q) の Message.ID = %q, want 空でないこと",
				body,
				msg.ID,
			)
		}
	})
}

// serveHelper はハンドラに 1 回リクエストを渡す
func serveHelper(
	tb testing.TB,
	h http.Handler,
	method string,
	body string,
) *httptest.ResponseRecorder {
	tb.Helper()

	req := httptest.NewRequest(
		method,
		"/pubsub/push",
		strings.NewReader(body),
	)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	return res
}

// envelopeHelper は push が送る形の JSON を組み立てる
//
// 項目名を 1 か所に集め、ケースごとに message の中身と deliveryAttempt だけを差し替える
func envelopeHelper(message, deliveryAttempt string) string {
	return `{"message": {` + message + `}, ` + deliveryAttempt + `"subscription": "projects/demo/subscriptions/allocate"}`
}

// discardLoggerHelper は出力を捨てるロガーを返す
func discardLoggerHelper() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
