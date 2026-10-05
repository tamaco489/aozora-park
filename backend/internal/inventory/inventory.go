// Package inventory は枠在庫を扱う
package inventory

import (
	gcpfirestore "cloud.google.com/go/firestore"

	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/inventory/v1/inventoryv1connect"
	inventoryhandler "github.com/tamaco489/aozora-park/backend/internal/inventory/handler"
	inventoryfirestore "github.com/tamaco489/aozora-park/backend/internal/inventory/infrastructure/firestore"
	inventoryusecase "github.com/tamaco489/aozora-park/backend/internal/inventory/usecase"
)

// NewConnectHandler は connect の入口までを組み立てる
//
// usecase と infrastructure と handler の結線はここに閉じる
func NewConnectHandler(client *gcpfirestore.Client) inventoryv1connect.InventoryServiceHandler {
	inventories := inventoryfirestore.NewRepository(client)

	return inventoryhandler.NewConnect(
		inventoryusecase.NewUpdateDateInventory(inventories, inventories),
		inventoryusecase.NewGetDateInventory(inventories),
		inventoryusecase.NewListTimeSlots(inventories),
	)
}
