package model

import (
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := map[string]struct {
		name                 string
		defaultDailyCapacity int32
		inventoryDays        int32
		wantErr              error
	}{
		"正常系_すべて範囲内の場合_パークが作られること": {
			name:                 "Aozora Park",
			defaultDailyCapacity: 1000,
			inventoryDays:        30,
		},
		"境界値_表示名が1文字の場合_パークが作られること": {
			name:                 "あ",
			defaultDailyCapacity: 1,
			inventoryDays:        1,
		},
		"境界値_表示名が100文字の場合_パークが作られること": {
			name:                 strings.Repeat("あ", 100),
			defaultDailyCapacity: 1,
			inventoryDays:        90,
		},
		"境界値_表示名が101文字の場合_ErrInvalidNameになること": {
			name:                 strings.Repeat("あ", 101),
			defaultDailyCapacity: 1,
			inventoryDays:        1,
			wantErr:              ErrInvalidName,
		},
		"異常系_表示名が空の場合_ErrInvalidNameになること": {
			name:                 "",
			defaultDailyCapacity: 1,
			inventoryDays:        1,
			wantErr:              ErrInvalidName,
		},
		"境界値_上限人数が0の場合_ErrInvalidDailyCapacityになること": {
			name:                 "Aozora Park",
			defaultDailyCapacity: 0,
			inventoryDays:        1,
			wantErr:              ErrInvalidDailyCapacity,
		},
		"異常系_上限人数が負の場合_ErrInvalidDailyCapacityになること": {
			name:                 "Aozora Park",
			defaultDailyCapacity: -1,
			inventoryDays:        1,
			wantErr:              ErrInvalidDailyCapacity,
		},
		"境界値_生成日数が0の場合_ErrInvalidInventoryDaysになること": {
			name:                 "Aozora Park",
			defaultDailyCapacity: 1,
			inventoryDays:        0,
			wantErr:              ErrInvalidInventoryDays,
		},
		"境界値_生成日数が91日の場合_ErrInvalidInventoryDaysになること": {
			name:                 "Aozora Park",
			defaultDailyCapacity: 1,
			inventoryDays:        91,
			wantErr:              ErrInvalidInventoryDays,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.name, tt.defaultDailyCapacity, tt.inventoryDays)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New(%q, %d, %d) のエラー = %v, want %v", tt.name, tt.defaultDailyCapacity, tt.inventoryDays, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("New(%q, %d, %d) = %v, want nil", tt.name, tt.defaultDailyCapacity, tt.inventoryDays, got)
				}
				return
			}

			if got.ID() == "" {
				t.Error("New() の ID が空、採番されていない")
			}
			if got.Name() != tt.name {
				t.Errorf("New() の Name = %q, want %q", got.Name(), tt.name)
			}
			if got.DefaultDailyCapacity() != tt.defaultDailyCapacity {
				t.Errorf("New() の DefaultDailyCapacity = %d, want %d", got.DefaultDailyCapacity(), tt.defaultDailyCapacity)
			}
			if got.InventoryDays() != tt.inventoryDays {
				t.Errorf("New() の InventoryDays = %d, want %d", got.InventoryDays(), tt.inventoryDays)
			}
		})
	}
}

func TestNewGeneratesDistinctIDs(t *testing.T) {
	first, err := New("Aozora Park", 1000, 30)
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	second, err := New("Aozora Park", 1000, 30)
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	if first.ID() == second.ID() {
		t.Errorf("New() の ID = %q, want 異なる値", first.ID())
	}
}

func TestRestore(t *testing.T) {
	tests := map[string]struct {
		id      ParkID
		wantErr error
	}{
		"正常系_識別子がある場合_パークが組み立てられること": {
			id: "park-1",
		},
		"異常系_識別子が空の場合_ErrInvalidIDになること": {
			id:      "",
			wantErr: ErrInvalidID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Restore(tt.id, "Aozora Park", 1000, 30)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Restore(%q, ...) のエラー = %v, want %v", tt.id, err, tt.wantErr)
			}

			if tt.wantErr == nil && got.ID() != tt.id {
				t.Errorf("Restore(%q, ...) の ID = %q, want %q", tt.id, got.ID(), tt.id)
			}
		})
	}
}

func TestPark_Update(t *testing.T) {
	tests := map[string]struct {
		name                 string
		defaultDailyCapacity int32
		inventoryDays        int32
		wantErr              error
	}{
		"正常系_すべて範囲内の場合_値が入れ替わること": {
			name:                 "Aozora Park 2",
			defaultDailyCapacity: 2000,
			inventoryDays:        60,
		},
		"異常系_生成日数が範囲外の場合_ErrInvalidInventoryDaysになること": {
			name:                 "Aozora Park 2",
			defaultDailyCapacity: 2000,
			inventoryDays:        91,
			wantErr:              ErrInvalidInventoryDays,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			park, err := Restore("park-1", "Aozora Park", 1000, 30)
			if err != nil {
				t.Fatalf("Restore() = %v, want nil", err)
			}

			err = park.Update(tt.name, tt.defaultDailyCapacity, tt.inventoryDays)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Park.Update(%q, %d, %d) = %v, want %v", tt.name, tt.defaultDailyCapacity, tt.inventoryDays, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				// 検証に失敗したときは元の値を保つ
				if park.Name() != "Aozora Park" || park.DefaultDailyCapacity() != 1000 || park.InventoryDays() != 30 {
					t.Errorf("Park.Update() の失敗後 = (%q, %d, %d), want (%q, %d, %d)",
						park.Name(), park.DefaultDailyCapacity(), park.InventoryDays(), "Aozora Park", 1000, 30)
				}
				return
			}

			if park.Name() != tt.name {
				t.Errorf("Park.Update() 後の Name = %q, want %q", park.Name(), tt.name)
			}
			if park.DefaultDailyCapacity() != tt.defaultDailyCapacity {
				t.Errorf("Park.Update() 後の DefaultDailyCapacity = %d, want %d", park.DefaultDailyCapacity(), tt.defaultDailyCapacity)
			}
			if park.InventoryDays() != tt.inventoryDays {
				t.Errorf("Park.Update() 後の InventoryDays = %d, want %d", park.InventoryDays(), tt.inventoryDays)
			}
			if park.ID() != "park-1" {
				t.Errorf("Park.Update() 後の ID = %q, want %q", park.ID(), "park-1")
			}
		})
	}
}
