// Package handler は枠在庫の入口を持つ
//
// RPC 1 つを 1 ファイルに置き、このファイルには全 RPC で共有するものだけを置く
package handler

import (
	inventoryv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1"
	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1/inventoryv1connect"
	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryusecase "github.com/tamaco489/aozora-park/backend/internal/inventory/usecase"
)

// Connect は InventoryService の connect ハンドラ
//
// 入出力の変換だけを行い、エラーはインターセプタが変換するのでそのまま返す
type Connect struct {
	updateDateInventory *inventoryusecase.UpdateDateInventory
	getDateInventory    *inventoryusecase.GetDateInventory
	listTimeSlots       *inventoryusecase.ListTimeSlots
}

var _ inventoryv1connect.InventoryServiceHandler = (*Connect)(nil)

func NewConnect(
	updateDateInventory *inventoryusecase.UpdateDateInventory,
	getDateInventory *inventoryusecase.GetDateInventory,
	listTimeSlots *inventoryusecase.ListTimeSlots,
) *Connect {
	return &Connect{
		updateDateInventory: updateDateInventory,
		getDateInventory:    getDateInventory,
		listTimeSlots:       listTimeSlots,
	}
}

func toDateInventoryProto(inventory *inventorymodel.DateInventory) *inventoryv1.DateInventory {
	return &inventoryv1.DateInventory{
		ParkId:    inventory.ParkID().String(),
		Date:      inventory.Date().String(),
		Capacity:  inventory.Capacity(),
		Remaining: inventory.Remaining(),
	}
}

func toTimeSlotProto(slot *inventorymodel.TimeSlot) *inventoryv1.TimeSlot {
	return &inventoryv1.TimeSlot{
		ParkId:       slot.ParkID().String(),
		AttractionId: slot.AttractionID().String(),
		TimeSlotId:   slot.ID().String(),
		Date:         slot.Date().String(),
		StartTime:    slot.StartTime(),
		Capacity:     slot.Capacity(),
		Remaining:    slot.Remaining(),
	}
}
