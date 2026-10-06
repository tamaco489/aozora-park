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

func TestConnectGetPriorityPass(t *testing.T) {
	handler := newHandlerHelper(t, newStoredRepositoryHelper(t))

	in := &prioritypassv1.GetPriorityPassRequest{PassId: "pass-1"}

	res, err := handler.GetPriorityPass(t.Context(), connect.NewRequest(in))
	if err != nil {
		t.Fatalf("Connect.GetPriorityPass(%v) = %v, want nil",
			in,
			err,
		)
	}

	want := &prioritypassv1.PriorityPass{
		PassId:       "pass-1",
		ParkId:       "park-1",
		TicketId:     "ticket-1",
		AttractionId: "attraction-1",
		TimeSlotId:   "20261005_1000",
		Status:       prioritypassv1.PriorityPassStatus_PRIORITY_PASS_STATUS_REQUESTED,
	}
	if diff := cmp.Diff(want, res.Msg.GetPriorityPass(), protocmp.Transform()); diff != "" {
		t.Errorf("Connect.GetPriorityPass(%v) の差分 (-want +got):\n%s",
			in,
			diff,
		)
	}
}

func TestConnectGetPriorityPassNotFound(t *testing.T) {
	handler := newHandlerHelper(t, newEmptyRepositoryHelper())

	in := &prioritypassv1.GetPriorityPassRequest{PassId: "pass-1"}

	// ハンドラはエラーを素通しする、connect の形への変換はインターセプタが行う
	_, err := handler.GetPriorityPass(t.Context(), connect.NewRequest(in))
	if !errors.Is(err, prioritypassmodel.ErrPriorityPassNotFound) {
		t.Fatalf("Connect.GetPriorityPass(%v) = %v, want %v",
			in,
			err,
			prioritypassmodel.ErrPriorityPassNotFound,
		)
	}
}
