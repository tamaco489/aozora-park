// Package model はパークのエンティティと検証規則を持つ
package model

import (
	"unicode/utf8"
	"uuid"
)

// 表示名と枠の初期値が満たす範囲、proto の制約と同じ値を置く
//
// 入口を通らない経路 (ジョブ・再実行) があるため、ここでも検査する
const (
	nameMaxLen       = 100
	inventoryDaysMin = 1
	inventoryDaysMax = 90
	dailyCapacityMin = 1
)

// ParkID はパークの識別子
type ParkID string

func (id ParkID) String() string { return string(id) }

// Park は来園予約の最上位となるパーク
type Park struct {
	id                   ParkID
	name                 string
	defaultDailyCapacity int32
	inventoryDays        int32
}

// New はパークを新しく作る
//
// 識別子は UUID v7 で採番する、生成順に並ぶので一覧の並びが安定する
// 呼び出し側は戻り値から識別子を取れるため、採番の差し替えは用意しない
func New(name string, defaultDailyCapacity, inventoryDays int32) (*Park, error) {
	return newPark(
		ParkID(uuid.NewV7().String()),
		name,
		defaultDailyCapacity,
		inventoryDays,
	)
}

// Restore は保存済みのパークを組み立てる
//
// infrastructure が読み出した値を入れる、保存されている値も検証を通す
func Restore(id ParkID, name string, defaultDailyCapacity, inventoryDays int32) (*Park, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	return newPark(id, name, defaultDailyCapacity, inventoryDays)
}

func newPark(id ParkID, name string, defaultDailyCapacity, inventoryDays int32) (*Park, error) {
	p := &Park{id: id}
	if err := p.apply(name, defaultDailyCapacity, inventoryDays); err != nil {
		return nil, err
	}
	return p, nil
}

// Update は表示名と枠の生成に使う初期値を差し替える
//
// 検証に失敗したときは元の値を保つ
func (p *Park) Update(name string, defaultDailyCapacity, inventoryDays int32) error {
	return p.apply(name, defaultDailyCapacity, inventoryDays)
}

func (p *Park) apply(name string, defaultDailyCapacity, inventoryDays int32) error {
	err := validate(
		name,
		defaultDailyCapacity,
		inventoryDays,
	)
	if err != nil {
		return err
	}

	p.name = name
	p.defaultDailyCapacity = defaultDailyCapacity
	p.inventoryDays = inventoryDays

	return nil
}

func validate(name string, defaultDailyCapacity, inventoryDays int32) error {
	// 文字数で数えるのは、上限が表示の崩れを防ぐための歯止めでバイト数に意味がないため
	if l := utf8.RuneCountInString(name); l < 1 || l > nameMaxLen {
		return ErrInvalidName
	}

	if defaultDailyCapacity < dailyCapacityMin {
		return ErrInvalidDailyCapacity
	}

	if inventoryDays < inventoryDaysMin || inventoryDays > inventoryDaysMax {
		return ErrInvalidInventoryDays
	}

	return nil
}

func (p *Park) ID() ParkID { return p.id }

func (p *Park) Name() string { return p.name }

func (p *Park) DefaultDailyCapacity() int32 { return p.defaultDailyCapacity }

func (p *Park) InventoryDays() int32 { return p.inventoryDays }
