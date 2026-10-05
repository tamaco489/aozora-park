// Package handler はパークの入口を持つ
//
// RPC 1 つを 1 ファイルに置き、このファイルには全 RPC で共有するものだけを置く
package handler

import (
	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1/parkv1connect"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

// Connect は ParkService の connect ハンドラ
//
// 入出力の変換だけを行い、エラーはインターセプタが変換するのでそのまま返す
type Connect struct {
	create           *parkusecase.Create
	get              *parkusecase.Get
	update           *parkusecase.Update
	createAttraction *parkusecase.CreateAttraction
	updateAttraction *parkusecase.UpdateAttraction
	createTicketType *parkusecase.CreateTicketType
	updateTicketType *parkusecase.UpdateTicketType
}

var _ parkv1connect.ParkServiceHandler = (*Connect)(nil)

func NewConnect(
	create *parkusecase.Create,
	get *parkusecase.Get,
	update *parkusecase.Update,
	createAttraction *parkusecase.CreateAttraction,
	updateAttraction *parkusecase.UpdateAttraction,
	createTicketType *parkusecase.CreateTicketType,
	updateTicketType *parkusecase.UpdateTicketType,
) *Connect {
	return &Connect{
		create:           create,
		get:              get,
		update:           update,
		createAttraction: createAttraction,
		updateAttraction: updateAttraction,
		createTicketType: createTicketType,
		updateTicketType: updateTicketType,
	}
}

func toProto(park *parkmodel.Park) *parkv1.Park {
	return &parkv1.Park{
		ParkId:               park.ID().String(),
		Name:                 park.Name(),
		DefaultDailyCapacity: park.DefaultDailyCapacity(),
		InventoryDays:        park.InventoryDays(),
	}
}

func toAttractionProto(attraction *parkmodel.Attraction) *parkv1.Attraction {
	config := attraction.PriorityPassConfig()

	return &parkv1.Attraction{
		AttractionId: attraction.ID().String(),
		ParkId:       attraction.ParkID().String(),
		Name:         attraction.Name(),
		PriorityPassConfig: &parkv1.PriorityPassConfig{
			Enabled:         config.Enabled(),
			StartTime:       config.StartTime(),
			EndTime:         config.EndTime(),
			IntervalMinutes: config.IntervalMinutes(),
			CapacityPerSlot: config.CapacityPerSlot(),
		},
	}
}

func toTicketTypeProto(ticketType *parkmodel.TicketType) *parkv1.TicketType {
	return &parkv1.TicketType{
		TicketTypeId:  ticketType.ID().String(),
		ParkId:        ticketType.ParkID().String(),
		Name:          ticketType.Name(),
		Price:         ticketType.Price(),
		EntryTimeFrom: ticketType.EntryTimeFrom(),
		EntryTimeTo:   ticketType.EntryTimeTo(),
	}
}
