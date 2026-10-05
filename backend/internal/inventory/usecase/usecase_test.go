package usecase

import (
	"context"
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
)

// 保存済みの枠に使う値、テストの入力と区別できるよう 1 か所に置く
const (
	storedParkID       = inventorymodel.ParkID("park-1")
	storedAttractionID = inventorymodel.AttractionID("attraction-1")
	storedDate         = inventorymodel.Date("2026-10-05")
	storedCapacity     = int32(1000)
	storedRemaining    = int32(800)
)

// fakeRepository は Reader と Writer を満たすインメモリの保存先
type fakeRepository struct {
	inventories map[string]*inventorymodel.DateInventory
	slots       map[string][]*inventorymodel.TimeSlot
	updateErr   error
}

var (
	_ inventoryrepository.Reader = (*fakeRepository)(nil)
	_ inventoryrepository.Writer = (*fakeRepository)(nil)
)

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		inventories: map[string]*inventorymodel.DateInventory{},
		slots:       map[string][]*inventorymodel.TimeSlot{},
	}
}

func dateKey(parkID inventorymodel.ParkID, date inventorymodel.Date) string {
	return parkID.String() + "/" + date.String()
}

func slotKey(parkID inventorymodel.ParkID, attractionID inventorymodel.AttractionID, date inventorymodel.Date) string {
	return parkID.String() + "/" + attractionID.String() + "/" + date.String()
}

func (r *fakeRepository) GetDateInventory(_ context.Context, parkID inventorymodel.ParkID, date inventorymodel.Date) (*inventorymodel.DateInventory, error) {
	inventory, ok := r.inventories[dateKey(parkID, date)]
	if !ok {
		return nil, inventorymodel.ErrDateInventoryNotFound
	}

	// infrastructure は読み出すたびにドキュメントから組み立て直すため、フェイクも保存済みの実体を渡さない
	// そのまま渡すと呼び出し側の書き換えが保存先に及び、保存に失敗した場合でも値が戻らなくなる
	return inventorymodel.RestoreDateInventory(
		inventory.ParkID(),
		inventory.Date(),
		inventory.Capacity(),
		inventory.Remaining(),
	)
}

func (r *fakeRepository) ListTimeSlots(_ context.Context, parkID inventorymodel.ParkID, attractionID inventorymodel.AttractionID, date inventorymodel.Date) ([]*inventorymodel.TimeSlot, error) {
	return r.slots[slotKey(parkID, attractionID, date)], nil
}

func (r *fakeRepository) UpdateDateInventory(_ context.Context, inventory *inventorymodel.DateInventory) error {
	if r.updateErr != nil {
		return r.updateErr
	}

	key := dateKey(inventory.ParkID(), inventory.Date())
	if _, ok := r.inventories[key]; !ok {
		return inventorymodel.ErrDateInventoryNotFound
	}
	r.inventories[key] = inventory

	return nil
}

// storeDateInventory は保存済みの入場枠を 1 件用意する
func storeDateInventory(tb testing.TB, repo *fakeRepository) *inventorymodel.DateInventory {
	tb.Helper()

	inventory, err := inventorymodel.RestoreDateInventory(storedParkID, storedDate, storedCapacity, storedRemaining)
	if err != nil {
		tb.Fatalf("RestoreDateInventory() = %v, want nil", err)
	}
	repo.inventories[dateKey(storedParkID, storedDate)] = inventory

	return inventory
}

// storeTimeSlots は保存済みの時間帯枠を開始時刻の昇順で用意する
func storeTimeSlots(tb testing.TB, repo *fakeRepository) []*inventorymodel.TimeSlot {
	tb.Helper()

	slots := []*inventorymodel.TimeSlot{
		restoreTimeSlot(tb, "20261005_1000", "10:00"),
		restoreTimeSlot(tb, "20261005_1100", "11:00"),
	}
	repo.slots[slotKey(storedParkID, storedAttractionID, storedDate)] = slots

	return slots
}

func restoreTimeSlot(tb testing.TB, id inventorymodel.TimeSlotID, startTime string) *inventorymodel.TimeSlot {
	tb.Helper()

	slot, err := inventorymodel.RestoreTimeSlot(storedParkID, storedAttractionID, id, storedDate, startTime, 60, 30)
	if err != nil {
		tb.Fatalf("RestoreTimeSlot(%q) = %v, want nil", id, err)
	}

	return slot
}
