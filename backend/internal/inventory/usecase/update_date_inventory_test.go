package usecase

import (
	"context"
	"errors"
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestUpdateDateInventoryDo(t *testing.T) {
	tests := map[string]struct {
		stored    bool
		in        UpdateDateInventoryInput
		updateErr error
		wantErr   error
	}{
		"正常系_保存済みの場合_上限と残りが入れ替わること": {
			stored: true,
			in:     UpdateDateInventoryInput{ParkID: storedParkID, Date: storedDate, Capacity: 2000, Remaining: 1500},
		},
		"境界値_残りが上限と等しい場合_入れ替わること": {
			stored: true,
			in:     UpdateDateInventoryInput{ParkID: storedParkID, Date: storedDate, Capacity: 2000, Remaining: 2000},
		},
		"境界値_残りが上限を1超える場合_ErrInvalidRemainingになること": {
			stored:  true,
			in:      UpdateDateInventoryInput{ParkID: storedParkID, Date: storedDate, Capacity: 2000, Remaining: 2001},
			wantErr: inventorymodel.ErrInvalidRemaining,
		},
		"異常系_残りが負の場合_ErrInvalidRemainingになること": {
			stored:  true,
			in:      UpdateDateInventoryInput{ParkID: storedParkID, Date: storedDate, Capacity: 2000, Remaining: -1},
			wantErr: inventorymodel.ErrInvalidRemaining,
		},
		"境界値_上限が0の場合_ErrInvalidCapacityになること": {
			stored:  true,
			in:      UpdateDateInventoryInput{ParkID: storedParkID, Date: storedDate, Capacity: 0, Remaining: 0},
			wantErr: inventorymodel.ErrInvalidCapacity,
		},
		"異常系_保存されていない場合_ErrDateInventoryNotFoundになること": {
			in:      UpdateDateInventoryInput{ParkID: storedParkID, Date: storedDate, Capacity: 2000, Remaining: 1500},
			wantErr: inventorymodel.ErrDateInventoryNotFound,
		},
		"異常系_保存に失敗した場合_そのエラーが返ること": {
			stored:    true,
			in:        UpdateDateInventoryInput{ParkID: storedParkID, Date: storedDate, Capacity: 2000, Remaining: 1500},
			updateErr: inventorymodel.ErrDateInventoryNotFound,
			wantErr:   inventorymodel.ErrDateInventoryNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.updateErr = tt.updateErr

			if tt.stored {
				storeDateInventory(t, repo)
			}

			got, err := NewUpdateDateInventory(repo, repo).Do(context.Background(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateDateInventory.Do(%+v) のエラー = %v, want %v", tt.in, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if !tt.stored {
					return
				}
				// 検証に失敗した入場枠が書き換わっていないことを確かめる
				stored := repo.inventories[dateKey(storedParkID, storedDate)]
				if stored.Capacity() != storedCapacity || stored.Remaining() != storedRemaining {
					t.Errorf("UpdateDateInventory.Do(%+v) の失敗後 = (%d, %d), want (%d, %d)",
						tt.in, stored.Capacity(), stored.Remaining(), storedCapacity, storedRemaining)
				}
				return
			}

			if got.Capacity() != tt.in.Capacity || got.Remaining() != tt.in.Remaining {
				t.Errorf("UpdateDateInventory.Do(%+v) = (%d, %d), want (%d, %d)",
					tt.in, got.Capacity(), got.Remaining(), tt.in.Capacity, tt.in.Remaining)
			}

			stored := repo.inventories[dateKey(storedParkID, storedDate)]
			if stored.Capacity() != tt.in.Capacity || stored.Remaining() != tt.in.Remaining {
				t.Errorf("UpdateDateInventory.Do(%+v) の保存後 = (%d, %d), want (%d, %d)",
					tt.in, stored.Capacity(), stored.Remaining(), tt.in.Capacity, tt.in.Remaining)
			}
		})
	}
}
