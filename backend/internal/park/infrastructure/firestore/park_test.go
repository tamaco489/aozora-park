package firestore

import (
	"context"
	"errors"
	"os"
	"testing"

	gcpfirestore "cloud.google.com/go/firestore"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore/firestoretest"
)

func TestMain(m *testing.M) {
	os.Exit(firestoretest.Main(m))
}

// newRepository は保存先を用意し、テストが作ったドキュメントを片付ける
//
// テスト間の分離はドキュメント ID を分けて行う、同じコレクションを共有するため
func newRepository(tb testing.TB) (*Repository, *gcpfirestore.Client) {
	tb.Helper()

	client := firestoretest.Client(tb)

	return NewRepository(client), client
}

func cleanup(tb testing.TB, client *gcpfirestore.Client, id parkmodel.ParkID) {
	tb.Helper()

	tb.Cleanup(func() {
		if _, err := client.Collection(collection).Doc(id.String()).Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil", id, err)
		}
	})
}

func TestRepositoryCreateAndGet(t *testing.T) {
	repo, client := newRepository(t)
	ctx := context.Background()

	park, err := parkmodel.New("Aozora Park", 1000, 30)
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}
	cleanup(t, client, park.ID())

	if err := repo.Create(ctx, park); err != nil {
		t.Fatalf("Repository.Create(%q) = %v, want nil", park.ID(), err)
	}

	got, err := repo.Get(ctx, park.ID())
	if err != nil {
		t.Fatalf("Repository.Get(%q) = %v, want nil", park.ID(), err)
	}

	if got.ID() != park.ID() || got.Name() != park.Name() ||
		got.DefaultDailyCapacity() != park.DefaultDailyCapacity() || got.InventoryDays() != park.InventoryDays() {
		t.Errorf("Repository.Get(%q) = (%q, %q, %d, %d), want (%q, %q, %d, %d)",
			park.ID(), got.ID(), got.Name(), got.DefaultDailyCapacity(), got.InventoryDays(),
			park.ID(), park.Name(), park.DefaultDailyCapacity(), park.InventoryDays())
	}
}

func TestRepositoryCreateAlreadyExists(t *testing.T) {
	repo, client := newRepository(t)
	ctx := context.Background()

	park, err := parkmodel.New("Aozora Park", 1000, 30)
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}
	cleanup(t, client, park.ID())

	if err := repo.Create(ctx, park); err != nil {
		t.Fatalf("Repository.Create(%q) = %v, want nil", park.ID(), err)
	}

	if err := repo.Create(ctx, park); !errors.Is(err, parkmodel.ErrAlreadyExists) {
		t.Errorf("Repository.Create(%q) の 2 回目 = %v, want %v", park.ID(), err, parkmodel.ErrAlreadyExists)
	}
}

func TestRepositoryGetNotFound(t *testing.T) {
	repo, _ := newRepository(t)

	const id = parkmodel.ParkID("park-does-not-exist")

	if _, err := repo.Get(context.Background(), id); !errors.Is(err, parkmodel.ErrNotFound) {
		t.Errorf("Repository.Get(%q) = %v, want %v", id, err, parkmodel.ErrNotFound)
	}
}

func TestRepositoryUpdate(t *testing.T) {
	repo, client := newRepository(t)
	ctx := context.Background()

	park, err := parkmodel.New("Aozora Park", 1000, 30)
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}
	cleanup(t, client, park.ID())

	if err := repo.Create(ctx, park); err != nil {
		t.Fatalf("Repository.Create(%q) = %v, want nil", park.ID(), err)
	}

	if err := park.Update("Aozora Park 2", 2000, 60); err != nil {
		t.Fatalf("Park.Update() = %v, want nil", err)
	}

	if err := repo.Update(ctx, park); err != nil {
		t.Fatalf("Repository.Update(%q) = %v, want nil", park.ID(), err)
	}

	got, err := repo.Get(ctx, park.ID())
	if err != nil {
		t.Fatalf("Repository.Get(%q) = %v, want nil", park.ID(), err)
	}

	if got.Name() != "Aozora Park 2" || got.DefaultDailyCapacity() != 2000 || got.InventoryDays() != 60 {
		t.Errorf("Repository.Get(%q) = (%q, %d, %d), want (%q, %d, %d)",
			park.ID(), got.Name(), got.DefaultDailyCapacity(), got.InventoryDays(), "Aozora Park 2", 2000, 60)
	}
}

func TestRepositoryUpdateNotFound(t *testing.T) {
	repo, _ := newRepository(t)

	park, err := parkmodel.Restore("park-does-not-exist", "Aozora Park", 1000, 30)
	if err != nil {
		t.Fatalf("Restore() = %v, want nil", err)
	}

	if err := repo.Update(context.Background(), park); !errors.Is(err, parkmodel.ErrNotFound) {
		t.Errorf("Repository.Update(%q) = %v, want %v", park.ID(), err, parkmodel.ErrNotFound)
	}
}
