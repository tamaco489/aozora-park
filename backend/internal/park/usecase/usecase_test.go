package usecase

import (
	"context"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// fakeRepository は Reader と Writer を満たすインメモリの保存先
type fakeRepository struct {
	parks     map[parkmodel.ParkID]*parkmodel.Park
	createErr error
	updateErr error
}

var (
	_ parkrepository.Reader = (*fakeRepository)(nil)
	_ parkrepository.Writer = (*fakeRepository)(nil)
)

func newFakeRepository() *fakeRepository {
	return &fakeRepository{parks: map[parkmodel.ParkID]*parkmodel.Park{}}
}

func (r *fakeRepository) Get(_ context.Context, id parkmodel.ParkID) (*parkmodel.Park, error) {
	park, ok := r.parks[id]
	if !ok {
		return nil, parkmodel.ErrNotFound
	}
	return park, nil
}

func (r *fakeRepository) Create(_ context.Context, park *parkmodel.Park) error {
	if r.createErr != nil {
		return r.createErr
	}
	if _, ok := r.parks[park.ID()]; ok {
		return parkmodel.ErrAlreadyExists
	}
	r.parks[park.ID()] = park
	return nil
}

func (r *fakeRepository) Update(_ context.Context, park *parkmodel.Park) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	if _, ok := r.parks[park.ID()]; !ok {
		return parkmodel.ErrNotFound
	}
	r.parks[park.ID()] = park
	return nil
}

// store は保存済みのパークを 1 件用意する
func store(tb testing.TB, repo *fakeRepository) *parkmodel.Park {
	tb.Helper()

	park, err := parkmodel.Restore("park-1", "Aozora Park", 1000, 30)
	if err != nil {
		tb.Fatalf("Restore() = %v, want nil", err)
	}
	repo.parks[park.ID()] = park

	return park
}
