package usecase

import (
	"context"
	"errors"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestCreateTicketTypeDo(t *testing.T) {
	valid := CreateTicketTypeInput{
		ParkID:        "park-1",
		Name:          "1 デーパスポート",
		Price:         8000,
		EntryTimeFrom: "09:00",
		EntryTimeTo:   "21:00",
	}

	tests := map[string]struct {
		storedPark bool
		in         CreateTicketTypeInput
		createErr  error
		wantErr    error
	}{
		"正常系_パークが保存済みで入力が妥当な場合_保存されること": {
			storedPark: true,
			in:         valid,
		},
		"境界値_価格が0円の場合_保存されること": {
			storedPark: true,
			in: CreateTicketTypeInput{
				ParkID:        "park-1",
				Name:          "招待券",
				Price:         0,
				EntryTimeFrom: "09:00",
				EntryTimeTo:   "21:00",
			},
		},
		"異常系_パークが保存されていない場合_ErrNotFoundになること": {
			in:      valid,
			wantErr: parkmodel.ErrNotFound,
		},
		"異常系_価格が負の場合_ErrTicketTypeInvalidPriceになること": {
			storedPark: true,
			in: CreateTicketTypeInput{
				ParkID:        "park-1",
				Name:          "1 デーパスポート",
				Price:         -1,
				EntryTimeFrom: "09:00",
				EntryTimeTo:   "21:00",
			},
			wantErr: parkmodel.ErrTicketTypeInvalidPrice,
		},
		"異常系_入場できる時間帯が逆転している場合_ErrTicketTypeInvalidEntryTimeRangeになること": {
			storedPark: true,
			in: CreateTicketTypeInput{
				ParkID:        "park-1",
				Name:          "1 デーパスポート",
				Price:         8000,
				EntryTimeFrom: "21:00",
				EntryTimeTo:   "09:00",
			},
			wantErr: parkmodel.ErrTicketTypeInvalidEntryTimeRange,
		},
		"異常系_保存が失敗した場合_そのエラーが返ること": {
			storedPark: true,
			in:         valid,
			createErr:  parkmodel.ErrTicketTypeAlreadyExists,
			wantErr:    parkmodel.ErrTicketTypeAlreadyExists,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.createTicketTypeErr = tt.createErr

			if tt.storedPark {
				store(t, repo)
			}

			got, err := NewCreateTicketType(repo, repo).Do(context.Background(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateTicketType.Do(%+v) のエラー = %v, want %v", tt.in, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if len(repo.ticketTypes) != 0 {
					t.Errorf("CreateTicketType.Do(%+v) の後の保存件数 = %d, want 0", tt.in, len(repo.ticketTypes))
				}
				return
			}

			key := ticketTypeKey{parkID: got.ParkID(), id: got.ID()}
			if _, ok := repo.ticketTypes[key]; !ok {
				t.Errorf("CreateTicketType.Do(%+v) の後に %q が保存されていない", tt.in, got.ID())
			}
			if got.ID() == "" {
				t.Error("CreateTicketType.Do() の ID が空、採番されていない")
			}
			if got.Price() != tt.in.Price {
				t.Errorf("CreateTicketType.Do(%+v) の Price = %d, want %d", tt.in, got.Price(), tt.in.Price)
			}
		})
	}
}
