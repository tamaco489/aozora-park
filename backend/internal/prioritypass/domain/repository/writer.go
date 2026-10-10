package repository

import (
	"context"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// Writer は優先パスを更新する
type Writer interface {
	// CreatePriorityPass は優先パスを作成する、既にあるときは model.ErrPriorityPassAlreadyExists を返す
	CreatePriorityPass(ctx context.Context, pass *prioritypassmodel.PriorityPass) error

	// MarkPriorityPassPublished は申込を publish した時刻を記録する、見つからないときは model.ErrPriorityPassNotFound を返す
	MarkPriorityPassPublished(
		ctx context.Context,
		id prioritypassmodel.PassID,
		publishedAt time.Time,
	) error

	// AllocateTimeSlot は時間帯枠の残りを 1 つ減らして申込を issued にする
	//
	// 残りが無いときは減らさず sold_out にする
	// 申込中でない申込には何も書かず Changed が false の結果を返す、再配信で終端のものが届くため
	// 見つからないときは model.ErrPriorityPassNotFound、枠が見つからないときは model.ErrTimeSlotNotFound を返す
	AllocateTimeSlot(
		ctx context.Context,
		id prioritypassmodel.PassID,
		now time.Time,
	) (prioritypassmodel.Allocation, error)
}
