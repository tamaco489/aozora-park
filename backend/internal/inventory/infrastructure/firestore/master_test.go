package firestore

import (
	"context"
	"testing"

	gcpfirestore "cloud.google.com/go/firestore"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

// seedParkHelper は park が登録する代わりにパークを 1 件置く
func seedParkHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	parkID inventorymodel.ParkID,
	defaultDailyCapacity int32,
	inventoryDays int32,
) {
	tb.Helper()

	doc := client.Collection(parkCollection).Doc(parkID.String())

	data := parkDocument{
		DefaultDailyCapacity: defaultDailyCapacity,
		InventoryDays:        inventoryDays,
	}
	if _, err := doc.Set(tb.Context(), data); err != nil {
		tb.Fatalf("Set(%q) = %v, want nil",
			parkID,
			err,
		)
	}

	tb.Cleanup(func() {
		// t.Context() は Cleanup の直前に取り消されるため、後始末は取り消されないものを使う
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				parkID,
				err,
			)
		}
	})
}

// seedAttractionHelper は park が登録する代わりにアトラクションを 1 件置く
func seedAttractionHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	parkID inventorymodel.ParkID,
	attractionID inventorymodel.AttractionID,
	enabled bool,
) {
	tb.Helper()

	doc := client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(attractionCollection).
		Doc(attractionID.String())

	data := attractionDocument{
		PriorityPassConfig: priorityPassConfigDocument{
			Enabled:         enabled,
			StartTime:       "09:00",
			EndTime:         "17:00",
			IntervalMinutes: 60,
			CapacityPerSlot: 30,
		},
	}
	if _, err := doc.Set(tb.Context(), data); err != nil {
		tb.Fatalf("Set(%q) = %v, want nil",
			attractionID,
			err,
		)
	}

	tb.Cleanup(func() {
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				attractionID,
				err,
			)
		}
	})
}

func TestRepositoryListParks(t *testing.T) {
	repo, client := newRepositoryHelper(t)

	const parkID = inventorymodel.ParkID("park-list-parks")
	seedParkHelper(
		t,
		client,
		parkID,
		1000,
		14,
	)

	parks, err := repo.ListParks(t.Context())
	if err != nil {
		t.Fatalf("Repository.ListParks() = %v, want nil", err)
	}

	// コレクションを他のテストと共有するため、置いたパークが含まれることだけを確かめる
	var got *inventorymodel.ParkMaster
	for _, park := range parks {
		if park.ID() == parkID {
			got = park
			break
		}
	}

	if got == nil {
		t.Fatalf("Repository.ListParks() に %q が含まれていない", parkID)
	}

	if got.DefaultDailyCapacity() != 1000 || got.InventoryDays() != 14 {
		t.Errorf("Repository.ListParks() の %q = (%d, %d), want (%d, %d)",
			parkID,
			got.DefaultDailyCapacity(),
			got.InventoryDays(),
			1000,
			14,
		)
	}
}

func TestRepositoryListAttractions(t *testing.T) {
	repo, client := newRepositoryHelper(t)

	const (
		parkID     = inventorymodel.ParkID("park-list-attractions")
		enabledID  = inventorymodel.AttractionID("attraction-enabled")
		disabledID = inventorymodel.AttractionID("attraction-disabled")
	)
	seedAttractionHelper(
		t,
		client,
		parkID,
		enabledID,
		true,
	)
	seedAttractionHelper(
		t,
		client,
		parkID,
		disabledID,
		false,
	)

	got, err := repo.ListAttractions(t.Context(), parkID)
	if err != nil {
		t.Fatalf("Repository.ListAttractions(%q) = %v, want nil",
			parkID,
			err,
		)
	}

	if len(got) != 2 {
		t.Fatalf("Repository.ListAttractions(%q) の件数 = %d, want %d",
			parkID,
			len(got),
			2,
		)
	}

	// enabled は読み出した値をそのまま持つ、生成の対象を決めるのは usecase のため
	enabled := map[inventorymodel.AttractionID]bool{}
	for _, attraction := range got {
		enabled[attraction.ID()] = attraction.PriorityPassEnabled()
	}

	if !enabled[enabledID] || enabled[disabledID] {
		t.Errorf("Repository.ListAttractions(%q) の enabled = (%t, %t), want (%t, %t)",
			parkID,
			enabled[enabledID],
			enabled[disabledID],
			true,
			false,
		)
	}
}

func TestRepositoryListAttractionsEmpty(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	const parkID = inventorymodel.ParkID("park-list-attractions-empty")

	got, err := repo.ListAttractions(t.Context(), parkID)
	if err != nil {
		t.Fatalf("Repository.ListAttractions(%q) = %v, want nil",
			parkID,
			err,
		)
	}

	if len(got) != 0 {
		t.Errorf("Repository.ListAttractions(%q) の件数 = %d, want %d",
			parkID,
			len(got),
			0,
		)
	}
}
