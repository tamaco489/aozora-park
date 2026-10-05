package model

import (
	"errors"
	"strings"
	"testing"
)

func TestNewTicketType(t *testing.T) {
	tests := map[string]struct {
		parkID        ParkID
		name          string
		price         int64
		entryTimeFrom string
		entryTimeTo   string
		wantErr       error
	}{
		"正常系_すべて範囲内の場合_券種が生成されること": {
			parkID:        "park-1",
			name:          "1 デーパスポート",
			price:         8000,
			entryTimeFrom: "09:00",
			entryTimeTo:   "21:00",
		},
		"境界値_価格が0円の場合_券種が生成されること": {
			parkID:        "park-1",
			name:          "招待券",
			price:         0,
			entryTimeFrom: "00:00",
			entryTimeTo:   "23:59",
		},
		"境界値_表示名が100文字の場合_券種が生成されること": {
			parkID:        "park-1",
			name:          strings.Repeat("あ", 100),
			price:         8000,
			entryTimeFrom: "09:00",
			entryTimeTo:   "21:00",
		},
		"境界値_表示名が101文字の場合_ErrTicketTypeInvalidNameになること": {
			parkID:        "park-1",
			name:          strings.Repeat("あ", 101),
			price:         8000,
			entryTimeFrom: "09:00",
			entryTimeTo:   "21:00",
			wantErr:       ErrTicketTypeInvalidName,
		},
		"異常系_表示名が空の場合_ErrTicketTypeInvalidNameになること": {
			parkID:        "park-1",
			name:          "",
			price:         8000,
			entryTimeFrom: "09:00",
			entryTimeTo:   "21:00",
			wantErr:       ErrTicketTypeInvalidName,
		},
		"境界値_価格が負の場合_ErrTicketTypeInvalidPriceになること": {
			parkID:        "park-1",
			name:          "1 デーパスポート",
			price:         -1,
			entryTimeFrom: "09:00",
			entryTimeTo:   "21:00",
			wantErr:       ErrTicketTypeInvalidPrice,
		},
		"異常系_時刻がHHMMでない場合_ErrTicketTypeInvalidEntryTimeになること": {
			parkID:        "park-1",
			name:          "1 デーパスポート",
			price:         8000,
			entryTimeFrom: "9時",
			entryTimeTo:   "21:00",
			wantErr:       ErrTicketTypeInvalidEntryTime,
		},
		"境界値_終了時刻が開始時刻と等しい場合_ErrTicketTypeInvalidEntryTimeRangeになること": {
			parkID:        "park-1",
			name:          "1 デーパスポート",
			price:         8000,
			entryTimeFrom: "09:00",
			entryTimeTo:   "09:00",
			wantErr:       ErrTicketTypeInvalidEntryTimeRange,
		},
		"異常系_終了時刻が開始時刻より前の場合_ErrTicketTypeInvalidEntryTimeRangeになること": {
			parkID:        "park-1",
			name:          "1 デーパスポート",
			price:         8000,
			entryTimeFrom: "21:00",
			entryTimeTo:   "09:00",
			wantErr:       ErrTicketTypeInvalidEntryTimeRange,
		},
		"異常系_パークの識別子が空の場合_ErrInvalidIDになること": {
			parkID:        "",
			name:          "1 デーパスポート",
			price:         8000,
			entryTimeFrom: "09:00",
			entryTimeTo:   "21:00",
			wantErr:       ErrInvalidID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewTicketType(
				tt.parkID,
				tt.name,
				tt.price,
				tt.entryTimeFrom,
				tt.entryTimeTo,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewTicketType(%q, %q, %d, %q, %q) のエラー = %v, want %v",
					tt.parkID,
					tt.name,
					tt.price,
					tt.entryTimeFrom,
					tt.entryTimeTo,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("NewTicketType(...) = %v, want nil", got)
				}
				return
			}

			if got.ID() == "" {
				t.Error("NewTicketType() の ID が空、採番されていない")
			}
			if got.ParkID() != tt.parkID {
				t.Errorf("NewTicketType() の ParkID = %q, want %q",
					got.ParkID(),
					tt.parkID,
				)
			}
			if got.Name() != tt.name {
				t.Errorf("NewTicketType() の Name = %q, want %q",
					got.Name(),
					tt.name,
				)
			}
			if got.Price() != tt.price {
				t.Errorf("NewTicketType() の Price = %d, want %d",
					got.Price(),
					tt.price,
				)
			}
			if got.EntryTimeFrom() != tt.entryTimeFrom {
				t.Errorf("NewTicketType() の EntryTimeFrom = %q, want %q",
					got.EntryTimeFrom(),
					tt.entryTimeFrom,
				)
			}
			if got.EntryTimeTo() != tt.entryTimeTo {
				t.Errorf("NewTicketType() の EntryTimeTo = %q, want %q",
					got.EntryTimeTo(),
					tt.entryTimeTo,
				)
			}
		})
	}
}

func TestNewTicketTypeGeneratesDistinctIDs(t *testing.T) {
	first, err := NewTicketType(
		"park-1",
		"1 デーパスポート",
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		t.Fatalf("NewTicketType() = %v, want nil", err)
	}

	second, err := NewTicketType(
		"park-1",
		"1 デーパスポート",
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		t.Fatalf("NewTicketType() = %v, want nil", err)
	}

	if first.ID() == second.ID() {
		t.Errorf("NewTicketType() の ID = %q, want 異なる値", first.ID())
	}
}

func TestRestoreTicketType(t *testing.T) {
	tests := map[string]struct {
		id      TicketTypeID
		wantErr error
	}{
		"正常系_識別子がある場合_券種が組み立てられること": {
			id: "ticket-type-1",
		},
		"異常系_識別子が空の場合_ErrTicketTypeInvalidIDになること": {
			id:      "",
			wantErr: ErrTicketTypeInvalidID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RestoreTicketType(
				"park-1",
				tt.id,
				"1 デーパスポート",
				8000,
				"09:00",
				"21:00",
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RestoreTicketType(%q, ...) のエラー = %v, want %v",
					tt.id,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr == nil && got.ID() != tt.id {
				t.Errorf("RestoreTicketType(%q, ...) の ID = %q, want %q",
					tt.id,
					got.ID(),
					tt.id,
				)
			}
		})
	}
}

func TestTicketType_Update(t *testing.T) {
	tests := map[string]struct {
		name          string
		price         int64
		entryTimeFrom string
		entryTimeTo   string
		wantErr       error
	}{
		"正常系_すべて範囲内の場合_値が入れ替わること": {
			name:          "2 デーパスポート",
			price:         14000,
			entryTimeFrom: "10:00",
			entryTimeTo:   "20:00",
		},
		"異常系_入場できる時間帯が逆転している場合_ErrTicketTypeInvalidEntryTimeRangeになること": {
			name:          "2 デーパスポート",
			price:         14000,
			entryTimeFrom: "20:00",
			entryTimeTo:   "10:00",
			wantErr:       ErrTicketTypeInvalidEntryTimeRange,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ticketType, err := RestoreTicketType(
				"park-1",
				"ticket-type-1",
				"1 デーパスポート",
				8000,
				"09:00",
				"21:00",
			)
			if err != nil {
				t.Fatalf("RestoreTicketType() = %v, want nil", err)
			}

			err = ticketType.Update(
				tt.name,
				tt.price,
				tt.entryTimeFrom,
				tt.entryTimeTo,
			)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("TicketType.Update(%q, %d, %q, %q) = %v, want %v",
					tt.name,
					tt.price,
					tt.entryTimeFrom,
					tt.entryTimeTo,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				// 検証に失敗したときは元の値を保つ
				if ticketType.Name() != "1 デーパスポート" || ticketType.Price() != 8000 ||
					ticketType.EntryTimeFrom() != "09:00" || ticketType.EntryTimeTo() != "21:00" {
					t.Errorf("TicketType.Update() の失敗後 = (%q, %d, %q, %q), want (%q, %d, %q, %q)",
						ticketType.Name(),
						ticketType.Price(),
						ticketType.EntryTimeFrom(),
						ticketType.EntryTimeTo(),
						"1 デーパスポート",
						8000,
						"09:00",
						"21:00",
					)
				}
				return
			}

			if ticketType.Name() != tt.name || ticketType.Price() != tt.price ||
				ticketType.EntryTimeFrom() != tt.entryTimeFrom || ticketType.EntryTimeTo() != tt.entryTimeTo {
				t.Errorf("TicketType.Update() 後 = (%q, %d, %q, %q), want (%q, %d, %q, %q)",
					ticketType.Name(),
					ticketType.Price(),
					ticketType.EntryTimeFrom(),
					ticketType.EntryTimeTo(),
					tt.name,
					tt.price,
					tt.entryTimeFrom,
					tt.entryTimeTo,
				)
			}
			if ticketType.ID() != "ticket-type-1" || ticketType.ParkID() != "park-1" {
				t.Errorf("TicketType.Update() 後の識別子 = (%q, %q), want (%q, %q)",
					ticketType.ParkID(),
					ticketType.ID(),
					"park-1",
					"ticket-type-1",
				)
			}
		})
	}
}
