package usecase

import (
	"errors"
	"testing"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

func TestGetPriorityPassDo(t *testing.T) {
	repo := newFakeRepository()
	storePriorityPassHelper(t, repo)

	tests := map[string]struct {
		in      GetPriorityPassInput
		wantErr error
	}{
		"正常系_保存済みの場合_取得できること": {
			in: GetPriorityPassInput{PassID: storedPassID},
		},
		"異常系_保存されていない識別子の場合_ErrPriorityPassNotFoundになること": {
			in:      GetPriorityPassInput{PassID: "pass-2"},
			wantErr: prioritypassmodel.ErrPriorityPassNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewGetPriorityPass(repo).Do(t.Context(), tt.in)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetPriorityPass.Do(%+v) のエラー = %v, want %v",
					tt.in,
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				return
			}

			if got.ID() != tt.in.PassID {
				t.Errorf("GetPriorityPass.Do(%+v) = %q, want %q",
					tt.in,
					got.ID(),
					tt.in.PassID,
				)
			}
		})
	}
}
