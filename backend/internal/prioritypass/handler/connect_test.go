package handler

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
	prioritypassport "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase/port"
)

// 保存済みの優先パスに使う値
const (
	storedPassID       = prioritypassmodel.PassID("pass-1")
	storedParkID       = prioritypassmodel.ParkID("park-1")
	storedTicketID     = prioritypassmodel.TicketID("ticket-1")
	storedAttractionID = prioritypassmodel.AttractionID("attraction-1")
	storedTimeSlotID   = prioritypassmodel.TimeSlotID("20261005_1000")
)

// fixedNow は WithClock で差し替える時刻
var fixedNow = time.Date(
	2026, 10, 5,
	9, 0, 0, 0,
	time.UTC,
)

// fakeRepository は保存先を差し替えるためのインメモリ実装
//
// usecase は実物を通す、差し替えるのは domain/repository だけにする
type fakeRepository struct {
	passes map[prioritypassmodel.PassID]*prioritypassmodel.PriorityPass
}

var (
	_ prioritypassrepository.Reader = (*fakeRepository)(nil)
	_ prioritypassrepository.Writer = (*fakeRepository)(nil)
)

func (r *fakeRepository) GetPriorityPass(
	_ context.Context,
	id prioritypassmodel.PassID,
) (*prioritypassmodel.PriorityPass, error) {
	pass, ok := r.passes[id]
	if !ok {
		return nil, prioritypassmodel.ErrPriorityPassNotFound
	}
	return pass, nil
}

func (r *fakeRepository) CreatePriorityPass(_ context.Context, pass *prioritypassmodel.PriorityPass) error {
	if _, ok := r.passes[pass.ID()]; ok {
		return prioritypassmodel.ErrPriorityPassAlreadyExists
	}
	r.passes[pass.ID()] = pass
	return nil
}

func (r *fakeRepository) MarkPriorityPassPublished(
	_ context.Context,
	id prioritypassmodel.PassID,
	_ time.Time,
) error {
	if _, ok := r.passes[id]; !ok {
		return prioritypassmodel.ErrPriorityPassNotFound
	}
	return nil
}

// AllocateTimeSlot は connect の入口が割当を呼ばないため、Writer を満たすためだけに置く
func (r *fakeRepository) AllocateTimeSlot(
	_ context.Context,
	id prioritypassmodel.PassID,
	_ time.Time,
) (prioritypassmodel.Allocation, error) {
	pass, ok := r.passes[id]
	if !ok {
		return prioritypassmodel.Allocation{}, prioritypassmodel.ErrPriorityPassNotFound
	}

	return prioritypassmodel.Allocation{Pass: pass}, nil
}

// fakePublisher は送り先を差し替えるためのインメモリ実装
type fakePublisher struct{}

var _ prioritypassport.Publisher = (*fakePublisher)(nil)

func (p *fakePublisher) PublishRequested(_ context.Context, _ *prioritypassmodel.PriorityPass) error {
	return nil
}

// newHandlerHelper は渡した保存先でハンドラを組み立てる
func newHandlerHelper(tb testing.TB, repo *fakeRepository) *Connect {
	tb.Helper()

	return NewConnect(
		prioritypassusecase.NewRequestPriorityPass(
			repo,
			&fakePublisher{},
			slog.New(slog.NewTextHandler(io.Discard, nil)),
			prioritypassusecase.WithClock(func() time.Time { return fixedNow }),
		),
		prioritypassusecase.NewGetPriorityPass(repo),
	)
}

func newEmptyRepositoryHelper() *fakeRepository {
	return &fakeRepository{passes: map[prioritypassmodel.PassID]*prioritypassmodel.PriorityPass{}}
}

// newStoredRepositoryHelper は優先パスを 1 件保存済みにする
func newStoredRepositoryHelper(tb testing.TB) *fakeRepository {
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

	repo := newEmptyRepositoryHelper()
	repo.passes[storedPassID] = pass

	return repo
}

func TestToStatusProto(t *testing.T) {
	tests := map[string]struct {
		status prioritypassmodel.Status
		want   string
	}{
		"正常系_requestedの場合_REQUESTEDになること": {
			status: prioritypassmodel.StatusRequested,
			want:   "PRIORITY_PASS_STATUS_REQUESTED",
		},
		"正常系_issuedの場合_ISSUEDになること": {
			status: prioritypassmodel.StatusIssued,
			want:   "PRIORITY_PASS_STATUS_ISSUED",
		},
		"正常系_sold_outの場合_SOLD_OUTになること": {
			status: prioritypassmodel.StatusSoldOut,
			want:   "PRIORITY_PASS_STATUS_SOLD_OUT",
		},
		"異常系_既知でない値の場合_UNSPECIFIEDになること": {
			status: prioritypassmodel.Status("rejected"),
			want:   "PRIORITY_PASS_STATUS_UNSPECIFIED",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := toStatusProto(tt.status).String(); got != tt.want {
				t.Errorf("toStatusProto(%q) = %q, want %q",
					tt.status,
					got,
					tt.want,
				)
			}
		})
	}
}
