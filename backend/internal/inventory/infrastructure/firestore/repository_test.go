package firestore

import (
	"context"
	"os"
	"testing"

	gcpfirestore "cloud.google.com/go/firestore"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore/firestoretest"
)

func TestMain(m *testing.M) {
	os.Exit(firestoretest.Main(m))
}

// newRepository は保存先を用意する
//
// テスト間の分離はパークの識別子を分けて行う、同じコレクションを共有するため
func newRepository(tb testing.TB) (*Repository, *gcpfirestore.Client) {
	tb.Helper()

	client := firestoretest.Client(tb)

	return NewRepository(client), client
}

// seedDateInventory は枠を作成するジョブの代わりに入場枠を 1 件置く
func seedDateInventory(
	tb testing.TB,
	client *gcpfirestore.Client,
	parkID inventorymodel.ParkID,
	date inventorymodel.Date,
	capacity int32,
	remaining int32,
) {
	tb.Helper()

	doc := client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(dateInventoryCollection).
		Doc(date.String())

	data := dateInventoryDocument{
		Date:      date.String(),
		Capacity:  capacity,
		Remaining: remaining,
	}
	if _, err := doc.Set(tb.Context(), data); err != nil {
		tb.Fatalf("Set(%q %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}

	tb.Cleanup(func() {
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q %q) = %v, want nil",
				parkID,
				date,
				err,
			)
		}
	})
}

// seedTimeSlot は枠を作成するジョブの代わりに時間帯枠を 1 件置く
func seedTimeSlot(
	tb testing.TB,
	client *gcpfirestore.Client,
	parkID inventorymodel.ParkID,
	attractionID inventorymodel.AttractionID,
	id inventorymodel.TimeSlotID,
	date inventorymodel.Date,
	startTime string,
) {
	tb.Helper()

	doc := client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(attractionCollection).
		Doc(attractionID.String()).
		Collection(timeSlotCollection).
		Doc(id.String())

	data := timeSlotDocument{
		AttractionID: attractionID.String(),
		Date:         date.String(),
		StartTime:    startTime,
		Capacity:     60,
		Remaining:    30,
	}
	if _, err := doc.Set(tb.Context(), data); err != nil {
		tb.Fatalf("Set(%q) = %v, want nil",
			id,
			err,
		)
	}

	tb.Cleanup(func() {
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				id,
				err,
			)
		}
	})
}

// cleanupDateInventory は検査対象が作成する入場枠の後始末を登録する
func cleanupDateInventory(
	tb testing.TB,
	client *gcpfirestore.Client,
	parkID inventorymodel.ParkID,
	date inventorymodel.Date,
) {
	tb.Helper()

	doc := client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(dateInventoryCollection).
		Doc(date.String())

	tb.Cleanup(func() {
		// t.Context() は Cleanup の直前に取り消されるため、後始末は取り消されないものを使う
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q %q) = %v, want nil",
				parkID,
				date,
				err,
			)
		}
	})
}

// cleanupTimeSlot は検査対象が作成する時間帯枠の後始末を登録する
func cleanupTimeSlot(
	tb testing.TB,
	client *gcpfirestore.Client,
	parkID inventorymodel.ParkID,
	attractionID inventorymodel.AttractionID,
	id inventorymodel.TimeSlotID,
) {
	tb.Helper()

	doc := client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(attractionCollection).
		Doc(attractionID.String()).
		Collection(timeSlotCollection).
		Doc(id.String())

	tb.Cleanup(func() {
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				id,
				err,
			)
		}
	})
}
