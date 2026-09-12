// Package park はパークのマスタを扱う
package park

import (
	gcpfirestore "cloud.google.com/go/firestore"

	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1/parkv1connect"
	parkhandler "github.com/tamaco489/aozora-park/backend/internal/park/handler"
	parkfirestore "github.com/tamaco489/aozora-park/backend/internal/park/infrastructure/firestore"
	parkusecase "github.com/tamaco489/aozora-park/backend/internal/park/usecase"
)

// NewConnectHandler は connect の入口までを組み立てる
//
// usecase と infrastructure と handler の結線はここに閉じる
func NewConnectHandler(client *gcpfirestore.Client) parkv1connect.ParkServiceHandler {
	parks := parkfirestore.NewRepository(client)

	return parkhandler.NewConnect(
		parkusecase.NewCreate(parks),
		parkusecase.NewGet(parks),
		parkusecase.NewUpdate(parks, parks),
	)
}
