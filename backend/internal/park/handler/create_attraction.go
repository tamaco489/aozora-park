package handler

import (
	"context"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

func (h *Connect) CreateAttraction(ctx context.Context, req *connect.Request[parkv1.CreateAttractionRequest]) (*connect.Response[parkv1.CreateAttractionResponse], error) {
	msg := req.Msg

	attraction, err := h.createAttraction.Do(ctx, parkusecase.CreateAttractionInput{
		ParkID:          parkmodel.ParkID(msg.GetParkId()),
		Name:            msg.GetName(),
		Enabled:         msg.GetEnabled(),
		StartTime:       msg.GetStartTime(),
		EndTime:         msg.GetEndTime(),
		IntervalMinutes: msg.GetIntervalMinutes(),
		CapacityPerSlot: msg.GetCapacityPerSlot(),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&parkv1.CreateAttractionResponse{Attraction: toAttractionProto(attraction)}), nil
}
