package handler

import (
	"context"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

func (h *Connect) UpdateTicketType(ctx context.Context, req *connect.Request[parkv1.UpdateTicketTypeRequest]) (*connect.Response[parkv1.UpdateTicketTypeResponse], error) {
	msg := req.Msg

	ticketType, err := h.updateTicketType.Do(ctx, parkusecase.UpdateTicketTypeInput{
		ParkID:        parkmodel.ParkID(msg.GetParkId()),
		ID:            parkmodel.TicketTypeID(msg.GetTicketTypeId()),
		Name:          msg.GetName(),
		Price:         msg.GetPrice(),
		EntryTimeFrom: msg.GetEntryTimeFrom(),
		EntryTimeTo:   msg.GetEntryTimeTo(),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&parkv1.UpdateTicketTypeResponse{TicketType: toTicketTypeProto(ticketType)}), nil
}
