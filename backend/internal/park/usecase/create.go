// Package usecase はパークの業務の手順を持つ
package usecase

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// CreateInput は CreatePark の入力
type CreateInput struct {
	Name                 string
	DefaultDailyCapacity int32
	InventoryDays        int32
}

// Create はパークを新しく登録する
type Create struct {
	parks parkrepository.Writer
}

func NewCreate(parks parkrepository.Writer) *Create {
	return &Create{parks: parks}
}

func (u *Create) Do(ctx context.Context, in CreateInput) (*parkmodel.Park, error) {
	park, err := parkmodel.New(
		in.Name,
		in.DefaultDailyCapacity,
		in.InventoryDays,
	)
	if err != nil {
		return nil, err
	}

	if err := u.parks.Create(ctx, park); err != nil {
		return nil, err
	}

	return park, nil
}
