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
	parks       map[parkmodel.ParkID]*parkmodel.Park
	attractions map[attractionKeyHelper]*parkmodel.Attraction
	ticketTypes map[ticketTypeKeyHelper]*parkmodel.TicketType
}

// attractionKeyHelper はアトラクションを親のパークごとに分けて持つための鍵
type attractionKeyHelper struct {
	parkID parkmodel.ParkID
	id     parkmodel.AttractionID
}

// ticketTypeKeyHelper は券種を親のパークごとに分けて持つための鍵
type ticketTypeKeyHelper struct {
	parkID parkmodel.ParkID
	id     parkmodel.TicketTypeID
}

var (
	_ parkrepository.Reader           = (*fakeRepository)(nil)
	_ parkrepository.Writer           = (*fakeRepository)(nil)
	_ parkrepository.AttractionReader = (*fakeRepository)(nil)
	_ parkrepository.AttractionWriter = (*fakeRepository)(nil)
	_ parkrepository.TicketTypeReader = (*fakeRepository)(nil)
	_ parkrepository.TicketTypeWriter = (*fakeRepository)(nil)
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

// newHandlerHelper は渡したパークを保存済みにしたハンドラを組み立てる
func newHandlerHelper(tb testing.TB, stored ...*parkmodel.Park) (*Connect, *fakeRepository) {
	tb.Helper()

	repo := &fakeRepository{
		parks:       map[parkmodel.ParkID]*parkmodel.Park{},
		attractions: map[attractionKeyHelper]*parkmodel.Attraction{},
		ticketTypes: map[ticketTypeKeyHelper]*parkmodel.TicketType{},
	}
	for _, park := range stored {
		repo.parks[park.ID()] = park
	}

	handler := NewConnect(
		parkusecase.NewCreate(repo),
		parkusecase.NewGet(repo),
		parkusecase.NewUpdate(repo, repo),
		parkusecase.NewCreateAttraction(repo, repo, repo),
		parkusecase.NewUpdateAttraction(repo, repo),
		parkusecase.NewCreateTicketType(repo, repo, repo),
		parkusecase.NewUpdateTicketType(repo, repo),
	)

	return handler, repo
}

func restoreHelper(tb testing.TB) *parkmodel.Park {
	tb.Helper()

	park, err := parkmodel.Restore(
		"park-1",
		"Aozora Park",
		1000,
		30,
	)
	if err != nil {
		tb.Fatalf("Restore() = %v, want nil", err)
	}
	return park
}

func (r *fakeRepository) GetAttraction(
	_ context.Context,
	parkID parkmodel.ParkID,
	id parkmodel.AttractionID,
) (*parkmodel.Attraction, error) {
	attraction, ok := r.attractions[attractionKeyHelper{
		parkID: parkID,
		id:     id,
	}]
	if !ok {
		return nil, parkmodel.ErrAttractionNotFound
	}
	return attraction, nil
}

func (r *fakeRepository) FindAttractionByName(
	_ context.Context,
	parkID parkmodel.ParkID,
	name string,
) (*parkmodel.Attraction, error) {
	// 実装は索引で 1 件に絞るが、フェイクは件数が少ないため総当たりで足りる
	for key, attraction := range r.attractions {
		if key.parkID == parkID && attraction.Name() == name {
			return attraction, nil
		}
	}

	return nil, parkmodel.ErrAttractionNotFound
}

func (r *fakeRepository) CreateAttraction(_ context.Context, attraction *parkmodel.Attraction) error {
	r.attractions[attractionKeyHelper{
		parkID: attraction.ParkID(),
		id:     attraction.ID(),
	}] = attraction
	return nil
}

func (r *fakeRepository) UpdateAttraction(_ context.Context, attraction *parkmodel.Attraction) error {
	key := attractionKeyHelper{
		parkID: attraction.ParkID(),
		id:     attraction.ID(),
	}
	if _, ok := r.attractions[key]; !ok {
		return parkmodel.ErrAttractionNotFound
	}
	r.attractions[key] = attraction

	return nil
}

func (r *fakeRepository) GetTicketType(
	_ context.Context,
	parkID parkmodel.ParkID,
	id parkmodel.TicketTypeID,
) (*parkmodel.TicketType, error) {
	ticketType, ok := r.ticketTypes[ticketTypeKeyHelper{
		parkID: parkID,
		id:     id,
	}]
	if !ok {
		return nil, parkmodel.ErrTicketTypeNotFound
	}
	return ticketType, nil
}

func (r *fakeRepository) FindTicketTypeByName(
	_ context.Context,
	parkID parkmodel.ParkID,
	name string,
) (*parkmodel.TicketType, error) {
	// 実装は索引で 1 件に絞るが、フェイクは件数が少ないため総当たりで足りる
	for key, ticketType := range r.ticketTypes {
		if key.parkID == parkID && ticketType.Name() == name {
			return ticketType, nil
		}
	}

	return nil, parkmodel.ErrTicketTypeNotFound
}

func (r *fakeRepository) CreateTicketType(_ context.Context, ticketType *parkmodel.TicketType) error {
	r.ticketTypes[ticketTypeKeyHelper{
		parkID: ticketType.ParkID(),
		id:     ticketType.ID(),
	}] = ticketType
	return nil
}

func (r *fakeRepository) UpdateTicketType(_ context.Context, ticketType *parkmodel.TicketType) error {
	key := ticketTypeKeyHelper{
		parkID: ticketType.ParkID(),
		id:     ticketType.ID(),
	}
	if _, ok := r.ticketTypes[key]; !ok {
		return parkmodel.ErrTicketTypeNotFound
	}
	r.ticketTypes[key] = ticketType

	return nil
}

// storeAttractionHelper は保存済みのアトラクションを 1 件用意する
func storeAttractionHelper(tb testing.TB, repo *fakeRepository) *parkmodel.Attraction {
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

	attraction, err := parkmodel.RestoreAttraction(
		"park-1",
		"attraction-1",
		"ジェットコースター",
		config,
	)
	if err != nil {
		tb.Fatalf("RestoreAttraction() = %v, want nil", err)
	}
	repo.attractions[attractionKeyHelper{
		parkID: attraction.ParkID(),
		id:     attraction.ID(),
	}] = attraction

	return attraction
}

// storeTicketTypeHelper は保存済みの券種を 1 件用意する
func storeTicketTypeHelper(tb testing.TB, repo *fakeRepository) *parkmodel.TicketType {
	tb.Helper()

	ticketType, err := parkmodel.RestoreTicketType(
		"park-1",
		"ticket-type-1",
		"1 デーパスポート",
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		tb.Fatalf("RestoreTicketType() = %v, want nil", err)
	}
	repo.ticketTypes[ticketTypeKeyHelper{
		parkID: ticketType.ParkID(),
		id:     ticketType.ID(),
	}] = ticketType

	return ticketType
}
