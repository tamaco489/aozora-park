package handler

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	prioritypassv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1"
	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

func TestConnectRequestPriorityPass(t *testing.T) {
	handler := newHandlerHelper(t, newEmptyRepositoryHelper())

	in := &prioritypassv1.RequestPriorityPassRequest{
		ParkId:       "park-1",
		TicketId:     "ticket-1",
		AttractionId: "attraction-1",
		TimeSlotId:   "20261005_1000",
	}

	res, err := handler.RequestPriorityPass(t.Context(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.RequestPriorityPass(%v) = %v, want nil",
			in,
			err,
		)
	}

	got := res.Msg.GetPriorityPass()
	if got.GetPassId() == "" {
		t.Fatalf("Connect.RequestPriorityPass(%v) の識別子 = %q, want 空でない値",
			in,
			got.GetPassId(),
		)
	}

	// 識別子は採番されるため、比較の前に応答が返した値を入れる
	want := &prioritypassv1.PriorityPass{
		PassId:       got.GetPassId(),
		ParkId:       "park-1",
		TicketId:     "ticket-1",
		AttractionId: "attraction-1",
		TimeSlotId:   "20261005_1000",
		Status:       prioritypassv1.PriorityPassStatus_PRIORITY_PASS_STATUS_REQUESTED,
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("Connect.RequestPriorityPass(%v) の差分 (-want +got):\n%s",
			in,
			diff,
		)
	}
}

func TestConnectRequestPriorityPassInvalidTimeSlotID(t *testing.T) {
	handler := newHandlerHelper(t, newEmptyRepositoryHelper())

	// protovalidate はインターセプタが見るため、ハンドラを直接呼ぶと domain の検証に掛かる
	in := &prioritypassv1.RequestPriorityPassRequest{
		ParkId:       "park-1",
		TicketId:     "ticket-1",
		AttractionId: "attraction-1",
		TimeSlotId:   "20261005-1000",
	}

	_, err := handler.RequestPriorityPass(t.Context(), connect.NewRequest(in))
	if !errors.Is(err, prioritypassmodel.ErrInvalidTimeSlotID) {
		t.Fatalf("Connect.RequestPriorityPass(%v) = %v, want %v",
			in,
			err,
			prioritypassmodel.ErrInvalidTimeSlotID,
		)
	}
}
