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

func TestConnectUpdateAttraction(t *testing.T) {
	handler, repo := newHandler(t, restore(t))
	storeAttraction(t, repo)

	in := &parkv1.UpdateAttractionRequest{
		ParkId:          "park-1",
		AttractionId:    "attraction-1",
		Name:            "ジェットコースター 2",
		Enabled:         false,
		StartTime:       "10:00",
		EndTime:         "20:00",
		IntervalMinutes: 15,
		CapacityPerSlot: 20,
	}

	res, err := handler.UpdateAttraction(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.UpdateAttraction(%v) = %v, want nil",
			in,
			err,
		)
	}

	want := &parkv1.Attraction{
		AttractionId: "attraction-1",
		ParkId:       "park-1",
		Name:         "ジェットコースター 2",
		PriorityPassConfig: &parkv1.PriorityPassConfig{
			Enabled:         false,
			StartTime:       "10:00",
			EndTime:         "20:00",
			IntervalMinutes: 15,
			CapacityPerSlot: 20,
		},
	}
	if diff := cmp.Diff(want, res.Msg.GetAttraction(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.UpdateAttraction(%v) の差分 (-want +got):\n%s",
			in,
			diff,
		)
	}

	stored := repo.attractions[attractionKey{
		parkID: "park-1",
		id:     "attraction-1",
	}]
	if stored.Name() != "ジェットコースター 2" {
		t.Errorf("Connect.UpdateAttraction(%v) の後の Name = %q, want %q",
			in,
			stored.Name(),
			"ジェットコースター 2",
		)
	}
}

func TestConnectUpdateAttractionNotFound(t *testing.T) {
	handler, _ := newHandler(t, restore(t))

	in := &parkv1.UpdateAttractionRequest{
		ParkId:          "park-1",
		AttractionId:    "attraction-2",
		Name:            "ジェットコースター 2",
		Enabled:         true,
		StartTime:       "10:00",
		EndTime:         "20:00",
		IntervalMinutes: 15,
		CapacityPerSlot: 20,
	}

	_, err := handler.UpdateAttraction(context.Background(), connect.NewRequest(in))
	if !errors.Is(err, parkmodel.ErrAttractionNotFound) {
		t.Fatalf("Connect.UpdateAttraction(%v) = %v, want %v",
			in,
			err,
			parkmodel.ErrAttractionNotFound,
		)
	}
}
