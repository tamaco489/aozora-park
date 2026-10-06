package model

import (
	"errors"
	"testing"
	"time"
)

// 申込に使う妥当な値、検証の対象にする項目だけをケースごとに差し替える
const (
	validParkID       = ParkID("park-1")
	validTicketID     = TicketID("ticket-1")
	validAttractionID = AttractionID("attraction-1")
	validTimeSlotID   = TimeSlotID("20261005_1000")
)

var validNow = time.Date(
	2026, 10, 5,
	9, 0, 0, 0,
	time.UTC,
)

func TestNewPriorityPass(t *testing.T) {
	tests := map[string]struct {
		parkID       ParkID
		ticketID     TicketID
		attractionID AttractionID
		timeSlotID   TimeSlotID
		wantErr      error
	}{
		"正常系_すべて妥当な場合_エラーにならないこと": {
			parkID:       validParkID,
			ticketID:     validTicketID,
			attractionID: validAttractionID,
			timeSlotID:   validTimeSlotID,
		},
		"異常系_パークの識別子が空の場合_ErrInvalidParkIDになること": {
			parkID:       "",
			ticketID:     validTicketID,
			attractionID: validAttractionID,
			timeSlotID:   validTimeSlotID,
			wantErr:      ErrInvalidParkID,
		},
		"異常系_券の識別子が空の場合_ErrInvalidTicketIDになること": {
			parkID:       validParkID,
			ticketID:     "",
			attractionID: validAttractionID,
			timeSlotID:   validTimeSlotID,
			wantErr:      ErrInvalidTicketID,
		},
		"異常系_アトラクションの識別子が空の場合_ErrInvalidAttractionIDになること": {
			parkID:       validParkID,
			ticketID:     validTicketID,
			attractionID: "",
			timeSlotID:   validTimeSlotID,
			wantErr:      ErrInvalidAttractionID,
		},
		"異常系_時間帯枠の識別子が形式に合わない場合_ErrInvalidTimeSlotIDになること": {
			parkID:       validParkID,
			ticketID:     validTicketID,
			attractionID: validAttractionID,
			timeSlotID:   "20261005-1000",
			wantErr:      ErrInvalidTimeSlotID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewPriorityPass(
				tt.parkID,
				tt.ticketID,
				tt.attractionID,
				tt.timeSlotID,
				validNow,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewPriorityPass(%q, %q, %q, %q) のエラー = %v, want %v",
					tt.parkID,
					tt.ticketID,
					tt.attractionID,
					tt.timeSlotID,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				return
			}

			if got.ID() == "" {
				t.Errorf("NewPriorityPass(%q, %q, %q, %q) の識別子 = %q, want 空でない値",
					tt.parkID,
					tt.ticketID,
					tt.attractionID,
					tt.timeSlotID,
					got.ID(),
				)
			}

			if got.Status() != StatusRequested {
				t.Errorf("NewPriorityPass(%q, %q, %q, %q) の状態 = %q, want %q",
					tt.parkID,
					tt.ticketID,
					tt.attractionID,
					tt.timeSlotID,
					got.Status(),
					StatusRequested,
				)
			}

			if !got.CreatedAt().Equal(validNow) || !got.UpdatedAt().Equal(validNow) {
				t.Errorf("NewPriorityPass(%q, %q, %q, %q) の時刻 = (%v, %v), want (%v, %v)",
					tt.parkID,
					tt.ticketID,
					tt.attractionID,
					tt.timeSlotID,
					got.CreatedAt(),
					got.UpdatedAt(),
					validNow,
					validNow,
				)
			}
		})
	}
}

// TestNewPriorityPassIDIsUnique は採番が呼び出しごとに異なることを確かめる
func TestNewPriorityPassIDIsUnique(t *testing.T) {
	first := newValidPriorityPassHelper(t)
	second := newValidPriorityPassHelper(t)

	if first.ID() == second.ID() {
		t.Errorf("NewPriorityPass() の識別子 = %q, want %q と異なる値",
			second.ID(),
			first.ID(),
		)
	}
}

func TestRestorePriorityPass(t *testing.T) {
	tests := map[string]struct {
		id         PassID
		timeSlotID TimeSlotID
		status     Status
		wantErr    error
	}{
		"正常系_すべて妥当な場合_エラーにならないこと": {
			id:         "pass-1",
			timeSlotID: validTimeSlotID,
			status:     StatusIssued,
		},
		"正常系_sold_outの場合_エラーにならないこと": {
			id:         "pass-1",
			timeSlotID: validTimeSlotID,
			status:     StatusSoldOut,
		},
		"異常系_識別子が空の場合_ErrInvalidIDになること": {
			id:         "",
			timeSlotID: validTimeSlotID,
			status:     StatusIssued,
			wantErr:    ErrInvalidID,
		},
		"異常系_状態が既知でない場合_ErrInvalidStatusになること": {
			id:         "pass-1",
			timeSlotID: validTimeSlotID,
			status:     Status("rejected"),
			wantErr:    ErrInvalidStatus,
		},
		"異常系_時間帯枠の識別子が形式に合わない場合_ErrInvalidTimeSlotIDになること": {
			id:         "pass-1",
			timeSlotID: "",
			status:     StatusIssued,
			wantErr:    ErrInvalidTimeSlotID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RestorePriorityPass(
				tt.id,
				validParkID,
				validTicketID,
				validAttractionID,
				tt.timeSlotID,
				tt.status,
				validNow,
				validNow,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RestorePriorityPass(%q, %q, %q) のエラー = %v, want %v",
					tt.id,
					tt.timeSlotID,
					tt.status,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				return
			}

			if got.ID() != tt.id || got.Status() != tt.status {
				t.Errorf("RestorePriorityPass(%q, %q, %q) = (%q, %q), want (%q, %q)",
					tt.id,
					tt.timeSlotID,
					tt.status,
					got.ID(),
					got.Status(),
					tt.id,
					tt.status,
				)
			}
		})
	}
}

// newValidPriorityPassHelper は妥当な値で優先パスを 1 件生成する
func newValidPriorityPassHelper(tb testing.TB) *PriorityPass {
	tb.Helper()

	pass, err := NewPriorityPass(
		validParkID,
		validTicketID,
		validAttractionID,
		validTimeSlotID,
		validNow,
	)
	if err != nil {
		tb.Fatalf("NewPriorityPass() = %v, want nil", err)
	}

	return pass
}
