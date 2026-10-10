package usecase

import (
	"context"
	"errors"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// UpdateTicketTypeInput は UpdateTicketType の入力
type UpdateTicketTypeInput struct {
	ParkID        parkmodel.ParkID
	ID            parkmodel.TicketTypeID
	Name          string
	Price         int64
	EntryTimeFrom string
	EntryTimeTo   string
}

// UpdateTicketType は表示名と価格と入場できる時間帯を更新する
type UpdateTicketType struct {
	reader parkrepository.TicketTypeReader
	writer parkrepository.TicketTypeWriter
}

func NewUpdateTicketType(
	reader parkrepository.TicketTypeReader,
	writer parkrepository.TicketTypeWriter,
) *UpdateTicketType {
	return &UpdateTicketType{
		reader: reader,
		writer: writer,
	}
}

// Do は保存済みの券種を読んでから書き換える
//
// 読んでから書くのは、更新後の値が不変条件を満たすかを TicketType 自身に判断させるため
func (u *UpdateTicketType) Do(ctx context.Context, in UpdateTicketTypeInput) (*parkmodel.TicketType, error) {
	ticketType, err := u.reader.GetTicketType(
		ctx,
		in.ParkID,
		in.ID,
	)
	if err != nil {
		return nil, err
	}

	if err := u.ensureNameFree(
		ctx,
		in.ParkID,
		in.Name,
		in.ID,
	); err != nil {
		return nil, err
	}

	err = ticketType.Update(
		in.Name,
		in.Price,
		in.EntryTimeFrom,
		in.EntryTimeTo,
	)
	if err != nil {
		return nil, err
	}

	if err := u.writer.UpdateTicketType(ctx, ticketType); err != nil {
		return nil, err
	}

	return ticketType, nil
}

// ensureNameFree は同じパークに同じ表示名の券種が無いことを確かめる
//
//   - 自分自身は除く、表示名を変えない更新を弾かないため
//   - 競合の扱いは CreateTicketType.ensureNameFree と同じ
func (u *UpdateTicketType) ensureNameFree(
	ctx context.Context,
	parkID parkmodel.ParkID,
	name string,
	self parkmodel.TicketTypeID,
) error {
	found, err := u.reader.FindTicketTypeByName(ctx, parkID, name)
	if errors.Is(err, parkmodel.ErrTicketTypeNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if found.ID() == self {
		return nil
	}

	return parkmodel.ErrTicketTypeNameTaken
}
