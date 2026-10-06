// Package repository はパークの永続化のインタフェースを持つ
package repository

import (
	"context"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

// Reader はパークを参照する
type Reader interface {
	// Get は識別子でパークを 1 件返す、見つからないときは model.ErrNotFound を返す
	Get(ctx context.Context, id parkmodel.ParkID) (*parkmodel.Park, error)
}

// AttractionReader はアトラクションを参照する
type AttractionReader interface {
	// GetAttraction は識別子でアトラクションを 1 件返す、見つからないときは model.ErrAttractionNotFound を返す
	GetAttraction(
		ctx context.Context,
		parkID parkmodel.ParkID,
		id parkmodel.AttractionID,
	) (*parkmodel.Attraction, error)

	// FindAttractionByName は表示名でアトラクションを 1 件返す、見つからないときは model.ErrAttractionNotFound を返す
	FindAttractionByName(
		ctx context.Context,
		parkID parkmodel.ParkID,
		name string,
	) (*parkmodel.Attraction, error)
}

// TicketTypeReader は券種を参照する
type TicketTypeReader interface {
	// GetTicketType は識別子で券種を 1 件返す、見つからないときは model.ErrTicketTypeNotFound を返す
	GetTicketType(
		ctx context.Context,
		parkID parkmodel.ParkID,
		id parkmodel.TicketTypeID,
	) (*parkmodel.TicketType, error)

	// FindTicketTypeByName は表示名で券種を 1 件返す、見つからないときは model.ErrTicketTypeNotFound を返す
	FindTicketTypeByName(
		ctx context.Context,
		parkID parkmodel.ParkID,
		name string,
	) (*parkmodel.TicketType, error)
}
