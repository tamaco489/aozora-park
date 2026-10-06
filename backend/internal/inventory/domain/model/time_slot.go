package model

import (
	"strings"
)

// TimeSlot はアトラクションの時間帯ごとの枠
type TimeSlot struct {
	parkID       ParkID
	attractionID AttractionID
	id           TimeSlotID
	date         Date
	startTime    string
	capacity     int32
	remaining    int32
}

// NewTimeSlotID は日付と開始時刻から時間帯枠の識別子を組み立てる
//
// 同じ日の同じ開始時刻には必ず同じ識別子が出るため、二重に作成しても上書きにならない
func NewTimeSlotID(date Date, startTime string) TimeSlotID {
	return TimeSlotID(strings.ReplaceAll(date.String(), "-", "") +
		"_" +
		strings.ReplaceAll(startTime, ":", ""))
}

// NewTimeSlot は枠を作成するジョブが時間帯枠を新しく生成する
//
// 残りは上限と同じ値から始まる、まだ誰も申し込んでいないため
func NewTimeSlot(
	parkID ParkID,
	attractionID AttractionID,
	date Date,
	startTime string,
	capacity int32,
) (*TimeSlot, error) {
	return RestoreTimeSlot(
		parkID,
		attractionID,
		NewTimeSlotID(date, startTime),
		date,
		startTime,
		capacity,
		capacity,
	)
}

// RestoreTimeSlot は保存済みの時間帯枠を組み立てる
func RestoreTimeSlot(
	parkID ParkID,
	attractionID AttractionID,
	id TimeSlotID,
	date Date,
	startTime string,
	capacity, remaining int32,
) (*TimeSlot, error) {
	if parkID == "" {
		return nil, ErrInvalidParkID
	}

	if attractionID == "" {
		return nil, ErrInvalidAttractionID
	}

	if id == "" {
		return nil, ErrInvalidTimeSlotID
	}

	if err := validateDate(date); err != nil {
		return nil, err
	}

	if err := validateStartTime(startTime); err != nil {
		return nil, err
	}

	if err := validateQuantity(capacity, remaining); err != nil {
		return nil, err
	}

	return &TimeSlot{
		parkID:       parkID,
		attractionID: attractionID,
		id:           id,
		date:         date,
		startTime:    startTime,
		capacity:     capacity,
		remaining:    remaining,
	}, nil
}

func (t *TimeSlot) ParkID() ParkID { return t.parkID }

func (t *TimeSlot) AttractionID() AttractionID { return t.attractionID }

func (t *TimeSlot) ID() TimeSlotID { return t.id }

func (t *TimeSlot) Date() Date { return t.date }

func (t *TimeSlot) StartTime() string { return t.startTime }

func (t *TimeSlot) Capacity() int32 { return t.capacity }

func (t *TimeSlot) Remaining() int32 { return t.remaining }
