package handler

import (
	"context"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

func (h *Connect) UpdatePark(ctx context.Context, req *connect.Request[parkv1.UpdateParkRequest]) (*connect.Response[parkv1.UpdateParkResponse], error) {
	msg := req.Msg

	park, err := h.update.Do(ctx, parkusecase.UpdateInput{
		ID:                   parkmodel.ParkID(msg.GetParkId()),
		Name:                 msg.GetName(),
		DefaultDailyCapacity: msg.GetDefaultDailyCapacity(),
		InventoryDays:        msg.GetInventoryDays(),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&parkv1.UpdateParkResponse{Park: toProto(park)}), nil
}
