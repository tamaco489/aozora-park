package model

import (
	"errors"
	"strings"
	"testing"
)

// validConfigHelper はテストで使う妥当な優先パスの条件を組み立てる
func validConfigHelper(tb testing.TB) PriorityPassConfig {
	tb.Helper()

	config, err := NewPriorityPassConfig(
		true,
		"09:00",
		"18:00",
		30,
		10,
	)
	if err != nil {
		tb.Fatalf("NewPriorityPassConfig() = %v, want nil", err)
	}

	return config
}

func TestNewPriorityPassConfig(t *testing.T) {
	tests := map[string]struct {
		enabled         bool
		startTime       string
		endTime         string
		intervalMinutes int32
		capacityPerSlot int32
		wantErr         error
	}{
		"正常系_すべて範囲内の場合_条件が生成されること": {
			enabled:         true,
			startTime:       "09:00",
			endTime:         "18:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
		},
		"正常系_無効でも時刻が妥当な場合_条件が生成されること": {
			enabled:         false,
			startTime:       "09:00",
			endTime:         "18:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
		},
		"境界値_時刻が0時0分と23時59分の場合_条件が生成されること": {
			enabled:         true,
			startTime:       "00:00",
			endTime:         "23:59",
			intervalMinutes: 1,
			capacityPerSlot: 1,
		},
		"境界値_終了時刻が開始時刻と等しい場合_ErrAttractionInvalidTimeRangeになること": {
			enabled:         true,
			startTime:       "09:00",
			endTime:         "09:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
			wantErr:         ErrAttractionInvalidTimeRange,
		},
		"異常系_終了時刻が開始時刻より前の場合_ErrAttractionInvalidTimeRangeになること": {
			enabled:         true,
			startTime:       "18:00",
			endTime:         "09:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
			wantErr:         ErrAttractionInvalidTimeRange,
		},
		"異常系_時刻が空の場合_ErrAttractionInvalidTimeになること": {
			enabled:         true,
			startTime:       "",
			endTime:         "18:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
			wantErr:         ErrAttractionInvalidTime,
		},
		"異常系_時刻の時が24の場合_ErrAttractionInvalidTimeになること": {
			enabled:         true,
			startTime:       "09:00",
			endTime:         "24:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
			wantErr:         ErrAttractionInvalidTime,
		},
		"異常系_時刻の分が60の場合_ErrAttractionInvalidTimeになること": {
			enabled:         true,
			startTime:       "09:60",
			endTime:         "18:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
			wantErr:         ErrAttractionInvalidTime,
		},
		"異常系_時刻が1桁の場合_ErrAttractionInvalidTimeになること": {
			enabled:         true,
			startTime:       "9:00",
			endTime:         "18:00",
			intervalMinutes: 30,
			capacityPerSlot: 10,
			wantErr:         ErrAttractionInvalidTime,
		},
		"境界値_間隔が0分の場合_ErrAttractionInvalidIntervalMinutesになること": {
			enabled:         true,
			startTime:       "09:00",
			endTime:         "18:00",
			intervalMinutes: 0,
			capacityPerSlot: 10,
			wantErr:         ErrAttractionInvalidIntervalMinutes,
		},
		"境界値_1枠あたりの上限が0枚の場合_ErrAttractionInvalidCapacityPerSlotになること": {
			enabled:         true,
			startTime:       "09:00",
			endTime:         "18:00",
			intervalMinutes: 30,
			capacityPerSlot: 0,
			wantErr:         ErrAttractionInvalidCapacityPerSlot,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewPriorityPassConfig(
				tt.enabled,
				tt.startTime,
				tt.endTime,
				tt.intervalMinutes,
				tt.capacityPerSlot,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewPriorityPassConfig(%t, %q, %q, %d, %d) のエラー = %v, want %v",
					tt.enabled,
					tt.startTime,
					tt.endTime,
					tt.intervalMinutes,
					tt.capacityPerSlot,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != (PriorityPassConfig{}) {
					t.Errorf("NewPriorityPassConfig(...) = %+v, want ゼロ値", got)
				}
				return
			}

			if got.Enabled() != tt.enabled {
				t.Errorf("NewPriorityPassConfig(...) の Enabled = %t, want %t",
					got.Enabled(),
					tt.enabled,
				)
			}
			if got.StartTime() != tt.startTime {
				t.Errorf("NewPriorityPassConfig(...) の StartTime = %q, want %q",
					got.StartTime(),
					tt.startTime,
				)
			}
			if got.EndTime() != tt.endTime {
				t.Errorf("NewPriorityPassConfig(...) の EndTime = %q, want %q",
					got.EndTime(),
					tt.endTime,
				)
			}
			if got.IntervalMinutes() != tt.intervalMinutes {
				t.Errorf("NewPriorityPassConfig(...) の IntervalMinutes = %d, want %d",
					got.IntervalMinutes(),
					tt.intervalMinutes,
				)
			}
			if got.CapacityPerSlot() != tt.capacityPerSlot {
				t.Errorf("NewPriorityPassConfig(...) の CapacityPerSlot = %d, want %d",
					got.CapacityPerSlot(),
					tt.capacityPerSlot,
				)
			}
		})
	}
}

func TestNewAttraction(t *testing.T) {
	tests := map[string]struct {
		parkID  ParkID
		name    string
		wantErr error
	}{
		"正常系_すべて範囲内の場合_アトラクションが生成されること": {
			parkID: "park-1",
			name:   "ジェットコースター",
		},
		"境界値_表示名が100文字の場合_アトラクションが生成されること": {
			parkID: "park-1",
			name:   strings.Repeat("あ", 100),
		},
		"境界値_表示名が101文字の場合_ErrAttractionInvalidNameになること": {
			parkID:  "park-1",
			name:    strings.Repeat("あ", 101),
			wantErr: ErrAttractionInvalidName,
		},
		"異常系_表示名が空の場合_ErrAttractionInvalidNameになること": {
			parkID:  "park-1",
			name:    "",
			wantErr: ErrAttractionInvalidName,
		},
		"異常系_パークの識別子が空の場合_ErrInvalidIDになること": {
			parkID:  "",
			name:    "ジェットコースター",
			wantErr: ErrInvalidID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewAttraction(
				tt.parkID,
				tt.name,
				validConfigHelper(t),
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewAttraction(%q, %q, ...) のエラー = %v, want %v",
					tt.parkID,
					tt.name,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("NewAttraction(%q, %q, ...) = %v, want nil",
						tt.parkID,
						tt.name,
						got,
					)
				}
				return
			}

			if got.ID() == "" {
				t.Error("NewAttraction() の ID が空、採番されていない")
			}
			if got.ParkID() != tt.parkID {
				t.Errorf("NewAttraction() の ParkID = %q, want %q",
					got.ParkID(),
					tt.parkID,
				)
			}
			if got.Name() != tt.name {
				t.Errorf("NewAttraction() の Name = %q, want %q",
					got.Name(),
					tt.name,
				)
			}
		})
	}
}

func TestNewAttractionRejectsZeroConfig(t *testing.T) {
	// ゼロ値の条件は検証を通っていないため、Attraction が受け取った時点で弾く
	got, err := NewAttraction(
		"park-1",
		"ジェットコースター",
		PriorityPassConfig{},
	)

	if !errors.Is(err, ErrAttractionInvalidTime) {
		t.Fatalf("NewAttraction(..., PriorityPassConfig{}) のエラー = %v, want %v",
			err,
			ErrAttractionInvalidTime,
		)
	}
	if got != nil {
		t.Errorf("NewAttraction(..., PriorityPassConfig{}) = %v, want nil", got)
	}
}

func TestNewAttractionGeneratesDistinctIDs(t *testing.T) {
	config := validConfigHelper(t)

	first, err := NewAttraction(
		"park-1",
		"ジェットコースター",
		config,
	)
	if err != nil {
		t.Fatalf("NewAttraction() = %v, want nil", err)
	}

	second, err := NewAttraction(
		"park-1",
		"ジェットコースター",
		config,
	)
	if err != nil {
		t.Fatalf("NewAttraction() = %v, want nil", err)
	}

	if first.ID() == second.ID() {
		t.Errorf("NewAttraction() の ID = %q, want 異なる値", first.ID())
	}
}

func TestRestoreAttraction(t *testing.T) {
	tests := map[string]struct {
		id      AttractionID
		wantErr error
	}{
		"正常系_識別子がある場合_アトラクションが組み立てられること": {
			id: "attraction-1",
		},
		"異常系_識別子が空の場合_ErrAttractionInvalidIDになること": {
			id:      "",
			wantErr: ErrAttractionInvalidID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RestoreAttraction(
				"park-1",
				tt.id,
				"ジェットコースター",
				validConfigHelper(t),
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RestoreAttraction(%q, ...) のエラー = %v, want %v",
					tt.id,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr == nil && got.ID() != tt.id {
				t.Errorf("RestoreAttraction(%q, ...) の ID = %q, want %q",
					tt.id,
					got.ID(),
					tt.id,
				)
			}
		})
	}
}

func TestAttraction_Update(t *testing.T) {
	tests := map[string]struct {
		name            string
		startTime       string
		endTime         string
		intervalMinutes int32
		capacityPerSlot int32
		wantErr         error
	}{
		"正常系_すべて範囲内の場合_値が入れ替わること": {
			name:            "ジェットコースター 2",
			startTime:       "10:00",
			endTime:         "20:00",
			intervalMinutes: 15,
			capacityPerSlot: 20,
		},
		"異常系_表示名が空の場合_ErrAttractionInvalidNameになること": {
			name:            "",
			startTime:       "10:00",
			endTime:         "20:00",
			intervalMinutes: 15,
			capacityPerSlot: 20,
			wantErr:         ErrAttractionInvalidName,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			attraction, err := RestoreAttraction(
				"park-1",
				"attraction-1",
				"ジェットコースター",
				validConfigHelper(t),
			)
			if err != nil {
				t.Fatalf("RestoreAttraction() = %v, want nil", err)
			}

			config, err := NewPriorityPassConfig(
				false,
				tt.startTime,
				tt.endTime,
				tt.intervalMinutes,
				tt.capacityPerSlot,
			)
			if err != nil {
				t.Fatalf("NewPriorityPassConfig() = %v, want nil", err)
			}

			err = attraction.Update(tt.name, config)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Attraction.Update(%q, %+v) = %v, want %v",
					tt.name,
					config,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				// 検証に失敗したときは元の値を保つ
				if attraction.Name() != "ジェットコースター" || attraction.PriorityPassConfig() != validConfigHelper(t) {
					t.Errorf("Attraction.Update() の失敗後 = (%q, %+v), want (%q, %+v)",
						attraction.Name(),
						attraction.PriorityPassConfig(),
						"ジェットコースター",
						validConfigHelper(t),
					)
				}
				return
			}

			if attraction.Name() != tt.name {
				t.Errorf("Attraction.Update() 後の Name = %q, want %q",
					attraction.Name(),
					tt.name,
				)
			}
			if attraction.PriorityPassConfig() != config {
				t.Errorf("Attraction.Update() 後の PriorityPassConfig = %+v, want %+v",
					attraction.PriorityPassConfig(),
					config,
				)
			}
			if attraction.ID() != "attraction-1" || attraction.ParkID() != "park-1" {
				t.Errorf("Attraction.Update() 後の識別子 = (%q, %q), want (%q, %q)",
					attraction.ParkID(),
					attraction.ID(),
					"park-1",
					"attraction-1",
				)
			}
		})
	}
}
