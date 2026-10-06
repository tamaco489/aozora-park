// Package prioritypass は優先パスを扱う
package prioritypass

import (
	gcpfirestore "cloud.google.com/go/firestore"

	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/prioritypass/v1/prioritypassv1connect"
	prioritypasshandler "github.com/tamaco489/aozora-park/backend/internal/prioritypass/handler"
	prioritypassfirestore "github.com/tamaco489/aozora-park/backend/internal/prioritypass/infrastructure/firestore"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

// NewConnectHandler は connect の入口までを組み立てる
//
// usecase と infrastructure と handler の結線はここに閉じる
func NewConnectHandler(client *gcpfirestore.Client) prioritypassv1connect.PriorityPassServiceHandler {
	passes := prioritypassfirestore.NewRepository(client)

	return prioritypasshandler.NewConnect(
		prioritypassusecase.NewRequestPriorityPass(passes),
		prioritypassusecase.NewGetPriorityPass(passes),
	)
}
