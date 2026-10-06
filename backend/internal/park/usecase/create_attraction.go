package usecase

import (
	"context"
	"errors"

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
	reader      parkrepository.AttractionReader
	attractions parkrepository.AttractionWriter
}

func NewCreateAttraction(
	parks parkrepository.Reader,
	reader parkrepository.AttractionReader,
	attractions parkrepository.AttractionWriter,
) *CreateAttraction {
	return &CreateAttraction{
		parks:       parks,
		reader:      reader,
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

	if err := u.ensureNameFree(ctx, in.ParkID, in.Name); err != nil {
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

// ensureNameFree は同じパークに同じ表示名のアトラクションが無いことを確かめる
//
// 同名が並ぶと来園者が見分けられず、どちらの枠を申し込んだのかも分からなくなる
//
// 読んでから書くため、同名の登録が同時に来ると両方が通る
// 運営の内部操作で同時に起きる状況が考えにくいため、ここではトランザクションで束ねない
func (u *CreateAttraction) ensureNameFree(
	ctx context.Context,
	parkID parkmodel.ParkID,
	name string,
) error {
	_, err := u.reader.FindAttractionByName(ctx, parkID, name)
	if errors.Is(err, parkmodel.ErrAttractionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return parkmodel.ErrAttractionNameTaken
}
