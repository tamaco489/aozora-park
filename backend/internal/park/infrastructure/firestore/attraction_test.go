package firestore

import (
	"context"
	"errors"
	"testing"

	gcpfirestore "cloud.google.com/go/firestore"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

// attractionParkID はアトラクションのテストが使う親のパーク、他のテストとドキュメントを分ける
const attractionParkID = parkmodel.ParkID("park-for-attraction-test")

func newAttractionConfigHelper(tb testing.TB) parkmodel.PriorityPassConfig {
	tb.Helper()

	config, err := parkmodel.NewPriorityPassConfig(
		true,
		"09:00",
		"18:00",
		30,
		10,
	)
	if err != nil {
		tb.Fatalf("NewPriorityPassConfig() = %v, want nil", err)
	}

	return config
}

func cleanupAttractionHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	attraction *parkmodel.Attraction,
) {
	tb.Helper()

	tb.Cleanup(func() {
		doc := client.Collection(collection).Doc(attraction.ParkID().String()).
			Collection(attractionCollection).Doc(attraction.ID().String())
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				attraction.ID(),
				err,
			)
		}
	})
}

func TestRepositoryCreateAndGetAttraction(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	attraction, err := parkmodel.NewAttraction(
		attractionParkID,
		"ジェットコースター",
		newAttractionConfigHelper(t),
	)
	if err != nil {
		t.Fatalf("NewAttraction() = %v, want nil", err)
	}
	cleanupAttractionHelper(
		t,
		client,
		attraction,
	)

	if err := repo.CreateAttraction(ctx, attraction); err != nil {
		t.Fatalf("Repository.CreateAttraction(%q) = %v, want nil",
			attraction.ID(),
			err,
		)
	}

	got, err := repo.GetAttraction(
		ctx,
		attraction.ParkID(),
		attraction.ID(),
	)
	if err != nil {
		t.Fatalf("Repository.GetAttraction(%q) = %v, want nil",
			attraction.ID(),
			err,
		)
	}

	if got.ID() != attraction.ID() || got.ParkID() != attraction.ParkID() || got.Name() != attraction.Name() {
		t.Errorf("Repository.GetAttraction(%q) = (%q, %q, %q), want (%q, %q, %q)",
			attraction.ID(),
			got.ParkID(),
			got.ID(),
			got.Name(),
			attraction.ParkID(),
			attraction.ID(),
			attraction.Name(),
		)
	}
	if got.PriorityPassConfig() != attraction.PriorityPassConfig() {
		t.Errorf("Repository.GetAttraction(%q) の PriorityPassConfig = %+v, want %+v",
			attraction.ID(),
			got.PriorityPassConfig(),
			attraction.PriorityPassConfig(),
		)
	}
}

func TestRepositoryCreateAttractionAlreadyExists(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	attraction, err := parkmodel.NewAttraction(
		attractionParkID,
		"ジェットコースター",
		newAttractionConfigHelper(t),
	)
	if err != nil {
		t.Fatalf("NewAttraction() = %v, want nil", err)
	}
	cleanupAttractionHelper(
		t,
		client,
		attraction,
	)

	if err := repo.CreateAttraction(ctx, attraction); err != nil {
		t.Fatalf("Repository.CreateAttraction(%q) = %v, want nil",
			attraction.ID(),
			err,
		)
	}

	if err := repo.CreateAttraction(ctx, attraction); !errors.Is(err, parkmodel.ErrAttractionAlreadyExists) {
		t.Errorf("Repository.CreateAttraction(%q) の 2 回目 = %v, want %v",
			attraction.ID(),
			err,
			parkmodel.ErrAttractionAlreadyExists,
		)
	}
}

func TestRepositoryGetAttractionNotFound(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	const id = parkmodel.AttractionID("attraction-does-not-exist")

	_, err := repo.GetAttraction(
		t.Context(),
		attractionParkID,
		id,
	)
	if !errors.Is(err, parkmodel.ErrAttractionNotFound) {
		t.Errorf("Repository.GetAttraction(%q) = %v, want %v",
			id,
			err,
			parkmodel.ErrAttractionNotFound,
		)
	}
}

func TestRepositoryUpdateAttraction(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	attraction, err := parkmodel.NewAttraction(
		attractionParkID,
		"ジェットコースター",
		newAttractionConfigHelper(t),
	)
	if err != nil {
		t.Fatalf("NewAttraction() = %v, want nil", err)
	}
	cleanupAttractionHelper(
		t,
		client,
		attraction,
	)

	if err := repo.CreateAttraction(ctx, attraction); err != nil {
		t.Fatalf("Repository.CreateAttraction(%q) = %v, want nil",
			attraction.ID(),
			err,
		)
	}

	updated, err := parkmodel.NewPriorityPassConfig(
		false,
		"10:00",
		"20:00",
		15,
		20,
	)
	if err != nil {
		t.Fatalf("NewPriorityPassConfig() = %v, want nil", err)
	}

	if err := attraction.Update("ジェットコースター 2", updated); err != nil {
		t.Fatalf("Attraction.Update() = %v, want nil", err)
	}

	if err := repo.UpdateAttraction(ctx, attraction); err != nil {
		t.Fatalf("Repository.UpdateAttraction(%q) = %v, want nil",
			attraction.ID(),
			err,
		)
	}

	got, err := repo.GetAttraction(
		ctx,
		attraction.ParkID(),
		attraction.ID(),
	)
	if err != nil {
		t.Fatalf("Repository.GetAttraction(%q) = %v, want nil",
			attraction.ID(),
			err,
		)
	}

	if got.Name() != "ジェットコースター 2" || got.PriorityPassConfig() != updated {
		t.Errorf("Repository.GetAttraction(%q) = (%q, %+v), want (%q, %+v)",
			attraction.ID(),
			got.Name(),
			got.PriorityPassConfig(),
			"ジェットコースター 2",
			updated,
		)
	}
}

func TestRepositoryUpdateAttractionNotFound(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	attraction, err := parkmodel.RestoreAttraction(
		attractionParkID,
		"attraction-does-not-exist",
		"ジェットコースター",
		newAttractionConfigHelper(t),
	)
	if err != nil {
		t.Fatalf("RestoreAttraction() = %v, want nil", err)
	}

	if err := repo.UpdateAttraction(t.Context(), attraction); !errors.Is(err, parkmodel.ErrAttractionNotFound) {
		t.Errorf("Repository.UpdateAttraction(%q) = %v, want %v",
			attraction.ID(),
			err,
			parkmodel.ErrAttractionNotFound,
		)
	}
}

func TestRepositoryFindAttractionByName(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	// 他のテストが作るアトラクションと表示名が重ならないようにする、親のパークを共有しているため
	const name = "名前引きのテスト用アトラクション"

	attraction, err := parkmodel.NewAttraction(
		attractionParkID,
		name,
		newAttractionConfigHelper(t),
	)
	if err != nil {
		t.Fatalf("NewAttraction(%q) = %v, want nil",
			name,
			err,
		)
	}
	cleanupAttractionHelper(t, client, attraction)

	if err := repo.CreateAttraction(ctx, attraction); err != nil {
		t.Fatalf("Repository.CreateAttraction(%q) = %v, want nil",
			attraction.ID(),
			err,
		)
	}

	got, err := repo.FindAttractionByName(
		ctx,
		attractionParkID,
		name,
	)
	if err != nil {
		t.Fatalf("Repository.FindAttractionByName(%q) = %v, want nil",
			name,
			err,
		)
	}

	if got.ID() != attraction.ID() {
		t.Errorf("Repository.FindAttractionByName(%q) の ID = %q, want %q",
			name,
			got.ID(),
			attraction.ID(),
		)
	}
}

func TestRepositoryFindAttractionByNameNotFound(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	const name = "保存されていないアトラクションの表示名"

	_, err := repo.FindAttractionByName(
		t.Context(),
		attractionParkID,
		name,
	)
	if !errors.Is(err, parkmodel.ErrAttractionNotFound) {
		t.Errorf("Repository.FindAttractionByName(%q) = %v, want %v",
			name,
			err,
			parkmodel.ErrAttractionNotFound,
		)
	}
}
