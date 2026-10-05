// Package repository は枠在庫の永続化のインタフェースを持つ
package repository

import (
	"context"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

// Reader は枠在庫を参照する
//
// 枠の種類ごとにメソッドを分ける、同じ日付でも入場枠と時間帯枠は別の集約のため
type Reader interface {
	// GetDateInventory は特定の日の入場枠を 1 件返す、見つからないときは model.ErrDateInventoryNotFound を返す
	GetDateInventory(ctx context.Context, parkID inventorymodel.ParkID, date inventorymodel.Date) (*inventorymodel.DateInventory, error)

	// ListTimeSlots は特定の日の時間帯枠を開始時刻の昇順で返す、1 件もないときは空のスライスを返す
	ListTimeSlots(ctx context.Context, parkID inventorymodel.ParkID, attractionID inventorymodel.AttractionID, date inventorymodel.Date) ([]*inventorymodel.TimeSlot, error)
}
