package model

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRestoreParkMaster(t *testing.T) {
	tests := map[string]struct {
		id                   ParkID
		defaultDailyCapacity int32
		inventoryDays        int32
		wantErr              error
	}{
		"正常系_すべて範囲内の場合_パークの設定が組み立てられること": {
			id:                   "park-1",
			defaultDailyCapacity: 1000,
			inventoryDays:        14,
		},
		"境界値_日数が1の場合_パークの設定が組み立てられること": {
			id:                   "park-1",
			defaultDailyCapacity: 1000,
			inventoryDays:        1,
		},
		"境界値_日数が90の場合_パークの設定が組み立てられること": {
			id:                   "park-1",
			defaultDailyCapacity: 1000,
			inventoryDays:        90,
		},
		"境界値_日数が0の場合_ErrInvalidInventoryDaysになること": {
			id:                   "park-1",
			defaultDailyCapacity: 1000,
			inventoryDays:        0,
			wantErr:              ErrInvalidInventoryDays,
		},
		"境界値_日数が91の場合_ErrInvalidInventoryDaysになること": {
			id:                   "park-1",
			defaultDailyCapacity: 1000,
			inventoryDays:        91,
			wantErr:              ErrInvalidInventoryDays,
		},
		"境界値_1日の上限が0の場合_ErrInvalidCapacityになること": {
			id:                   "park-1",
			defaultDailyCapacity: 0,
			inventoryDays:        14,
			wantErr:              ErrInvalidCapacity,
		},
		"異常系_パークの識別子が空の場合_ErrInvalidParkIDになること": {
			defaultDailyCapacity: 1000,
			inventoryDays:        14,
			wantErr:              ErrInvalidParkID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RestoreParkMaster(
				tt.id,
				tt.defaultDailyCapacity,
				tt.inventoryDays,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RestoreParkMaster(%q, %d, %d) のエラー = %v, want %v",
					tt.id,
					tt.defaultDailyCapacity,
					tt.inventoryDays,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("RestoreParkMaster(%q, %d, %d) = %v, want nil",
						tt.id,
						tt.defaultDailyCapacity,
						tt.inventoryDays,
						got,
					)
				}
				return
			}

			if got.ID() != tt.id || got.DefaultDailyCapacity() != tt.defaultDailyCapacity ||
				got.InventoryDays() != tt.inventoryDays {
				t.Errorf("RestoreParkMaster(%q, %d, %d) = (%q, %d, %d), want 入力と同じ値",
					tt.id,
					tt.defaultDailyCapacity,
					tt.inventoryDays,
					got.ID(),
					got.DefaultDailyCapacity(),
					got.InventoryDays(),
				)
			}
		})
	}
}

func TestRestoreAttractionMaster(t *testing.T) {
	tests := map[string]struct {
		parkID          ParkID
		id              AttractionID
		enabled         bool
		startTime       string
		endTime         string
		intervalMinutes int32
		capacityPerSlot int32
		wantErr         error
	}{
		"正常系_すべて範囲内の場合_アトラクションの設定が組み立てられること": {
			parkID:          "park-1",
			id:              "attraction-1",
			enabled:         true,
			startTime:       "09:00",
			endTime:         "17:00",
			intervalMinutes: 60,
			capacityPerSlot: 30,
		},
		"正常系_無効な場合も_時刻の条件が検証されること": {
			parkID:          "park-1",
			id:              "attraction-1",
			enabled:         false,
			startTime:       "17:00",
			endTime:         "09:00",
			intervalMinutes: 60,
			capacityPerSlot: 30,
			wantErr:         ErrInvalidTimeRange,
		},
		"境界値_開始時刻と終了時刻が同じ場合_ErrInvalidTimeRangeになること": {
			parkID:          "park-1",
			id:              "attraction-1",
			enabled:         true,
			startTime:       "09:00",
			endTime:         "09:00",
			intervalMinutes: 60,
			capacityPerSlot: 30,
			wantErr:         ErrInvalidTimeRange,
		},
		"境界値_間隔が0の場合_ErrInvalidIntervalMinutesになること": {
			parkID:          "park-1",
			id:              "attraction-1",
			enabled:         true,
			startTime:       "09:00",
			endTime:         "17:00",
			intervalMinutes: 0,
			capacityPerSlot: 30,
			wantErr:         ErrInvalidIntervalMinutes,
		},
		"境界値_1枠の上限が0の場合_ErrInvalidCapacityになること": {
			parkID:          "park-1",
			id:              "attraction-1",
			enabled:         true,
			startTime:       "09:00",
			endTime:         "17:00",
			intervalMinutes: 60,
			capacityPerSlot: 0,
			wantErr:         ErrInvalidCapacity,
		},
		"異常系_パークの識別子が空の場合_ErrInvalidParkIDになること": {
			id:              "attraction-1",
			enabled:         true,
			startTime:       "09:00",
			endTime:         "17:00",
			intervalMinutes: 60,
			capacityPerSlot: 30,
			wantErr:         ErrInvalidParkID,
		},
		"異常系_アトラクションの識別子が空の場合_ErrInvalidAttractionIDになること": {
			parkID:          "park-1",
			enabled:         true,
			startTime:       "09:00",
			endTime:         "17:00",
			intervalMinutes: 60,
			capacityPerSlot: 30,
			wantErr:         ErrInvalidAttractionID,
		},
		"異常系_開始時刻がゼロ埋めされていない場合_ErrInvalidStartTimeになること": {
			parkID:          "park-1",
			id:              "attraction-1",
			enabled:         true,
			startTime:       "9:00",
			endTime:         "17:00",
			intervalMinutes: 60,
			capacityPerSlot: 30,
			wantErr:         ErrInvalidStartTime,
		},
		"異常系_終了時刻の形式が違う場合_ErrInvalidEndTimeになること": {
			parkID:          "park-1",
			id:              "attraction-1",
			enabled:         true,
			startTime:       "09:00",
			endTime:         "1700",
			intervalMinutes: 60,
			capacityPerSlot: 30,
			wantErr:         ErrInvalidEndTime,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RestoreAttractionMaster(
				tt.parkID,
				tt.id,
				tt.enabled,
				tt.startTime,
				tt.endTime,
				tt.intervalMinutes,
				tt.capacityPerSlot,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RestoreAttractionMaster(%q, %q, %t, %q, %q, %d, %d) のエラー = %v, want %v",
					tt.parkID,
					tt.id,
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
				if got != nil {
					t.Errorf("RestoreAttractionMaster(%q, %q, %t, %q, %q, %d, %d) = %v, want nil",
						tt.parkID,
						tt.id,
						tt.enabled,
						tt.startTime,
						tt.endTime,
						tt.intervalMinutes,
						tt.capacityPerSlot,
						got,
					)
				}
				return
			}

			if got.ParkID() != tt.parkID || got.ID() != tt.id ||
				got.PriorityPassEnabled() != tt.enabled ||
				got.StartTime() != tt.startTime || got.EndTime() != tt.endTime ||
				got.IntervalMinutes() != tt.intervalMinutes || got.CapacityPerSlot() != tt.capacityPerSlot {
				t.Errorf("RestoreAttractionMaster(%q, %q, %t, %q, %q, %d, %d) = (%q, %q, %t, %q, %q, %d, %d), want 入力と同じ値",
					tt.parkID,
					tt.id,
					tt.enabled,
					tt.startTime,
					tt.endTime,
					tt.intervalMinutes,
					tt.capacityPerSlot,
					got.ParkID(),
					got.ID(),
					got.PriorityPassEnabled(),
					got.StartTime(),
					got.EndTime(),
					got.IntervalMinutes(),
					got.CapacityPerSlot(),
				)
			}
		})
	}
}

func TestAttractionMasterStartTimes(t *testing.T) {
	tests := map[string]struct {
		startTime       string
		endTime         string
		intervalMinutes int32
		want            []string
	}{
		"正常系_9時から17時を60分で刻む場合_終了時刻を含まない8枠になること": {
			startTime:       "09:00",
			endTime:         "17:00",
			intervalMinutes: 60,
			want: []string{
				"09:00",
				"10:00",
				"11:00",
				"12:00",
				"13:00",
				"14:00",
				"15:00",
				"16:00",
			},
		},
		"正常系_30分で刻む場合_ゼロ埋めされた開始時刻が昇順に並ぶこと": {
			startTime:       "09:00",
			endTime:         "10:30",
			intervalMinutes: 30,
			want: []string{
				"09:00",
				"09:30",
				"10:00",
			},
		},
		"境界値_間隔で割り切れない場合_終了時刻を越える枠を作らないこと": {
			startTime:       "09:00",
			endTime:         "10:00",
			intervalMinutes: 45,
			want: []string{
				"09:00",
				"09:45",
			},
		},
		"境界値_間隔が終了時刻までの長さと同じ場合_1枠になること": {
			startTime:       "09:00",
			endTime:         "10:00",
			intervalMinutes: 60,
			want:            []string{"09:00"},
		},
		"境界値_間隔が終了時刻までの長さより長い場合_開始時刻の1枠だけになること": {
			startTime:       "09:00",
			endTime:         "10:00",
			intervalMinutes: 120,
			want:            []string{"09:00"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			attraction, err := RestoreAttractionMaster(
				"park-1",
				"attraction-1",
				true,
				tt.startTime,
				tt.endTime,
				tt.intervalMinutes,
				30,
			)
			if err != nil {
				t.Fatalf("RestoreAttractionMaster(%q, %q, %d) = %v, want nil",
					tt.startTime,
					tt.endTime,
					tt.intervalMinutes,
					err,
				)
			}

			got := attraction.StartTimes()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AttractionMaster.StartTimes() の差分 (-want +got):\n%s", diff)
			}
		})
	}
}
