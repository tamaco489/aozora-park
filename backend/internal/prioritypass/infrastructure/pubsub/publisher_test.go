package pubsub

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// TestNewRequestedMessage は送る本文と属性が購読側の読む形になっていることを確かめる
//
// 形が崩れると割当の worker がメッセージを解釈できなくなるため、項目名まで含めて確かめる
func TestNewRequestedMessage(t *testing.T) {
	createdAt := time.Date(
		2026,
		4,
		1,
		10,
		30,
		0,
		0,
		time.UTC,
	)

	pass, err := prioritypassmodel.RestorePriorityPass(
		"01a09793-928d-7716-9725-5bb1173c6309",
		"01a09793-928d-7716-9725-5bb1173c6310",
		"01a09793-928d-7716-9725-5bb1173c6311",
		"01a09793-928d-7716-9725-5bb1173c6312",
		"20260401_1030",
		prioritypassmodel.StatusRequested,
		createdAt,
		createdAt,
	)
	if err != nil {
		t.Fatalf("RestorePriorityPass() のエラー = %v, want %v",
			err,
			nil,
		)
	}

	got, err := newRequestedMessage(pass)
	if err != nil {
		t.Fatalf("newRequestedMessage(%q) のエラー = %v, want %v",
			pass.ID(),
			err,
			nil,
		)
	}

	var decoded requestedMessage
	if err := json.Unmarshal(got.Data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(%s) のエラー = %v, want %v",
			got.Data,
			err,
			nil,
		)
	}

	want := requestedMessage{
		PassID:       "01a09793-928d-7716-9725-5bb1173c6309",
		ParkID:       "01a09793-928d-7716-9725-5bb1173c6310",
		TicketID:     "01a09793-928d-7716-9725-5bb1173c6311",
		AttractionID: "01a09793-928d-7716-9725-5bb1173c6312",
		TimeSlotID:   "20260401_1030",
		RequestedAt:  createdAt,
	}
	if diff := cmp.Diff(want, decoded); diff != "" {
		t.Errorf("newRequestedMessage(%q) の本文の差分 (-want +got):\n%s",
			pass.ID(),
			diff,
		)
	}

	// 購読側が本文を読まずに対象を特定できるようにしているため、属性も確かめる
	wantAttributes := map[string]string{"passId": "01a09793-928d-7716-9725-5bb1173c6309"}
	if diff := cmp.Diff(wantAttributes, got.Attributes); diff != "" {
		t.Errorf("newRequestedMessage(%q) の属性の差分 (-want +got):\n%s",
			pass.ID(),
			diff,
		)
	}

	// 項目名の変更に気づけるよう、JSON のキーそのものも確かめる
	var keys map[string]any
	if err := json.Unmarshal(got.Data, &keys); err != nil {
		t.Fatalf("json.Unmarshal(%s) のエラー = %v, want %v",
			got.Data,
			err,
			nil,
		)
	}

	wantKeys := []string{
		"attractionId",
		"parkId",
		"passId",
		"requestedAt",
		"ticketId",
		"timeSlotId",
	}
	gotKeys := slices.Sorted(maps.Keys(keys))
	if diff := cmp.Diff(wantKeys, gotKeys); diff != "" {
		t.Errorf("newRequestedMessage(%q) の項目名の差分 (-want +got):\n%s",
			pass.ID(),
			diff,
		)
	}
}
