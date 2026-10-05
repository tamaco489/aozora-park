package handler

import (
	"context"

	"connectrpc.com/connect"

	inventoryv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1"
	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryusecase "github.com/tamaco489/aozora-park/backend/internal/inventory/usecase"
)

func (h *Connect) ListTimeSlots(ctx context.Context, req *connect.Request[inventoryv1.ListTimeSlotsRequest]) (*connect.Response[inventoryv1.ListTimeSlotsResponse], error) {
	msg := req.Msg

	slots, err := h.listTimeSlots.Do(ctx, inventoryusecase.ListTimeSlotsInput{
		ParkID:       inventorymodel.ParkID(msg.GetParkId()),
		AttractionID: inventorymodel.AttractionID(msg.GetAttractionId()),
		Date:         inventorymodel.Date(msg.GetDate()),
	})
	if err != nil {
		return nil, err
	}

	res := &inventoryv1.ListTimeSlotsResponse{
		TimeSlots: make(
			[]*inventoryv1.TimeSlot,
			0,
			len(slots),
		),
	}
	for _, slot := range slots {
		res.TimeSlots = append(res.TimeSlots, toTimeSlotProto(slot))
	}

	return connect.NewResponse(res), nil
}
