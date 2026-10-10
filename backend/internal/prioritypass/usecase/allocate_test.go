package usecase

import (
	"errors"
	"testing"
	"time"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// allocatedAt は WithAllocateClock で差し替える時刻、申込の時刻と区別できる値にする
var allocatedAt = fixedNow.Add(time.Hour)

func TestAllocateTimeSlotDo(t *testing.T) {
	tests := map[string]struct {
		remaining     int32
		wantStatus    prioritypassmodel.Status
		wantRemaining int32
	}{
		"正常系_残りがある場合_issuedになり残りが1減ること": {
			remaining:     2,
			wantStatus:    prioritypassmodel.StatusIssued,
			wantRemaining: 1,
		},
		"境界値_残りが1の場合_issuedになり残りが0になること": {
			remaining:     1,
			wantStatus:    prioritypassmodel.StatusIssued,
			wantRemaining: 0,
		},
		"境界値_残りが0の場合_sold_outになり残りが変わらないこと": {
			remaining:     0,
			wantStatus:    prioritypassmodel.StatusSoldOut,
			wantRemaining: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository()
			storePriorityPassHelper(t, repo)
			storeTimeSlotHelper(t, repo, tt.remaining)

			got, err := newAllocateTimeSlotHelper(t, repo).Do(
				t.Context(),
				AllocateTimeSlotInput{PassID: storedPassID},
			)
			if err != nil {
				t.Fatalf("AllocateTimeSlot.Do(%q) = %v, want nil",
					storedPassID,
					err,
				)
			}

			if got.Status() != tt.wantStatus {
				t.Errorf("AllocateTimeSlot.Do(%q) の状態 = %q, want %q",
					storedPassID,
					got.Status(),
					tt.wantStatus,
				)
			}

			if !got.UpdatedAt().Equal(allocatedAt) {
				t.Errorf("AllocateTimeSlot.Do(%q) の更新の時刻 = %v, want %v",
					storedPassID,
					got.UpdatedAt(),
					allocatedAt,
				)
			}

			if repo.remaining[storedTimeSlotID] != tt.wantRemaining {
				t.Errorf("AllocateTimeSlot.Do(%q) の後の残り = %d, want %d",
					storedPassID,
					repo.remaining[storedTimeSlotID],
					tt.wantRemaining,
				)
			}

			if repo.events[storedPassID] != 1 {
				t.Errorf("AllocateTimeSlot.Do(%q) の後のイベントの件数 = %d, want %d",
					storedPassID,
					repo.events[storedPassID],
					1,
				)
			}
		})
	}
}

// TestAllocateTimeSlotDoTwice は再配信で 2 回処理しても残りが 1 しか減らないことを確かめる
func TestAllocateTimeSlotDoTwice(t *testing.T) {
	repo := newFakeRepository()
	storePriorityPassHelper(t, repo)
	storeTimeSlotHelper(t, repo, 2)

	usecase := newAllocateTimeSlotHelper(t, repo)
	in := AllocateTimeSlotInput{PassID: storedPassID}

	if _, err := usecase.Do(t.Context(), in); err != nil {
		t.Fatalf("1 回目の AllocateTimeSlot.Do(%q) = %v, want nil",
			storedPassID,
			err,
		)
	}

	got, err := usecase.Do(t.Context(), in)
	if err != nil {
		t.Fatalf("2 回目の AllocateTimeSlot.Do(%q) = %v, want nil",
			storedPassID,
			err,
		)
	}

	if got.Status() != prioritypassmodel.StatusIssued {
		t.Errorf("2 回目の AllocateTimeSlot.Do(%q) の状態 = %q, want %q",
			storedPassID,
			got.Status(),
			prioritypassmodel.StatusIssued,
		)
	}

	if repo.remaining[storedTimeSlotID] != 1 {
		t.Errorf("2 回目の AllocateTimeSlot.Do(%q) の後の残り = %d, want %d",
			storedPassID,
			repo.remaining[storedTimeSlotID],
			1,
		)
	}

	if repo.events[storedPassID] != 1 {
		t.Errorf("2 回目の AllocateTimeSlot.Do(%q) の後のイベントの件数 = %d, want %d",
			storedPassID,
			repo.events[storedPassID],
			1,
		)
	}
}

func TestAllocateTimeSlotDoNotFound(t *testing.T) {
	repo := newFakeRepository()
	storeTimeSlotHelper(t, repo, 1)

	_, err := newAllocateTimeSlotHelper(t, repo).Do(
		t.Context(),
		AllocateTimeSlotInput{PassID: storedPassID},
	)
	if !errors.Is(err, prioritypassmodel.ErrPriorityPassNotFound) {
		t.Errorf("AllocateTimeSlot.Do(%q) = %v, want %v",
			storedPassID,
			err,
			prioritypassmodel.ErrPriorityPassNotFound,
		)
	}
}

// TestAllocateTimeSlotDoTimeSlotNotFound は枠が無い場合に売り切れへ畳まないことを確かめる
func TestAllocateTimeSlotDoTimeSlotNotFound(t *testing.T) {
	repo := newFakeRepository()
	storePriorityPassHelper(t, repo)

	_, err := newAllocateTimeSlotHelper(t, repo).Do(
		t.Context(),
		AllocateTimeSlotInput{PassID: storedPassID},
	)
	if !errors.Is(err, prioritypassmodel.ErrTimeSlotNotFound) {
		t.Fatalf("AllocateTimeSlot.Do(%q) = %v, want %v",
			storedPassID,
			err,
			prioritypassmodel.ErrTimeSlotNotFound,
		)
	}

	if repo.events[storedPassID] != 0 {
		t.Errorf("AllocateTimeSlot.Do(%q) の後のイベントの件数 = %d, want %d",
			storedPassID,
			repo.events[storedPassID],
			0,
		)
	}
}

// newAllocateTimeSlotHelper は時刻を固定した割当のユースケースを組み立てる
func newAllocateTimeSlotHelper(tb testing.TB, repo *fakeRepository) *AllocateTimeSlot {
	tb.Helper()

	return NewAllocateTimeSlot(
		repo,
		discardLoggerHelper(),
		WithAllocateClock(func() time.Time { return allocatedAt }),
	)
}

// storeTimeSlotHelper は申込が指す時間帯枠の残りを用意する
func storeTimeSlotHelper(
	tb testing.TB,
	repo *fakeRepository,
	remaining int32,
) {
	tb.Helper()

	repo.remaining[storedTimeSlotID] = remaining
}
