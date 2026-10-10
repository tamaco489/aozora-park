package handler

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

// allocatedAt は割当の時刻、申込の時刻と区別できる値にする
var allocatedAt = fixedNow.Add(time.Hour)

func TestNewSubscriberAllocates(t *testing.T) {
	repo := newFakeAllocateRepositoryHelper(t)
	res := serveEnvelopeHelper(t, repo, envelopeHelper(`{"passId": "pass-1"}`, ""))

	if res.Code != http.StatusNoContent {
		t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) のコード = %d, want %d",
			res.Code,
			http.StatusNoContent,
		)
	}

	if repo.allocated != 1 {
		t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) の割当の回数 = %d, want %d",
			repo.allocated,
			1,
		)
	}

	if got := repo.passes[storedPassID].Status(); got != prioritypassmodel.StatusIssued {
		t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) の状態 = %q, want %q",
			got,
			prioritypassmodel.StatusIssued,
		)
	}
}

// TestNewSubscriberIsIdempotent は同じメッセージが再配信されても状態が動かないことを確かめる
//
// Pub/Sub は at-least-once のため、2 回目は何も書かずに ack する必要がある
func TestNewSubscriberIsIdempotent(t *testing.T) {
	repo := newFakeAllocateRepositoryHelper(t)
	body := envelopeHelper(`{"passId": "pass-1"}`, "")

	first := serveEnvelopeHelper(t, repo, body)
	second := serveEnvelopeHelper(t, repo, body)

	for i, res := range []*httptest.ResponseRecorder{first, second} {
		if res.Code != http.StatusNoContent {
			t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) の %d 回目のコード = %d, want %d",
				i+1,
				res.Code,
				http.StatusNoContent,
			)
		}
	}

	if repo.calls != 2 {
		t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) を 2 回呼んだときのユースケースの呼び出し回数 = %d, want %d",
			repo.calls,
			2,
		)
	}

	// 2 回目は遷移できないため、割当は 1 回しか起きない
	if repo.allocated != 1 {
		t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) を 2 回呼んだときの割当の回数 = %d, want %d",
			repo.allocated,
			1,
		)
	}

	if got := repo.passes[storedPassID].UpdatedAt(); !got.Equal(allocatedAt) {
		t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) を 2 回呼んだときの updatedAt = %v, want %v",
			got,
			allocatedAt,
		)
	}
}

func TestNewSubscriberAcksWhatCannotBeRetried(t *testing.T) {
	tests := map[string]struct {
		body  string
		calls int // calls はユースケースまで届くべき回数
	}{
		"異常系_本文がJSONでない場合_割当を呼ばずackすること": {
			body:  envelopeHelper(`not json`, ""),
			calls: 0,
		},
		"異常系_passIdが空の場合_割当を呼ばずackすること": {
			body:  envelopeHelper(`{"passId": ""}`, ""),
			calls: 0,
		},
		"異常系_申込が存在しない場合_呼んだうえでackすること": {
			body:  envelopeHelper(`{"passId": "pass-404"}`, ""),
			calls: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeAllocateRepositoryHelper(t)

			res := serveEnvelopeHelper(t, repo, tt.body)
			if res.Code != http.StatusNoContent {
				t.Errorf("NewSubscriber(...).ServeHTTP(%s) のコード = %d, want %d",
					tt.body,
					res.Code,
					http.StatusNoContent,
				)
			}

			if repo.calls != tt.calls {
				t.Errorf("NewSubscriber(...).ServeHTTP(%s) のユースケースの呼び出し回数 = %d, want %d",
					tt.body,
					repo.calls,
					tt.calls,
				)
			}

			if repo.allocated != 0 {
				t.Errorf("NewSubscriber(...).ServeHTTP(%s) の遷移の回数 = %d, want %d",
					tt.body,
					repo.allocated,
					0,
				)
			}
		})
	}
}

// TestNewSubscriberRetriesWhatCanSucceedLater は枠がまだ無いときに再配信させることを確かめる
//
// 枠の生成が追いついていないだけの場合があり、ack すると申込が requested のまま残る
func TestNewSubscriberRetriesWhatCanSucceedLater(t *testing.T) {
	repo := newFakeAllocateRepositoryHelper(t)
	repo.allocateErr = prioritypassmodel.ErrTimeSlotNotFound

	res := serveEnvelopeHelper(t, repo, envelopeHelper(`{"passId": "pass-1"}`, ""))
	if res.Code != http.StatusInternalServerError {
		t.Errorf("NewSubscriber(...).ServeHTTP(pass-1) のコード = %d, want %d",
			res.Code,
			http.StatusInternalServerError,
		)
	}
}

// TestNewSubscriberAcceptsDeliveryAttempt は DLQ を経由したエンベロープでも扱えることを確かめる
func TestNewSubscriberAcceptsDeliveryAttempt(t *testing.T) {
	repo := newFakeAllocateRepositoryHelper(t)

	res := serveEnvelopeHelper(t, repo, envelopeHelper(`{"passId": "pass-1"}`, `"deliveryAttempt": 3,`))
	if res.Code != http.StatusNoContent {
		t.Errorf("NewSubscriber(...).ServeHTTP(deliveryAttempt=3) のコード = %d, want %d",
			res.Code,
			http.StatusNoContent,
		)
	}

	if repo.allocated != 1 {
		t.Errorf("NewSubscriber(...).ServeHTTP(deliveryAttempt=3) の割当の回数 = %d, want %d",
			repo.allocated,
			1,
		)
	}
}

// envelopeHelper は push が送る形の JSON を組み立てる
//
// data は本文をそのまま埋め込める形にし、ケースごとに壊れた本文も渡せるようにする
func envelopeHelper(data, deliveryAttempt string) string {
	return `{"message": {"messageId": "1518", "data": "` + base64Helper(data) + `"}, ` +
		deliveryAttempt + `"subscription": "projects/demo/subscriptions/allocate"}`
}

// serveEnvelopeHelper は受け口に 1 回リクエストを渡す
func serveEnvelopeHelper(
	tb testing.TB,
	repo *fakeAllocateRepository,
	body string,
) *httptest.ResponseRecorder {
	tb.Helper()

	handler := NewSubscriber(
		discardLoggerHelper(),
		prioritypassusecase.NewAllocateTimeSlot(
			repo,
			discardLoggerHelper(),
			prioritypassusecase.WithAllocateClock(func() time.Time { return allocatedAt }),
		),
	)

	req := httptest.NewRequest(http.MethodPost,
		"/pubsub/push",
		strings.NewReader(body),
	)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	return res
}

// fakeAllocateRepository は割当の回数を数える保存先
//
// connect の入口が使う fakeRepository は遷移を行わないため、受け口の検証には別に置く
type fakeAllocateRepository struct {
	passes      map[prioritypassmodel.PassID]*prioritypassmodel.PriorityPass
	calls       int // calls は AllocateTimeSlot が呼ばれた回数、本文を弾けているかを見る
	allocated   int // allocated は遷移まで進んだ回数、冪等性を見る
	allocateErr error
}

var _ prioritypassrepository.Writer = (*fakeAllocateRepository)(nil)

func (r *fakeAllocateRepository) CreatePriorityPass(_ context.Context, pass *prioritypassmodel.PriorityPass) error {
	r.passes[pass.ID()] = pass

	return nil
}

func (r *fakeAllocateRepository) MarkPriorityPassPublished(
	_ context.Context,
	_ prioritypassmodel.PassID,
	_ time.Time,
) error {
	return nil
}

func (r *fakeAllocateRepository) AllocateTimeSlot(
	_ context.Context,
	id prioritypassmodel.PassID,
	now time.Time,
) (prioritypassmodel.Allocation, error) {
	r.calls++

	if r.allocateErr != nil {
		return prioritypassmodel.Allocation{}, r.allocateErr
	}

	pass, ok := r.passes[id]
	if !ok {
		return prioritypassmodel.Allocation{}, prioritypassmodel.ErrPriorityPassNotFound
	}

	// 遷移できないものは何も書かずに返す、infrastructure の冪等の判定と同じ形にする
	if !pass.IsRequested() {
		return prioritypassmodel.Allocation{Pass: pass}, nil
	}

	if err := pass.Issue(now); err != nil {
		return prioritypassmodel.Allocation{}, err
	}
	r.allocated++

	return prioritypassmodel.Allocation{
		Pass:    pass,
		Changed: true,
	}, nil
}

// newFakeAllocateRepositoryHelper は申込中の優先パスを 1 件だけ持つ保存先を返す
func newFakeAllocateRepositoryHelper(tb testing.TB) *fakeAllocateRepository {
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
		tb.Fatalf("RestorePriorityPass() = %v, want %v",
			err,
			nil,
		)
	}

	return &fakeAllocateRepository{
		passes: map[prioritypassmodel.PassID]*prioritypassmodel.PriorityPass{storedPassID: pass},
	}
}

// base64Helper はエンベロープの data に埋める形に変換する
func base64Helper(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// discardLoggerHelper は出力を捨てるロガーを返す
func discardLoggerHelper() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
