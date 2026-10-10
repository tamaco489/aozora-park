package prioritypass

import (
	"log/slog"
	"net/http"

	gcpfirestore "cloud.google.com/go/firestore"

	prioritypasshandler "github.com/tamaco489/aozora-park/backend/internal/prioritypass/handler"
	prioritypassfirestore "github.com/tamaco489/aozora-park/backend/internal/prioritypass/infrastructure/firestore"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

// NewPushHandler は Pub/Sub push の入口までを組み立てる
//
// 割当は Firestore しか触らないため、publish と違って Pub/Sub のクライアントを要さない
func NewPushHandler(
	firestoreClient *gcpfirestore.Client,
	logger *slog.Logger,
) http.Handler {
	passes := prioritypassfirestore.NewRepository(firestoreClient)

	return prioritypasshandler.NewSubscriber(
		logger,
		prioritypassusecase.NewAllocateTimeSlot(passes, logger),
	)
}
