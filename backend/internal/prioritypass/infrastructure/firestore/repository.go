// Package firestore は優先パスの永続化を Firestore で実装する
package firestore

import (
	gcpfirestore "cloud.google.com/go/firestore"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassrepository "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/repository"
)

// 優先パスを置くコレクションと、状態の変化を残す子コレクション
//
// パークのサブコレクションにしないのは、識別子だけで 1 件を引く参照がこの機能の中心のため
const (
	collection      = "priorityPasses"
	eventCollection = "events"
)

// Repository は Reader と Writer の両方を満たす
//
// 分けるのは受け取る側の都合なので、実装は 1 つで足りる
type Repository struct {
	client *gcpfirestore.Client
}

var (
	_ prioritypassrepository.Reader = (*Repository)(nil)
	_ prioritypassrepository.Writer = (*Repository)(nil)
)

func NewRepository(client *gcpfirestore.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) doc(id prioritypassmodel.PassID) *gcpfirestore.DocumentRef {
	return r.client.Collection(collection).Doc(id.String())
}
