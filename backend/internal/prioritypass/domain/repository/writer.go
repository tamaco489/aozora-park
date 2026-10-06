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
}
