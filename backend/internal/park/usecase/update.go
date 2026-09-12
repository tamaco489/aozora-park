package usecase

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

type UpdateInput struct {
	ID                   parkmodel.ParkID
	Name                 string
	DefaultDailyCapacity int32
	InventoryDays        int32
}

// Update は表示名と枠の生成に使う初期値を更新する
type Update struct {
	reader parkrepository.Reader
	writer parkrepository.Writer
}

func NewUpdate(
	reader parkrepository.Reader,
	writer parkrepository.Writer,
) *Update {
	return &Update{
		reader: reader,
		writer: writer,
	}
}

// Do は保存済みのパークを読んでから書き換える
//
// 読んでから書くのは、更新後の値が不変条件を満たすかを Park 自身に判断させるため
func (u *Update) Do(ctx context.Context, in UpdateInput) (*parkmodel.Park, error) {
	park, err := u.reader.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	err = park.Update(
		in.Name,
		in.DefaultDailyCapacity,
		in.InventoryDays,
	)
	if err != nil {
		return nil, err
	}

	if err := u.writer.Update(ctx, park); err != nil {
		return nil, err
	}

	return park, nil
}
