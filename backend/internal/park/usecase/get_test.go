package usecase

import (
	"context"
	"errors"
	"testing"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

func TestGetDo(t *testing.T) {
	repo := newFakeRepository()
	store(t, repo)

	tests := map[string]struct {
		id      parkmodel.ParkID
		wantErr error
	}{
		"正常系_保存済みの場合_取得できること": {
			id: "park-1",
		},
		"異常系_保存されていない場合_ErrNotFoundになること": {
			id:      "park-2",
			wantErr: parkmodel.ErrNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewGet(repo).Do(context.Background(), tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Get.Do(%q) のエラー = %v, want %v", tt.id, err, tt.wantErr)
			}

			if tt.wantErr == nil && got.ID() != tt.id {
				t.Errorf("Get.Do(%q) の ID = %q, want %q", tt.id, got.ID(), tt.id)
			}
		})
	}
}
