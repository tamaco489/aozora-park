package firestore

import (
	"errors"
	"testing"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

func TestRepositoryGetDateInventory(t *testing.T) {
	repo, client := newRepositoryHelper(t)

	const (
		parkID = inventorymodel.ParkID("park-get-date-inventory")
		date   = inventorymodel.Date("2026-10-05")
	)
	seedDateInventoryHelper(
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
	repo, _ := newRepositoryHelper(t)

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
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const (
		parkID = inventorymodel.ParkID("park-update-date-inventory")
		date   = inventorymodel.Date("2026-10-05")
	)
	seedDateInventoryHelper(
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
	repo, _ := newRepositoryHelper(t)

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
