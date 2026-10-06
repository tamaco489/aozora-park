package handler

import (
	"context"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

func (h *Connect) CreateTicketType(ctx context.Context, req *connect.Request[parkv1.CreateTicketTypeRequest]) (*connect.Response[parkv1.CreateTicketTypeResponse], error) {
	msg := req.Msg

	ticketType, err := h.createTicketType.Do(ctx, parkusecase.CreateTicketTypeInput{
		ParkID:        parkmodel.ParkID(msg.GetParkId()),
		Name:          msg.GetName(),
		Price:         msg.GetPrice(),
		EntryTimeFrom: msg.GetEntryTimeFrom(),
		EntryTimeTo:   msg.GetEntryTimeTo(),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&parkv1.CreateTicketTypeResponse{TicketType: toTicketTypeProto(ticketType)}), nil
}
