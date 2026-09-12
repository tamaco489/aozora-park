package handler

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
)

func TestConnectUpdatePark(t *testing.T) {
	handler, repo := newHandler(t, restore(t))

	in := &parkv1.UpdateParkRequest{ParkId: "park-1", Name: "Aozora Park 2", DefaultDailyCapacity: 2000, InventoryDays: 60}

	res, err := handler.UpdatePark(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.UpdatePark(%v) = %v, want nil", in, err)
	}

	want := &parkv1.Park{
		ParkId:               "park-1",
		Name:                 "Aozora Park 2",
		DefaultDailyCapacity: 2000,
		InventoryDays:        60,
	}
	if diff := cmp.Diff(want, res.Msg.GetPark(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.UpdatePark(%v) の差分 (-want +got):\n%s", in, diff)
	}

	if stored := repo.parks["park-1"]; stored.Name() != "Aozora Park 2" {
		t.Errorf("Connect.UpdatePark(%v) の後の Name = %q, want %q", in, stored.Name(), "Aozora Park 2")
	}
}
