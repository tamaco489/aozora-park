// Package handler は health の入口を持つ
package handler

import (
	"context"

	"connectrpc.com/connect"

	healthv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/health/v1"
	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/health/v1/healthv1connect"
)

// Connect は HealthService の connect ハンドラ
type Connect struct{}

var _ healthv1connect.HealthServiceHandler = (*Connect)(nil)

func NewConnect() *Connect {
	return &Connect{}
}

// Check は依存先を持たないため常に SERVING を返す
func (c *Connect) Check(context.Context, *connect.Request[healthv1.CheckRequest]) (*connect.Response[healthv1.CheckResponse], error) {
	return connect.NewResponse(&healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_SERVING,
	}), nil
}
