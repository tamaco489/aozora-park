package usecase

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// CreateAttractionInput は CreateAttraction の入力
type CreateAttractionInput struct {
	ParkID          parkmodel.ParkID
	Name            string
	Enabled         bool
	StartTime       string
	EndTime         string
	IntervalMinutes int32
	CapacityPerSlot int32
}

// CreateAttraction はパークにアトラクションを新しく登録する
type CreateAttraction struct {
	parks       parkrepository.Reader
	attractions parkrepository.AttractionWriter
}

func NewCreateAttraction(
	parks parkrepository.Reader,
	attractions parkrepository.AttractionWriter,
) *CreateAttraction {
	return &CreateAttraction{
		parks:       parks,
		attractions: attractions,
	}
}

// Do は親のパークを確かめてからアトラクションを保存する
//
// 子コレクションへの書き込みは親のドキュメントが無くても成功するため、ここで存在を確かめる
func (u *CreateAttraction) Do(ctx context.Context, in CreateAttractionInput) (*parkmodel.Attraction, error) {
	if _, err := u.parks.Get(ctx, in.ParkID); err != nil {
		return nil, err
	}

	config, err := parkmodel.NewPriorityPassConfig(
		in.Enabled,
		in.StartTime,
		in.EndTime,
		in.IntervalMinutes,
		in.CapacityPerSlot,
	)
	if err != nil {
		return nil, err
	}

	attraction, err := parkmodel.NewAttraction(
		in.ParkID,
		in.Name,
		config,
	)
	if err != nil {
		return nil, err
	}

	if err := u.attractions.CreateAttraction(ctx, attraction); err != nil {
		return nil, err
	}

	return attraction, nil
}
