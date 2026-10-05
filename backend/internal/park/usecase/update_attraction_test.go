package usecase

import (
	"context"
	"errors"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestUpdateAttractionDo(t *testing.T) {
	valid := UpdateAttractionInput{
		ParkID:          "park-1",
		ID:              "attraction-1",
		Name:            "ジェットコースター 2",
		Enabled:         false,
		StartTime:       "10:00",
		EndTime:         "20:00",
		IntervalMinutes: 15,
		CapacityPerSlot: 20,
	}

	tests := map[string]struct {
		stored    bool
		in        UpdateAttractionInput
		updateErr error
		wantErr   error
	}{
		"正常系_保存済みの場合_値が入れ替わること": {
			stored: true,
			in:     valid,
		},
		"異常系_保存されていない場合_ErrAttractionNotFoundになること": {
			in:      valid,
			wantErr: parkmodel.ErrAttractionNotFound,
		},
		"異常系_別のパークの識別子を渡した場合_ErrAttractionNotFoundになること": {
			stored: true,
			in: UpdateAttractionInput{
				ParkID:          "park-2",
				ID:              "attraction-1",
				Name:            "ジェットコースター 2",
				Enabled:         false,
				StartTime:       "10:00",
				EndTime:         "20:00",
				IntervalMinutes: 15,
				CapacityPerSlot: 20,
			},
			wantErr: parkmodel.ErrAttractionNotFound,
		},
		"境界値_間隔が0分の場合_ErrAttractionInvalidIntervalMinutesになること": {
			stored: true,
			in: UpdateAttractionInput{
				ParkID:          "park-1",
				ID:              "attraction-1",
				Name:            "ジェットコースター 2",
				Enabled:         true,
				StartTime:       "10:00",
				EndTime:         "20:00",
				IntervalMinutes: 0,
				CapacityPerSlot: 20,
			},
			wantErr: parkmodel.ErrAttractionInvalidIntervalMinutes,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.updateAttractionErr = tt.updateErr

			if tt.stored {
				storeAttraction(t, repo)
			}

			got, err := NewUpdateAttraction(repo, repo).Do(context.Background(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateAttraction.Do(%+v) のエラー = %v, want %v",
					tt.in,
					err,
					tt.wantErr,
				)
			}

			key := attractionKey{
				parkID: "park-1",
				id:     "attraction-1",
			}

			if tt.wantErr != nil {
				if tt.stored {
					// 保存済みの値が書き換わっていないことを確かめる
					if stored := repo.attractions[key]; stored.Name() != "ジェットコースター" {
						t.Errorf("UpdateAttraction.Do(%+v) の失敗後の Name = %q, want %q",
							tt.in,
							stored.Name(),
							"ジェットコースター",
						)
					}
				}
				return
			}

			if got.Name() != tt.in.Name {
				t.Errorf("UpdateAttraction.Do(%+v) の Name = %q, want %q",
					tt.in,
					got.Name(),
					tt.in.Name,
				)
			}

			stored := repo.attractions[key]
			if stored.PriorityPassConfig().Enabled() != tt.in.Enabled {
				t.Errorf("UpdateAttraction.Do(%+v) の後の Enabled = %t, want %t",
					tt.in,
					stored.PriorityPassConfig().Enabled(),
					tt.in.Enabled,
				)
			}
			if stored.PriorityPassConfig().IntervalMinutes() != tt.in.IntervalMinutes {
				t.Errorf("UpdateAttraction.Do(%+v) の後の IntervalMinutes = %d, want %d",
					tt.in,
					stored.PriorityPassConfig().IntervalMinutes(),
					tt.in.IntervalMinutes,
				)
			}
		})
	}
}
