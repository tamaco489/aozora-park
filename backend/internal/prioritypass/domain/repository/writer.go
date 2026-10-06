package repository

import (
	"context"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// Writer は優先パスを更新する
type Writer interface {
	// CreatePriorityPass は優先パスを作成する、既にあるときは model.ErrPriorityPassAlreadyExists を返す
	CreatePriorityPass(ctx context.Context, pass *prioritypassmodel.PriorityPass) error
}
