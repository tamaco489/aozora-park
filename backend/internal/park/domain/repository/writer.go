package repository

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

// Writer はパークを更新する
type Writer interface {
	// Create はパークを新しく保存する、既にあるときは model.ErrAlreadyExists を返す
	Create(ctx context.Context, park *parkmodel.Park) error

	// Update は保存済みのパークを書き換える、見つからないときは model.ErrNotFound を返す
	Update(ctx context.Context, park *parkmodel.Park) error
}
