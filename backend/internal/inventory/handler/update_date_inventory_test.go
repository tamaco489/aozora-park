package handler

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	inventoryv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1"
	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestConnectUpdateDateInventory(t *testing.T) {
	handler := newHandler(t, newStoredRepository(t))

	in := &inventoryv1.UpdateDateInventoryRequest{
		ParkId:    "park-1",
		Date:      "2026-10-05",
		Capacity:  2000,
		Remaining: 1500,
	}

	res, err := handler.UpdateDateInventory(t.Context(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.UpdateDateInventory(%v) = %v, want nil",
			in,
			err,
		)
	}

	want := &inventoryv1.DateInventory{
		ParkId:    "park-1",
		Date:      "2026-10-05",
		Capacity:  2000,
		Remaining: 1500,
	}
	if diff := cmp.Diff(want, res.Msg.GetDateInventory(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.UpdateDateInventory(%v) の差分 (-want +got):\n%s",
			in,
			diff,
		)
	}
}

func TestConnectUpdateDateInventoryInvalidRemaining(t *testing.T) {
	handler := newHandler(t, newStoredRepository(t))

	// protovalidate は残りが上限を超えるかを見ないため、ここまで届く
	in := &inventoryv1.UpdateDateInventoryRequest{
		ParkId:    "park-1",
		Date:      "2026-10-05",
		Capacity:  2000,
		Remaining: 2001,
	}

	_, err := handler.UpdateDateInventory(t.Context(), connect.NewRequest(in))
	if !errors.Is(err, inventorymodel.ErrInvalidRemaining) {
		t.Fatalf("Connect.UpdateDateInventory(%v) = %v, want %v",
			in,
			err,
			inventorymodel.ErrInvalidRemaining,
		)
	}
}

func TestConnectUpdateDateInventoryNotFound(t *testing.T) {
	handler := newHandler(t, newEmptyRepository())

	in := &inventoryv1.UpdateDateInventoryRequest{
		ParkId:    "park-1",
		Date:      "2026-10-05",
		Capacity:  2000,
		Remaining: 1500,
	}

	_, err := handler.UpdateDateInventory(t.Context(), connect.NewRequest(in))
	if !errors.Is(err, inventorymodel.ErrDateInventoryNotFound) {
		t.Fatalf("Connect.UpdateDateInventory(%v) = %v, want %v",
			in,
			err,
			inventorymodel.ErrDateInventoryNotFound,
		)
	}
}
