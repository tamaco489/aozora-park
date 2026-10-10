package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"

	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1/inventoryv1connect"
	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1/parkv1connect"
	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1/prioritypassv1connect"
	"github.com/tamaco489/aozora-park/backend/internal/inventory"
	"github.com/tamaco489/aozora-park/backend/internal/park"
	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore"
	"github.com/tamaco489/aozora-park/backend/internal/platform/client/pubsub"
	"github.com/tamaco489/aozora-park/backend/internal/platform/config"
	"github.com/tamaco489/aozora-park/backend/internal/platform/observability/logging"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/httpx"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/interceptor"
	"github.com/tamaco489/aozora-park/backend/internal/prioritypass"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("api: %v", err)
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
	// 登録の逆順に閉じるため、依存される側から順に登録する
	app.Cleanup("firestore", func(context.Context) error { return firestoreClient.Close() })

	pubsubClient, err := pubsub.New(ctx, cfg.ProjectID)
	if err != nil {
		return err
	}
	app.Cleanup("pubsub", func(context.Context) error { return pubsubClient.Close() })

	// 機能側が持つ publisher はクライアントより先に止める、送り残しの送信にクライアントを使うため
	prioritypassHandler, stopPriorityPassPublisher := prioritypass.NewConnectHandler(
		firestoreClient,
		pubsubClient,
		logger,
	)
	app.Cleanup("prioritypass publisher", func(context.Context) error {
		stopPriorityPassPublisher()
		return nil
	})

	// 共通処理は 1 つにまとめてすべての connect ハンドラに渡す (機能ごとに組み立てない)
	opts := interceptor.All(logger)

	mux := http.NewServeMux()

	// Cloud Run のプローブが叩く口、grpcurl と grpc-health-probe からも同じ形で呼べる
	// 引数のサービス名を増やすと、そのサービス単位でも状態を答えられる
	mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker(
		parkv1connect.ParkServiceName,
		inventoryv1connect.InventoryServiceName,
		prioritypassv1connect.PriorityPassServiceName,
	), opts))

	// 結線は機能パッケージ側に閉じるため、ここは組み立て関数を呼んで登録するだけにする
	mux.Handle(parkv1connect.NewParkServiceHandler(park.NewConnectHandler(firestoreClient), opts))
	mux.Handle(inventoryv1connect.NewInventoryServiceHandler(inventory.NewConnectHandler(firestoreClient), opts))
	mux.Handle(prioritypassv1connect.NewPriorityPassServiceHandler(prioritypassHandler, opts))

	// grpcui と buf curl がサービス一覧を引けるようにする
	reflector := grpcreflect.NewStaticReflector(
		parkv1connect.ParkServiceName,
		inventoryv1connect.InventoryServiceName,
		prioritypassv1connect.PriorityPassServiceName,
		grpchealth.HealthV1ServiceName,
	)

	// v1 と v1alpha の両方を公開するのは、v1alpha しか呼ばないクライアントがまだあるため
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	return app.Serve(
		ctx,
		":"+cfg.Port,
		mux,
	)
}
