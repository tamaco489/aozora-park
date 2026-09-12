package usecase

import (
	"context"
	"errors"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestCreateDo(t *testing.T) {
	tests := map[string]struct {
		in        CreateInput
		createErr error
		wantErr   error
	}{
		"正常系_入力が妥当な場合_保存されること": {
			in: CreateInput{Name: "Aozora Park", DefaultDailyCapacity: 1000, InventoryDays: 30},
		},
		"異常系_表示名が空の場合_保存されないこと": {
			in:      CreateInput{Name: "", DefaultDailyCapacity: 1000, InventoryDays: 30},
			wantErr: parkmodel.ErrInvalidName,
		},
		"異常系_保存が失敗した場合_そのエラーが返ること": {
			in:        CreateInput{Name: "Aozora Park", DefaultDailyCapacity: 1000, InventoryDays: 30},
			createErr: parkmodel.ErrAlreadyExists,
			wantErr:   parkmodel.ErrAlreadyExists,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.createErr = tt.createErr

			got, err := NewCreate(repo).Do(context.Background(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create.Do(%+v) のエラー = %v, want %v", tt.in, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if len(repo.parks) != 0 {
					t.Errorf("Create.Do(%+v) の後の保存件数 = %d, want 0", tt.in, len(repo.parks))
				}
				return
			}

			if _, ok := repo.parks[got.ID()]; !ok {
				t.Errorf("Create.Do(%+v) の後に %q が保存されていない", tt.in, got.ID())
			}
		})
	}
}
