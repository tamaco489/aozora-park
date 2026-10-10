// Package firestore は枠在庫の永続化を Firestore で実装する
package firestore

import (
	gcpfirestore "cloud.google.com/go/firestore"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
	inventoryrepository "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/repository"
)

// 枠を置くコレクション、どちらもパークのサブコレクションに入れ子にする
//
//   - timeSlots のパスは prioritypass も優先パスの割当で残りを減らすために持つ
//   - internal/prioritypass/infrastructure/firestore/repository.go と 2 か所にある、片方を変更したらもう片方も修正する
const (
	parkCollection          = "parks"
	dateInventoryCollection = "dateInventories"
	attractionCollection    = "attractions"
	timeSlotCollection      = "timeSlots"
)

// Repository は Reader と Writer の両方を満たす
//
// 分けるのは受け取る側の都合なので、実装は 1 つで足りる
type Repository struct {
	client *gcpfirestore.Client
}

var (
	_ inventoryrepository.Reader       = (*Repository)(nil)
	_ inventoryrepository.Writer       = (*Repository)(nil)
	_ inventoryrepository.MasterReader = (*Repository)(nil)
	_ inventoryrepository.Creator      = (*Repository)(nil)
)

func NewRepository(client *gcpfirestore.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) dateInventories(parkID inventorymodel.ParkID) *gcpfirestore.CollectionRef {
	return r.client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(dateInventoryCollection)
}

func (r *Repository) timeSlots(parkID inventorymodel.ParkID, attractionID inventorymodel.AttractionID) *gcpfirestore.CollectionRef {
	return r.client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(attractionCollection).
		Doc(attractionID.String()).
		Collection(timeSlotCollection)
}
