// Package firestoretest はテスト用に Firestore のエミュレータを用意する
//
// infrastructure のテストから使う (`.claude/rules/go/testing.md`)
package firestoretest

import (
	"context"
	"fmt"
	"os"
	"testing"

	gcpfirestore "cloud.google.com/go/firestore"
	"github.com/testcontainers/testcontainers-go"
	tcfirestore "github.com/testcontainers/testcontainers-go/modules/gcloud/firestore"

	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore"
)

// ProjectID はエミュレータに使うプロジェクト ID (demo- で始まる ID は SDK が本物の Google Cloud への接続を拒む)
const ProjectID = "demo-aozora-park"

// emulatorImage は Firestore エミュレータを含む Cloud SDK のイメージ (エミュレータの挙動が版で変わるため、タグを固定する)
const emulatorImage = "gcr.io/google.com/cloudsdktool/cloud-sdk:584.0.0-emulators"

// gate はエミュレータを起動するかを決める環境変数 (Docker が無い環境でも go test ./... が通るようにするため、既定では起動しない)
const gate = "DOCKER_TESTS"

// client は Main が用意したクライアント、起動しなかった場合は nil のまま
var client *gcpfirestore.Client

// Main は TestMain から呼び、エミュレータを起こしてからテストを走らせる (コンテナはパッケージごとに 1 つにする、テストごとの起動は遅すぎるため)
func Main(m *testing.M) int {
	if os.Getenv(gate) == "" {
		return m.Run()
	}

	stop, err := start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "firestoretest: %v\n", err)
		return 1
	}
	defer stop()

	return m.Run()
}

// Client は Main が用意したクライアントを返す (エミュレータを起動していないときはテストを飛ばす)
func Client(tb testing.TB) *gcpfirestore.Client {
	tb.Helper()

	if client == nil {
		tb.Skip("run with " + gate + "=1")
	}

	return client
}

func start() (func(), error) {
	ctx := context.Background()

	container, err := tcfirestore.Run(ctx, emulatorImage, tcfirestore.WithProjectID(ProjectID))
	if err != nil {
		return nil, fmt.Errorf("run emulator: %w", err)
	}

	// SDK は環境変数で接続先を決めるため、クライアントを作る前に設定する
	if err := os.Setenv("FIRESTORE_EMULATOR_HOST", container.URI()); err != nil {
		return nil, fmt.Errorf("set emulator host: %w", err)
	}

	client, err = firestore.New(ctx, ProjectID)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	return func() {
		if err := client.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "firestoretest: close client: %v\n", err)
		}
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Fprintf(os.Stderr, "firestoretest: terminate container: %v\n", err)
		}
	}, nil
}
