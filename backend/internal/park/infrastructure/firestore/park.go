// Package firestore はパークの永続化を Firestore で実装する
package firestore

import (
	"context"
	"fmt"

	gcpfirestore "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// collection はパークを置くコレクション
const collection = "parks"

// document は Firestore に保存する形
//
// 識別子はドキュメント ID が持つため、フィールドには持たない
type document struct {
	Name                 string `firestore:"name"`
	DefaultDailyCapacity int32  `firestore:"defaultDailyCapacity"`
	InventoryDays        int32  `firestore:"inventoryDays"`
}

// Repository は Reader と Writer の両方を満たす
//
// 分けるのは受け取る側の都合なので、実装は 1 つで足りる
type Repository struct {
	client *gcpfirestore.Client
}

var (
	_ parkrepository.Reader = (*Repository)(nil)
	_ parkrepository.Writer = (*Repository)(nil)
)

func NewRepository(client *gcpfirestore.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) Get(ctx context.Context, id parkmodel.ParkID) (*parkmodel.Park, error) {
	snapshot, err := r.client.Collection(collection).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, parkmodel.ErrNotFound
		}
		return nil, fmt.Errorf("get park %q: %w", id, err)
	}

	var doc document
	if err := snapshot.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("decode park %q: %w", id, err)
	}

	park, err := parkmodel.Restore(id, doc.Name, doc.DefaultDailyCapacity, doc.InventoryDays)
	if err != nil {
		return nil, fmt.Errorf("restore park %q: %w", id, err)
	}

	return park, nil
}

func (r *Repository) Create(ctx context.Context, park *parkmodel.Park) error {
	// Create は既にあると失敗するため、採番が衝突した場合も上書きにならない
	if _, err := r.doc(park).Create(ctx, toDocument(park)); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return parkmodel.ErrAlreadyExists
		}
		return fmt.Errorf("create park %q: %w", park.ID(), err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, park *parkmodel.Park) error {
	// Set は存在しなくても作ってしまうため、消えた相手への更新を弾ける Update を使う
	updates := []gcpfirestore.Update{
		{Path: "name", Value: park.Name()},
		{Path: "defaultDailyCapacity", Value: park.DefaultDailyCapacity()},
		{Path: "inventoryDays", Value: park.InventoryDays()},
	}

	if _, err := r.doc(park).Update(ctx, updates); err != nil {
		if status.Code(err) == codes.NotFound {
			return parkmodel.ErrNotFound
		}
		return fmt.Errorf("update park %q: %w", park.ID(), err)
	}

	return nil
}

func (r *Repository) doc(park *parkmodel.Park) *gcpfirestore.DocumentRef {
	return r.client.Collection(collection).Doc(park.ID().String())
}

func toDocument(park *parkmodel.Park) document {
	return document{
		Name:                 park.Name(),
		DefaultDailyCapacity: park.DefaultDailyCapacity(),
		InventoryDays:        park.InventoryDays(),
	}
}
