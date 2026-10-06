package firestore

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

func TestRepositoryCreateAndGetPriorityPass(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const id = prioritypassmodel.PassID("pass-create-and-get")
	cleanupPriorityPassHelper(t, client, id)

	pass := restorePriorityPassHelper(t, id, prioritypassmodel.StatusRequested)

	if err := repo.CreatePriorityPass(ctx, pass); err != nil {
		t.Fatalf("Repository.CreatePriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}

	got, err := repo.GetPriorityPass(ctx, id)
	if err != nil {
		t.Fatalf("Repository.GetPriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}

	if got.ID() != id ||
		got.ParkID() != testParkID ||
		got.TicketID() != testTicketID ||
		got.AttractionID() != testAttractionID ||
		got.TimeSlotID() != testTimeSlotID ||
		got.Status() != prioritypassmodel.StatusRequested {
		t.Errorf("Repository.GetPriorityPass(%q) = (%q, %q, %q, %q, %q, %q), want (%q, %q, %q, %q, %q, %q)",
			id,
			got.ID(),
			got.ParkID(),
			got.TicketID(),
			got.AttractionID(),
			got.TimeSlotID(),
			got.Status(),
			id,
			testParkID,
			testTicketID,
			testAttractionID,
			testTimeSlotID,
			prioritypassmodel.StatusRequested,
		)
	}

	if !got.CreatedAt().Equal(testNow) || !got.UpdatedAt().Equal(testNow) {
		t.Errorf("Repository.GetPriorityPass(%q) の時刻 = (%v, %v), want (%v, %v)",
			id,
			got.CreatedAt(),
			got.UpdatedAt(),
			testNow,
			testNow,
		)
	}
}

// TestRepositoryCreatePriorityPassWritesPublishedAt は publish 前の印として null を置くことを確かめる
func TestRepositoryCreatePriorityPassWritesPublishedAt(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const id = prioritypassmodel.PassID("pass-published-at")
	cleanupPriorityPassHelper(t, client, id)

	if err := repo.CreatePriorityPass(ctx, restorePriorityPassHelper(t, id, prioritypassmodel.StatusRequested)); err != nil {
		t.Fatalf("Repository.CreatePriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}

	snapshot, err := client.Collection(collection).Doc(id.String()).Get(ctx)
	if err != nil {
		t.Fatalf("Get(%q) = %v, want nil",
			id,
			err,
		)
	}

	data := snapshot.Data()
	publishedAt, ok := data["publishedAt"]
	if !ok {
		t.Fatalf("Get(%q) の publishedAt = 無し, want nil の値",
			id,
		)
	}
	if publishedAt != nil {
		t.Errorf("Get(%q) の publishedAt = %v, want %v",
			id,
			publishedAt,
			nil,
		)
	}
}

// TestRepositoryCreatePriorityPassWritesEvent は親と同じトランザクションで作成の記録が残ることを確かめる
func TestRepositoryCreatePriorityPassWritesEvent(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const id = prioritypassmodel.PassID("pass-event")
	cleanupPriorityPassHelper(t, client, id)

	if err := repo.CreatePriorityPass(ctx, restorePriorityPassHelper(t, id, prioritypassmodel.StatusRequested)); err != nil {
		t.Fatalf("Repository.CreatePriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}

	events, err := client.Collection(collection).
		Doc(id.String()).
		Collection(eventCollection).
		Documents(ctx).
		GetAll()
	if err != nil {
		t.Fatalf("GetAll(%q の events) = %v, want nil",
			id,
			err,
		)
	}

	if len(events) != 1 {
		t.Fatalf("GetAll(%q の events) の件数 = %d, want %d",
			id,
			len(events),
			1,
		)
	}

	var got eventDocument
	if err := events[0].DataTo(&got); err != nil {
		t.Fatalf("DataTo(%q の events) = %v, want nil",
			id,
			err,
		)
	}

	want := eventDocument{
		At:     testNow,
		Actor:  eventActorAPI,
		Action: eventActionCreated,
		Changes: map[string]string{
			"parkId":       testParkID.String(),
			"ticketId":     testTicketID.String(),
			"attractionId": testAttractionID.String(),
			"timeSlotId":   testTimeSlotID.String(),
			"status":       prioritypassmodel.StatusRequested.String(),
		},
		Cause: eventCauseRequested,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Repository.CreatePriorityPass(%q) のイベントの差分 (-want +got):\n%s",
			id,
			diff,
		)
	}
}

func TestRepositoryCreatePriorityPassAlreadyExists(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const id = prioritypassmodel.PassID("pass-already-exists")
	cleanupPriorityPassHelper(t, client, id)

	if err := repo.CreatePriorityPass(ctx, restorePriorityPassHelper(t, id, prioritypassmodel.StatusRequested)); err != nil {
		t.Fatalf("1 回目の Repository.CreatePriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}

	err := repo.CreatePriorityPass(ctx, restorePriorityPassHelper(t, id, prioritypassmodel.StatusIssued))
	if !errors.Is(err, prioritypassmodel.ErrPriorityPassAlreadyExists) {
		t.Fatalf("2 回目の Repository.CreatePriorityPass(%q) = %v, want %v",
			id,
			err,
			prioritypassmodel.ErrPriorityPassAlreadyExists,
		)
	}

	// 2 回目が弾かれたことで、状態が後から来た値に置き換わっていないことを確かめる
	got, err := repo.GetPriorityPass(ctx, id)
	if err != nil {
		t.Fatalf("Repository.GetPriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}
	if got.Status() != prioritypassmodel.StatusRequested {
		t.Errorf("Repository.GetPriorityPass(%q) の状態 = %q, want %q",
			id,
			got.Status(),
			prioritypassmodel.StatusRequested,
		)
	}
}

func TestRepositoryGetPriorityPassNotFound(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	const id = prioritypassmodel.PassID("pass-not-found")

	_, err := repo.GetPriorityPass(t.Context(), id)
	if !errors.Is(err, prioritypassmodel.ErrPriorityPassNotFound) {
		t.Errorf("Repository.GetPriorityPass(%q) = %v, want %v",
			id,
			err,
			prioritypassmodel.ErrPriorityPassNotFound,
		)
	}
}
