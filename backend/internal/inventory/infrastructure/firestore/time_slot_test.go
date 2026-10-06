package firestore

import (
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestRepositoryListTimeSlots(t *testing.T) {
	repo, client := newRepositoryHelper(t)

	const (
		parkID       = inventorymodel.ParkID("park-list-time-slots")
		attractionID = inventorymodel.AttractionID("attraction-1")
		date         = inventorymodel.Date("2026-10-05")
		otherDate    = inventorymodel.Date("2026-10-06")
	)

	// 開始時刻の昇順で並ぶことを確かめるため、保存は昇順にしない
	seedTimeSlotHelper(
		t,
		client,
		parkID,
		attractionID,
		"20261005_1100",
		date,
		"11:00",
	)
	seedTimeSlotHelper(
		t,
		client,
		parkID,
		attractionID,
		"20261005_1000",
		date,
		"10:00",
	)
	// date で絞られることを確かめるための、別の日の枠
	seedTimeSlotHelper(
		t,
		client,
		parkID,
		attractionID,
		"20261006_1000",
		otherDate,
		"10:00",
	)

	got, err := repo.ListTimeSlots(
		t.Context(),
		parkID,
		attractionID,
		date,
	)
	if err != nil {
		t.Fatalf("Repository.ListTimeSlots(%q, %q, %q) = %v, want nil",
			parkID,
			attractionID,
			date,
			err,
		)
	}

	want := []string{
		"10:00",
		"11:00",
	}
	if len(got) != len(want) {
		t.Fatalf("Repository.ListTimeSlots(%q, %q, %q) の件数 = %d, want %d",
			parkID,
			attractionID,
			date,
			len(got),
			len(want),
		)
	}

	for i, slot := range got {
		if slot.StartTime() != want[i] {
			t.Errorf("Repository.ListTimeSlots(%q, %q, %q) の %d 件目の StartTime = %q, want %q",
				parkID,
				attractionID,
				date,
				i,
				slot.StartTime(),
				want[i],
			)
		}
		if slot.ParkID() != parkID || slot.AttractionID() != attractionID || slot.Date() != date {
			t.Errorf("Repository.ListTimeSlots(%q, %q, %q) の %d 件目 = (%q, %q, %q), want (%q, %q, %q)",
				parkID,
				attractionID,
				date,
				i,
				slot.ParkID(),
				slot.AttractionID(),
				slot.Date(),
				parkID,
				attractionID,
				date,
			)
		}
	}
}

func TestRepositoryListTimeSlotsEmpty(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	const (
		parkID       = inventorymodel.ParkID("park-list-time-slots-empty")
		attractionID = inventorymodel.AttractionID("attraction-1")
		date         = inventorymodel.Date("2026-10-05")
	)

	got, err := repo.ListTimeSlots(
		t.Context(),
		parkID,
		attractionID,
		date,
	)
	if err != nil {
		t.Fatalf("Repository.ListTimeSlots(%q, %q, %q) = %v, want nil",
			parkID,
			attractionID,
			date,
			err,
		)
	}

	if len(got) != 0 {
		t.Errorf("Repository.ListTimeSlots(%q, %q, %q) の件数 = %d, want %d",
			parkID,
			attractionID,
			date,
			len(got),
			0,
		)
	}
}
