package handler

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestConnectCreatePark(t *testing.T) {
	handler, repo := newHandler(t)

	in := &parkv1.CreateParkRequest{Name: "Aozora Park", DefaultDailyCapacity: 1000, InventoryDays: 30}

	res, err := handler.CreatePark(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.CreatePark(%v) = %v, want nil", in, err)
	}

	got := res.Msg.GetPark()
	if got.GetParkId() == "" {
		t.Error("Connect.CreatePark() の park_id が空")
	}

	want := &parkv1.Park{
		ParkId:               got.GetParkId(),
		Name:                 "Aozora Park",
		DefaultDailyCapacity: 1000,
		InventoryDays:        30,
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("Connect.CreatePark() の差分 (-want +got):\n%s", diff)
	}

	if _, ok := repo.parks[parkmodel.ParkID(got.GetParkId())]; !ok {
		t.Errorf("Connect.CreatePark() の後に %q が保存されていない", got.GetParkId())
	}
}
