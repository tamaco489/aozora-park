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

func TestConnectGetPark(t *testing.T) {
	handler, _ := newHandler(t, restore(t))

	in := &parkv1.GetParkRequest{ParkId: "park-1"}

	res, err := handler.GetPark(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.GetPark(%v) = %v, want nil", in, err)
	}

	want := &parkv1.Park{
		ParkId:               "park-1",
		Name:                 "Aozora Park",
		DefaultDailyCapacity: 1000,
		InventoryDays:        30,
	}
	if diff := cmp.Diff(want, res.Msg.GetPark(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.GetPark(%v) の差分 (-want +got):\n%s", in, diff)
	}
}

func TestConnectGetParkNotFound(t *testing.T) {
	handler, _ := newHandler(t)

	in := &parkv1.GetParkRequest{ParkId: "park-2"}

	// ハンドラはエラーを素通しする、connect の形への変換はインターセプタが行う
	_, err := handler.GetPark(context.Background(), connect.NewRequest(in))
	if !errors.Is(err, parkmodel.ErrNotFound) {
		t.Fatalf("Connect.GetPark(%v) = %v, want %v", in, err, parkmodel.ErrNotFound)
	}
}
