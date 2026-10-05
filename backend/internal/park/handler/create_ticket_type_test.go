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

func TestConnectCreateTicketType(t *testing.T) {
	handler, repo := newHandler(t, restore(t))

	in := &parkv1.CreateTicketTypeRequest{
		ParkId:        "park-1",
		Name:          "1 デーパスポート",
		Price:         8000,
		EntryTimeFrom: "09:00",
		EntryTimeTo:   "21:00",
	}

	res, err := handler.CreateTicketType(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.CreateTicketType(%v) = %v, want nil", in, err)
	}

	got := res.Msg.GetTicketType()
	if got.GetTicketTypeId() == "" {
		t.Error("Connect.CreateTicketType() の ticket_type_id が空")
	}

	want := &parkv1.TicketType{
		TicketTypeId:  got.GetTicketTypeId(),
		ParkId:        "park-1",
		Name:          "1 デーパスポート",
		Price:         8000,
		EntryTimeFrom: "09:00",
		EntryTimeTo:   "21:00",
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("Connect.CreateTicketType() の差分 (-want +got):\n%s", diff)
	}

	key := ticketTypeKey{parkID: "park-1", id: parkmodel.TicketTypeID(got.GetTicketTypeId())}
	if _, ok := repo.ticketTypes[key]; !ok {
		t.Errorf("Connect.CreateTicketType() の後に %q が保存されていない", got.GetTicketTypeId())
	}
}

func TestConnectCreateTicketTypeParkNotFound(t *testing.T) {
	handler, _ := newHandler(t)

	in := &parkv1.CreateTicketTypeRequest{
		ParkId:        "park-2",
		Name:          "1 デーパスポート",
		Price:         8000,
		EntryTimeFrom: "09:00",
		EntryTimeTo:   "21:00",
	}

	// ハンドラはエラーを素通しする、connect の形への変換はインターセプタが行う
	_, err := handler.CreateTicketType(context.Background(), connect.NewRequest(in))
	if !errors.Is(err, parkmodel.ErrNotFound) {
		t.Fatalf("Connect.CreateTicketType(%v) = %v, want %v", in, err, parkmodel.ErrNotFound)
	}
}
