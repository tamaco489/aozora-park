package firestore

import (
	"context"
	"fmt"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

// parkDocument は枠の生成に使うパークの設定を読む形
//
// park が保存したドキュメントを読む、機能パッケージ同士は import しないため形だけを inventory 側にも置く
// 表示名のように生成に使わない項目は持たない、DataTo は構造体にない項目を捨てる
type parkDocument struct {
	DefaultDailyCapacity int32 `firestore:"defaultDailyCapacity"`
	InventoryDays        int32 `firestore:"inventoryDays"`
}

// attractionDocument は枠の生成に使うアトラクションの設定を読む形
type attractionDocument struct {
	PriorityPassConfig priorityPassConfigDocument `firestore:"priorityPassConfig"`
}

// priorityPassConfigDocument は入れ子のマップとして保存されている優先パスの条件
type priorityPassConfigDocument struct {
	Enabled         bool   `firestore:"enabled"`
	StartTime       string `firestore:"startTime"`
	EndTime         string `firestore:"endTime"`
	IntervalMinutes int32  `firestore:"intervalMinutes"`
	CapacityPerSlot int32  `firestore:"capacityPerSlot"`
}

func (r *Repository) ListParks(ctx context.Context) ([]*inventorymodel.ParkMaster, error) {
	snapshots, err := r.client.Collection(parkCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("list parks: %w", err)
	}

	parks := make(
		[]*inventorymodel.ParkMaster,
		0,
		len(snapshots),
	)
	for _, snapshot := range snapshots {
		var doc parkDocument
		if err := snapshot.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("decode park %q: %w",
				snapshot.Ref.ID,
				err,
			)
		}

		park, err := inventorymodel.RestoreParkMaster(
			inventorymodel.ParkID(snapshot.Ref.ID),
			doc.DefaultDailyCapacity,
			doc.InventoryDays,
		)
		if err != nil {
			return nil, fmt.Errorf("restore park %q: %w",
				snapshot.Ref.ID,
				err,
			)
		}

		parks = append(parks, park)
	}

	return parks, nil
}

func (r *Repository) ListAttractions(
	ctx context.Context,
	parkID inventorymodel.ParkID,
) ([]*inventorymodel.AttractionMaster, error) {
	collection := r.client.Collection(parkCollection).
		Doc(parkID.String()).
		Collection(attractionCollection)

	snapshots, err := collection.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("list attractions of park %q: %w",
			parkID,
			err,
		)
	}

	attractions := make(
		[]*inventorymodel.AttractionMaster,
		0,
		len(snapshots),
	)
	for _, snapshot := range snapshots {
		var doc attractionDocument
		if err := snapshot.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("decode attraction %q: %w",
				snapshot.Ref.ID,
				err,
			)
		}

		attraction, err := inventorymodel.RestoreAttractionMaster(
			parkID,
			inventorymodel.AttractionID(snapshot.Ref.ID),
			doc.PriorityPassConfig.Enabled,
			doc.PriorityPassConfig.StartTime,
			doc.PriorityPassConfig.EndTime,
			doc.PriorityPassConfig.IntervalMinutes,
			doc.PriorityPassConfig.CapacityPerSlot,
		)
		if err != nil {
			return nil, fmt.Errorf("restore attraction %q: %w",
				snapshot.Ref.ID,
				err,
			)
		}

		attractions = append(attractions, attraction)
	}

	return attractions, nil
}
