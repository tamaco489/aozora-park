package model

import (
	"uuid"
)

// 価格が満たす範囲、proto の制約と同じ値を置く
//
// 入口を通らない経路 (ジョブ・再実行) があるため、ここでも検査する
const priceMin = 0

// TicketTypeID は券種の識別子
type TicketTypeID string

func (id TicketTypeID) String() string { return string(id) }

// TicketType はパークが販売する券種
type TicketType struct {
	parkID        ParkID
	id            TicketTypeID
	name          string
	price         int64  // price の単位は円
	entryTimeFrom string // entryTimeFrom は入場できる時間帯の開始時刻、HH:MM の 24 時間表記
	entryTimeTo   string // entryTimeTo は入場できる時間帯の終了時刻、HH:MM の 24 時間表記
}

// NewTicketType は券種を新しく生成する
//
// 識別子は UUID v7 で採番する、生成順に並ぶので一覧の並びが安定する
func NewTicketType(parkID ParkID, name string, price int64, entryTimeFrom, entryTimeTo string) (*TicketType, error) {
	return newTicketType(
		parkID,
		TicketTypeID(uuid.NewV7().String()),
		name,
		price,
		entryTimeFrom,
		entryTimeTo,
	)
}

// RestoreTicketType は保存済みの券種を組み立てる
//
// infrastructure が読み出した値を入れる、保存されている値も検証を通す
func RestoreTicketType(parkID ParkID, id TicketTypeID, name string, price int64, entryTimeFrom, entryTimeTo string) (*TicketType, error) {
	if id == "" {
		return nil, ErrTicketTypeInvalidID
	}
	return newTicketType(parkID, id, name, price, entryTimeFrom, entryTimeTo)
}

func newTicketType(parkID ParkID, id TicketTypeID, name string, price int64, entryTimeFrom, entryTimeTo string) (*TicketType, error) {
	if parkID == "" {
		return nil, ErrInvalidID
	}

	t := &TicketType{parkID: parkID, id: id}
	if err := t.apply(name, price, entryTimeFrom, entryTimeTo); err != nil {
		return nil, err
	}

	return t, nil
}

// Update は表示名と価格と入場できる時間帯を差し替える
//
// 検証に失敗したときは元の値を保つ
func (t *TicketType) Update(name string, price int64, entryTimeFrom, entryTimeTo string) error {
	return t.apply(name, price, entryTimeFrom, entryTimeTo)
}

func (t *TicketType) apply(name string, price int64, entryTimeFrom, entryTimeTo string) error {
	err := validateTicketType(name, price, entryTimeFrom, entryTimeTo)
	if err != nil {
		return err
	}

	t.name = name
	t.price = price
	t.entryTimeFrom = entryTimeFrom
	t.entryTimeTo = entryTimeTo

	return nil
}

func validateTicketType(name string, price int64, entryTimeFrom, entryTimeTo string) error {
	if !isValidName(name) {
		return ErrTicketTypeInvalidName
	}

	if price < priceMin {
		return ErrTicketTypeInvalidPrice
	}

	if !isValidTimeOfDay(entryTimeFrom) || !isValidTimeOfDay(entryTimeTo) {
		return ErrTicketTypeInvalidEntryTime
	}

	// 等しい場合も入場できる時間が無くなるため弾く
	if !isBefore(entryTimeFrom, entryTimeTo) {
		return ErrTicketTypeInvalidEntryTimeRange
	}

	return nil
}

func (t *TicketType) ParkID() ParkID { return t.parkID }

func (t *TicketType) ID() TicketTypeID { return t.id }

func (t *TicketType) Name() string { return t.name }

func (t *TicketType) Price() int64 { return t.price }

func (t *TicketType) EntryTimeFrom() string { return t.entryTimeFrom }

func (t *TicketType) EntryTimeTo() string { return t.entryTimeTo }
