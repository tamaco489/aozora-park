// Package usecase は枠在庫の業務の手順を持つ
package usecase

import (
	"context"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
)

// GetDateInventoryInput は GetDateInventory の入力
type GetDateInventoryInput struct {
	ParkID inventorymodel.ParkID
	Date   inventorymodel.Date
}

// GetDateInventory は特定の日の入場枠を 1 件取得する
type GetDateInventory struct {
	inventories inventoryrepository.Reader
}

func NewGetDateInventory(inventories inventoryrepository.Reader) *GetDateInventory {
	return &GetDateInventory{inventories: inventories}
}

func (u *GetDateInventory) Do(ctx context.Context, in GetDateInventoryInput) (*inventorymodel.DateInventory, error) {
	return u.inventories.GetDateInventory(ctx, in.ParkID, in.Date)
}
