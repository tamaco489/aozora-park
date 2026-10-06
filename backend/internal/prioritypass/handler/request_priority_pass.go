package handler

import (
	"context"

	"connectrpc.com/connect"

	prioritypassv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1"
	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

func (h *Connect) RequestPriorityPass(ctx context.Context, req *connect.Request[prioritypassv1.RequestPriorityPassRequest]) (*connect.Response[prioritypassv1.RequestPriorityPassResponse], error) {
	msg := req.Msg

	pass, err := h.requestPriorityPass.Do(ctx, prioritypassusecase.RequestPriorityPassInput{
		ParkID:       prioritypassmodel.ParkID(msg.GetParkId()),
		TicketID:     prioritypassmodel.TicketID(msg.GetTicketId()),
		AttractionID: prioritypassmodel.AttractionID(msg.GetAttractionId()),
		TimeSlotID:   prioritypassmodel.TimeSlotID(msg.GetTimeSlotId()),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&prioritypassv1.RequestPriorityPassResponse{
		PriorityPass: toPriorityPassProto(pass),
	}), nil
}
