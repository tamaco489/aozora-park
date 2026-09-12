// Package firestore は Firestore のクライアントを組み立てる
package firestore

import (
	"context"
	"fmt"

	gcpfirestore "cloud.google.com/go/firestore"
)

// New は Firestore のクライアントを作る
//
// FIRESTORE_EMULATOR_HOST が設定されていれば SDK がエミュレータへ繋ぐ
// 呼び出し側は終了処理を App に登録する
func New(ctx context.Context, projectID string) (*gcpfirestore.Client, error) {
	client, err := gcpfirestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("new firestore client: %w", err)
	}
	return client, nil
}
