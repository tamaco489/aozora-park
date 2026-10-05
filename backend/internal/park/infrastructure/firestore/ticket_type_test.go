package firestore

import (
	"context"
	"errors"
	"testing"

	gcpfirestore "cloud.google.com/go/firestore"

	parkmodel "github.com/tamaco489/aozora-park/backend/internal/park/domain/model"
)

// ticketTypeParkID は券種のテストが使う親のパーク、他のテストとドキュメントを分ける
const ticketTypeParkID = parkmodel.ParkID("park-for-ticket-type-test")

func cleanupTicketType(
	tb testing.TB,
	client *gcpfirestore.Client,
	ticketType *parkmodel.TicketType,
) {
	tb.Helper()

	tb.Cleanup(func() {
		doc := client.Collection(collection).Doc(ticketType.ParkID().String()).
			Collection(ticketTypeCollection).Doc(ticketType.ID().String())
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				ticketType.ID(),
				err,
			)
		}
	})
}

func TestRepositoryCreateAndGetTicketType(t *testing.T) {
	repo, client := newRepository(t)
	ctx := t.Context()

	ticketType, err := parkmodel.NewTicketType(
		ticketTypeParkID,
		"1 デーパスポート",
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		t.Fatalf("NewTicketType() = %v, want nil", err)
	}
	cleanupTicketType(
		t,
		client,
		ticketType,
	)

	if err := repo.CreateTicketType(ctx, ticketType); err != nil {
		t.Fatalf("Repository.CreateTicketType(%q) = %v, want nil",
			ticketType.ID(),
			err,
		)
	}

	got, err := repo.GetTicketType(
		ctx,
		ticketType.ParkID(),
		ticketType.ID(),
	)
	if err != nil {
		t.Fatalf("Repository.GetTicketType(%q) = %v, want nil",
			ticketType.ID(),
			err,
		)
	}

	if got.ID() != ticketType.ID() || got.ParkID() != ticketType.ParkID() || got.Name() != ticketType.Name() ||
		got.Price() != ticketType.Price() || got.EntryTimeFrom() != ticketType.EntryTimeFrom() ||
		got.EntryTimeTo() != ticketType.EntryTimeTo() {
		t.Errorf("Repository.GetTicketType(%q) = (%q, %q, %q, %d, %q, %q), want (%q, %q, %q, %d, %q, %q)",
			ticketType.ID(),
			got.ParkID(),
			got.ID(),
			got.Name(),
			got.Price(),
			got.EntryTimeFrom(),
			got.EntryTimeTo(),
			ticketType.ParkID(),
			ticketType.ID(),
			ticketType.Name(),
			ticketType.Price(),
			ticketType.EntryTimeFrom(),
			ticketType.EntryTimeTo(),
		)
	}
}

func TestRepositoryCreateTicketTypeAlreadyExists(t *testing.T) {
	repo, client := newRepository(t)
	ctx := t.Context()

	ticketType, err := parkmodel.NewTicketType(
		ticketTypeParkID,
		"1 デーパスポート",
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		t.Fatalf("NewTicketType() = %v, want nil", err)
	}
	cleanupTicketType(
		t,
		client,
		ticketType,
	)

	if err := repo.CreateTicketType(ctx, ticketType); err != nil {
		t.Fatalf("Repository.CreateTicketType(%q) = %v, want nil",
			ticketType.ID(),
			err,
		)
	}

	if err := repo.CreateTicketType(ctx, ticketType); !errors.Is(err, parkmodel.ErrTicketTypeAlreadyExists) {
		t.Errorf("Repository.CreateTicketType(%q) の 2 回目 = %v, want %v",
			ticketType.ID(),
			err,
			parkmodel.ErrTicketTypeAlreadyExists,
		)
	}
}

func TestRepositoryGetTicketTypeNotFound(t *testing.T) {
	repo, _ := newRepository(t)

	const id = parkmodel.TicketTypeID("ticket-type-does-not-exist")

	_, err := repo.GetTicketType(
		t.Context(),
		ticketTypeParkID,
		id,
	)
	if !errors.Is(err, parkmodel.ErrTicketTypeNotFound) {
		t.Errorf("Repository.GetTicketType(%q) = %v, want %v",
			id,
			err,
			parkmodel.ErrTicketTypeNotFound,
		)
	}
}

func TestRepositoryUpdateTicketType(t *testing.T) {
	repo, client := newRepository(t)
	ctx := t.Context()

	ticketType, err := parkmodel.NewTicketType(
		ticketTypeParkID,
		"1 デーパスポート",
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		t.Fatalf("NewTicketType() = %v, want nil", err)
	}
	cleanupTicketType(
		t,
		client,
		ticketType,
	)

	if err := repo.CreateTicketType(ctx, ticketType); err != nil {
		t.Fatalf("Repository.CreateTicketType(%q) = %v, want nil",
			ticketType.ID(),
			err,
		)
	}

	if err := ticketType.Update("2 デーパスポート", 14000, "10:00", "20:00"); err != nil {
		t.Fatalf("TicketType.Update() = %v, want nil", err)
	}

	if err := repo.UpdateTicketType(ctx, ticketType); err != nil {
		t.Fatalf("Repository.UpdateTicketType(%q) = %v, want nil",
			ticketType.ID(),
			err,
		)
	}

	got, err := repo.GetTicketType(
		ctx,
		ticketType.ParkID(),
		ticketType.ID(),
	)
	if err != nil {
		t.Fatalf("Repository.GetTicketType(%q) = %v, want nil",
			ticketType.ID(),
			err,
		)
	}

	if got.Name() != "2 デーパスポート" || got.Price() != 14000 ||
		got.EntryTimeFrom() != "10:00" || got.EntryTimeTo() != "20:00" {
		t.Errorf("Repository.GetTicketType(%q) = (%q, %d, %q, %q), want (%q, %d, %q, %q)",
			ticketType.ID(),
			got.Name(),
			got.Price(),
			got.EntryTimeFrom(),
			got.EntryTimeTo(),
			"2 デーパスポート",
			14000,
			"10:00",
			"20:00",
		)
	}
}

func TestRepositoryUpdateTicketTypeNotFound(t *testing.T) {
	repo, _ := newRepository(t)

	ticketType, err := parkmodel.RestoreTicketType(
		ticketTypeParkID,
		"ticket-type-does-not-exist",
		"1 デーパスポート",
		8000,
		"09:00",
		"21:00",
	)
	if err != nil {
		t.Fatalf("RestoreTicketType() = %v, want nil", err)
	}

	if err := repo.UpdateTicketType(t.Context(), ticketType); !errors.Is(err, parkmodel.ErrTicketTypeNotFound) {
		t.Errorf("Repository.UpdateTicketType(%q) = %v, want %v",
			ticketType.ID(),
			err,
			parkmodel.ErrTicketTypeNotFound,
		)
	}
}
