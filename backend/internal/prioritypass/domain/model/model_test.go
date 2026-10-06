package model

import (
	"errors"
	"testing"
)

func TestValidateTimeSlotID(t *testing.T) {
	tests := map[string]struct {
		id      TimeSlotID
		wantErr error
	}{
		"正常系_YYYYMMDD_HHMMの場合_エラーにならないこと": {
			id: "20261005_1000",
		},
		"境界値_0時0分の場合_エラーにならないこと": {
			id: "20261005_0000",
		},
		"異常系_区切りがない場合_ErrInvalidTimeSlotIDになること": {
			id:      "202610051000",
			wantErr: ErrInvalidTimeSlotID,
		},
		"異常系_区切りがハイフンの場合_ErrInvalidTimeSlotIDになること": {
			id:      "20261005-1000",
			wantErr: ErrInvalidTimeSlotID,
		},
		"異常系_時刻の桁が足りない場合_ErrInvalidTimeSlotIDになること": {
			id:      "20261005_100",
			wantErr: ErrInvalidTimeSlotID,
		},
		"異常系_日付の桁が多い場合_ErrInvalidTimeSlotIDになること": {
			id:      "202610051_1000",
			wantErr: ErrInvalidTimeSlotID,
		},
		"異常系_数字以外を含む場合_ErrInvalidTimeSlotIDになること": {
			id:      "2026100a_1000",
			wantErr: ErrInvalidTimeSlotID,
		},
		"異常系_空の場合_ErrInvalidTimeSlotIDになること": {
			id:      "",
			wantErr: ErrInvalidTimeSlotID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateTimeSlotID(tt.id); !errors.Is(err, tt.wantErr) {
				t.Errorf("validateTimeSlotID(%q) = %v, want %v",
					tt.id,
					err,
					tt.wantErr,
				)
			}
		})
	}
}
