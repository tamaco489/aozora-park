package usecase

import (
	"context"
	"errors"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestUpdateDo(t *testing.T) {
	tests := map[string]struct {
		stored    bool
		in        UpdateInput
		updateErr error
		wantErr   error
	}{
		"正常系_保存済みの場合_値が入れ替わること": {
			stored: true,
			in:     UpdateInput{ID: "park-1", Name: "Aozora Park 2", DefaultDailyCapacity: 2000, InventoryDays: 60},
		},
		"異常系_保存されていない場合_ErrNotFoundになること": {
			in:      UpdateInput{ID: "park-1", Name: "Aozora Park 2", DefaultDailyCapacity: 2000, InventoryDays: 60},
			wantErr: parkmodel.ErrNotFound,
		},
		"異常系_生成日数が範囲外の場合_保存されないこと": {
			stored:  true,
			in:      UpdateInput{ID: "park-1", Name: "Aozora Park 2", DefaultDailyCapacity: 2000, InventoryDays: 91},
			wantErr: parkmodel.ErrInvalidInventoryDays,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.updateErr = tt.updateErr

			if tt.stored {
				store(t, repo)
			}

			got, err := NewUpdate(repo, repo).Do(context.Background(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Update.Do(%+v) のエラー = %v, want %v", tt.in, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if tt.stored {
					// 保存済みの値が書き換わっていないことを確かめる
					if stored := repo.parks["park-1"]; stored.Name() != "Aozora Park" {
						t.Errorf("Update.Do(%+v) の失敗後の Name = %q, want %q", tt.in, stored.Name(), "Aozora Park")
					}
				}
				return
			}

			if got.Name() != tt.in.Name {
				t.Errorf("Update.Do(%+v) の Name = %q, want %q", tt.in, got.Name(), tt.in.Name)
			}
			if stored := repo.parks["park-1"]; stored.InventoryDays() != tt.in.InventoryDays {
				t.Errorf("Update.Do(%+v) の後の InventoryDays = %d, want %d", tt.in, stored.InventoryDays(), tt.in.InventoryDays)
			}
		})
	}
}
