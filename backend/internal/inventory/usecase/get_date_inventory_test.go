package usecase

import (
	"errors"
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestGetDateInventoryDo(t *testing.T) {
	repo := newFakeRepository()
	storeDateInventoryHelper(t, repo)

	tests := map[string]struct {
		in      GetDateInventoryInput
		wantErr error
	}{
		"正常系_保存済みの場合_取得できること": {
			in: GetDateInventoryInput{
				ParkID: storedParkID,
				Date:   storedDate,
			},
		},
		"異常系_保存されていない日付の場合_ErrDateInventoryNotFoundになること": {
			in: GetDateInventoryInput{
				ParkID: storedParkID,
				Date:   "2026-10-06",
			},
			wantErr: inventorymodel.ErrDateInventoryNotFound,
		},
		"異常系_保存されていないパークの場合_ErrDateInventoryNotFoundになること": {
			in: GetDateInventoryInput{
				ParkID: "park-2",
				Date:   storedDate,
			},
			wantErr: inventorymodel.ErrDateInventoryNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewGetDateInventory(repo).Do(t.Context(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetDateInventory.Do(%+v) のエラー = %v, want %v",
					tt.in,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				return
			}

			if got.ParkID() != tt.in.ParkID || got.Date() != tt.in.Date {
				t.Errorf("GetDateInventory.Do(%+v) = (%q, %q), want (%q, %q)",
					tt.in,
					got.ParkID(),
					got.Date(),
					tt.in.ParkID,
					tt.in.Date,
				)
			}
		})
	}
}
