package usecase

import (
	"errors"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestCreateAttractionDo(t *testing.T) {
	valid := CreateAttractionInput{
		ParkID:          "park-1",
		Name:            "ジェットコースター",
		Enabled:         true,
		StartTime:       "09:00",
		EndTime:         "18:00",
		IntervalMinutes: 30,
		CapacityPerSlot: 10,
	}

	tests := map[string]struct {
		storedPark bool
		in         CreateAttractionInput
		createErr  error
		wantErr    error
	}{
		"正常系_パークが保存済みで入力が妥当な場合_保存されること": {
			storedPark: true,
			in:         valid,
		},
		"異常系_パークが保存されていない場合_ErrNotFoundになること": {
			in:      valid,
			wantErr: parkmodel.ErrNotFound,
		},
		"異常系_表示名が空の場合_保存されないこと": {
			storedPark: true,
			in: CreateAttractionInput{
				ParkID:          "park-1",
				Name:            "",
				Enabled:         true,
				StartTime:       "09:00",
				EndTime:         "18:00",
				IntervalMinutes: 30,
				CapacityPerSlot: 10,
			},
			wantErr: parkmodel.ErrAttractionInvalidName,
		},
		"境界値_終了時刻が開始時刻と等しい場合_ErrAttractionInvalidTimeRangeになること": {
			storedPark: true,
			in: CreateAttractionInput{
				ParkID:          "park-1",
				Name:            "ジェットコースター",
				Enabled:         true,
				StartTime:       "09:00",
				EndTime:         "09:00",
				IntervalMinutes: 30,
				CapacityPerSlot: 10,
			},
			wantErr: parkmodel.ErrAttractionInvalidTimeRange,
		},
		"異常系_保存が失敗した場合_そのエラーが返ること": {
			storedPark: true,
			in:         valid,
			createErr:  parkmodel.ErrAttractionAlreadyExists,
			wantErr:    parkmodel.ErrAttractionAlreadyExists,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.createAttractionErr = tt.createErr

			if tt.storedPark {
				store(t, repo)
			}

			got, err := NewCreateAttraction(repo, repo).Do(t.Context(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateAttraction.Do(%+v) のエラー = %v, want %v",
					tt.in,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if len(repo.attractions) != 0 {
					t.Errorf("CreateAttraction.Do(%+v) の後の保存件数 = %d, want 0",
						tt.in,
						len(repo.attractions),
					)
				}
				return
			}

			key := attractionKey{
				parkID: got.ParkID(),
				id:     got.ID(),
			}
			if _, ok := repo.attractions[key]; !ok {
				t.Errorf("CreateAttraction.Do(%+v) の後に %q が保存されていない",
					tt.in,
					got.ID(),
				)
			}
			if got.ID() == "" {
				t.Error("CreateAttraction.Do() の ID が空、採番されていない")
			}
			if got.PriorityPassConfig().CapacityPerSlot() != tt.in.CapacityPerSlot {
				t.Errorf("CreateAttraction.Do(%+v) の CapacityPerSlot = %d, want %d",
					tt.in,
					got.PriorityPassConfig().CapacityPerSlot(),
					tt.in.CapacityPerSlot,
				)
			}
		})
	}
}
