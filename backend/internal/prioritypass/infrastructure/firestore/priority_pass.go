package firestore

import (
	"context"
	"fmt"
	"time"

	gcpfirestore "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// document は優先パスを保存する形
//
// 識別子はドキュメント ID が持つため、フィールドには持たない
type document struct {
	ParkID       string     `firestore:"parkId"`
	TicketID     string     `firestore:"ticketId"`
	AttractionID string     `firestore:"attractionId"`
	TimeSlotID   string     `firestore:"timeSlotId"`
	Status       string     `firestore:"status"`
	PublishedAt  *time.Time `firestore:"publishedAt"`
	CreatedAt    time.Time  `firestore:"createdAt"`
	UpdatedAt    time.Time  `firestore:"updatedAt"`
}

// eventDocument は状態の変化を 1 件ずつ残す形
//
// 追跡の識別子はまだ持たない、ログとの突き合わせを始めるときに足す
type eventDocument struct {
	At      time.Time         `firestore:"at"`
	Actor   string            `firestore:"actor"`
	Action  string            `firestore:"action"`
	Changes map[string]string `firestore:"changes"`
	Cause   string            `firestore:"cause"`
}

// 作成の記録に使う値
const (
	eventActorAPI       = "api"
	eventActionCreated  = "created"
	eventCauseRequested = "guest_requested"
)

func (r *Repository) GetPriorityPass(
	ctx context.Context,
	id prioritypassmodel.PassID,
) (*prioritypassmodel.PriorityPass, error) {
	snapshot, err := r.doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, prioritypassmodel.ErrPriorityPassNotFound
		}
		return nil, fmt.Errorf("get priority pass %q: %w",
			id,
			err,
		)
	}

	var doc document
	if err := snapshot.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("decode priority pass %q: %w",
			id,
			err,
		)
	}

	pass, err := prioritypassmodel.RestorePriorityPass(
		id,
		prioritypassmodel.ParkID(doc.ParkID),
		prioritypassmodel.TicketID(doc.TicketID),
		prioritypassmodel.AttractionID(doc.AttractionID),
		prioritypassmodel.TimeSlotID(doc.TimeSlotID),
		prioritypassmodel.Status(doc.Status),
		doc.CreatedAt,
		doc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("restore priority pass %q: %w",
			id,
			err,
		)
	}

	return pass, nil
}

// CreatePriorityPass は優先パスと、その作成を記録したイベントを同時に保存する
//
// 片方だけが残ると状態の履歴が追えなくなるため、同じトランザクションで書く
func (r *Repository) CreatePriorityPass(ctx context.Context, pass *prioritypassmodel.PriorityPass) error {
	doc := r.doc(pass.ID())
	data := toDocument(pass)
	event := toCreatedEvent(pass)

	err := r.client.RunTransaction(ctx, func(_ context.Context, tx *gcpfirestore.Transaction) error {
		// Create は既にあると失敗するため、採番が衝突した場合も上書きにならない
		if err := tx.Create(doc, data); err != nil {
			return err
		}

		return tx.Create(doc.Collection(eventCollection).NewDoc(), event)
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return prioritypassmodel.ErrPriorityPassAlreadyExists
		}
		return fmt.Errorf("create priority pass %q: %w",
			pass.ID(),
			err,
		)
	}

	return nil
}

func toDocument(pass *prioritypassmodel.PriorityPass) document {
	return document{
		ParkID:       pass.ParkID().String(),
		TicketID:     pass.TicketID().String(),
		AttractionID: pass.AttractionID().String(),
		TimeSlotID:   pass.TimeSlotID().String(),
		Status:       pass.Status().String(),
		// 申込の時点ではまだ publish していないため、後から埋める場所として null を置く
		PublishedAt: nil,
		CreatedAt:   pass.CreatedAt(),
		UpdatedAt:   pass.UpdatedAt(),
	}
}

func toCreatedEvent(pass *prioritypassmodel.PriorityPass) eventDocument {
	return eventDocument{
		At:     pass.CreatedAt(),
		Actor:  eventActorAPI,
		Action: eventActionCreated,
		Changes: map[string]string{
			"parkId":       pass.ParkID().String(),
			"ticketId":     pass.TicketID().String(),
			"attractionId": pass.AttractionID().String(),
			"timeSlotId":   pass.TimeSlotID().String(),
			"status":       pass.Status().String(),
		},
		Cause: eventCauseRequested,
	}
}
