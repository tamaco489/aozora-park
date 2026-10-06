package usecase

import (
	"context"
	"log/slog"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
)

// AllocateTimeSlotInput は AllocateTimeSlot の入力
type AllocateTimeSlotInput struct {
	PassID prioritypassmodel.PassID
}

// AllocateTimeSlot は申込に時間帯枠を割り当てる
//
// 残りの確認と減算と遷移は 1 つのトランザクションで行う必要があるため、手順は infrastructure に閉じる
type AllocateTimeSlot struct {
	passes prioritypassrepository.Writer
	logger *slog.Logger
	now    func() time.Time
}

// AllocateTimeSlotOption は AllocateTimeSlot の既定値を差し替える
type AllocateTimeSlotOption func(*AllocateTimeSlot)

// WithAllocateClock は遷移の時刻の基準になる時刻を差し替える
func WithAllocateClock(now func() time.Time) AllocateTimeSlotOption {
	return func(u *AllocateTimeSlot) { u.now = now }
}

func NewAllocateTimeSlot(
	passes prioritypassrepository.Writer,
	logger *slog.Logger,
	opts ...AllocateTimeSlotOption,
) *AllocateTimeSlot {
	u := &AllocateTimeSlot{
		passes: passes,
		logger: logger,
		now:    time.Now,
	}
	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Do は申込の状態を進める
//
// 再配信で既に遷移済みのものが届いてもエラーにしない、入口が ack できるようにするため
func (u *AllocateTimeSlot) Do(
	ctx context.Context,
	in AllocateTimeSlotInput,
) (*prioritypassmodel.PriorityPass, error) {
	allocation, err := u.passes.AllocateTimeSlot(ctx, in.PassID, u.now().UTC())
	if err != nil {
		return nil, err
	}

	if !allocation.Changed {
		u.logger.InfoContext(ctx, "優先パスは遷移済みのため割当を行わない",
			slog.String("passId", allocation.Pass.ID().String()),
			slog.String("status", allocation.Pass.Status().String()),
		)

		return allocation.Pass, nil
	}

	u.logger.InfoContext(ctx, "優先パスの割当を行った",
		slog.String("passId", allocation.Pass.ID().String()),
		slog.String("status", allocation.Pass.Status().String()),
	)

	return allocation.Pass, nil
}
