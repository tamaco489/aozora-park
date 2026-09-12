package handler

import (
	"context"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func (h *Connect) GetPark(ctx context.Context, req *connect.Request[parkv1.GetParkRequest]) (*connect.Response[parkv1.GetParkResponse], error) {
	park, err := h.get.Do(ctx, parkmodel.ParkID(req.Msg.GetParkId()))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&parkv1.GetParkResponse{Park: toProto(park)}), nil
}
