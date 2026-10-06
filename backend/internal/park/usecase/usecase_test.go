package usecase

import (
	"context"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

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

// fakeRepository は Reader と Writer を満たすインメモリの保存先
type fakeRepository struct {
	parks     map[parkmodel.ParkID]*parkmodel.Park
	createErr error
	updateErr error

	attractions         map[attractionKeyHelper]*parkmodel.Attraction
	createAttractionErr error
	updateAttractionErr error

	ticketTypes         map[ticketTypeKeyHelper]*parkmodel.TicketType
	createTicketTypeErr error
	updateTicketTypeErr error
}

var (
	_ parkrepository.Reader           = (*fakeRepository)(nil)
	_ parkrepository.Writer           = (*fakeRepository)(nil)
	_ parkrepository.AttractionReader = (*fakeRepository)(nil)
	_ parkrepository.AttractionWriter = (*fakeRepository)(nil)
	_ parkrepository.TicketTypeReader = (*fakeRepository)(nil)
	_ parkrepository.TicketTypeWriter = (*fakeRepository)(nil)
)

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		parks:       map[parkmodel.ParkID]*parkmodel.Park{},
		attractions: map[attractionKeyHelper]*parkmodel.Attraction{},
		ticketTypes: map[ticketTypeKeyHelper]*parkmodel.TicketType{},
	}
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
func storeHelper(tb testing.TB, repo *fakeRepository) *parkmodel.Park {
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
	repo.parks[park.ID()] = park

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
	if r.createAttractionErr != nil {
		return r.createAttractionErr
	}

	key := attractionKeyHelper{
		parkID: attraction.ParkID(),
		id:     attraction.ID(),
	}
	if _, ok := r.attractions[key]; ok {
		return parkmodel.ErrAttractionAlreadyExists
	}
	r.attractions[key] = attraction

	return nil
}

func (r *fakeRepository) UpdateAttraction(_ context.Context, attraction *parkmodel.Attraction) error {
	if r.updateAttractionErr != nil {
		return r.updateAttractionErr
	}

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
	if r.createTicketTypeErr != nil {
		return r.createTicketTypeErr
	}

	key := ticketTypeKeyHelper{
		parkID: ticketType.ParkID(),
		id:     ticketType.ID(),
	}
	if _, ok := r.ticketTypes[key]; ok {
		return parkmodel.ErrTicketTypeAlreadyExists
	}
	r.ticketTypes[key] = ticketType

	return nil
}

func (r *fakeRepository) UpdateTicketType(_ context.Context, ticketType *parkmodel.TicketType) error {
	if r.updateTicketTypeErr != nil {
		return r.updateTicketTypeErr
	}

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

// priorityPassConfig はテストで使う優先パスの条件を組み立てる
func priorityPassConfigHelper(tb testing.TB) parkmodel.PriorityPassConfig {
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

// storeAttraction は保存済みのアトラクションを 1 件用意する
func storeAttractionHelper(tb testing.TB, repo *fakeRepository) *parkmodel.Attraction {
	tb.Helper()

	attraction, err := parkmodel.RestoreAttraction(
		"park-1",
		"attraction-1",
		"ジェットコースター",
		priorityPassConfigHelper(tb),
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

// storeTicketType は保存済みの券種を 1 件用意する
// storeAttractionAs は表示名の重複を確かめるために、親と識別子と表示名を指定して 1 件積む
func storeAttractionAsHelper(
	tb testing.TB,
	repo *fakeRepository,
	parkID parkmodel.ParkID,
	id parkmodel.AttractionID,
	name string,
) *parkmodel.Attraction {
	tb.Helper()

	attraction, err := parkmodel.RestoreAttraction(
		parkID,
		id,
		name,
		priorityPassConfigHelper(tb),
	)
	if err != nil {
		tb.Fatalf("RestoreAttraction() = %v, want nil", err)
	}
	repo.attractions[attractionKeyHelper{
		parkID: parkID,
		id:     id,
	}] = attraction

	return attraction
}

// storeTicketTypeAs は storeAttractionAs と同じ目的で券種を 1 件積む
func storeTicketTypeAsHelper(
	tb testing.TB,
	repo *fakeRepository,
	parkID parkmodel.ParkID,
	id parkmodel.TicketTypeID,
	name string,
) *parkmodel.TicketType {
	tb.Helper()

	ticketType, err := parkmodel.RestoreTicketType(
		parkID,
		id,
		name,
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		tb.Fatalf("RestoreTicketType() = %v, want nil", err)
	}
	repo.ticketTypes[ticketTypeKeyHelper{
		parkID: parkID,
		id:     id,
	}] = ticketType

	return ticketType
}

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
