package handler

import (
	"context"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

// fakeRepository は保存先を差し替えるためのインメモリ実装
//
// usecase は実物を通す、差し替えるのは domain/repository だけにする
type fakeRepository struct {
	parks map[parkmodel.ParkID]*parkmodel.Park
}

var (
	_ parkrepository.Reader = (*fakeRepository)(nil)
	_ parkrepository.Writer = (*fakeRepository)(nil)
)

func (r *fakeRepository) Get(_ context.Context, id parkmodel.ParkID) (*parkmodel.Park, error) {
	park, ok := r.parks[id]
	if !ok {
		return nil, parkmodel.ErrNotFound
	}
	return park, nil
}

func (r *fakeRepository) Create(_ context.Context, park *parkmodel.Park) error {
	r.parks[park.ID()] = park
	return nil
}

func (r *fakeRepository) Update(_ context.Context, park *parkmodel.Park) error {
	if _, ok := r.parks[park.ID()]; !ok {
		return parkmodel.ErrNotFound
	}
	r.parks[park.ID()] = park
	return nil
}

// newHandler は渡したパークを保存済みにしたハンドラを組み立てる
func newHandler(tb testing.TB, stored ...*parkmodel.Park) (*Connect, *fakeRepository) {
	tb.Helper()

	repo := &fakeRepository{parks: map[parkmodel.ParkID]*parkmodel.Park{}}
	for _, park := range stored {
		repo.parks[park.ID()] = park
	}

	handler := NewConnect(
		parkusecase.NewCreate(repo),
		parkusecase.NewGet(repo),
		parkusecase.NewUpdate(repo, repo),
	)

	return handler, repo
}

func restore(tb testing.TB) *parkmodel.Park {
	tb.Helper()

	park, err := parkmodel.Restore("park-1", "Aozora Park", 1000, 30)
	if err != nil {
		tb.Fatalf("Restore() = %v, want nil", err)
	}
	return park
}
