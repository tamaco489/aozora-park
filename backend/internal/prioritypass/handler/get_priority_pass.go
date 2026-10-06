package handler

import (
	"context"

	"connectrpc.com/connect"

	prioritypassv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1"
	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

func (h *Connect) GetPriorityPass(ctx context.Context, req *connect.Request[prioritypassv1.GetPriorityPassRequest]) (*connect.Response[prioritypassv1.GetPriorityPassResponse], error) {
	msg := req.Msg

	pass, err := h.getPriorityPass.Do(ctx, prioritypassusecase.GetPriorityPassInput{
		PassID: prioritypassmodel.PassID(msg.GetPassId()),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&prioritypassv1.GetPriorityPassResponse{
		PriorityPass: toPriorityPassProto(pass),
	}), nil
}
