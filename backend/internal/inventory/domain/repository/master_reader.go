package repository

import (
	"context"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

// MasterReader は枠の生成に使うマスタを参照する
//
// 枠を作成するジョブだけが使うため Reader と分ける、参照の RPC に走査の手段を渡さない
type MasterReader interface {
	// ListParks はすべてのパークを返す、1 件もないときは空のスライスを返す
	ListParks(ctx context.Context) ([]*inventorymodel.ParkMaster, error)

	// ListAttractions はパークに属するアトラクションを返す、1 件もないときは空のスライスを返す
	ListAttractions(
		ctx context.Context,
		parkID inventorymodel.ParkID,
	) ([]*inventorymodel.AttractionMaster, error)
}
