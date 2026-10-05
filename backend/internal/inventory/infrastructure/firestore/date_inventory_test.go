package firestore

import (
	"errors"
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestRepositoryGetDateInventory(t *testing.T) {
	repo, client := newRepository(t)

	const (
		parkID = inventorymodel.ParkID("park-get-date-inventory")
		date   = inventorymodel.Date("2026-10-05")
	)
	seedDateInventory(
		t,
		client,
		parkID,
		date,
		1000,
		800,
	)

	got, err := repo.GetDateInventory(
		t.Context(),
		parkID,
		date,
	)
	if err != nil {
		t.Fatalf("Repository.GetDateInventory(%q, %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}

	if got.ParkID() != parkID || got.Date() != date || got.Capacity() != 1000 || got.Remaining() != 800 {
		t.Errorf("Repository.GetDateInventory(%q, %q) = (%q, %q, %d, %d), want (%q, %q, %d, %d)",
			parkID,
			date,
			got.ParkID(),
			got.Date(),
			got.Capacity(),
			got.Remaining(),
			parkID,
			date,
			1000,
			800,
		)
	}
}

func TestRepositoryGetDateInventoryNotFound(t *testing.T) {
	repo, _ := newRepository(t)

	const (
		parkID = inventorymodel.ParkID("park-get-date-inventory-not-found")
		date   = inventorymodel.Date("2026-10-05")
	)

	_, err := repo.GetDateInventory(
		t.Context(),
		parkID,
		date,
	)
	if !errors.Is(err, inventorymodel.ErrDateInventoryNotFound) {
		t.Errorf("Repository.GetDateInventory(%q, %q) = %v, want %v",
			parkID,
			date,
			err,
			inventorymodel.ErrDateInventoryNotFound,
		)
	}
}

func TestRepositoryUpdateDateInventory(t *testing.T) {
	repo, client := newRepository(t)
	ctx := t.Context()

	const (
		parkID = inventorymodel.ParkID("park-update-date-inventory")
		date   = inventorymodel.Date("2026-10-05")
	)
	seedDateInventory(
		t,
		client,
		parkID,
		date,
		1000,
		800,
	)

	inventory, err := inventorymodel.RestoreDateInventory(
		parkID,
		date,
		2000,
		1500,
	)
	if err != nil {
		t.Fatalf("RestoreDateInventory() = %v, want nil", err)
	}

	if err := repo.UpdateDateInventory(ctx, inventory); err != nil {
		t.Fatalf("Repository.UpdateDateInventory(%q %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}

	got, err := repo.GetDateInventory(
		ctx,
		parkID,
		date,
	)
	if err != nil {
		t.Fatalf("Repository.GetDateInventory(%q, %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}

	if got.Capacity() != 2000 || got.Remaining() != 1500 {
		t.Errorf("Repository.GetDateInventory(%q, %q) = (%d, %d), want (%d, %d)",
			parkID,
			date,
			got.Capacity(),
			got.Remaining(),
			2000,
			1500,
		)
	}
}

func TestRepositoryUpdateDateInventoryNotFound(t *testing.T) {
	repo, _ := newRepository(t)

	const (
		parkID = inventorymodel.ParkID("park-update-date-inventory-not-found")
		date   = inventorymodel.Date("2026-10-05")
	)

	inventory, err := inventorymodel.RestoreDateInventory(
		parkID,
		date,
		2000,
		1500,
	)
	if err != nil {
		t.Fatalf("RestoreDateInventory() = %v, want nil", err)
	}

	err = repo.UpdateDateInventory(t.Context(), inventory)
	if !errors.Is(err, inventorymodel.ErrDateInventoryNotFound) {
		t.Errorf("Repository.UpdateDateInventory(%q %q) = %v, want %v",
			parkID,
			date,
			err,
			inventorymodel.ErrDateInventoryNotFound,
		)
	}
}

func TestRepositoryCreateDateInventoryIfAbsent(t *testing.T) {
	repo, client := newRepository(t)
	ctx := t.Context()

	const (
		parkID   = inventorymodel.ParkID("park-create-date-inventory")
		date     = inventorymodel.Date("2026-10-05")
		capacity = int32(1000)
	)
	cleanupDateInventory(
		t,
		client,
		parkID,
		date,
	)

	inventory, err := inventorymodel.NewDateInventory(
		parkID,
		date,
		capacity,
	)
	if err != nil {
		t.Fatalf("NewDateInventory(%q, %q, %d) = %v, want nil",
			parkID,
			date,
			capacity,
			err,
		)
	}

	created, err := repo.CreateDateInventoryIfAbsent(ctx, inventory)
	if err != nil {
		t.Fatalf("1 回目の Repository.CreateDateInventoryIfAbsent(%q, %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}
	if !created {
		t.Errorf("1 回目の Repository.CreateDateInventoryIfAbsent(%q, %q) = %t, want %t",
			parkID,
			date,
			created,
			true,
		)
	}

	// 運営が残りを減らした状態で再実行しても、値が初期値に戻らないことを確かめる
	if err := inventory.Overwrite(capacity, 1); err != nil {
		t.Fatalf("DateInventory.Overwrite(%d, 1) = %v, want nil",
			capacity,
			err,
		)
	}
	if err := repo.UpdateDateInventory(ctx, inventory); err != nil {
		t.Fatalf("Repository.UpdateDateInventory(%q, %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}

	again, err := inventorymodel.NewDateInventory(
		parkID,
		date,
		capacity,
	)
	if err != nil {
		t.Fatalf("NewDateInventory(%q, %q, %d) = %v, want nil",
			parkID,
			date,
			capacity,
			err,
		)
	}

	created, err = repo.CreateDateInventoryIfAbsent(ctx, again)
	if err != nil {
		t.Fatalf("2 回目の Repository.CreateDateInventoryIfAbsent(%q, %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}
	if created {
		t.Errorf("2 回目の Repository.CreateDateInventoryIfAbsent(%q, %q) = %t, want %t",
			parkID,
			date,
			created,
			false,
		)
	}

	stored, err := repo.GetDateInventory(
		ctx,
		parkID,
		date,
	)
	if err != nil {
		t.Fatalf("Repository.GetDateInventory(%q, %q) = %v, want nil",
			parkID,
			date,
			err,
		)
	}

	if stored.Remaining() != 1 {
		t.Errorf("Repository.GetDateInventory(%q, %q) の残り = %d, want %d",
			parkID,
			date,
			stored.Remaining(),
			1,
		)
	}
}
