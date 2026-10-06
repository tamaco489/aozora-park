package firestore

import (
	"context"
	"errors"
	"fmt"

	gcpfirestore "cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
	parkrepository "github.com/tamaco489/aozora-park/backend/internal/park/domain/repository"
)

// attractionCollection はアトラクションを置く parks の子コレクション
const attractionCollection = "attractions"

// attractionDocument は Firestore に保存する形
//
// 識別子はドキュメント ID と親のパスが持つため、フィールドには持たない
type attractionDocument struct {
	Name               string                     `firestore:"name"`
	PriorityPassConfig priorityPassConfigDocument `firestore:"priorityPassConfig"`
}

// priorityPassConfigDocument はネストしたマップとして保存する優先パスの条件
type priorityPassConfigDocument struct {
	Enabled         bool   `firestore:"enabled"`
	StartTime       string `firestore:"startTime"`
	EndTime         string `firestore:"endTime"`
	IntervalMinutes int32  `firestore:"intervalMinutes"`
	CapacityPerSlot int32  `firestore:"capacityPerSlot"`
}

var (
	_ parkrepository.AttractionReader = (*Repository)(nil)
	_ parkrepository.AttractionWriter = (*Repository)(nil)
)

func (r *Repository) GetAttraction(
	ctx context.Context,
	parkID parkmodel.ParkID,
	id parkmodel.AttractionID,
) (*parkmodel.Attraction, error) {
	snapshot, err := r.attractionDoc(parkID, id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, parkmodel.ErrAttractionNotFound
		}
		return nil, fmt.Errorf("get attraction %q of park %q: %w",
			id,
			parkID,
			err,
		)
	}

	return toAttraction(parkID, id, snapshot)
}

// FindAttractionByName は同じパークの中を表示名で引く
//
// 表示名の重複を usecase が判断するために使う、単一フィールドの等価条件のため索引は自動で作られる
func (r *Repository) FindAttractionByName(
	ctx context.Context,
	parkID parkmodel.ParkID,
	name string,
) (*parkmodel.Attraction, error) {
	iter := r.attractionsRef(parkID).Where("name", "==", name).Limit(1).Documents(ctx)
	defer iter.Stop()

	snapshot, err := iter.Next()
	if errors.Is(err, iterator.Done) {
		return nil, parkmodel.ErrAttractionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find attraction by name %q of park %q: %w",
			name,
			parkID,
			err,
		)
	}

	return toAttraction(
		parkID,
		parkmodel.AttractionID(snapshot.Ref.ID),
		snapshot,
	)
}

// toAttraction は保存されている 1 件をドメインの型に組み立て直す
func toAttraction(
	parkID parkmodel.ParkID,
	id parkmodel.AttractionID,
	snapshot *gcpfirestore.DocumentSnapshot,
) (*parkmodel.Attraction, error) {
	var doc attractionDocument
	if err := snapshot.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("decode attraction %q of park %q: %w",
			id,
			parkID,
			err,
		)
	}

	config, err := parkmodel.NewPriorityPassConfig(
		doc.PriorityPassConfig.Enabled,
		doc.PriorityPassConfig.StartTime,
		doc.PriorityPassConfig.EndTime,
		doc.PriorityPassConfig.IntervalMinutes,
		doc.PriorityPassConfig.CapacityPerSlot,
	)
	if err != nil {
		return nil, fmt.Errorf("restore priority pass config of attraction %q: %w",
			id,
			err,
		)
	}

	attraction, err := parkmodel.RestoreAttraction(
		parkID,
		id,
		doc.Name,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("restore attraction %q of park %q: %w",
			id,
			parkID,
			err,
		)
	}

	return attraction, nil
}

func (r *Repository) CreateAttraction(ctx context.Context, attraction *parkmodel.Attraction) error {
	doc := r.attractionDoc(attraction.ParkID(), attraction.ID())

	// Create は既にあると失敗するため、採番が衝突した場合も上書きにならない
	if _, err := doc.Create(ctx, toAttractionDocument(attraction)); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return parkmodel.ErrAttractionAlreadyExists
		}
		return fmt.Errorf("create attraction %q of park %q: %w",
			attraction.ID(),
			attraction.ParkID(),
			err,
		)
	}

	return nil
}

func (r *Repository) UpdateAttraction(ctx context.Context, attraction *parkmodel.Attraction) error {
	// Set は存在しなくても作成してしまうため、削除された相手への更新を弾ける Update を使う
	// priorityPassConfig はマップごと差し替える、条件の組は常にまとめて更新するため
	updates := []gcpfirestore.Update{
		{Path: "name", Value: attraction.Name()},
		{Path: "priorityPassConfig", Value: toPriorityPassConfigDocument(attraction.PriorityPassConfig())},
	}

	doc := r.attractionDoc(attraction.ParkID(), attraction.ID())
	if _, err := doc.Update(ctx, updates); err != nil {
		if status.Code(err) == codes.NotFound {
			return parkmodel.ErrAttractionNotFound
		}
		return fmt.Errorf("update attraction %q of park %q: %w",
			attraction.ID(),
			attraction.ParkID(),
			err,
		)
	}

	return nil
}

func (r *Repository) attractionDoc(parkID parkmodel.ParkID, id parkmodel.AttractionID) *gcpfirestore.DocumentRef {
	return r.attractionsRef(parkID).Doc(id.String())
}

func (r *Repository) attractionsRef(parkID parkmodel.ParkID) *gcpfirestore.CollectionRef {
	return r.client.Collection(collection).Doc(parkID.String()).Collection(attractionCollection)
}

func toAttractionDocument(attraction *parkmodel.Attraction) attractionDocument {
	return attractionDocument{
		Name:               attraction.Name(),
		PriorityPassConfig: toPriorityPassConfigDocument(attraction.PriorityPassConfig()),
	}
}

func toPriorityPassConfigDocument(config parkmodel.PriorityPassConfig) priorityPassConfigDocument {
	return priorityPassConfigDocument{
		Enabled:         config.Enabled(),
		StartTime:       config.StartTime(),
		EndTime:         config.EndTime(),
		IntervalMinutes: config.IntervalMinutes(),
		CapacityPerSlot: config.CapacityPerSlot(),
	}
}
