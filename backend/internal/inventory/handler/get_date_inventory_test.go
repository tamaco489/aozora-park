package handler

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	inventoryv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1"
	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestConnectGetDateInventory(t *testing.T) {
	handler := newHandler(t, newStoredRepository(t))

	in := &inventoryv1.GetDateInventoryRequest{ParkId: "park-1", Date: "2026-10-05"}

	res, err := handler.GetDateInventory(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.GetDateInventory(%v) = %v, want nil", in, err)
	}

	want := &inventoryv1.DateInventory{
		ParkId:    "park-1",
		Date:      "2026-10-05",
		Capacity:  1000,
		Remaining: 800,
	}
	if diff := cmp.Diff(want, res.Msg.GetDateInventory(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.GetDateInventory(%v) の差分 (-want +got):\n%s", in, diff)
	}
}

func TestConnectGetDateInventoryNotFound(t *testing.T) {
	handler := newHandler(t, newEmptyRepository())

	in := &inventoryv1.GetDateInventoryRequest{ParkId: "park-1", Date: "2026-10-05"}

	// ハンドラはエラーを素通しする、connect の形への変換はインターセプタが行う
	_, err := handler.GetDateInventory(context.Background(), connect.NewRequest(in))
	if !errors.Is(err, inventorymodel.ErrDateInventoryNotFound) {
		t.Fatalf("Connect.GetDateInventory(%v) = %v, want %v", in, err, inventorymodel.ErrDateInventoryNotFound)
	}
}
