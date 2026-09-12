package usecase

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// Get はパークを 1 件取得する
type Get struct {
	parks parkrepository.Reader
}

func NewGet(parks parkrepository.Reader) *Get {
	return &Get{parks: parks}
}

func (u *Get) Do(ctx context.Context, id parkmodel.ParkID) (*parkmodel.Park, error) {
	return u.parks.Get(ctx, id)
}
