// Package usecase は優先パスの業務の手順を持つ
package usecase

import (
	"context"
	"log/slog"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
	prioritypassport "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase/port"
)

// publishTimeout は publish と publishedAt の記録に使える時間の上限
//
// 失敗しても publishedAt が null で残るだけなので、申込を返すのを長く待たせる価値がない
const publishTimeout = 3 * time.Second

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
	passes    prioritypassrepository.Writer
	publisher prioritypassport.Publisher
	logger    *slog.Logger
	now       func() time.Time
}

// RequestPriorityPassOption は RequestPriorityPass の既定値を差し替える
type RequestPriorityPassOption func(*RequestPriorityPass)

// WithClock は作成時刻の基準になる時刻を差し替える
func WithClock(now func() time.Time) RequestPriorityPassOption {
	return func(u *RequestPriorityPass) { u.now = now }
}

func NewRequestPriorityPass(
	passes prioritypassrepository.Writer,
	publisher prioritypassport.Publisher,
	logger *slog.Logger,
	opts ...RequestPriorityPassOption,
) *RequestPriorityPass {
	u := &RequestPriorityPass{
		passes:    passes,
		publisher: publisher,
		logger:    logger,
		now:       time.Now,
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

	u.publish(ctx, pass)

	return pass, nil
}

// publish は申込を送り、送れたことを publishedAt に残す
//
//   - 失敗しても申込は巻き戻さずエラーも返さない、作成済みのものを呼び出し側が作り直すと二重の申込になるため
//   - publishedAt が null のまま残ったものは reconciliation が検出して送り直す
func (u *RequestPriorityPass) publish(ctx context.Context, pass *prioritypassmodel.PriorityPass) {
	// 申込は作成済みのため、呼び出し元が切断しても送り切る
	// 宛先に届かない間 SDK がリトライを繰り返すため、呼び出し元を待たせる上限をこちらで決める
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), publishTimeout)
	defer cancel()

	// トランザクションの外で送る、外部への送信を含めると保持の時間が伸びて衝突しやすくなるため
	if err := u.publisher.PublishRequested(ctx, pass); err != nil {
		u.logger.ErrorContext(ctx, "優先パスの申込の publish に失敗",
			slog.String("passId", pass.ID().String()),
			slog.Any("error", err),
		)
		return
	}

	if err := u.passes.MarkPriorityPassPublished(ctx, pass.ID(), u.now().UTC()); err != nil {
		// 送信そのものは済んでいるため、再送しても購読側が冪等に捌く
		u.logger.ErrorContext(ctx, "優先パスの publishedAt の記録に失敗",
			slog.String("passId", pass.ID().String()),
			slog.Any("error", err),
		)
	}
}
