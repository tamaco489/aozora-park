package usecase

import (
	"errors"
	"testing"
	"time"

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

			got, err := newRequestPriorityPassHelper(repo).Do(t.Context(), tt.in)

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

	_, err := newRequestPriorityPassHelper(repo).Do(t.Context(), in)
	if !errors.Is(err, prioritypassmodel.ErrPriorityPassAlreadyExists) {
		t.Errorf("RequestPriorityPass.Do(%+v) = %v, want %v",
			in,
			err,
			prioritypassmodel.ErrPriorityPassAlreadyExists,
		)
	}
}

// newRequestPriorityPassHelper は時刻を固定したユースケースを組み立てる
func newRequestPriorityPassHelper(repo *fakeRepository) *RequestPriorityPass {
	return NewRequestPriorityPass(repo, WithClock(func() time.Time { return fixedNow }))
}
