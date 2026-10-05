package usecase

import (
	"errors"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestUpdateTicketTypeDo(t *testing.T) {
	valid := UpdateTicketTypeInput{
		ParkID:        "park-1",
		ID:            "ticket-type-1",
		Name:          "2 デーパスポート",
		Price:         14000,
		EntryTimeFrom: "10:00",
		EntryTimeTo:   "20:00",
	}

	tests := map[string]struct {
		stored    bool
		in        UpdateTicketTypeInput
		updateErr error
		wantErr   error
	}{
		"正常系_保存済みの場合_値が入れ替わること": {
			stored: true,
			in:     valid,
		},
		"異常系_保存されていない場合_ErrTicketTypeNotFoundになること": {
			in:      valid,
			wantErr: parkmodel.ErrTicketTypeNotFound,
		},
		"異常系_別のパークの識別子を渡した場合_ErrTicketTypeNotFoundになること": {
			stored: true,
			in: UpdateTicketTypeInput{
				ParkID:        "park-2",
				ID:            "ticket-type-1",
				Name:          "2 デーパスポート",
				Price:         14000,
				EntryTimeFrom: "10:00",
				EntryTimeTo:   "20:00",
			},
			wantErr: parkmodel.ErrTicketTypeNotFound,
		},
		"異常系_時刻がHHMMでない場合_ErrTicketTypeInvalidEntryTimeになること": {
			stored: true,
			in: UpdateTicketTypeInput{
				ParkID:        "park-1",
				ID:            "ticket-type-1",
				Name:          "2 デーパスポート",
				Price:         14000,
				EntryTimeFrom: "10時",
				EntryTimeTo:   "20:00",
			},
			wantErr: parkmodel.ErrTicketTypeInvalidEntryTime,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.updateTicketTypeErr = tt.updateErr

			if tt.stored {
				storeTicketType(t, repo)
			}

			got, err := NewUpdateTicketType(repo, repo).Do(t.Context(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateTicketType.Do(%+v) のエラー = %v, want %v",
					tt.in,
					err,
					tt.wantErr,
				)
			}

			key := ticketTypeKey{
				parkID: "park-1",
				id:     "ticket-type-1",
			}

			if tt.wantErr != nil {
				if tt.stored {
					// 保存済みの値が書き換わっていないことを確かめる
					if stored := repo.ticketTypes[key]; stored.Price() != 8000 {
						t.Errorf("UpdateTicketType.Do(%+v) の失敗後の Price = %d, want %d",
							tt.in,
							stored.Price(),
							8000,
						)
					}
				}
				return
			}

			if got.Name() != tt.in.Name {
				t.Errorf("UpdateTicketType.Do(%+v) の Name = %q, want %q",
					tt.in,
					got.Name(),
					tt.in.Name,
				)
			}
			if stored := repo.ticketTypes[key]; stored.Price() != tt.in.Price {
				t.Errorf("UpdateTicketType.Do(%+v) の後の Price = %d, want %d",
					tt.in,
					stored.Price(),
					tt.in.Price,
				)
			}
		})
	}
}
