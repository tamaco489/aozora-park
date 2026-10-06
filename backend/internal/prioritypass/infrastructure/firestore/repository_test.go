package firestore

import (
	"context"
	"os"
	"testing"
	"time"

	gcpfirestore "cloud.google.com/go/firestore"

	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore/firestoretest"
	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// 保存する優先パスに使う値、識別子はテストごとに分ける
const (
	testParkID       = prioritypassmodel.ParkID("park-1")
	testTicketID     = prioritypassmodel.TicketID("ticket-1")
	testAttractionID = prioritypassmodel.AttractionID("attraction-1")
	testTimeSlotID   = prioritypassmodel.TimeSlotID("20261005_1000")
)

// testNow は保存する時刻、Firestore はマイクロ秒までしか持たないため端数を置かない
var testNow = time.Date(
	2026, 10, 5,
	9, 0, 0, 0,
	time.UTC,
)

func TestMain(m *testing.M) {
	os.Exit(firestoretest.Main(m))
}

// newRepositoryHelper は保存先を用意する
//
// テスト間の分離は優先パスの識別子を分けて行う、同じコレクションを共有するため
func newRepositoryHelper(tb testing.TB) (*Repository, *gcpfirestore.Client) {
	tb.Helper()

	client := firestoretest.Client(tb)

	return NewRepository(client), client
}

// restorePriorityPassHelper は保存する優先パスを組み立てる
func restorePriorityPassHelper(
	tb testing.TB,
	id prioritypassmodel.PassID,
	status prioritypassmodel.Status,
) *prioritypassmodel.PriorityPass {
	tb.Helper()

	pass, err := prioritypassmodel.RestorePriorityPass(
		id,
		testParkID,
		testTicketID,
		testAttractionID,
		testTimeSlotID,
		status,
		testNow,
		testNow,
	)
	if err != nil {
		tb.Fatalf("RestorePriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}

	return pass
}

// cleanupPriorityPass は検査対象が作成する優先パスとイベントの後始末を登録する
func cleanupPriorityPass(
	tb testing.TB,
	client *gcpfirestore.Client,
	id prioritypassmodel.PassID,
) {
	tb.Helper()

	doc := client.Collection(collection).Doc(id.String())

	tb.Cleanup(func() {
		// t.Context() は Cleanup の直前に取り消されるため、後始末は取り消されないものを使う
		ctx := context.Background()

		// 子コレクションは親を消しても残るため、先に消す
		events, err := doc.Collection(eventCollection).Documents(ctx).GetAll()
		if err != nil {
			tb.Errorf("GetAll(%q の events) = %v, want nil",
				id,
				err,
			)
		}
		for _, event := range events {
			if _, err := event.Ref.Delete(ctx); err != nil {
				tb.Errorf("Delete(%q の events) = %v, want nil",
					id,
					err,
				)
			}
		}

		if _, err := doc.Delete(ctx); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				id,
				err,
			)
		}
	})
}
