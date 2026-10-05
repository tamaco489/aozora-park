// Package model は枠在庫のエンティティと検証規則を持つ
package model

import (
	"time"
)

// 枠の値が満たす範囲と形式
//
// 入口を通らない経路 (ジョブ・再実行) があるため、ここでも検査する
const (
	capacityMin = 1
	// 日付と開始時刻は桁を詰めた表記を許さない、ドキュメント ID の組み立てに桁数が揃っていることを使うため
	dateLen      = 10
	startTimeLen = 5
	// startTimeLayout は HH:MM、time には対応する定数がない
	startTimeLayout = "15:04"
)

// ParkID は枠が属するパークの識別子
type ParkID string

func (id ParkID) String() string { return string(id) }

// AttractionID は枠が属するアトラクションの識別子
type AttractionID string

func (id AttractionID) String() string { return string(id) }

// TimeSlotID は時間帯枠の識別子、YYYYMMDD_HHMM の形
type TimeSlotID string

func (id TimeSlotID) String() string { return string(id) }

// Date は枠の日付、YYYY-MM-DD の形
type Date string

func (d Date) String() string { return string(d) }

func validateDate(date Date) error {
	// time.Parse は 1 桁の月日も受け入れるため、桁数も見る
	if len(date) != dateLen {
		return ErrInvalidDate
	}

	if _, err := time.Parse(time.DateOnly, date.String()); err != nil {
		return ErrInvalidDate
	}

	return nil
}

func validateStartTime(startTime string) error {
	if len(startTime) != startTimeLen {
		return ErrInvalidStartTime
	}

	if _, err := time.Parse(startTimeLayout, startTime); err != nil {
		return ErrInvalidStartTime
	}

	return nil
}

// validateQuantity は枠の不変条件を検査する、残りは 0 以上かつ上限以下
func validateQuantity(capacity, remaining int32) error {
	if capacity < capacityMin {
		return ErrInvalidCapacity
	}

	if remaining < 0 || remaining > capacity {
		return ErrInvalidRemaining
	}

	return nil
}
