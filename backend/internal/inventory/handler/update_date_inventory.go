package handler

import (
	"context"

	"connectrpc.com/connect"

	inventoryv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1"
	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryusecase "github.com/tamaco489/aozora-park/backend/internal/inventory/usecase"
)

func (h *Connect) UpdateDateInventory(ctx context.Context, req *connect.Request[inventoryv1.UpdateDateInventoryRequest]) (*connect.Response[inventoryv1.UpdateDateInventoryResponse], error) {
	msg := req.Msg

	inventory, err := h.updateDateInventory.Do(ctx, inventoryusecase.UpdateDateInventoryInput{
		ParkID:    inventorymodel.ParkID(msg.GetParkId()),
		Date:      inventorymodel.Date(msg.GetDate()),
		Capacity:  msg.GetCapacity(),
		Remaining: msg.GetRemaining(),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&inventoryv1.UpdateDateInventoryResponse{
		DateInventory: toDateInventoryProto(inventory),
	}), nil
}
