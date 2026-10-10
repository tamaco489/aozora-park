// Package pubsub は Pub/Sub のクライアントを組み立てる
package pubsub

import (
	"context"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub/v2"
)

// New は Pub/Sub のクライアントを生成する
//
//   - PUBSUB_EMULATOR_HOST が設定されていれば SDK がエミュレータへ繋ぐ
//   - トピックごとの Publisher は使う側が client.Publisher で作る、トピック名は業務の判断なのでここでは持たない
//   - 呼び出し側は終了処理を App に登録する
func New(ctx context.Context, projectID string) (*gcppubsub.Client, error) {
	client, err := gcppubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("new pubsub client: %w", err)
	}
	return client, nil
}
