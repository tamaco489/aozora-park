package usecase

import (
	"context"
	"errors"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// CreateTicketTypeInput は CreateTicketType の入力
type CreateTicketTypeInput struct {
	ParkID        parkmodel.ParkID
	Name          string
	Price         int64
	EntryTimeFrom string
	EntryTimeTo   string
}

// CreateTicketType はパークに券種を新しく登録する
type CreateTicketType struct {
	parks       parkrepository.Reader
	reader      parkrepository.TicketTypeReader
	ticketTypes parkrepository.TicketTypeWriter
}

func NewCreateTicketType(
	parks parkrepository.Reader,
	reader parkrepository.TicketTypeReader,
	ticketTypes parkrepository.TicketTypeWriter,
) *CreateTicketType {
	return &CreateTicketType{
		parks:       parks,
		reader:      reader,
		ticketTypes: ticketTypes,
	}
}

// Do は親のパークを確かめてから券種を保存する
//
// 子コレクションへの書き込みは親のドキュメントが無くても成功するため、ここで存在を確かめる
func (u *CreateTicketType) Do(ctx context.Context, in CreateTicketTypeInput) (*parkmodel.TicketType, error) {
	if _, err := u.parks.Get(ctx, in.ParkID); err != nil {
		return nil, err
	}

	if err := u.ensureNameFree(ctx, in.ParkID, in.Name); err != nil {
		return nil, err
	}

	ticketType, err := parkmodel.NewTicketType(
		in.ParkID,
		in.Name,
		in.Price,
		in.EntryTimeFrom,
		in.EntryTimeTo,
	)
	if err != nil {
		return nil, err
	}

	if err := u.ticketTypes.CreateTicketType(ctx, ticketType); err != nil {
		return nil, err
	}

	return ticketType, nil
}

// ensureNameFree は同じパークに同じ表示名の券種が無いことを確かめる
//
// 同名が並ぶと購入者がどちらを選んだのか分からなくなる
//
// 読んでから書くため、同名の登録が同時に来ると両方が通る
// 運営の内部操作で同時に起きる状況が考えにくいため、ここではトランザクションで束ねない
func (u *CreateTicketType) ensureNameFree(
	ctx context.Context,
	parkID parkmodel.ParkID,
	name string,
) error {
	_, err := u.reader.FindTicketTypeByName(ctx, parkID, name)
	if errors.Is(err, parkmodel.ErrTicketTypeNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return parkmodel.ErrTicketTypeNameTaken
}
