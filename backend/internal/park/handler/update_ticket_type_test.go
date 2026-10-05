package handler

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestConnectUpdateTicketType(t *testing.T) {
	handler, repo := newHandler(t, restore(t))
	storeTicketType(t, repo)

	in := &parkv1.UpdateTicketTypeRequest{
		ParkId:        "park-1",
		TicketTypeId:  "ticket-type-1",
		Name:          "2 デーパスポート",
		Price:         14000,
		EntryTimeFrom: "10:00",
		EntryTimeTo:   "20:00",
	}

	res, err := handler.UpdateTicketType(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.UpdateTicketType(%v) = %v, want nil",
			in,
			err,
		)
	}

	want := &parkv1.TicketType{
		TicketTypeId:  "ticket-type-1",
		ParkId:        "park-1",
		Name:          "2 デーパスポート",
		Price:         14000,
		EntryTimeFrom: "10:00",
		EntryTimeTo:   "20:00",
	}
	if diff := cmp.Diff(want, res.Msg.GetTicketType(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.UpdateTicketType(%v) の差分 (-want +got):\n%s",
			in,
			diff,
		)
	}

	stored := repo.ticketTypes[ticketTypeKey{
		parkID: "park-1",
		id:     "ticket-type-1",
	}]
	if stored.Price() != 14000 {
		t.Errorf("Connect.UpdateTicketType(%v) の後の Price = %d, want %d",
			in,
			stored.Price(),
			14000,
		)
	}
}

func TestConnectUpdateTicketTypeNotFound(t *testing.T) {
	handler, _ := newHandler(t, restore(t))

	in := &parkv1.UpdateTicketTypeRequest{
		ParkId:        "park-1",
		TicketTypeId:  "ticket-type-2",
		Name:          "2 デーパスポート",
		Price:         14000,
		EntryTimeFrom: "10:00",
		EntryTimeTo:   "20:00",
	}

	_, err := handler.UpdateTicketType(context.Background(), connect.NewRequest(in))
	if !errors.Is(err, parkmodel.ErrTicketTypeNotFound) {
		t.Fatalf("Connect.UpdateTicketType(%v) = %v, want %v",
			in,
			err,
			parkmodel.ErrTicketTypeNotFound,
		)
	}
}
