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

// Creator は枠を作成する
//
//   - 既にある枠は上書きせず、作成したかどうかを戻り値で返す
//   - 運営が減らした残りや手で変更した上限を、ジョブの再実行で初期値に戻さないため
type Creator interface {
	// CreateDateInventoryIfAbsent は入場枠がまだ無いときだけ作成する、作成したときに true を返す
	CreateDateInventoryIfAbsent(ctx context.Context, inventory *inventorymodel.DateInventory) (bool, error)

	// CreateTimeSlotIfAbsent は時間帯枠がまだ無いときだけ作成する、作成したときに true を返す
	CreateTimeSlotIfAbsent(ctx context.Context, slot *inventorymodel.TimeSlot) (bool, error)
}
