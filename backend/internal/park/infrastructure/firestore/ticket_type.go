package firestore

import (
	"context"
	"fmt"

	gcpfirestore "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// ticketTypeCollection は券種を置く parks の子コレクション
const ticketTypeCollection = "ticketTypes"

// ticketTypeDocument は Firestore に保存する形
//
// 識別子はドキュメント ID と親のパスが持つため、フィールドには持たない
type ticketTypeDocument struct {
	Name          string `firestore:"name"`
	Price         int64  `firestore:"price"`
	EntryTimeFrom string `firestore:"entryTimeFrom"`
	EntryTimeTo   string `firestore:"entryTimeTo"`
}

var (
	_ parkrepository.TicketTypeReader = (*Repository)(nil)
	_ parkrepository.TicketTypeWriter = (*Repository)(nil)
)

func (r *Repository) GetTicketType(
	ctx context.Context,
	parkID parkmodel.ParkID,
	id parkmodel.TicketTypeID,
) (*parkmodel.TicketType, error) {
	snapshot, err := r.ticketTypeDoc(parkID, id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, parkmodel.ErrTicketTypeNotFound
		}
		return nil, fmt.Errorf("get ticket type %q of park %q: %w", id, parkID, err)
	}

	var doc ticketTypeDocument
	if err := snapshot.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("decode ticket type %q of park %q: %w", id, parkID, err)
	}

	ticketType, err := parkmodel.RestoreTicketType(parkID, id, doc.Name, doc.Price, doc.EntryTimeFrom, doc.EntryTimeTo)
	if err != nil {
		return nil, fmt.Errorf("restore ticket type %q of park %q: %w", id, parkID, err)
	}

	return ticketType, nil
}

func (r *Repository) CreateTicketType(ctx context.Context, ticketType *parkmodel.TicketType) error {
	doc := r.ticketTypeDoc(ticketType.ParkID(), ticketType.ID())

	// Create は既にあると失敗するため、採番が衝突した場合も上書きにならない
	if _, err := doc.Create(ctx, toTicketTypeDocument(ticketType)); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return parkmodel.ErrTicketTypeAlreadyExists
		}
		return fmt.Errorf("create ticket type %q of park %q: %w", ticketType.ID(), ticketType.ParkID(), err)
	}

	return nil
}

func (r *Repository) UpdateTicketType(ctx context.Context, ticketType *parkmodel.TicketType) error {
	// Set は存在しなくても作成してしまうため、削除された相手への更新を弾ける Update を使う
	updates := []gcpfirestore.Update{
		{Path: "name", Value: ticketType.Name()},
		{Path: "price", Value: ticketType.Price()},
		{Path: "entryTimeFrom", Value: ticketType.EntryTimeFrom()},
		{Path: "entryTimeTo", Value: ticketType.EntryTimeTo()},
	}

	doc := r.ticketTypeDoc(ticketType.ParkID(), ticketType.ID())
	if _, err := doc.Update(ctx, updates); err != nil {
		if status.Code(err) == codes.NotFound {
			return parkmodel.ErrTicketTypeNotFound
		}
		return fmt.Errorf("update ticket type %q of park %q: %w", ticketType.ID(), ticketType.ParkID(), err)
	}

	return nil
}

func (r *Repository) ticketTypeDoc(parkID parkmodel.ParkID, id parkmodel.TicketTypeID) *gcpfirestore.DocumentRef {
	return r.client.Collection(collection).Doc(parkID.String()).Collection(ticketTypeCollection).Doc(id.String())
}

func toTicketTypeDocument(ticketType *parkmodel.TicketType) ticketTypeDocument {
	return ticketTypeDocument{
		Name:          ticketType.Name(),
		Price:         ticketType.Price(),
		EntryTimeFrom: ticketType.EntryTimeFrom(),
		EntryTimeTo:   ticketType.EntryTimeTo(),
	}
}
