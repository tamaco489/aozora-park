package repository

import (
	"context"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

// Writer は枠在庫を更新する
type Writer interface {
	// UpdateDateInventory は保存済みの入場枠を書き換える、見つからないときは model.ErrDateInventoryNotFound を返す
	UpdateDateInventory(ctx context.Context, inventory *inventorymodel.DateInventory) error
}
