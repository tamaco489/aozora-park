package usecase

import (
	"context"
	"testing"
	"time"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
)

// 枠の生成に使うマスタの値、生成した枠の期待値と突き合わせるため 1 か所に置く
const (
	masterParkID          = inventorymodel.ParkID("park-generate")
	masterAttractionID    = inventorymodel.AttractionID("attraction-generate")
	masterDailyCapacity   = int32(1000)
	masterInventoryDays   = int32(14)
	masterStartTime       = "09:00"
	masterEndTime         = "17:00"
	masterIntervalMinutes = int32(60)
	masterCapacityPerSlot = int32(30)
)

// fakeMasterStore は MasterReader と Creator を満たすインメモリの保存先
//
// 作成済みの枠を識別子で持ち、2 回目の作成を弾くところまで Firestore の Create と同じ振る舞いにする
type fakeMasterStore struct {
	parks       []*inventorymodel.ParkMaster
	attractions []*inventorymodel.AttractionMaster
	inventories map[string]*inventorymodel.DateInventory
	slots       map[string]*inventorymodel.TimeSlot
}

var (
	_ inventoryrepository.MasterReader = (*fakeMasterStore)(nil)
	_ inventoryrepository.Creator      = (*fakeMasterStore)(nil)
)

func newFakeMasterStore(tb testing.TB, priorityPassEnabled bool) *fakeMasterStore {
	tb.Helper()

	park, err := inventorymodel.RestoreParkMaster(
		masterParkID,
		masterDailyCapacity,
		masterInventoryDays,
	)
	if err != nil {
		tb.Fatalf("RestoreParkMaster() = %v, want nil", err)
	}

	attraction, err := inventorymodel.RestoreAttractionMaster(
		masterParkID,
		masterAttractionID,
		priorityPassEnabled,
		masterStartTime,
		masterEndTime,
		masterIntervalMinutes,
		masterCapacityPerSlot,
	)
	if err != nil {
		tb.Fatalf("RestoreAttractionMaster() = %v, want nil", err)
	}

	return &fakeMasterStore{
		parks:       []*inventorymodel.ParkMaster{park},
		attractions: []*inventorymodel.AttractionMaster{attraction},
		inventories: map[string]*inventorymodel.DateInventory{},
		slots:       map[string]*inventorymodel.TimeSlot{},
	}
}

func (s *fakeMasterStore) ListParks(_ context.Context) ([]*inventorymodel.ParkMaster, error) {
	return s.parks, nil
}

func (s *fakeMasterStore) ListAttractions(
	_ context.Context,
	parkID inventorymodel.ParkID,
) ([]*inventorymodel.AttractionMaster, error) {
	attractions := make(
		[]*inventorymodel.AttractionMaster,
		0,
		len(s.attractions),
	)
	for _, attraction := range s.attractions {
		if attraction.ParkID() == parkID {
			attractions = append(attractions, attraction)
		}
	}

	return attractions, nil
}

func (s *fakeMasterStore) CreateDateInventoryIfAbsent(
	_ context.Context,
	inventory *inventorymodel.DateInventory,
) (bool, error) {
	key := dateKey(inventory.ParkID(), inventory.Date())
	if _, ok := s.inventories[key]; ok {
		return false, nil
	}
	s.inventories[key] = inventory

	return true, nil
}

func (s *fakeMasterStore) CreateTimeSlotIfAbsent(
	_ context.Context,
	slot *inventorymodel.TimeSlot,
) (bool, error) {
	key := slotKey(
		slot.ParkID(),
		slot.AttractionID(),
		slot.Date(),
	) + "/" + slot.ID().String()
	if _, ok := s.slots[key]; ok {
		return false, nil
	}
	s.slots[key] = slot

	return true, nil
}

// generatedAt は枠の日付の基準にする時刻
//
// UTC では 2026-10-04 だが JST では 2026-10-05 になる時刻を選び、日付が JST で決まることを確かめる
var generatedAt = time.Date(
	2026,
	time.October,
	4,
	16,
	0,
	0,
	0,
	time.UTC,
)

func newGenerateInventory(store *fakeMasterStore) *GenerateInventory {
	return NewGenerateInventory(
		store,
		store,
		WithClock(func() time.Time { return generatedAt }),
	)
}

func TestGenerateInventoryDo(t *testing.T) {
	store := newFakeMasterStore(t, true)

	got, err := newGenerateInventory(store).Do(t.Context())
	if err != nil {
		t.Fatalf("GenerateInventory.Do() のエラー = %v, want nil", err)
	}

	// 9 時から 17 時を 60 分で刻むと、終了時刻に始まる枠を除いて 1 日 8 枠になる
	want := GenerateInventoryResult{
		DateInventories: 14,
		TimeSlots:       8 * 14,
	}
	if got != want {
		t.Errorf("GenerateInventory.Do() = %+v, want %+v",
			got,
			want,
		)
	}

	// 今日を含む 14 日分のため、末尾は 10-18 になる
	first := inventorymodel.Date("2026-10-05")
	last := inventorymodel.Date("2026-10-18")
	for _, date := range []inventorymodel.Date{first, last} {
		inventory, ok := store.inventories[dateKey(masterParkID, date)]
		if !ok {
			t.Fatalf("GenerateInventory.Do() が %q の入場枠を作成していない", date)
		}

		if inventory.Capacity() != masterDailyCapacity || inventory.Remaining() != masterDailyCapacity {
			t.Errorf("GenerateInventory.Do() の %q の枠 = (%d, %d), want (%d, %d)",
				date,
				inventory.Capacity(),
				inventory.Remaining(),
				masterDailyCapacity,
				masterDailyCapacity,
			)
		}
	}

	// 開始時刻はゼロ埋めする、文字列の昇順が時刻の昇順と一致しないと一覧の並びが壊れる
	slotKeyOfFirst := slotKey(
		masterParkID,
		masterAttractionID,
		first,
	) + "/20261005_0900"
	slot, ok := store.slots[slotKeyOfFirst]
	if !ok {
		t.Fatalf("GenerateInventory.Do() が %q の時間帯枠を作成していない", slotKeyOfFirst)
	}

	if slot.StartTime() != masterStartTime || slot.Capacity() != masterCapacityPerSlot {
		t.Errorf("GenerateInventory.Do() の %q の枠 = (%q, %d), want (%q, %d)",
			slotKeyOfFirst,
			slot.StartTime(),
			slot.Capacity(),
			masterStartTime,
			masterCapacityPerSlot,
		)
	}
}

func TestGenerateInventoryDoTwice(t *testing.T) {
	store := newFakeMasterStore(t, true)
	generator := newGenerateInventory(store)
	ctx := t.Context()

	if _, err := generator.Do(ctx); err != nil {
		t.Fatalf("1 回目の GenerateInventory.Do() のエラー = %v, want nil", err)
	}

	// 既にある枠の残りを運営が減らしても、再実行で初期値に戻らないことを確かめる
	date := inventorymodel.Date("2026-10-05")
	inventory := store.inventories[dateKey(masterParkID, date)]
	if err := inventory.Overwrite(masterDailyCapacity, 1); err != nil {
		t.Fatalf("DateInventory.Overwrite(%d, 1) = %v, want nil",
			masterDailyCapacity,
			err,
		)
	}

	beforeInventories := len(store.inventories)
	beforeSlots := len(store.slots)

	got, err := generator.Do(ctx)
	if err != nil {
		t.Fatalf("2 回目の GenerateInventory.Do() のエラー = %v, want nil", err)
	}

	if got != (GenerateInventoryResult{}) {
		t.Errorf("2 回目の GenerateInventory.Do() = %+v, want %+v",
			got,
			GenerateInventoryResult{},
		)
	}

	if len(store.inventories) != beforeInventories || len(store.slots) != beforeSlots {
		t.Errorf("2 回目の GenerateInventory.Do() の件数 = (%d, %d), want (%d, %d)",
			len(store.inventories),
			len(store.slots),
			beforeInventories,
			beforeSlots,
		)
	}

	if got := store.inventories[dateKey(masterParkID, date)].Remaining(); got != 1 {
		t.Errorf("2 回目の GenerateInventory.Do() の %q の残り = %d, want %d",
			date,
			got,
			1,
		)
	}
}

func TestGenerateInventoryDoPriorityPassDisabled(t *testing.T) {
	store := newFakeMasterStore(t, false)

	got, err := newGenerateInventory(store).Do(t.Context())
	if err != nil {
		t.Fatalf("GenerateInventory.Do() のエラー = %v, want nil", err)
	}

	want := GenerateInventoryResult{
		DateInventories: 14,
		TimeSlots:       0,
	}
	if got != want {
		t.Errorf("GenerateInventory.Do() = %+v, want %+v",
			got,
			want,
		)
	}

	if len(store.slots) != 0 {
		t.Errorf("GenerateInventory.Do() の時間帯枠の件数 = %d, want %d",
			len(store.slots),
			0,
		)
	}
}
