package model

import (
	"time"
	"uuid"
)

// PriorityPass はアトラクションの時間帯枠に対する優先パスの申込
type PriorityPass struct {
	id           PassID
	parkID       ParkID
	ticketID     TicketID
	attractionID AttractionID
	timeSlotID   TimeSlotID
	status       Status
	createdAt    time.Time
	updatedAt    time.Time
}

// NewPriorityPass は申込を受け付けた優先パスを新しく生成する
//
// 識別子は UUID v7 で採番する、生成順に並ぶので一覧の並びが安定する
// 枠を確保するのは割当の処理のため、ここでは requested で作る
func NewPriorityPass(
	parkID ParkID,
	ticketID TicketID,
	attractionID AttractionID,
	timeSlotID TimeSlotID,
	now time.Time,
) (*PriorityPass, error) {
	return newPriorityPass(
		PassID(uuid.NewV7().String()),
		parkID,
		ticketID,
		attractionID,
		timeSlotID,
		StatusRequested,
		now,
		now,
	)
}

// RestorePriorityPass は保存済みの優先パスを組み立てる
//
// infrastructure が読み出した値を入れる、保存されている値も検証を通す
func RestorePriorityPass(
	id PassID,
	parkID ParkID,
	ticketID TicketID,
	attractionID AttractionID,
	timeSlotID TimeSlotID,
	status Status,
	createdAt time.Time,
	updatedAt time.Time,
) (*PriorityPass, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	return newPriorityPass(
		id,
		parkID,
		ticketID,
		attractionID,
		timeSlotID,
		status,
		createdAt,
		updatedAt,
	)
}

func newPriorityPass(
	id PassID,
	parkID ParkID,
	ticketID TicketID,
	attractionID AttractionID,
	timeSlotID TimeSlotID,
	status Status,
	createdAt time.Time,
	updatedAt time.Time,
) (*PriorityPass, error) {
	if err := validate(
		parkID,
		ticketID,
		attractionID,
		timeSlotID,
		status,
	); err != nil {
		return nil, err
	}

	return &PriorityPass{
		id:           id,
		parkID:       parkID,
		ticketID:     ticketID,
		attractionID: attractionID,
		timeSlotID:   timeSlotID,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

func validate(
	parkID ParkID,
	ticketID TicketID,
	attractionID AttractionID,
	timeSlotID TimeSlotID,
	status Status,
) error {
	if parkID == "" {
		return ErrInvalidParkID
	}

	if ticketID == "" {
		return ErrInvalidTicketID
	}

	if attractionID == "" {
		return ErrInvalidAttractionID
	}

	if err := validateTimeSlotID(timeSlotID); err != nil {
		return err
	}

	if !status.IsValid() {
		return ErrInvalidStatus
	}

	return nil
}

// Issue は枠を確保できた申込を発行済みにする
func (p *PriorityPass) Issue(now time.Time) error {
	return p.transit(StatusIssued, now)
}

// MarkSoldOut は枠が残っていなかった申込を売り切れにする
func (p *PriorityPass) MarkSoldOut(now time.Time) error {
	return p.transit(StatusSoldOut, now)
}

// transit は申込中からの遷移だけを許す
//
// 割当のメッセージは再配信されるため、終端まで進んだものが申込中に戻らないようにする
func (p *PriorityPass) transit(status Status, now time.Time) error {
	if !p.IsRequested() {
		return ErrNotRequested
	}

	p.status = status
	p.updatedAt = now

	return nil
}

// IsRequested は申込が割当を待っている状態かを返す
func (p *PriorityPass) IsRequested() bool { return p.status == StatusRequested }

func (p *PriorityPass) ID() PassID { return p.id }

func (p *PriorityPass) ParkID() ParkID { return p.parkID }

func (p *PriorityPass) TicketID() TicketID { return p.ticketID }

func (p *PriorityPass) AttractionID() AttractionID { return p.attractionID }

func (p *PriorityPass) TimeSlotID() TimeSlotID { return p.timeSlotID }

func (p *PriorityPass) Status() Status { return p.status }

func (p *PriorityPass) CreatedAt() time.Time { return p.createdAt }

func (p *PriorityPass) UpdatedAt() time.Time { return p.updatedAt }
