package usecase

import (
	"context"
	"time"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
)

// jst は枠の日付を決めるタイムゾーン
//
// パークは日本にあるため、ジョブを動かす環境の設定で「今日」がずれないよう固定する
var jst = time.FixedZone("JST", 9*60*60)

// GenerateInventoryResult は実際に作成した枠の件数
//
// 既にある枠は数えないため、2 回目の実行では 0 になる
type GenerateInventoryResult struct {
	DateInventories int
	TimeSlots       int
}

// GenerateInventory は今日から inventoryDays 日分の枠を、まだ無いものだけ作成する
type GenerateInventory struct {
	masters inventoryrepository.MasterReader
	creator inventoryrepository.Creator
	now     func() time.Time
}

// GenerateInventoryOption は GenerateInventory の既定値を差し替える
type GenerateInventoryOption func(*GenerateInventory)

// WithClock は枠の日付の基準になる時刻を差し替える
func WithClock(now func() time.Time) GenerateInventoryOption {
	return func(u *GenerateInventory) { u.now = now }
}

func NewGenerateInventory(
	masters inventoryrepository.MasterReader,
	creator inventoryrepository.Creator,
	opts ...GenerateInventoryOption,
) *GenerateInventory {
	u := &GenerateInventory{
		masters: masters,
		creator: creator,
		now:     time.Now,
	}
	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Do はすべてのパークを走査し、入場枠と時間帯枠を作成する
//
// 既にある枠には触れないため、何度実行しても件数と内容が変わらない
func (u *GenerateInventory) Do(ctx context.Context) (GenerateInventoryResult, error) {
	parks, err := u.masters.ListParks(ctx)
	if err != nil {
		return GenerateInventoryResult{}, err
	}

	today := u.now().In(jst)

	var result GenerateInventoryResult
	for _, park := range parks {
		dates := inventoryDates(today, park.InventoryDays())

		created, err := u.createDateInventories(ctx, park, dates)
		if err != nil {
			return GenerateInventoryResult{}, err
		}
		result.DateInventories += created

		created, err = u.createTimeSlots(ctx, park.ID(), dates)
		if err != nil {
			return GenerateInventoryResult{}, err
		}
		result.TimeSlots += created
	}

	return result, nil
}

func (u *GenerateInventory) createDateInventories(
	ctx context.Context,
	park *inventorymodel.ParkMaster,
	dates []inventorymodel.Date,
) (int, error) {
	count := 0
	for _, date := range dates {
		inventory, err := inventorymodel.NewDateInventory(
			park.ID(),
			date,
			park.DefaultDailyCapacity(),
		)
		if err != nil {
			return 0, err
		}

		created, err := u.creator.CreateDateInventoryIfAbsent(ctx, inventory)
		if err != nil {
			return 0, err
		}
		if created {
			count++
		}
	}

	return count, nil
}

func (u *GenerateInventory) createTimeSlots(
	ctx context.Context,
	parkID inventorymodel.ParkID,
	dates []inventorymodel.Date,
) (int, error) {
	attractions, err := u.masters.ListAttractions(ctx, parkID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, attraction := range attractions {
		// 無効なアトラクションには枠を作らない、優先パスの申込を受け付けないため
		if !attraction.PriorityPassEnabled() {
			continue
		}

		startTimes := attraction.StartTimes()
		for _, date := range dates {
			for _, startTime := range startTimes {
				slot, err := inventorymodel.NewTimeSlot(
					parkID,
					attraction.ID(),
					date,
					startTime,
					attraction.CapacityPerSlot(),
				)
				if err != nil {
					return 0, err
				}

				created, err := u.creator.CreateTimeSlotIfAbsent(ctx, slot)
				if err != nil {
					return 0, err
				}
				if created {
					count++
				}
			}
		}
	}

	return count, nil
}

// inventoryDates は今日を含む days 日分の日付を昇順で返す
func inventoryDates(today time.Time, days int32) []inventorymodel.Date {
	dates := make(
		[]inventorymodel.Date,
		0,
		days,
	)
	for i := range int(days) {
		date := today.AddDate(0, 0, i).Format(time.DateOnly)
		dates = append(dates, inventorymodel.Date(date))
	}

	return dates
}
