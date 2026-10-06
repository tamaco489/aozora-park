package usecase

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
	prioritypassport "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase/port"
)

// 申込に使う値、テストの入力と区別できるよう 1 か所に置く
const (
	storedPassID       = prioritypassmodel.PassID("pass-1")
	storedParkID       = prioritypassmodel.ParkID("park-1")
	storedTicketID     = prioritypassmodel.TicketID("ticket-1")
	storedAttractionID = prioritypassmodel.AttractionID("attraction-1")
	storedTimeSlotID   = prioritypassmodel.TimeSlotID("20261005_1000")
)

// fixedNow は WithClock で差し替える時刻、createdAt の値そのものを検証するため固定する
var fixedNow = time.Date(
	2026, 10, 5,
	9, 0, 0, 0,
	time.UTC,
)

// fakeRepository は Reader と Writer を満たすインメモリの保存先
type fakeRepository struct {
	passes      map[prioritypassmodel.PassID]*prioritypassmodel.PriorityPass
	publishedAt map[prioritypassmodel.PassID]time.Time
	createErr   error
	markErr     error
}

var (
	_ prioritypassrepository.Reader = (*fakeRepository)(nil)
	_ prioritypassrepository.Writer = (*fakeRepository)(nil)
)

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		passes:      map[prioritypassmodel.PassID]*prioritypassmodel.PriorityPass{},
		publishedAt: map[prioritypassmodel.PassID]time.Time{},
	}
}

func (r *fakeRepository) GetPriorityPass(
	_ context.Context,
	id prioritypassmodel.PassID,
) (*prioritypassmodel.PriorityPass, error) {
	pass, ok := r.passes[id]
	if !ok {
		return nil, prioritypassmodel.ErrPriorityPassNotFound
	}

	// infrastructure は読み出すたびにドキュメントから組み立て直すため、フェイクも保存済みの実体を渡さない
	return prioritypassmodel.RestorePriorityPass(
		pass.ID(),
		pass.ParkID(),
		pass.TicketID(),
		pass.AttractionID(),
		pass.TimeSlotID(),
		pass.Status(),
		pass.CreatedAt(),
		pass.UpdatedAt(),
	)
}

func (r *fakeRepository) CreatePriorityPass(_ context.Context, pass *prioritypassmodel.PriorityPass) error {
	if r.createErr != nil {
		return r.createErr
	}

	if _, ok := r.passes[pass.ID()]; ok {
		return prioritypassmodel.ErrPriorityPassAlreadyExists
	}
	r.passes[pass.ID()] = pass

	return nil
}

func (r *fakeRepository) MarkPriorityPassPublished(
	_ context.Context,
	id prioritypassmodel.PassID,
	publishedAt time.Time,
) error {
	if r.markErr != nil {
		return r.markErr
	}

	if _, ok := r.passes[id]; !ok {
		return prioritypassmodel.ErrPriorityPassNotFound
	}
	r.publishedAt[id] = publishedAt

	return nil
}

// fakePublisher は送り先を差し替えるためのインメモリ実装
type fakePublisher struct {
	published   []prioritypassmodel.PassID
	err         error
	ctxErr      error // ctxErr は呼び出された時点の ctx の取り消しの状態
	hasDeadline bool  // hasDeadline は呼び出された ctx に期限が付いていたか
}

var _ prioritypassport.Publisher = (*fakePublisher)(nil)

func (p *fakePublisher) PublishRequested(ctx context.Context, pass *prioritypassmodel.PriorityPass) error {
	p.ctxErr = ctx.Err()
	_, p.hasDeadline = ctx.Deadline()

	if p.err != nil {
		return p.err
	}
	p.published = append(p.published, pass.ID())

	return nil
}

// discardLoggerHelper は検査に使わないログの出力先を捨てる
func discardLoggerHelper() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// storePriorityPassHelper は保存済みの優先パスを 1 件用意する
func storePriorityPassHelper(tb testing.TB, repo *fakeRepository) *prioritypassmodel.PriorityPass {
	tb.Helper()

	pass, err := prioritypassmodel.RestorePriorityPass(
		storedPassID,
		storedParkID,
		storedTicketID,
		storedAttractionID,
		storedTimeSlotID,
		prioritypassmodel.StatusRequested,
		fixedNow,
		fixedNow,
	)
	if err != nil {
		tb.Fatalf("RestorePriorityPass() = %v, want nil", err)
	}
	repo.passes[storedPassID] = pass

	return pass
}
