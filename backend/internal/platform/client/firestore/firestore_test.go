package firestore_test

import (
	"context"
	"os"
	"testing"

	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore/firestoretest"
)

func TestMain(m *testing.M) {
	os.Exit(firestoretest.Main(m))
}

// New が作ったクライアントでエミュレータに読み書きできることを確かめる
func TestNew(t *testing.T) {
	client := firestoretest.Client(t)
	ctx := context.Background()

	doc := client.Collection("connectivity").Doc(t.Name())
	t.Cleanup(func() {
		if _, err := doc.Delete(ctx); err != nil {
			t.Errorf("doc.Delete() = %v, want nil", err)
		}
	})

	want := map[string]any{"name": "Aozora Park"}
	if _, err := doc.Set(ctx, want); err != nil {
		t.Fatalf("doc.Set(%v) = %v, want nil", want, err)
	}

	snapshot, err := doc.Get(ctx)
	if err != nil {
		t.Fatalf("doc.Get() = %v, want nil", err)
	}

	if got := snapshot.Data()["name"]; got != want["name"] {
		t.Errorf("doc.Get() の name = %v, want %v", got, want["name"])
	}
}
