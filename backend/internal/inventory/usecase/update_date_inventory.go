package usecase

import (
	"context"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
)

// UpdateDateInventoryInput は UpdateDateInventory の入力
type UpdateDateInventoryInput struct {
	ParkID    inventorymodel.ParkID
	Date      inventorymodel.Date
	Capacity  int32
	Remaining int32
}

// UpdateDateInventory は運営が特定の日の上限人数と残りの人数を上書きする
type UpdateDateInventory struct {
	reader inventoryrepository.Reader
	writer inventoryrepository.Writer
}

func NewUpdateDateInventory(
	reader inventoryrepository.Reader,
	writer inventoryrepository.Writer,
) *UpdateDateInventory {
	return &UpdateDateInventory{
		reader: reader,
		writer: writer,
	}
}

// Do は保存済みの入場枠を読んでから書き換える
//
// 読んでから書くのは、上書き後の値が不変条件を満たすかを DateInventory 自身に判断させるため
func (u *UpdateDateInventory) Do(ctx context.Context, in UpdateDateInventoryInput) (*inventorymodel.DateInventory, error) {
	inventory, err := u.reader.GetDateInventory(
		ctx,
		in.ParkID,
		in.Date,
	)
	if err != nil {
		return nil, err
	}

	if err := inventory.Overwrite(in.Capacity, in.Remaining); err != nil {
		return nil, err
	}

	if err := u.writer.UpdateDateInventory(ctx, inventory); err != nil {
		return nil, err
	}

	return inventory, nil
}
