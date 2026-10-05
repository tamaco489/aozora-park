package usecase

import (
	"context"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
)

// ListTimeSlotsInput は ListTimeSlots の入力
type ListTimeSlotsInput struct {
	ParkID       inventorymodel.ParkID
	AttractionID inventorymodel.AttractionID
	Date         inventorymodel.Date
}

// ListTimeSlots はアトラクションの特定の日の時間帯枠を取得する
type ListTimeSlots struct {
	inventories inventoryrepository.Reader
}

func NewListTimeSlots(inventories inventoryrepository.Reader) *ListTimeSlots {
	return &ListTimeSlots{inventories: inventories}
}

func (u *ListTimeSlots) Do(ctx context.Context, in ListTimeSlotsInput) ([]*inventorymodel.TimeSlot, error) {
	return u.inventories.ListTimeSlots(
		ctx,
		in.ParkID,
		in.AttractionID,
		in.Date,
	)
}
