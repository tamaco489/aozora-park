package usecase

import (
	"context"

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
	ticketTypes parkrepository.TicketTypeWriter
}

func NewCreateTicketType(
	parks parkrepository.Reader,
	ticketTypes parkrepository.TicketTypeWriter,
) *CreateTicketType {
	return &CreateTicketType{
		parks:       parks,
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
