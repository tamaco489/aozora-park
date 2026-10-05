package model

// 枠の生成に使う設定が満たす範囲、park 側の検証規則と同じ値を置く
//
// 機能パッケージ同士は import しないため、inventory は park のマスタを読み取り専用の型として自前で持つ
const (
	inventoryDaysMin   = 1
	inventoryDaysMax   = 90
	intervalMinutesMin = 1
)

// ParkMaster は枠の生成に使うパークの設定
//
// 枠を作成するジョブだけが読む、更新する手段は持たない
type ParkMaster struct {
	id                   ParkID
	defaultDailyCapacity int32
	inventoryDays        int32
}

// RestoreParkMaster は保存済みのパークの設定を組み立てる
//
// 枠の日数と初期値が壊れていると生成そのものが誤るため、読み出した値も検証を通す
func RestoreParkMaster(
	id ParkID,
	defaultDailyCapacity int32,
	inventoryDays int32,
) (*ParkMaster, error) {
	if id == "" {
		return nil, ErrInvalidParkID
	}

	if defaultDailyCapacity < capacityMin {
		return nil, ErrInvalidCapacity
	}

	if inventoryDays < inventoryDaysMin || inventoryDays > inventoryDaysMax {
		return nil, ErrInvalidInventoryDays
	}

	return &ParkMaster{
		id:                   id,
		defaultDailyCapacity: defaultDailyCapacity,
		inventoryDays:        inventoryDays,
	}, nil
}

func (p *ParkMaster) ID() ParkID { return p.id }

func (p *ParkMaster) DefaultDailyCapacity() int32 { return p.defaultDailyCapacity }

func (p *ParkMaster) InventoryDays() int32 { return p.inventoryDays }

// AttractionMaster は枠の生成に使うアトラクションの優先パスの条件
//
// 開始時刻と終了時刻は 0 時からの分で持つ、刻みの計算と表記の組み立てを 1 か所に寄せるため
type AttractionMaster struct {
	parkID              ParkID
	id                  AttractionID
	priorityPassEnabled bool
	startMinutes        int32
	endMinutes          int32
	intervalMinutes     int32
	capacityPerSlot     int32
}

// RestoreAttractionMaster は保存済みのアトラクションの条件を組み立てる
//
// 無効にしていても後から有効にするため、時刻と枠の条件は enabled によらず検証する
func RestoreAttractionMaster(
	parkID ParkID,
	id AttractionID,
	priorityPassEnabled bool,
	startTime string,
	endTime string,
	intervalMinutes int32,
	capacityPerSlot int32,
) (*AttractionMaster, error) {
	if parkID == "" {
		return nil, ErrInvalidParkID
	}

	if id == "" {
		return nil, ErrInvalidAttractionID
	}

	startMinutes, ok := parseMinutes(startTime)
	if !ok {
		return nil, ErrInvalidStartTime
	}

	endMinutes, ok := parseMinutes(endTime)
	if !ok {
		return nil, ErrInvalidEndTime
	}

	// 等しい場合も枠が 1 つも作れないため弾く
	if startMinutes >= endMinutes {
		return nil, ErrInvalidTimeRange
	}

	// 0 以下だと開始時刻から先に進めず、枠を刻み終えられない
	if intervalMinutes < intervalMinutesMin {
		return nil, ErrInvalidIntervalMinutes
	}

	if capacityPerSlot < capacityMin {
		return nil, ErrInvalidCapacity
	}

	return &AttractionMaster{
		parkID:              parkID,
		id:                  id,
		priorityPassEnabled: priorityPassEnabled,
		startMinutes:        startMinutes,
		endMinutes:          endMinutes,
		intervalMinutes:     intervalMinutes,
		capacityPerSlot:     capacityPerSlot,
	}, nil
}

// StartTimes は時間帯枠の開始時刻を HH:MM で昇順に返す
//
// 終了時刻に始まる枠は作らない、終了時刻は枠の終わりであって始まりではないため
func (a *AttractionMaster) StartTimes() []string {
	startTimes := make(
		[]string,
		0,
		(a.endMinutes-a.startMinutes)/a.intervalMinutes+1,
	)
	for minutes := a.startMinutes; minutes < a.endMinutes; minutes += a.intervalMinutes {
		startTimes = append(startTimes, formatMinutes(minutes))
	}

	return startTimes
}

func (a *AttractionMaster) ParkID() ParkID { return a.parkID }

func (a *AttractionMaster) ID() AttractionID { return a.id }

func (a *AttractionMaster) PriorityPassEnabled() bool { return a.priorityPassEnabled }

func (a *AttractionMaster) StartTime() string { return formatMinutes(a.startMinutes) }

func (a *AttractionMaster) EndTime() string { return formatMinutes(a.endMinutes) }

func (a *AttractionMaster) IntervalMinutes() int32 { return a.intervalMinutes }

func (a *AttractionMaster) CapacityPerSlot() int32 { return a.capacityPerSlot }
