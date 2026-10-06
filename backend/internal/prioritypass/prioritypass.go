// Package prioritypass は優先パスを扱う
package prioritypass

import (
	"log/slog"

	gcpfirestore "cloud.google.com/go/firestore"
	gcppubsub "cloud.google.com/go/pubsub/v2"

	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1/prioritypassv1connect"
	prioritypasshandler "github.com/tamaco489/aozora-park/backend/internal/prioritypass/handler"
	prioritypassfirestore "github.com/tamaco489/aozora-park/backend/internal/prioritypass/infrastructure/firestore"
	prioritypasspubsub "github.com/tamaco489/aozora-park/backend/internal/prioritypass/infrastructure/pubsub"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

// NewConnectHandler は connect の入口までを組み立てる
//
// usecase と infrastructure と handler の結線はここに閉じる
// 2 つ目の戻り値は publish の goroutine を止める処理で、呼び出し側が App に登録する
func NewConnectHandler(
	firestoreClient *gcpfirestore.Client,
	pubsubClient *gcppubsub.Client,
	logger *slog.Logger,
) (prioritypassv1connect.PriorityPassServiceHandler, func()) {
	passes := prioritypassfirestore.NewRepository(firestoreClient)
	publisher := prioritypasspubsub.NewPublisher(pubsubClient)

	handler := prioritypasshandler.NewConnect(
		prioritypassusecase.NewRequestPriorityPass(
			passes,
			publisher,
			logger,
		),
		prioritypassusecase.NewGetPriorityPass(passes),
	)

	return handler, publisher.Stop
}
