// Package repository はパークの永続化のインタフェースを持つ
package repository

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

// Reader はパークを参照する
type Reader interface {
	// Get は識別子でパークを 1 件返す、見つからないときは model.ErrNotFound を返す
	Get(ctx context.Context, id parkmodel.ParkID) (*parkmodel.Park, error)
}
