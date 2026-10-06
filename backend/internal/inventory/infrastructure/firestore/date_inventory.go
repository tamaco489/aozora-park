package firestore

import (
	"context"
	"fmt"

	gcpfirestore "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

// dateInventoryDocument は入場枠を保存する形
//
// パークの識別子はパスが持つためフィールドに持たない、日付はドキュメント ID と同じ値をコレクショングループで絞るために持つ
type dateInventoryDocument struct {
	Date      string `firestore:"date"`
	Capacity  int32  `firestore:"capacity"`
	Remaining int32  `firestore:"remaining"`
}

func (r *Repository) GetDateInventory(
	ctx context.Context,
	parkID inventorymodel.ParkID,
	date inventorymodel.Date,
) (*inventorymodel.DateInventory, error) {
	snapshot, err := r.dateInventories(parkID).Doc(date.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, inventorymodel.ErrDateInventoryNotFound
		}
		return nil, fmt.Errorf("get date inventory %q %q: %w",
			parkID,
			date,
			err,
		)
	}

	var doc dateInventoryDocument
	if err := snapshot.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("decode date inventory %q %q: %w",
			parkID,
			date,
			err,
		)
	}

	inventory, err := inventorymodel.RestoreDateInventory(
		parkID,
		inventorymodel.Date(doc.Date),
		doc.Capacity,
		doc.Remaining,
	)
	if err != nil {
		return nil, fmt.Errorf("restore date inventory %q %q: %w",
			parkID,
			date,
			err,
		)
	}

	return inventory, nil
}

func (r *Repository) UpdateDateInventory(ctx context.Context, inventory *inventorymodel.DateInventory) error {
	// Set は存在しなくても作成してしまうため、ジョブが作成していない日への上書きを弾ける Update を使う
	updates := []gcpfirestore.Update{
		{Path: "capacity", Value: inventory.Capacity()},
		{Path: "remaining", Value: inventory.Remaining()},
	}

	doc := r.dateInventories(inventory.ParkID()).Doc(inventory.Date().String())
	if _, err := doc.Update(ctx, updates); err != nil {
		if status.Code(err) == codes.NotFound {
			return inventorymodel.ErrDateInventoryNotFound
		}
		return fmt.Errorf("update date inventory %q %q: %w",
			inventory.ParkID(),
			inventory.Date(),
			err,
		)
	}

	return nil
}

func (r *Repository) CreateDateInventoryIfAbsent(
	ctx context.Context,
	inventory *inventorymodel.DateInventory,
) (bool, error) {
	data := dateInventoryDocument{
		Date:      inventory.Date().String(),
		Capacity:  inventory.Capacity(),
		Remaining: inventory.Remaining(),
	}

	// Create は既にあると失敗するため、運営が上書きした上限や購入で減った残りを初期値に戻さない
	doc := r.dateInventories(inventory.ParkID()).Doc(inventory.Date().String())
	if _, err := doc.Create(ctx, data); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return false, nil
		}
		return false, fmt.Errorf("create date inventory %q %q: %w",
			inventory.ParkID(),
			inventory.Date(),
			err,
		)
	}

	return true, nil
}
