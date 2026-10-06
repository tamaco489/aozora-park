package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

func TestRequestPriorityPassDo(t *testing.T) {
	tests := map[string]struct {
		in      RequestPriorityPassInput
		wantErr error
	}{
		"正常系_すべて妥当な場合_requestedで作成されること": {
			in: RequestPriorityPassInput{
				ParkID:       storedParkID,
				TicketID:     storedTicketID,
				AttractionID: storedAttractionID,
				TimeSlotID:   storedTimeSlotID,
			},
		},
		"異常系_パークの識別子が空の場合_ErrInvalidParkIDになること": {
			in: RequestPriorityPassInput{
				TicketID:     storedTicketID,
				AttractionID: storedAttractionID,
				TimeSlotID:   storedTimeSlotID,
			},
			wantErr: prioritypassmodel.ErrInvalidParkID,
		},
		"異常系_券の識別子が空の場合_ErrInvalidTicketIDになること": {
			in: RequestPriorityPassInput{
				ParkID:       storedParkID,
				AttractionID: storedAttractionID,
				TimeSlotID:   storedTimeSlotID,
			},
			wantErr: prioritypassmodel.ErrInvalidTicketID,
		},
		"異常系_時間帯枠の識別子が形式に合わない場合_ErrInvalidTimeSlotIDになること": {
			in: RequestPriorityPassInput{
				ParkID:       storedParkID,
				TicketID:     storedTicketID,
				AttractionID: storedAttractionID,
				TimeSlotID:   "20261005-1000",
			},
			wantErr: prioritypassmodel.ErrInvalidTimeSlotID,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			publisher := &fakePublisher{}

			got, err := newRequestPriorityPassHelper(repo, publisher).Do(t.Context(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RequestPriorityPass.Do(%+v) のエラー = %v, want %v",
					tt.in,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if len(repo.passes) != 0 {
					t.Errorf("RequestPriorityPass.Do(%+v) の保存件数 = %d, want %d",
						tt.in,
						len(repo.passes),
						0,
					)
				}
				return
			}

			if got.Status() != prioritypassmodel.StatusRequested {
				t.Errorf("RequestPriorityPass.Do(%+v) の状態 = %q, want %q",
					tt.in,
					got.Status(),
					prioritypassmodel.StatusRequested,
				)
			}

			if !got.CreatedAt().Equal(fixedNow) {
				t.Errorf("RequestPriorityPass.Do(%+v) の作成時刻 = %v, want %v",
					tt.in,
					got.CreatedAt(),
					fixedNow,
				)
			}

			if _, ok := repo.passes[got.ID()]; !ok {
				t.Errorf("RequestPriorityPass.Do(%+v) の保存済みの識別子 = 無し, want %q",
					tt.in,
					got.ID(),
				)
			}

			wantPublished := []prioritypassmodel.PassID{got.ID()}
			if diff := cmp.Diff(wantPublished, publisher.published); diff != "" {
				t.Errorf("RequestPriorityPass.Do(%+v) の publish の差分 (-want +got):\n%s",
					tt.in,
					diff,
				)
			}

			if publishedAt, ok := repo.publishedAt[got.ID()]; !ok || !publishedAt.Equal(fixedNow) {
				t.Errorf("RequestPriorityPass.Do(%+v) の publishedAt = (%v, %t), want (%v, %t)",
					tt.in,
					publishedAt,
					ok,
					fixedNow,
					true,
				)
			}
		})
	}
}

// TestRequestPriorityPassDoCreateFails は保存に失敗したときにエラーを素通しすることを確かめる
func TestRequestPriorityPassDoCreateFails(t *testing.T) {
	repo := newFakeRepository()
	repo.createErr = prioritypassmodel.ErrPriorityPassAlreadyExists

	in := RequestPriorityPassInput{
		ParkID:       storedParkID,
		TicketID:     storedTicketID,
		AttractionID: storedAttractionID,
		TimeSlotID:   storedTimeSlotID,
	}

	publisher := &fakePublisher{}

	_, err := newRequestPriorityPassHelper(repo, publisher).Do(t.Context(), in)
	if !errors.Is(err, prioritypassmodel.ErrPriorityPassAlreadyExists) {
		t.Errorf("RequestPriorityPass.Do(%+v) = %v, want %v",
			in,
			err,
			prioritypassmodel.ErrPriorityPassAlreadyExists,
		)
	}

	if len(publisher.published) != 0 {
		t.Errorf("RequestPriorityPass.Do(%+v) の publish の件数 = %d, want %d",
			in,
			len(publisher.published),
			0,
		)
	}
}

// TestRequestPriorityPassDoPublishFails は publish に失敗しても申込を巻き戻さないことを確かめる
//
// publishedAt を書かずに残したものは reconciliation が拾う
func TestRequestPriorityPassDoPublishFails(t *testing.T) {
	repo := newFakeRepository()
	publisher := &fakePublisher{err: errors.New("publish failed")}

	in := RequestPriorityPassInput{
		ParkID:       storedParkID,
		TicketID:     storedTicketID,
		AttractionID: storedAttractionID,
		TimeSlotID:   storedTimeSlotID,
	}

	got, err := newRequestPriorityPassHelper(repo, publisher).Do(t.Context(), in)
	if err != nil {
		t.Fatalf("RequestPriorityPass.Do(%+v) = %v, want %v",
			in,
			err,
			nil,
		)
	}

	if _, ok := repo.passes[got.ID()]; !ok {
		t.Errorf("RequestPriorityPass.Do(%+v) の保存済みの識別子 = 無し, want %q",
			in,
			got.ID(),
		)
	}

	if publishedAt, ok := repo.publishedAt[got.ID()]; ok {
		t.Errorf("RequestPriorityPass.Do(%+v) の publishedAt = (%v, %t), want (%v, %t)",
			in,
			publishedAt,
			ok,
			time.Time{},
			false,
		)
	}
}

// TestRequestPriorityPassDoMarkPublishedFails は publishedAt の記録に失敗しても申込を返すことを確かめる
func TestRequestPriorityPassDoMarkPublishedFails(t *testing.T) {
	repo := newFakeRepository()
	repo.markErr = errors.New("mark failed")
	publisher := &fakePublisher{}

	in := RequestPriorityPassInput{
		ParkID:       storedParkID,
		TicketID:     storedTicketID,
		AttractionID: storedAttractionID,
		TimeSlotID:   storedTimeSlotID,
	}

	got, err := newRequestPriorityPassHelper(repo, publisher).Do(t.Context(), in)
	if err != nil {
		t.Fatalf("RequestPriorityPass.Do(%+v) = %v, want %v",
			in,
			err,
			nil,
		)
	}

	wantPublished := []prioritypassmodel.PassID{got.ID()}
	if diff := cmp.Diff(wantPublished, publisher.published); diff != "" {
		t.Errorf("RequestPriorityPass.Do(%+v) の publish の差分 (-want +got):\n%s",
			in,
			diff,
		)
	}
}

// newRequestPriorityPassHelper は時刻を固定したユースケースを組み立てる
func newRequestPriorityPassHelper(
	repo *fakeRepository,
	publisher *fakePublisher,
) *RequestPriorityPass {
	return NewRequestPriorityPass(
		repo,
		publisher,
		discardLoggerHelper(),
		WithClock(func() time.Time { return fixedNow }),
	)
}

// TestRequestPriorityPassDoPublishesAfterCallerCancels は呼び出し元が切断しても publish を行うことを確かめる
//
// 申込は作成済みのため、送らずに終わると publishedAt が null のまま残り reconciliation の対象が無用に増える
func TestRequestPriorityPassDoPublishesAfterCallerCancels(t *testing.T) {
	repo := newFakeRepository()
	publisher := &fakePublisher{}

	// 呼び出し元が離れた状態を作る
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	in := RequestPriorityPassInput{
		ParkID:       storedParkID,
		TicketID:     storedTicketID,
		AttractionID: storedAttractionID,
		TimeSlotID:   storedTimeSlotID,
	}

	got, err := newRequestPriorityPassHelper(repo, publisher).Do(ctx, in)
	if err != nil {
		t.Fatalf("RequestPriorityPass.Do(%+v) = %v, want %v",
			in,
			err,
			nil,
		)
	}

	if publisher.ctxErr != nil {
		t.Errorf("RequestPriorityPass.Do(%+v) の publish 時の ctx.Err() = %v, want %v",
			in,
			publisher.ctxErr,
			nil,
		)
	}

	// 届かない宛先でリトライが続くため、待つ上限が付いていることを確かめる
	if !publisher.hasDeadline {
		t.Errorf("RequestPriorityPass.Do(%+v) の publish 時の ctx の期限 = %t, want %t",
			in,
			publisher.hasDeadline,
			true,
		)
	}

	if _, ok := repo.publishedAt[got.ID()]; !ok {
		t.Errorf("RequestPriorityPass.Do(%+v) の publishedAt = 無し, want %v",
			in,
			fixedNow,
		)
	}
}
