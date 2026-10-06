package firestore

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	gcpfirestore "cloud.google.com/go/firestore"
	"github.com/google/go-cmp/cmp"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// allocatedAt は割当の時刻、申込の時刻と区別できる値にする
var allocatedAt = testNow.Add(time.Hour)

// storedTimeSlotDocument は inventory が保存する時間帯枠の形
//
// 割当が読まない項目も置き、残りだけを読めていることを確かめる
type storedTimeSlotDocument struct {
	AttractionID string `firestore:"attractionId"`
	Date         string `firestore:"date"`
	StartTime    string `firestore:"startTime"`
	Capacity     int32  `firestore:"capacity"`
	Remaining    int32  `firestore:"remaining"`
}

func TestRepositoryAllocateTimeSlot(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const (
		id         = prioritypassmodel.PassID("pass-allocate")
		timeSlotID = prioritypassmodel.TimeSlotID("20261005_1100")
	)
	cleanupPriorityPassHelper(t, client, id)
	storeTimeSlotHelper(t, client, timeSlotID, 2)
	createPriorityPassHelper(t, repo, id, timeSlotID)

	got, err := repo.AllocateTimeSlot(ctx, id, allocatedAt)
	if err != nil {
		t.Fatalf("Repository.AllocateTimeSlot(%q, %v) = %v, want nil",
			id,
			allocatedAt,
			err,
		)
	}

	if !got.Changed || got.Pass.Status() != prioritypassmodel.StatusIssued {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) = (%v, %q), want (%v, %q)",
			id,
			allocatedAt,
			got.Changed,
			got.Pass.Status(),
			true,
			prioritypassmodel.StatusIssued,
		)
	}

	if remaining := remainingHelper(t, client, timeSlotID); remaining != 1 {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) の後の残り = %d, want %d",
			id,
			allocatedAt,
			remaining,
			1,
		)
	}

	wantDocument := document{
		ParkID:       testParkID.String(),
		TicketID:     testTicketID.String(),
		AttractionID: testAttractionID.String(),
		TimeSlotID:   timeSlotID.String(),
		Status:       prioritypassmodel.StatusIssued.String(),
		PublishedAt:  nil,
		CreatedAt:    testNow,
		UpdatedAt:    allocatedAt,
	}
	if diff := cmp.Diff(wantDocument, documentHelper(t, client, id)); diff != "" {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) の差分 (-want +got):\n%s",
			id,
			allocatedAt,
			diff,
		)
	}

	wantEvent := eventDocument{
		At:     allocatedAt,
		Actor:  eventActorIssuer,
		Action: eventActionStatusChanged,
		Changes: map[string]string{
			"status": prioritypassmodel.StatusIssued.String(),
		},
		Cause: eventCauseSlotAllocated,
	}
	if diff := cmp.Diff(wantEvent, eventHelper(t, client, id, eventCauseSlotAllocated)); diff != "" {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) のイベントの差分 (-want +got):\n%s",
			id,
			allocatedAt,
			diff,
		)
	}
}

// TestRepositoryAllocateTimeSlotTwice は再配信で 2 回処理しても残りが 1 しか減らないことを確かめる
func TestRepositoryAllocateTimeSlotTwice(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const (
		id         = prioritypassmodel.PassID("pass-allocate-twice")
		timeSlotID = prioritypassmodel.TimeSlotID("20261005_1200")
	)
	cleanupPriorityPassHelper(t, client, id)
	storeTimeSlotHelper(t, client, timeSlotID, 2)
	createPriorityPassHelper(t, repo, id, timeSlotID)

	if _, err := repo.AllocateTimeSlot(ctx, id, allocatedAt); err != nil {
		t.Fatalf("1 回目の Repository.AllocateTimeSlot(%q, %v) = %v, want nil",
			id,
			allocatedAt,
			err,
		)
	}

	// 2 回目は別の時刻を渡し、遷移していないことを更新の時刻でも確かめる
	retriedAt := allocatedAt.Add(time.Minute)

	got, err := repo.AllocateTimeSlot(ctx, id, retriedAt)
	if err != nil {
		t.Fatalf("2 回目の Repository.AllocateTimeSlot(%q, %v) = %v, want nil",
			id,
			retriedAt,
			err,
		)
	}

	if got.Changed || got.Pass.Status() != prioritypassmodel.StatusIssued {
		t.Errorf("2 回目の Repository.AllocateTimeSlot(%q, %v) = (%v, %q), want (%v, %q)",
			id,
			retriedAt,
			got.Changed,
			got.Pass.Status(),
			false,
			prioritypassmodel.StatusIssued,
		)
	}

	if remaining := remainingHelper(t, client, timeSlotID); remaining != 1 {
		t.Errorf("2 回目の Repository.AllocateTimeSlot(%q, %v) の後の残り = %d, want %d",
			id,
			retriedAt,
			remaining,
			1,
		)
	}

	if updatedAt := documentHelper(t, client, id).UpdatedAt; !updatedAt.Equal(allocatedAt) {
		t.Errorf("2 回目の Repository.AllocateTimeSlot(%q, %v) の後の更新の時刻 = %v, want %v",
			id,
			retriedAt,
			updatedAt,
			allocatedAt,
		)
	}

	// 作成と 1 回目の遷移の 2 件だけが残る
	if count := eventCountHelper(t, client, id); count != 2 {
		t.Errorf("2 回目の Repository.AllocateTimeSlot(%q, %v) の後のイベントの件数 = %d, want %d",
			id,
			retriedAt,
			count,
			2,
		)
	}
}

// TestRepositoryAllocateTimeSlotSoldOut は上限まで申し込むと以降が sold_out になり残りが負にならないことを確かめる
func TestRepositoryAllocateTimeSlotSoldOut(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const (
		firstID    = prioritypassmodel.PassID("pass-sold-out-first")
		secondID   = prioritypassmodel.PassID("pass-sold-out-second")
		timeSlotID = prioritypassmodel.TimeSlotID("20261005_1300")
	)
	cleanupPriorityPassHelper(t, client, firstID)
	cleanupPriorityPassHelper(t, client, secondID)
	storeTimeSlotHelper(t, client, timeSlotID, 1)
	createPriorityPassHelper(t, repo, firstID, timeSlotID)
	createPriorityPassHelper(t, repo, secondID, timeSlotID)

	if _, err := repo.AllocateTimeSlot(ctx, firstID, allocatedAt); err != nil {
		t.Fatalf("Repository.AllocateTimeSlot(%q, %v) = %v, want nil",
			firstID,
			allocatedAt,
			err,
		)
	}

	got, err := repo.AllocateTimeSlot(ctx, secondID, allocatedAt)
	if err != nil {
		t.Fatalf("Repository.AllocateTimeSlot(%q, %v) = %v, want nil",
			secondID,
			allocatedAt,
			err,
		)
	}

	if !got.Changed || got.Pass.Status() != prioritypassmodel.StatusSoldOut {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) = (%v, %q), want (%v, %q)",
			secondID,
			allocatedAt,
			got.Changed,
			got.Pass.Status(),
			true,
			prioritypassmodel.StatusSoldOut,
		)
	}

	if remaining := remainingHelper(t, client, timeSlotID); remaining != 0 {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) の後の残り = %d, want %d",
			secondID,
			allocatedAt,
			remaining,
			0,
		)
	}

	wantEvent := eventDocument{
		At:     allocatedAt,
		Actor:  eventActorIssuer,
		Action: eventActionStatusChanged,
		Changes: map[string]string{
			"status": prioritypassmodel.StatusSoldOut.String(),
		},
		Cause: eventCauseSlotSoldOut,
	}
	if diff := cmp.Diff(wantEvent, eventHelper(t, client, secondID, eventCauseSlotSoldOut)); diff != "" {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) のイベントの差分 (-want +got):\n%s",
			secondID,
			allocatedAt,
			diff,
		)
	}
}

// TestRepositoryAllocateTimeSlotConcurrent は同時に処理しても上限を超えて配らないことを確かめる
//
// 同じ枠を読んだトランザクションは衝突して再試行されるため、残りは負にならず遷移も上限までに収まる
func TestRepositoryAllocateTimeSlotConcurrent(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const (
		timeSlotID = prioritypassmodel.TimeSlotID("20261005_1500")
		capacity   = 1
		passes     = 3
	)
	ids := []prioritypassmodel.PassID{
		"pass-concurrent-1",
		"pass-concurrent-2",
		"pass-concurrent-3",
	}

	storeTimeSlotHelper(t, client, timeSlotID, capacity)
	for _, id := range ids {
		cleanupPriorityPassHelper(t, client, id)
		createPriorityPassHelper(t, repo, id, timeSlotID)
	}

	statuses := make([]prioritypassmodel.Status, passes)
	errs := make([]error, passes)

	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Go(func() {
			allocation, err := repo.AllocateTimeSlot(ctx, id, allocatedAt)
			if err != nil {
				errs[i] = err
				return
			}
			statuses[i] = allocation.Pass.Status()
		})
	}
	wg.Wait()

	issued := 0
	for i, id := range ids {
		// goroutine の中では失敗を伝えられないため、ここで検査する
		if errs[i] != nil {
			t.Fatalf("Repository.AllocateTimeSlot(%q, %v) = %v, want nil",
				id,
				allocatedAt,
				errs[i],
			)
		}

		if statuses[i] == prioritypassmodel.StatusIssued {
			issued++
		}
	}

	if issued != capacity {
		t.Errorf("同時に %d 件の Repository.AllocateTimeSlot() を行った後の issued の件数 = %d, want %d",
			passes,
			issued,
			capacity,
		)
	}

	if remaining := remainingHelper(t, client, timeSlotID); remaining != 0 {
		t.Errorf("同時に %d 件の Repository.AllocateTimeSlot() を行った後の残り = %d, want %d",
			passes,
			remaining,
			0,
		)
	}
}

func TestRepositoryAllocateTimeSlotNotFound(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	const id = prioritypassmodel.PassID("pass-allocate-not-found")

	_, err := repo.AllocateTimeSlot(t.Context(), id, allocatedAt)
	if !errors.Is(err, prioritypassmodel.ErrPriorityPassNotFound) {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) = %v, want %v",
			id,
			allocatedAt,
			err,
			prioritypassmodel.ErrPriorityPassNotFound,
		)
	}
}

// TestRepositoryAllocateTimeSlotTimeSlotNotFound は枠が無い場合に売り切れへ畳まず、遷移も起きないことを確かめる
func TestRepositoryAllocateTimeSlotTimeSlotNotFound(t *testing.T) {
	repo, client := newRepositoryHelper(t)
	ctx := t.Context()

	const (
		id         = prioritypassmodel.PassID("pass-allocate-slot-not-found")
		timeSlotID = prioritypassmodel.TimeSlotID("20261005_1400")
	)
	cleanupPriorityPassHelper(t, client, id)
	createPriorityPassHelper(t, repo, id, timeSlotID)

	_, err := repo.AllocateTimeSlot(ctx, id, allocatedAt)
	if !errors.Is(err, prioritypassmodel.ErrTimeSlotNotFound) {
		t.Fatalf("Repository.AllocateTimeSlot(%q, %v) = %v, want %v",
			id,
			allocatedAt,
			err,
			prioritypassmodel.ErrTimeSlotNotFound,
		)
	}

	if status := documentHelper(t, client, id).Status; status != prioritypassmodel.StatusRequested.String() {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) の後の状態 = %q, want %q",
			id,
			allocatedAt,
			status,
			prioritypassmodel.StatusRequested,
		)
	}

	// 作成の 1 件だけが残る
	if count := eventCountHelper(t, client, id); count != 1 {
		t.Errorf("Repository.AllocateTimeSlot(%q, %v) の後のイベントの件数 = %d, want %d",
			id,
			allocatedAt,
			count,
			1,
		)
	}
}

// createPriorityPassHelper は割当の対象になる申込を保存する
//
// 枠をテストごとに分けるため、時間帯枠の識別子を受け取る
func createPriorityPassHelper(
	tb testing.TB,
	repo *Repository,
	id prioritypassmodel.PassID,
	timeSlotID prioritypassmodel.TimeSlotID,
) {
	tb.Helper()

	pass, err := prioritypassmodel.RestorePriorityPass(
		id,
		testParkID,
		testTicketID,
		testAttractionID,
		timeSlotID,
		prioritypassmodel.StatusRequested,
		testNow,
		testNow,
	)
	if err != nil {
		tb.Fatalf("RestorePriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}

	if err := repo.CreatePriorityPass(tb.Context(), pass); err != nil {
		tb.Fatalf("Repository.CreatePriorityPass(%q) = %v, want nil",
			id,
			err,
		)
	}
}

// storeTimeSlotHelper は inventory が作成する時間帯枠を用意し、後始末を登録する
func storeTimeSlotHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	timeSlotID prioritypassmodel.TimeSlotID,
	remaining int32,
) {
	tb.Helper()

	data := storedTimeSlotDocument{
		AttractionID: testAttractionID.String(),
		Date:         "2026-10-05",
		StartTime:    "11:00",
		Capacity:     remaining,
		Remaining:    remaining,
	}

	doc := timeSlotDocHelper(client, timeSlotID)
	if _, err := doc.Create(tb.Context(), data); err != nil {
		tb.Fatalf("Create(%q) = %v, want nil",
			timeSlotID,
			err,
		)
	}

	tb.Cleanup(func() {
		// t.Context() は Cleanup の直前に取り消されるため、後始末は取り消されないものを使う
		if _, err := doc.Delete(context.Background()); err != nil {
			tb.Errorf("Delete(%q) = %v, want nil",
				timeSlotID,
				err,
			)
		}
	})
}

// remainingHelper は保存されている時間帯枠の残りを読む
func remainingHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	timeSlotID prioritypassmodel.TimeSlotID,
) int32 {
	tb.Helper()

	snapshot, err := timeSlotDocHelper(client, timeSlotID).Get(tb.Context())
	if err != nil {
		tb.Fatalf("Get(%q) = %v, want nil",
			timeSlotID,
			err,
		)
	}

	var doc storedTimeSlotDocument
	if err := snapshot.DataTo(&doc); err != nil {
		tb.Fatalf("DataTo(%q) = %v, want nil",
			timeSlotID,
			err,
		)
	}

	return doc.Remaining
}

// timeSlotDocHelper は実装と同じパスを組み立てる
func timeSlotDocHelper(
	client *gcpfirestore.Client,
	timeSlotID prioritypassmodel.TimeSlotID,
) *gcpfirestore.DocumentRef {
	return client.Collection(parkCollection).
		Doc(testParkID.String()).
		Collection(attractionCollection).
		Doc(testAttractionID.String()).
		Collection(timeSlotCollection).
		Doc(timeSlotID.String())
}

// documentHelper は保存されている優先パスを読む
func documentHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	id prioritypassmodel.PassID,
) document {
	tb.Helper()

	snapshot, err := client.Collection(collection).Doc(id.String()).Get(tb.Context())
	if err != nil {
		tb.Fatalf("Get(%q) = %v, want nil",
			id,
			err,
		)
	}

	var doc document
	if err := snapshot.DataTo(&doc); err != nil {
		tb.Fatalf("DataTo(%q) = %v, want nil",
			id,
			err,
		)
	}

	return doc
}

// eventHelper は原因で 1 件に絞ってイベントを読む
func eventHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	id prioritypassmodel.PassID,
	cause string,
) eventDocument {
	tb.Helper()

	snapshots, err := client.Collection(collection).
		Doc(id.String()).
		Collection(eventCollection).
		Where("cause", "==", cause).
		Documents(tb.Context()).
		GetAll()
	if err != nil {
		tb.Fatalf("GetAll(%q の events) = %v, want nil",
			id,
			err,
		)
	}

	if len(snapshots) != 1 {
		tb.Fatalf("GetAll(%q の %q のイベント) の件数 = %d, want %d",
			id,
			cause,
			len(snapshots),
			1,
		)
	}

	var event eventDocument
	if err := snapshots[0].DataTo(&event); err != nil {
		tb.Fatalf("DataTo(%q の events) = %v, want nil",
			id,
			err,
		)
	}

	return event
}

// eventCountHelper は残っているイベントの件数を返す
func eventCountHelper(
	tb testing.TB,
	client *gcpfirestore.Client,
	id prioritypassmodel.PassID,
) int {
	tb.Helper()

	snapshots, err := client.Collection(collection).
		Doc(id.String()).
		Collection(eventCollection).
		Documents(tb.Context()).
		GetAll()
	if err != nil {
		tb.Fatalf("GetAll(%q の events) = %v, want nil",
			id,
			err,
		)
	}

	return len(snapshots)
}

// TestAllocationEventValues は events に残す値を固定する
//
// 他のテストは定数どうしを比べるため、値そのものを変更したことに気づけるようにする
// 運用がログと突き合わせる値のため、変更すると過去の記録と意味が揃わなくなる
func TestAllocationEventValues(t *testing.T) {
	tests := map[string]struct {
		got  string
		want string
	}{
		"正常系_actorがpriority-pass-issuerであること": {
			got:  eventActorIssuer,
			want: "system:priority-pass-issuer",
		},
		"正常系_actionがstatus_changedであること": {
			got:  eventActionStatusChanged,
			want: "status_changed",
		},
		"正常系_割当のcauseがslot_allocatedであること": {
			got:  eventCauseSlotAllocated,
			want: "slot_allocated",
		},
		"正常系_売り切れのcauseがslot_sold_outであること": {
			got:  eventCauseSlotSoldOut,
			want: "slot_sold_out",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("events に残す値 = %q, want %q",
					tt.got,
					tt.want,
				)
			}
		})
	}
}

// TestRepositoryTimeSlotDocPath は時間帯枠のパスをリテラルで固定する
//
// 同じパスを inventory も持つため、定数を書き換えても他のテストは追従してしまう
// 書き換えに気づけるよう、ここだけは組み立てた結果をリテラルと突き合わせる
func TestRepositoryTimeSlotDocPath(t *testing.T) {
	repo, _ := newRepositoryHelper(t)

	const (
		parkID       = prioritypassmodel.ParkID("park-9")
		attractionID = prioritypassmodel.AttractionID("attraction-9")
		timeSlotID   = prioritypassmodel.TimeSlotID("20261005_1600")
	)

	got := repo.timeSlotDoc(
		parkID,
		attractionID,
		timeSlotID,
	)

	want := "parks/park-9/attractions/attraction-9/timeSlots/20261005_1600"
	if !strings.HasSuffix(got.Path, "/documents/"+want) {
		t.Errorf("Repository.timeSlotDoc(%q, %q, %q) のパス = %q, want 末尾が %q",
			parkID,
			attractionID,
			timeSlotID,
			got.Path,
			want,
		)
	}
}
