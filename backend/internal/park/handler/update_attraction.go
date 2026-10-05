package handler

import (
	"context"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

func (h *Connect) UpdateAttraction(ctx context.Context, req *connect.Request[parkv1.UpdateAttractionRequest]) (*connect.Response[parkv1.UpdateAttractionResponse], error) {
	msg := req.Msg

	attraction, err := h.updateAttraction.Do(ctx, parkusecase.UpdateAttractionInput{
		ParkID:          parkmodel.ParkID(msg.GetParkId()),
		ID:              parkmodel.AttractionID(msg.GetAttractionId()),
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

	return connect.NewResponse(&parkv1.UpdateAttractionResponse{Attraction: toAttractionProto(attraction)}), nil
}
