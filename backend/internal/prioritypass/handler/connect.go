// Package handler は優先パスの入口を持つ
//
// RPC 1 つを 1 ファイルに置き、このファイルには全 RPC で共有するものだけを置く
package handler

import (
	prioritypassv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1"
	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1/prioritypassv1connect"
	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

// Connect は PriorityPassService の connect ハンドラ
//
// 入出力の変換だけを行い、エラーはインターセプタが変換するのでそのまま返す
type Connect struct {
	requestPriorityPass *prioritypassusecase.RequestPriorityPass
	getPriorityPass     *prioritypassusecase.GetPriorityPass
}

var _ prioritypassv1connect.PriorityPassServiceHandler = (*Connect)(nil)

func NewConnect(
	requestPriorityPass *prioritypassusecase.RequestPriorityPass,
	getPriorityPass *prioritypassusecase.GetPriorityPass,
) *Connect {
	return &Connect{
		requestPriorityPass: requestPriorityPass,
		getPriorityPass:     getPriorityPass,
	}
}

func toPriorityPassProto(pass *prioritypassmodel.PriorityPass) *prioritypassv1.PriorityPass {
	return &prioritypassv1.PriorityPass{
		PassId:       pass.ID().String(),
		ParkId:       pass.ParkID().String(),
		TicketId:     pass.TicketID().String(),
		AttractionId: pass.AttractionID().String(),
		TimeSlotId:   pass.TimeSlotID().String(),
		Status:       toStatusProto(pass.Status()),
	}
}

// toStatusProto は状態を proto の enum に変換する
//
// 既知でない値は UNSPECIFIED に落とす、保存済みの値が増えても画面に別の状態として映らないため
func toStatusProto(status prioritypassmodel.Status) prioritypassv1.PriorityPassStatus {
	switch status {
	case prioritypassmodel.StatusRequested:
		return prioritypassv1.PriorityPassStatus_PRIORITY_PASS_STATUS_REQUESTED
	case prioritypassmodel.StatusIssued:
		return prioritypassv1.PriorityPassStatus_PRIORITY_PASS_STATUS_ISSUED
	case prioritypassmodel.StatusSoldOut:
		return prioritypassv1.PriorityPassStatus_PRIORITY_PASS_STATUS_SOLD_OUT
	default:
		return prioritypassv1.PriorityPassStatus_PRIORITY_PASS_STATUS_UNSPECIFIED
	}
}
