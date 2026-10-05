package model

import (
	"errors"
	"testing"
)

func TestValidateDate(t *testing.T) {
	tests := map[string]struct {
		date    Date
		wantErr error
	}{
		"正常系_YYYY-MM-DDの場合_エラーにならないこと": {
			date: "2026-10-05",
		},
		"境界値_うるう年の2月29日の場合_エラーにならないこと": {
			date: "2028-02-29",
		},
		"境界値_うるう年でない2月29日の場合_ErrInvalidDateになること": {
			date:    "2026-02-29",
			wantErr: ErrInvalidDate,
		},
		"異常系_月日が1桁の場合_ErrInvalidDateになること": {
			date:    "2026-1-5",
			wantErr: ErrInvalidDate,
		},
		"異常系_区切りがない場合_ErrInvalidDateになること": {
			date:    "20261005",
			wantErr: ErrInvalidDate,
		},
		"異常系_13月の場合_ErrInvalidDateになること": {
			date:    "2026-13-01",
			wantErr: ErrInvalidDate,
		},
		"異常系_空の場合_ErrInvalidDateになること": {
			date:    "",
			wantErr: ErrInvalidDate,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateDate(tt.date); !errors.Is(err, tt.wantErr) {
				t.Errorf("validateDate(%q) = %v, want %v",
					tt.date,
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestValidateStartTime(t *testing.T) {
	tests := map[string]struct {
		startTime string
		wantErr   error
	}{
		"正常系_HH:MMの場合_エラーにならないこと": {
			startTime: "10:00",
		},
		"境界値_0時0分の場合_エラーにならないこと": {
			startTime: "00:00",
		},
		"境界値_23時59分の場合_エラーにならないこと": {
			startTime: "23:59",
		},
		"境界値_24時の場合_ErrInvalidStartTimeになること": {
			startTime: "24:00",
			wantErr:   ErrInvalidStartTime,
		},
		"異常系_時が1桁の場合_ErrInvalidStartTimeになること": {
			startTime: "9:00",
			wantErr:   ErrInvalidStartTime,
		},
		"異常系_秒まである場合_ErrInvalidStartTimeになること": {
			startTime: "10:00:00",
			wantErr:   ErrInvalidStartTime,
		},
		"異常系_空の場合_ErrInvalidStartTimeになること": {
			startTime: "",
			wantErr:   ErrInvalidStartTime,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateStartTime(tt.startTime); !errors.Is(err, tt.wantErr) {
				t.Errorf("validateStartTime(%q) = %v, want %v",
					tt.startTime,
					err,
					tt.wantErr,
				)
			}
		})
	}
}
