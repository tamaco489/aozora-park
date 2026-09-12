package handler

import (
	"context"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

func (h *Connect) CreatePark(ctx context.Context, req *connect.Request[parkv1.CreateParkRequest]) (*connect.Response[parkv1.CreateParkResponse], error) {
	msg := req.Msg

	park, err := h.create.Do(ctx, parkusecase.CreateInput{
		Name:                 msg.GetName(),
		DefaultDailyCapacity: msg.GetDefaultDailyCapacity(),
		InventoryDays:        msg.GetInventoryDays(),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&parkv1.CreateParkResponse{Park: toProto(park)}), nil
}
