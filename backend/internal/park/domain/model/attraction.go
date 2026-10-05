package model

import (
	"uuid"
)

// 優先パスの条件が満たす範囲、proto の制約と同じ値を置く
//
// 入口を通らない経路 (ジョブ・再実行) があるため、ここでも検査する
const (
	intervalMinutesMin = 1
	capacityPerSlotMin = 1
)

// AttractionID はアトラクションの識別子
type AttractionID string

func (id AttractionID) String() string { return string(id) }

// PriorityPassConfig は優先パスの時間帯枠を生成する条件
//
// 単体で検証規則を持ち、Attraction の生成と更新で同じ組を受け渡すため型にする
type PriorityPassConfig struct {
	enabled         bool
	startTime       string
	endTime         string
	intervalMinutes int32
	capacityPerSlot int32
}

// NewPriorityPassConfig は優先パスの条件を組み立てる
//
// 無効にしていても後から有効にするため、時刻と枠の条件は enabled によらず検証する
func NewPriorityPassConfig(enabled bool, startTime, endTime string, intervalMinutes, capacityPerSlot int32) (PriorityPassConfig, error) {
	c := PriorityPassConfig{
		enabled:         enabled,
		startTime:       startTime,
		endTime:         endTime,
		intervalMinutes: intervalMinutes,
		capacityPerSlot: capacityPerSlot,
	}
	if err := c.validate(); err != nil {
		return PriorityPassConfig{}, err
	}

	return c, nil
}

func (c PriorityPassConfig) validate() error {
	if !isValidTimeOfDay(c.startTime) || !isValidTimeOfDay(c.endTime) {
		return ErrAttractionInvalidTime
	}

	// 等しい場合も枠が 1 つも作れないため弾く
	if !isBefore(c.startTime, c.endTime) {
		return ErrAttractionInvalidTimeRange
	}

	if c.intervalMinutes < intervalMinutesMin {
		return ErrAttractionInvalidIntervalMinutes
	}

	if c.capacityPerSlot < capacityPerSlotMin {
		return ErrAttractionInvalidCapacityPerSlot
	}

	return nil
}

func (c PriorityPassConfig) Enabled() bool { return c.enabled }

func (c PriorityPassConfig) StartTime() string { return c.startTime }

func (c PriorityPassConfig) EndTime() string { return c.endTime }

func (c PriorityPassConfig) IntervalMinutes() int32 { return c.intervalMinutes }

func (c PriorityPassConfig) CapacityPerSlot() int32 { return c.capacityPerSlot }

// Attraction はパークに属するアトラクション
type Attraction struct {
	parkID             ParkID
	id                 AttractionID
	name               string
	priorityPassConfig PriorityPassConfig
}

// NewAttraction はアトラクションを新しく生成する
//
// 識別子は UUID v7 で採番する、生成順に並ぶので一覧の並びが安定する
func NewAttraction(parkID ParkID, name string, config PriorityPassConfig) (*Attraction, error) {
	return newAttraction(parkID, AttractionID(uuid.NewV7().String()), name, config)
}

// RestoreAttraction は保存済みのアトラクションを組み立てる
//
// infrastructure が読み出した値を入れる、保存されている値も検証を通す
func RestoreAttraction(parkID ParkID, id AttractionID, name string, config PriorityPassConfig) (*Attraction, error) {
	if id == "" {
		return nil, ErrAttractionInvalidID
	}
	return newAttraction(parkID, id, name, config)
}

func newAttraction(parkID ParkID, id AttractionID, name string, config PriorityPassConfig) (*Attraction, error) {
	if parkID == "" {
		return nil, ErrInvalidID
	}

	a := &Attraction{parkID: parkID, id: id}
	if err := a.apply(name, config); err != nil {
		return nil, err
	}

	return a, nil
}

// Update は表示名と優先パスの条件を差し替える
//
// 検証に失敗したときは元の値を保つ
func (a *Attraction) Update(name string, config PriorityPassConfig) error {
	return a.apply(name, config)
}

func (a *Attraction) apply(name string, config PriorityPassConfig) error {
	if !isValidName(name) {
		return ErrAttractionInvalidName
	}

	// ゼロ値の PriorityPassConfig を渡されても不変条件を満たすよう、組み立て済みの値もここで検査する
	if err := config.validate(); err != nil {
		return err
	}

	a.name = name
	a.priorityPassConfig = config

	return nil
}

func (a *Attraction) ParkID() ParkID { return a.parkID }

func (a *Attraction) ID() AttractionID { return a.id }

func (a *Attraction) Name() string { return a.name }

func (a *Attraction) PriorityPassConfig() PriorityPassConfig { return a.priorityPassConfig }
