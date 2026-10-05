package usecase

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// UpdateAttractionInput は UpdateAttraction の入力
type UpdateAttractionInput struct {
	ParkID          parkmodel.ParkID
	ID              parkmodel.AttractionID
	Name            string
	Enabled         bool
	StartTime       string
	EndTime         string
	IntervalMinutes int32
	CapacityPerSlot int32
}

// UpdateAttraction は表示名と優先パスの条件を更新する
type UpdateAttraction struct {
	reader parkrepository.AttractionReader
	writer parkrepository.AttractionWriter
}

func NewUpdateAttraction(
	reader parkrepository.AttractionReader,
	writer parkrepository.AttractionWriter,
) *UpdateAttraction {
	return &UpdateAttraction{
		reader: reader,
		writer: writer,
	}
}

// Do は保存済みのアトラクションを読んでから書き換える
//
// 読んでから書くのは、更新後の値が不変条件を満たすかを Attraction 自身に判断させるため
func (u *UpdateAttraction) Do(ctx context.Context, in UpdateAttractionInput) (*parkmodel.Attraction, error) {
	attraction, err := u.reader.GetAttraction(ctx, in.ParkID, in.ID)
	if err != nil {
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

	if err := attraction.Update(in.Name, config); err != nil {
		return nil, err
	}

	if err := u.writer.UpdateAttraction(ctx, attraction); err != nil {
		return nil, err
	}

	return attraction, nil
}
