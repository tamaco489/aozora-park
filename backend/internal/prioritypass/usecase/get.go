package usecase

import (
	"context"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
)

// GetPriorityPassInput は GetPriorityPass の入力
type GetPriorityPassInput struct {
	PassID prioritypassmodel.PassID
}

// GetPriorityPass は優先パスを 1 件取得する
type GetPriorityPass struct {
	passes prioritypassrepository.Reader
}

func NewGetPriorityPass(passes prioritypassrepository.Reader) *GetPriorityPass {
	return &GetPriorityPass{passes: passes}
}

func (u *GetPriorityPass) Do(ctx context.Context, in GetPriorityPassInput) (*prioritypassmodel.PriorityPass, error) {
	return u.passes.GetPriorityPass(ctx, in.PassID)
}
