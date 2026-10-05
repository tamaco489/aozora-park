package usecase

import (
	"context"
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestListTimeSlotsDo(t *testing.T) {
	repo := newFakeRepository()
	stored := storeTimeSlots(t, repo)

	tests := map[string]struct {
		in            ListTimeSlotsInput
		wantStartTime []string
	}{
		"正常系_保存済みの場合_開始時刻の昇順で取得できること": {
			in:            ListTimeSlotsInput{ParkID: storedParkID, AttractionID: storedAttractionID, Date: storedDate},
			wantStartTime: []string{stored[0].StartTime(), stored[1].StartTime()},
		},
		"正常系_保存されていない日付の場合_0件になること": {
			in:            ListTimeSlotsInput{ParkID: storedParkID, AttractionID: storedAttractionID, Date: "2026-10-06"},
			wantStartTime: nil,
		},
		"正常系_保存されていないアトラクションの場合_0件になること": {
			in:            ListTimeSlotsInput{ParkID: storedParkID, AttractionID: "attraction-2", Date: storedDate},
			wantStartTime: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewListTimeSlots(repo).Do(context.Background(), tt.in)
			if err != nil {
				t.Fatalf("ListTimeSlots.Do(%+v) = %v, want nil", tt.in, err)
			}

			if len(got) != len(tt.wantStartTime) {
				t.Fatalf("ListTimeSlots.Do(%+v) の件数 = %d, want %d", tt.in, len(got), len(tt.wantStartTime))
			}

			for i, slot := range got {
				if slot.StartTime() != tt.wantStartTime[i] {
					t.Errorf("ListTimeSlots.Do(%+v) の %d 件目の StartTime = %q, want %q",
						tt.in, i, slot.StartTime(), tt.wantStartTime[i])
				}
			}
		})
	}
}

// 並びの検証が空振りしていないことを確かめる、保存の順を入れ替えたら取得の順も入れ替わる
func TestListTimeSlotsDoKeepsRepositoryOrder(t *testing.T) {
	repo := newFakeRepository()
	repo.slots[slotKey(storedParkID, storedAttractionID, storedDate)] = []*inventorymodel.TimeSlot{
		restoreTimeSlot(t, "20261005_1100", "11:00"),
		restoreTimeSlot(t, "20261005_1000", "10:00"),
	}

	in := ListTimeSlotsInput{ParkID: storedParkID, AttractionID: storedAttractionID, Date: storedDate}

	got, err := NewListTimeSlots(repo).Do(context.Background(), in)
	if err != nil {
		t.Fatalf("ListTimeSlots.Do(%+v) = %v, want nil", in, err)
	}

	if len(got) != 2 || got[0].StartTime() != "11:00" {
		t.Errorf("ListTimeSlots.Do(%+v) の 1 件目の StartTime = %q, want %q", in, got[0].StartTime(), "11:00")
	}
}
