package main

import (
	"context"
	"log"
	"net/http"

	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore"
	"github.com/tamaco489/aozora-park/backend/internal/platform/config"
	"github.com/tamaco489/aozora-park/backend/internal/platform/observability/logging"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/httpx"
	"github.com/tamaco489/aozora-park/backend/internal/prioritypass"
)

// pushPath は Pub/Sub の push を受けるパス
//
// infra/modules/pubsub/main.tf の local.push_path と同じ値にする、片方を変更したらもう片方も修正する
const pushPath = "/pubsub/push"

func main() {
	if err := run(); err != nil {
		log.Fatalf("priority-pass-issuer: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := logging.New(cfg.LogLevel)
	app := httpx.NewApp(logger)

	ctx := context.Background()

	firestoreClient, err := firestore.New(ctx, cfg.ProjectID)
	if err != nil {
		return err
	}
	app.Cleanup("firestore", func(context.Context) error { return firestoreClient.Close() })

	mux := http.NewServeMux()

	// 呼び出せるのは Pub/Sub だけのため connect もヘルスチェックの RPC も置かない
	// Cloud Run の起動の判定はコンテナがポートを開いたかで行う
	mux.Handle(pushPath, prioritypass.NewPushHandler(firestoreClient, logger))

	return app.Serve(
		ctx,
		":"+cfg.Port,
		mux,
	)
}
