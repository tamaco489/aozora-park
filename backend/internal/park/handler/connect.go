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
	create *parkusecase.Create
	get    *parkusecase.Get
	update *parkusecase.Update
}

var _ parkv1connect.ParkServiceHandler = (*Connect)(nil)

func NewConnect(create *parkusecase.Create, get *parkusecase.Get, update *parkusecase.Update) *Connect {
	return &Connect{create: create, get: get, update: update}
}

func toProto(park *parkmodel.Park) *parkv1.Park {
	return &parkv1.Park{
		ParkId:               park.ID().String(),
		Name:                 park.Name(),
		DefaultDailyCapacity: park.DefaultDailyCapacity(),
		InventoryDays:        park.InventoryDays(),
	}
}
