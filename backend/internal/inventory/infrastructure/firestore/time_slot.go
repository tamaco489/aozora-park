package firestore

import (
	"context"
	"fmt"

	gcpfirestore "cloud.google.com/go/firestore"

	inventorymodel "github.com/tamaco489/aozora-park/backend/internal/inventory/domain/model"
)

// timeSlotDocument は時間帯枠を保存する形
//
// 識別子はドキュメント ID が持つ、アトラクションの識別子はコレクショングループで絞るためにフィールドにも持つ
type timeSlotDocument struct {
	AttractionID string `firestore:"attractionId"`
	Date         string `firestore:"date"`
	StartTime    string `firestore:"startTime"`
	Capacity     int32  `firestore:"capacity"`
	Remaining    int32  `firestore:"remaining"`
}

func (r *Repository) ListTimeSlots(ctx context.Context, parkID inventorymodel.ParkID, attractionID inventorymodel.AttractionID, date inventorymodel.Date) ([]*inventorymodel.TimeSlot, error) {
	// ドキュメント ID の範囲ではなく date の等価で絞る、問い合わせを ID の組み立て規則に依存させないため
	query := r.timeSlots(parkID, attractionID).
		Where("date", "==", date.String()).
		OrderBy("startTime", gcpfirestore.Asc)

	snapshots, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("list time slots %q %q %q: %w", parkID, attractionID, date, err)
	}

	slots := make([]*inventorymodel.TimeSlot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		var doc timeSlotDocument
		if err := snapshot.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("decode time slot %q: %w", snapshot.Ref.ID, err)
		}

		slot, err := inventorymodel.RestoreTimeSlot(
			parkID,
			inventorymodel.AttractionID(doc.AttractionID),
			inventorymodel.TimeSlotID(snapshot.Ref.ID),
			inventorymodel.Date(doc.Date),
			doc.StartTime,
			doc.Capacity,
			doc.Remaining,
		)
		if err != nil {
			return nil, fmt.Errorf("restore time slot %q: %w", snapshot.Ref.ID, err)
		}

		slots = append(slots, slot)
	}

	return slots, nil
}
