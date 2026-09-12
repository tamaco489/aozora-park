package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testOrigin = "http://localhost:5173"

// requestHeaders は connect-web がプリフライトで申告するヘッダ
//
// ブラウザは小文字かつ辞書順で送り、rs/cors はその順序を検査するため並びを崩さない
const requestHeaders = "connect-protocol-version,content-type"

// newCORSHandler は許可オリジンを与えて包んだハンドラを返す
func newCORSHandler(tb testing.TB, origins []string) http.Handler {
	tb.Helper()

	return WithCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), origins)
}

func TestWithCORSPreflight(t *testing.T) {
	tests := map[string]struct {
		origins    []string
		origin     string
		wantOrigin string
	}{
		"正常系_許可したオリジンからのプリフライトで許可ヘッダが返ること": {
			origins:    []string{testOrigin},
			origin:     testOrigin,
			wantOrigin: testOrigin,
		},
		"異常系_許可していないオリジンからは許可ヘッダが返らないこと": {
			origins:    []string{testOrigin},
			origin:     "http://evil.example.com",
			wantOrigin: "",
		},
		"境界値_許可オリジンが空なら何も包まないこと": {
			origins:    nil,
			origin:     testOrigin,
			wantOrigin: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/aozorapark.park.v1.ParkService/GetPark", nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			req.Header.Set("Access-Control-Request-Headers", requestHeaders)

			rec := httptest.NewRecorder()
			newCORSHandler(t, tt.origins).ServeHTTP(rec, req)

			got := rec.Header().Get("Access-Control-Allow-Origin")
			if got != tt.wantOrigin {
				t.Errorf("WithCORS(%v) のプリフライト Access-Control-Allow-Origin = %q, want %q", tt.origins, got, tt.wantOrigin)
			}

			if tt.wantOrigin != "" && rec.Code != http.StatusNoContent {
				t.Errorf("WithCORS(%v) のプリフライト = %d, want %d", tt.origins, rec.Code, http.StatusNoContent)
			}
		})
	}
}

func TestWithCORSAllowsConnectHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/aozorapark.park.v1.ParkService/GetPark", nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", requestHeaders)

	rec := httptest.NewRecorder()
	newCORSHandler(t, []string{testOrigin}).ServeHTTP(rec, req)

	// connect-web はこのヘッダを必ず送るため、通らないとブラウザからの呼び出しが全部落ちる
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Errorf("WithCORS() のプリフライト Access-Control-Allow-Headers = %q, want Connect-Protocol-Version を含む値", got)
	}
}

func TestWithCORSPassesThroughWithoutOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/aozorapark.park.v1.ParkService/GetPark", nil)

	rec := httptest.NewRecorder()
	newCORSHandler(t, []string{testOrigin}).ServeHTTP(rec, req)

	// buf curl のように Origin を送らない呼び出しを塞がないこと
	if rec.Code != http.StatusOK {
		t.Errorf("WithCORS() の Origin 無しリクエスト = %d, want %d", rec.Code, http.StatusOK)
	}
}
