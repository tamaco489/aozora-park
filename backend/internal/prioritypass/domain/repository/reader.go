// Package repository は優先パスの永続化のインタフェースを持つ
package repository

import (
	"context"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// Reader は優先パスを参照する
type Reader interface {
	// GetPriorityPass は優先パスを 1 件返す、見つからないときは model.ErrPriorityPassNotFound を返す
	GetPriorityPass(ctx context.Context, id prioritypassmodel.PassID) (*prioritypassmodel.PriorityPass, error)
}
