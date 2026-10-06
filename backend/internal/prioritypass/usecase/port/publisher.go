// Package port は優先パスの永続化以外の外部サービスのインタフェースを持つ
package port

import (
	"context"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// Publisher は優先パスの出来事を外へ送り出す
type Publisher interface {
	// PublishRequested は申込を受け付けたことを送り、相手が受け取るまで待つ
	PublishRequested(ctx context.Context, pass *prioritypassmodel.PriorityPass) error
}
