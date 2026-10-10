// Package model は優先パスのエンティティと検証規則を持つ
package model

import (
	"regexp"
)

// timeSlotID は時間帯枠の識別子、YYYYMMDD_HHMM の形で proto の制約と同じ値を置く
//
// 入口を通らない経路 (ジョブ・再実行) があるため、ここでも検査する
var timeSlotID = regexp.MustCompile(`^[0-9]{8}_[0-9]{4}$`)

// PassID は優先パスの識別子
type PassID string

func (id PassID) String() string { return string(id) }

// ParkID は優先パスが属するパークの識別子
type ParkID string

func (id ParkID) String() string { return string(id) }

// TicketID は申込に使った券の識別子
type TicketID string

func (id TicketID) String() string { return string(id) }

// AttractionID は優先パスの対象となるアトラクションの識別子
type AttractionID string

func (id AttractionID) String() string { return string(id) }

// TimeSlotID は希望する時間帯枠の識別子、YYYYMMDD_HHMM の形
type TimeSlotID string

func (id TimeSlotID) String() string { return string(id) }

// validateTimeSlotID は時間帯枠の識別子が YYYYMMDD_HHMM の形かを確かめる
//
// 日付と時刻の妥当性は枠そのものが持つため、ここでは形式だけを見る
func validateTimeSlotID(id TimeSlotID) error {
	if !timeSlotID.MatchString(id.String()) {
		return ErrInvalidTimeSlotID
	}

	return nil
}
