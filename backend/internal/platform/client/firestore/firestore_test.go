package firestore

import (
	"context"
	"os"
	"testing"
)

// エミュレータが要るため、接続先が設定されていない環境では飛ばす
func requireEmulator(tb testing.TB) string {
	tb.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		tb.Skip("run with FIRESTORE_EMULATOR_HOST set")
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		tb.Fatal("GOOGLE_CLOUD_PROJECT is required")
	}
	return projectID
}

func TestNew(t *testing.T) {
	projectID := requireEmulator(t)
	ctx := context.Background()

	client, err := New(ctx, projectID)
	if err != nil {
		t.Fatalf("New(ctx, %q) = %v, want nil", projectID, err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("client.Close() = %v, want nil", err)
		}
	})

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
