// Package usecase は優先パスの業務の手順を持つ
package usecase

import (
	"context"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
)

// RequestPriorityPassInput は RequestPriorityPass の入力
type RequestPriorityPassInput struct {
	ParkID       prioritypassmodel.ParkID
	TicketID     prioritypassmodel.TicketID
	AttractionID prioritypassmodel.AttractionID
	TimeSlotID   prioritypassmodel.TimeSlotID
}

// RequestPriorityPass は来園者の申込を受け付けて優先パスを作成する
//
// 枠の確保は割当の処理が行うため、ここでは requested のまま保存する
type RequestPriorityPass struct {
	passes prioritypassrepository.Writer
	now    func() time.Time
}

// RequestPriorityPassOption は RequestPriorityPass の既定値を差し替える
type RequestPriorityPassOption func(*RequestPriorityPass)

// WithClock は作成時刻の基準になる時刻を差し替える
func WithClock(now func() time.Time) RequestPriorityPassOption {
	return func(u *RequestPriorityPass) { u.now = now }
}

func NewRequestPriorityPass(
	passes prioritypassrepository.Writer,
	opts ...RequestPriorityPassOption,
) *RequestPriorityPass {
	u := &RequestPriorityPass{
		passes: passes,
		now:    time.Now,
	}
	for _, opt := range opts {
		opt(u)
	}

	return u
}

func (u *RequestPriorityPass) Do(ctx context.Context, in RequestPriorityPassInput) (*prioritypassmodel.PriorityPass, error) {
	// 保存する時刻を 1 回の呼び出しで揃えるため、作成時刻は先に読む
	now := u.now().UTC()

	pass, err := prioritypassmodel.NewPriorityPass(
		in.ParkID,
		in.TicketID,
		in.AttractionID,
		in.TimeSlotID,
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := u.passes.CreatePriorityPass(ctx, pass); err != nil {
		return nil, err
	}

	return pass, nil
}
