package repository

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

// Writer はパークを更新する
type Writer interface {
	// Create はパークを新しく保存する、既にあるときは model.ErrAlreadyExists を返す
	Create(ctx context.Context, park *parkmodel.Park) error

	// Update は保存済みのパークを書き換える、見つからないときは model.ErrNotFound を返す
	Update(ctx context.Context, park *parkmodel.Park) error
}

// AttractionWriter はアトラクションを更新する
type AttractionWriter interface {
	// CreateAttraction はアトラクションを新しく保存する、既にあるときは model.ErrAttractionAlreadyExists を返す
	CreateAttraction(ctx context.Context, attraction *parkmodel.Attraction) error

	// UpdateAttraction は保存済みのアトラクションを書き換える、見つからないときは model.ErrAttractionNotFound を返す
	UpdateAttraction(ctx context.Context, attraction *parkmodel.Attraction) error
}

// TicketTypeWriter は券種を更新する
type TicketTypeWriter interface {
	// CreateTicketType は券種を新しく保存する、既にあるときは model.ErrTicketTypeAlreadyExists を返す
	CreateTicketType(ctx context.Context, ticketType *parkmodel.TicketType) error

	// UpdateTicketType は保存済みの券種を書き換える、見つからないときは model.ErrTicketTypeNotFound を返す
	UpdateTicketType(ctx context.Context, ticketType *parkmodel.TicketType) error
}
