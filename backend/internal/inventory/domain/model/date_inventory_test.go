package model

import (
	"errors"
	"testing"
)

func TestRestoreDateInventory(t *testing.T) {
	tests := map[string]struct {
		parkID    ParkID
		date      Date
		capacity  int32
		remaining int32
		wantErr   error
	}{
		"正常系_すべて範囲内の場合_入場枠が組み立てられること": {
			parkID:    "park-1",
			date:      "2026-10-05",
			capacity:  1000,
			remaining: 800,
		},
		"境界値_残りが0の場合_入場枠が組み立てられること": {
			parkID:    "park-1",
			date:      "2026-10-05",
			capacity:  1000,
			remaining: 0,
		},
		"境界値_残りが上限と等しい場合_入場枠が組み立てられること": {
			parkID:    "park-1",
			date:      "2026-10-05",
			capacity:  1000,
			remaining: 1000,
		},
		"境界値_残りが上限を1超える場合_ErrInvalidRemainingになること": {
			parkID:    "park-1",
			date:      "2026-10-05",
			capacity:  1000,
			remaining: 1001,
			wantErr:   ErrInvalidRemaining,
		},
		"異常系_残りが負の場合_ErrInvalidRemainingになること": {
			parkID:    "park-1",
			date:      "2026-10-05",
			capacity:  1000,
			remaining: -1,
			wantErr:   ErrInvalidRemaining,
		},
		"境界値_上限が0の場合_ErrInvalidCapacityになること": {
			parkID:    "park-1",
			date:      "2026-10-05",
			capacity:  0,
			remaining: 0,
			wantErr:   ErrInvalidCapacity,
		},
		"異常系_パークの識別子が空の場合_ErrInvalidParkIDになること": {
			parkID:    "",
			date:      "2026-10-05",
			capacity:  1000,
			remaining: 800,
			wantErr:   ErrInvalidParkID,
		},
		"異常系_日付の形式が違う場合_ErrInvalidDateになること": {
			parkID:    "park-1",
			date:      "2026/10/05",
			capacity:  1000,
			remaining: 800,
			wantErr:   ErrInvalidDate,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RestoreDateInventory(
				tt.parkID,
				tt.date,
				tt.capacity,
				tt.remaining,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RestoreDateInventory(%q, %q, %d, %d) のエラー = %v, want %v",
					tt.parkID,
					tt.date,
					tt.capacity,
					tt.remaining,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("RestoreDateInventory(%q, %q, %d, %d) = %v, want nil",
						tt.parkID,
						tt.date,
						tt.capacity,
						tt.remaining,
						got,
					)
				}
				return
			}

			if got.ParkID() != tt.parkID || got.Date() != tt.date ||
				got.Capacity() != tt.capacity || got.Remaining() != tt.remaining {
				t.Errorf("RestoreDateInventory(%q, %q, %d, %d) = (%q, %q, %d, %d), want (%q, %q, %d, %d)",
					tt.parkID,
					tt.date,
					tt.capacity,
					tt.remaining,
					got.ParkID(),
					got.Date(),
					got.Capacity(),
					got.Remaining(),
					tt.parkID,
					tt.date,
					tt.capacity,
					tt.remaining,
				)
			}
		})
	}
}

func TestDateInventoryOverwrite(t *testing.T) {
	tests := map[string]struct {
		capacity  int32
		remaining int32
		wantErr   error
	}{
		"正常系_範囲内の場合_上限と残りが入れ替わること": {
			capacity:  2000,
			remaining: 1500,
		},
		"境界値_残りが上限と等しい場合_入れ替わること": {
			capacity:  2000,
			remaining: 2000,
		},
		"境界値_残りが上限を1超える場合_ErrInvalidRemainingになること": {
			capacity:  2000,
			remaining: 2001,
			wantErr:   ErrInvalidRemaining,
		},
		"異常系_残りが負の場合_ErrInvalidRemainingになること": {
			capacity:  2000,
			remaining: -1,
			wantErr:   ErrInvalidRemaining,
		},
		"境界値_上限が0の場合_ErrInvalidCapacityになること": {
			capacity:  0,
			remaining: 0,
			wantErr:   ErrInvalidCapacity,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inventory, err := RestoreDateInventory(
				"park-1",
				"2026-10-05",
				1000,
				800,
			)
			if err != nil {
				t.Fatalf("RestoreDateInventory() = %v, want nil", err)
			}

			err = inventory.Overwrite(tt.capacity, tt.remaining)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DateInventory.Overwrite(%d, %d) = %v, want %v",
					tt.capacity,
					tt.remaining,
					err,
					tt.wantErr,
				)
			}

			wantCapacity, wantRemaining := tt.capacity, tt.remaining
			if tt.wantErr != nil {
				// 検証に失敗したときは元の値を保つ
				wantCapacity, wantRemaining = 1000, 800
			}

			if inventory.Capacity() != wantCapacity || inventory.Remaining() != wantRemaining {
				t.Errorf("DateInventory.Overwrite(%d, %d) の後 = (%d, %d), want (%d, %d)",
					tt.capacity,
					tt.remaining,
					inventory.Capacity(),
					inventory.Remaining(),
					wantCapacity,
					wantRemaining,
				)
			}
		})
	}
}
