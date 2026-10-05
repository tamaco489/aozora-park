package handler

import (
	"context"

	"connectrpc.com/connect"

	inventoryv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1"
	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryusecase "github.com/tamaco489/aozora-park/backend/internal/inventory/usecase"
)

func (h *Connect) GetDateInventory(ctx context.Context, req *connect.Request[inventoryv1.GetDateInventoryRequest]) (*connect.Response[inventoryv1.GetDateInventoryResponse], error) {
	msg := req.Msg

	inventory, err := h.getDateInventory.Do(ctx, inventoryusecase.GetDateInventoryInput{
		ParkID: inventorymodel.ParkID(msg.GetParkId()),
		Date:   inventorymodel.Date(msg.GetDate()),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&inventoryv1.GetDateInventoryResponse{
		DateInventory: toDateInventoryProto(inventory),
	}), nil
}
