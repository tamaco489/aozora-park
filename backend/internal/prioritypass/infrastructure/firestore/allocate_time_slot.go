package firestore

import (
	"context"
	"errors"
	"fmt"
	"time"

	gcpfirestore "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// timeSlotDocument は時間帯枠の残りを読む形
//
// inventory が保存したドキュメントを読む、減算に使わない項目は持たない (DataTo は構造体にない項目を捨てる)
type timeSlotDocument struct {
	Remaining int32 `firestore:"remaining"`
}

// 割当で書き換えるフィールド (document と timeSlotDocument の firestore タグと同じ値にする)
const (
	fieldStatus    = "status"
	fieldUpdatedAt = "updatedAt"
	fieldRemaining = "remaining"
)

// 割当の記録に使う値
const (
	eventActorIssuer         = "system:priority-pass-issuer"
	eventActionStatusChanged = "status_changed"
	eventCauseSlotAllocated  = "slot_allocated"
	eventCauseSlotSoldOut    = "slot_sold_out"
)

// AllocateTimeSlot は時間帯枠の残りを 1 つ減らして申込を issued にする
//
//   - 枠の減算と状態の遷移が片方だけ残ると枠を二重に配ることになるため、集約をまたいで 1 つのトランザクションで書く
//   - 残りが無いときは減算せずに sold_out へ遷移させる
func (r *Repository) AllocateTimeSlot(
	ctx context.Context,
	id prioritypassmodel.PassID,
	now time.Time,
) (prioritypassmodel.Allocation, error) {
	var allocation prioritypassmodel.Allocation

	err := r.client.RunTransaction(ctx, func(_ context.Context, tx *gcpfirestore.Transaction) error {
		// 衝突して再試行されたときに前回の結果が残らないようにする
		allocation = prioritypassmodel.Allocation{}

		passDoc := r.doc(id)

		pass, err := getPriorityPassInTransaction(tx, passDoc, id)
		if err != nil {
			return err
		}

		// 再配信で終端まで進んだものが届くため、遷移できない申込には何も書かない
		if !pass.IsRequested() {
			allocation = prioritypassmodel.Allocation{Pass: pass}

			return nil
		}

		slotDoc := r.timeSlotDoc(
			pass.ParkID(),
			pass.AttractionID(),
			pass.TimeSlotID(),
		)

		remaining, err := getRemainingInTransaction(tx, slotDoc, pass.TimeSlotID())
		if err != nil {
			return err
		}

		// Firestore のトランザクションは読んだドキュメントが変更されると再試行されるため、読んだ値から引けば残りは負にならない
		var cause string
		if remaining > 0 {
			if err := pass.Issue(now); err != nil {
				return err
			}

			decrement := []gcpfirestore.Update{
				{
					Path:  fieldRemaining,
					Value: remaining - 1,
				},
			}
			if err := tx.Update(slotDoc, decrement); err != nil {
				return err
			}

			cause = eventCauseSlotAllocated
		} else {
			if err := pass.MarkSoldOut(now); err != nil {
				return err
			}

			cause = eventCauseSlotSoldOut
		}

		updates := []gcpfirestore.Update{
			{
				Path:  fieldStatus,
				Value: pass.Status().String(),
			},
			{
				Path:  fieldUpdatedAt,
				Value: pass.UpdatedAt(),
			},
		}
		if err := tx.Update(passDoc, updates); err != nil {
			return err
		}

		event := toAllocatedEvent(pass, cause)
		if err := tx.Create(passDoc.Collection(eventCollection).NewDoc(), event); err != nil {
			return err
		}

		allocation = prioritypassmodel.Allocation{
			Pass:    pass,
			Changed: true,
		}

		return nil
	})
	if err != nil {
		// 翻訳済みのセンチネルは呼び出し側が分類に使うため、文脈を足さずに返す
		if errors.Is(err, prioritypassmodel.ErrPriorityPassNotFound) ||
			errors.Is(err, prioritypassmodel.ErrTimeSlotNotFound) {
			return prioritypassmodel.Allocation{}, err
		}

		return prioritypassmodel.Allocation{}, fmt.Errorf("allocate time slot for priority pass %q: %w",
			id,
			err,
		)
	}

	return allocation, nil
}

func getPriorityPassInTransaction(
	tx *gcpfirestore.Transaction,
	doc *gcpfirestore.DocumentRef,
	id prioritypassmodel.PassID,
) (*prioritypassmodel.PriorityPass, error) {
	snapshot, err := tx.Get(doc)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, prioritypassmodel.ErrPriorityPassNotFound
		}

		return nil, err
	}

	return restoreFromSnapshot(snapshot, id)
}

// getRemainingInTransaction は時間帯枠の残りを読む
//
// 枠が無いのは生成のジョブが追いついていない場合のため、売り切れに畳まずエラーにする
func getRemainingInTransaction(
	tx *gcpfirestore.Transaction,
	doc *gcpfirestore.DocumentRef,
	timeSlotID prioritypassmodel.TimeSlotID,
) (int32, error) {
	snapshot, err := tx.Get(doc)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return 0, prioritypassmodel.ErrTimeSlotNotFound
		}

		return 0, err
	}

	var slot timeSlotDocument
	if err := snapshot.DataTo(&slot); err != nil {
		return 0, fmt.Errorf("decode time slot %q: %w",
			timeSlotID,
			err,
		)
	}

	return slot.Remaining, nil
}

func toAllocatedEvent(pass *prioritypassmodel.PriorityPass, cause string) eventDocument {
	return eventDocument{
		At:     pass.UpdatedAt(),
		Actor:  eventActorIssuer,
		Action: eventActionStatusChanged,
		Changes: map[string]string{
			"status": pass.Status().String(),
		},
		Cause: cause,
	}
}
