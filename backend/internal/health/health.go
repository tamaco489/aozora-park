// Package health は疎通確認の RPC を提供する
package health

import (
	"net/http"

	"connectrpc.com/connect"

	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/health/v1/healthv1connect"
	healthhandler "github.com/tamaco489/aozora-park/backend/internal/health/handler"
)

// NewConnectHandler は HealthService のパスとハンドラを返す
func NewConnectHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return healthv1connect.NewHealthServiceHandler(healthhandler.NewConnect(), opts...)
}
