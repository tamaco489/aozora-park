package model

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

// RestoreTimeSlot は保存済みの時間帯枠を組み立てる
//
// 枠の作成は専用のジョブが行うため、この機能パッケージは新規生成を持たない
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
