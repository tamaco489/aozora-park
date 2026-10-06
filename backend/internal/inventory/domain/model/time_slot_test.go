package model

import (
	"errors"
	"testing"
)

func TestRestoreTimeSlot(t *testing.T) {
	tests := map[string]struct {
		parkID       ParkID
		attractionID AttractionID
		id           TimeSlotID
		date         Date
		startTime    string
		capacity     int32
		remaining    int32
		wantErr      error
	}{
		"正常系_すべて範囲内の場合_時間帯枠が組み立てられること": {
			parkID:       "park-1",
			attractionID: "attraction-1",
			id:           "20261005_1000",
			date:         "2026-10-05",
			startTime:    "10:00",
			capacity:     60,
			remaining:    30,
		},
		"境界値_残りが0の場合_時間帯枠が組み立てられること": {
			parkID:       "park-1",
			attractionID: "attraction-1",
			id:           "20261005_1000",
			date:         "2026-10-05",
			startTime:    "10:00",
			capacity:     60,
			remaining:    0,
		},
		"境界値_残りが上限を1超える場合_ErrInvalidRemainingになること": {
			parkID:       "park-1",
			attractionID: "attraction-1",
			id:           "20261005_1000",
			date:         "2026-10-05",
			startTime:    "10:00",
			capacity:     60,
			remaining:    61,
			wantErr:      ErrInvalidRemaining,
		},
		"境界値_上限が0の場合_ErrInvalidCapacityになること": {
			parkID:       "park-1",
			attractionID: "attraction-1",
			id:           "20261005_1000",
			date:         "2026-10-05",
			startTime:    "10:00",
			capacity:     0,
			remaining:    0,
			wantErr:      ErrInvalidCapacity,
		},
		"異常系_パークの識別子が空の場合_ErrInvalidParkIDになること": {
			attractionID: "attraction-1",
			id:           "20261005_1000",
			date:         "2026-10-05",
			startTime:    "10:00",
			capacity:     60,
			remaining:    30,
			wantErr:      ErrInvalidParkID,
		},
		"異常系_アトラクションの識別子が空の場合_ErrInvalidAttractionIDになること": {
			parkID:    "park-1",
			id:        "20261005_1000",
			date:      "2026-10-05",
			startTime: "10:00",
			capacity:  60,
			remaining: 30,
			wantErr:   ErrInvalidAttractionID,
		},
		"異常系_時間帯枠の識別子が空の場合_ErrInvalidTimeSlotIDになること": {
			parkID:       "park-1",
			attractionID: "attraction-1",
			date:         "2026-10-05",
			startTime:    "10:00",
			capacity:     60,
			remaining:    30,
			wantErr:      ErrInvalidTimeSlotID,
		},
		"異常系_日付の形式が違う場合_ErrInvalidDateになること": {
			parkID:       "park-1",
			attractionID: "attraction-1",
			id:           "20261005_1000",
			date:         "2026-10-5",
			startTime:    "10:00",
			capacity:     60,
			remaining:    30,
			wantErr:      ErrInvalidDate,
		},
		"異常系_開始時刻の形式が違う場合_ErrInvalidStartTimeになること": {
			parkID:       "park-1",
			attractionID: "attraction-1",
			id:           "20261005_1000",
			date:         "2026-10-05",
			startTime:    "1000",
			capacity:     60,
			remaining:    30,
			wantErr:      ErrInvalidStartTime,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RestoreTimeSlot(
				tt.parkID,
				tt.attractionID,
				tt.id,
				tt.date,
				tt.startTime,
				tt.capacity,
				tt.remaining,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RestoreTimeSlot(%q, %q, %q, %q, %q, %d, %d) のエラー = %v, want %v",
					tt.parkID,
					tt.attractionID,
					tt.id,
					tt.date,
					tt.startTime,
					tt.capacity,
					tt.remaining,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("RestoreTimeSlot(%q, %q, %q, %q, %q, %d, %d) = %v, want nil",
						tt.parkID,
						tt.attractionID,
						tt.id,
						tt.date,
						tt.startTime,
						tt.capacity,
						tt.remaining,
						got,
					)
				}
				return
			}

			if got.ParkID() != tt.parkID || got.AttractionID() != tt.attractionID || got.ID() != tt.id ||
				got.Date() != tt.date || got.StartTime() != tt.startTime ||
				got.Capacity() != tt.capacity || got.Remaining() != tt.remaining {
				t.Errorf("RestoreTimeSlot(%q, %q, %q, %q, %q, %d, %d) = (%q, %q, %q, %q, %q, %d, %d), want 入力と同じ値",
					tt.parkID,
					tt.attractionID,
					tt.id,
					tt.date,
					tt.startTime,
					tt.capacity,
					tt.remaining,
					got.ParkID(),
					got.AttractionID(),
					got.ID(),
					got.Date(),
					got.StartTime(),
					got.Capacity(),
					got.Remaining(),
				)
			}
		})
	}
}

func TestNewTimeSlotID(t *testing.T) {
	tests := map[string]struct {
		date      Date
		startTime string
		want      TimeSlotID
	}{
		"正常系_日付と開始時刻がある場合_区切りを除いた識別子になること": {
			date:      "2026-10-05",
			startTime: "09:00",
			want:      "20261005_0900",
		},
		"境界値_0時0分の場合_ゼロ埋めが保たれること": {
			date:      "2026-10-05",
			startTime: "00:00",
			want:      "20261005_0000",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewTimeSlotID(tt.date, tt.startTime); got != tt.want {
				t.Errorf("NewTimeSlotID(%q, %q) = %q, want %q",
					tt.date,
					tt.startTime,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestNewTimeSlot(t *testing.T) {
	const (
		parkID       = ParkID("park-1")
		attractionID = AttractionID("attraction-1")
		date         = Date("2026-10-05")
		startTime    = "09:00"
		capacity     = int32(30)
	)

	got, err := NewTimeSlot(
		parkID,
		attractionID,
		date,
		startTime,
		capacity,
	)
	if err != nil {
		t.Fatalf("NewTimeSlot(%q, %q, %q, %q, %d) のエラー = %v, want nil",
			parkID,
			attractionID,
			date,
			startTime,
			capacity,
			err,
		)
	}

	// 識別子は日付と開始時刻から決まるため、同じ枠を二度生成しても同じ値になる
	wantID := TimeSlotID("20261005_0900")
	if got.ID() != wantID || got.Capacity() != capacity || got.Remaining() != capacity {
		t.Errorf("NewTimeSlot(%q, %q, %q, %q, %d) = (%q, %d, %d), want (%q, %d, %d)",
			parkID,
			attractionID,
			date,
			startTime,
			capacity,
			got.ID(),
			got.Capacity(),
			got.Remaining(),
			wantID,
			capacity,
			capacity,
		)
	}
}
