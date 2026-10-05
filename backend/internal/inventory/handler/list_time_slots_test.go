package handler

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	inventoryv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1"
)

func TestConnectListTimeSlots(t *testing.T) {
	handler := newHandler(t, newStoredRepository(t))

	in := &inventoryv1.ListTimeSlotsRequest{ParkId: "park-1", AttractionId: "attraction-1", Date: "2026-10-05"}

	res, err := handler.ListTimeSlots(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.ListTimeSlots(%v) = %v, want nil", in, err)
	}

	want := []*inventoryv1.TimeSlot{
		{
			ParkId:       "park-1",
			AttractionId: "attraction-1",
			TimeSlotId:   "20261005_1000",
			Date:         "2026-10-05",
			StartTime:    "10:00",
			Capacity:     60,
			Remaining:    30,
		},
		{
			ParkId:       "park-1",
			AttractionId: "attraction-1",
			TimeSlotId:   "20261005_1100",
			Date:         "2026-10-05",
			StartTime:    "11:00",
			Capacity:     60,
			Remaining:    30,
		},
	}
	if diff := cmp.Diff(want, res.Msg.GetTimeSlots(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.ListTimeSlots(%v) の差分 (-want +got):\n%s", in, diff)
	}
}

func TestConnectListTimeSlotsEmpty(t *testing.T) {
	handler := newHandler(t, newEmptyRepository())

	in := &inventoryv1.ListTimeSlotsRequest{ParkId: "park-1", AttractionId: "attraction-1", Date: "2026-10-05"}

	res, err := handler.ListTimeSlots(context.Background(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.ListTimeSlots(%v) = %v, want nil", in, err)
	}

	if got := len(res.Msg.GetTimeSlots()); got != 0 {
		t.Errorf("Connect.ListTimeSlots(%v) の件数 = %d, want %d", in, got, 0)
	}
}
