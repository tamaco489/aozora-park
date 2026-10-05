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

func TestConnectCreateAttraction(t *testing.T) {
	handler, repo := newHandler(t, restore(t))

	in := &parkv1.CreateAttractionRequest{
		ParkId:          "park-1",
		Name:            "ジェットコースター",
		Enabled:         true,
		StartTime:       "09:00",
		EndTime:         "18:00",
		IntervalMinutes: 30,
		CapacityPerSlot: 10,
	}

	res, err := handler.CreateAttraction(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.CreateAttraction(%v) = %v, want nil",
			in,
			err,
		)
	}

	got := res.Msg.GetAttraction()
	if got.GetAttractionId() == "" {
		t.Error("Connect.CreateAttraction() の attraction_id が空")
	}

	want := &parkv1.Attraction{
		AttractionId: got.GetAttractionId(),
		ParkId:       "park-1",
		Name:         "ジェットコースター",
		PriorityPassConfig: &parkv1.PriorityPassConfig{
			Enabled:         true,
			StartTime:       "09:00",
			EndTime:         "18:00",
			IntervalMinutes: 30,
			CapacityPerSlot: 10,
		},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("Connect.CreateAttraction() の差分 (-want +got):\n%s", diff)
	}

	key := attractionKey{
		parkID: "park-1",
		id:     parkmodel.AttractionID(got.GetAttractionId()),
	}
	if _, ok := repo.attractions[key]; !ok {
		t.Errorf("Connect.CreateAttraction() の後に %q が保存されていない", got.GetAttractionId())
	}
}

func TestConnectCreateAttractionParkNotFound(t *testing.T) {
	handler, _ := newHandler(t)

	in := &parkv1.CreateAttractionRequest{
		ParkId:          "park-2",
		Name:            "ジェットコースター",
		Enabled:         true,
		StartTime:       "09:00",
		EndTime:         "18:00",
		IntervalMinutes: 30,
		CapacityPerSlot: 10,
	}

	// ハンドラはエラーを素通しする、connect の形への変換はインターセプタが行う
	_, err := handler.CreateAttraction(context.Background(), connect.NewRequest(in))
	if !errors.Is(err, parkmodel.ErrNotFound) {
		t.Fatalf("Connect.CreateAttraction(%v) = %v, want %v",
			in,
			err,
			parkmodel.ErrNotFound,
		)
	}
}
