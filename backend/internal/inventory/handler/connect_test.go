package handler

import (
	"context"
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
	inventoryusecase "github.com/tamaco489/aozora-park/backend/internal/inventory/usecase"
)

// 保存済みの枠に使う値
const (
	storedParkID       = inventorymodel.ParkID("park-1")
	storedAttractionID = inventorymodel.AttractionID("attraction-1")
	storedDate         = inventorymodel.Date("2026-10-05")
)

// fakeRepository は保存先を差し替えるためのインメモリ実装
//
// usecase は実物を通す、差し替えるのは domain/repository だけにする
type fakeRepository struct {
	inventories map[inventorymodel.Date]*inventorymodel.DateInventory
	slots       []*inventorymodel.TimeSlot
}

var (
	_ inventoryrepository.Reader = (*fakeRepository)(nil)
	_ inventoryrepository.Writer = (*fakeRepository)(nil)
)

func (r *fakeRepository) GetDateInventory(
	_ context.Context,
	_ inventorymodel.ParkID,
	date inventorymodel.Date,
) (*inventorymodel.DateInventory, error) {
	inventory, ok := r.inventories[date]
	if !ok {
		return nil, inventorymodel.ErrDateInventoryNotFound
	}
	return inventory, nil
}

func (r *fakeRepository) ListTimeSlots(
	_ context.Context,
	_ inventorymodel.ParkID,
	_ inventorymodel.AttractionID,
	_ inventorymodel.Date,
) ([]*inventorymodel.TimeSlot, error) {
	return r.slots, nil
}

func (r *fakeRepository) UpdateDateInventory(_ context.Context, inventory *inventorymodel.DateInventory) error {
	if _, ok := r.inventories[inventory.Date()]; !ok {
		return inventorymodel.ErrDateInventoryNotFound
	}
	r.inventories[inventory.Date()] = inventory
	return nil
}

// newHandler は渡した枠を保存済みにしたハンドラを組み立てる
func newHandler(tb testing.TB, repo *fakeRepository) *Connect {
	tb.Helper()

	return NewConnect(
		inventoryusecase.NewUpdateDateInventory(repo, repo),
		inventoryusecase.NewGetDateInventory(repo),
		inventoryusecase.NewListTimeSlots(repo),
	)
}

func newEmptyRepository() *fakeRepository {
	return &fakeRepository{inventories: map[inventorymodel.Date]*inventorymodel.DateInventory{}}
}

// newStoredRepository は入場枠 1 件と時間帯枠 2 件を保存済みにする
func newStoredRepository(tb testing.TB) *fakeRepository {
	tb.Helper()

	inventory, err := inventorymodel.RestoreDateInventory(
		storedParkID,
		storedDate,
		1000,
		800,
	)
	if err != nil {
		tb.Fatalf("RestoreDateInventory() = %v, want nil", err)
	}

	repo := newEmptyRepository()
	repo.inventories[storedDate] = inventory
	repo.slots = []*inventorymodel.TimeSlot{
		restoreTimeSlot(
			tb,
			"20261005_1000",
			"10:00",
		),
		restoreTimeSlot(
			tb,
			"20261005_1100",
			"11:00",
		),
	}

	return repo
}

func restoreTimeSlot(
	tb testing.TB,
	id inventorymodel.TimeSlotID,
	startTime string,
) *inventorymodel.TimeSlot {
	tb.Helper()

	slot, err := inventorymodel.RestoreTimeSlot(
		storedParkID,
		storedAttractionID,
		id,
		storedDate,
		startTime,
		60,
		30,
	)
	if err != nil {
		tb.Fatalf("RestoreTimeSlot(%q) = %v, want nil",
			id,
			err,
		)
	}

	return slot
}
