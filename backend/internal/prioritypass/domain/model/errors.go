package model

import (
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

var ErrPriorityPassNotFound = apperr.New(
	apperr.KindNotFound,
	"PRIORITY_PASS_NOT_FOUND",
	"優先パスが見つからない",
)

var ErrPriorityPassAlreadyExists = apperr.New(
	apperr.KindConflict,
	"PRIORITY_PASS_ALREADY_EXISTS",
	"優先パスが既に存在する",
)

var ErrInvalidID = apperr.New(
	apperr.KindInvalidArgument,
	"PRIORITY_PASS_INVALID_ID",
	"優先パスの識別子が空",
)

var ErrInvalidParkID = apperr.New(
	apperr.KindInvalidArgument,
	"PRIORITY_PASS_INVALID_PARK_ID",
	"パークの識別子が空",
)

var ErrInvalidTicketID = apperr.New(
	apperr.KindInvalidArgument,
	"PRIORITY_PASS_INVALID_TICKET_ID",
	"券の識別子が空",
)

var ErrInvalidAttractionID = apperr.New(
	apperr.KindInvalidArgument,
	"PRIORITY_PASS_INVALID_ATTRACTION_ID",
	"アトラクションの識別子が空",
)

var ErrInvalidTimeSlotID = apperr.New(
	apperr.KindInvalidArgument,
	"PRIORITY_PASS_INVALID_TIME_SLOT_ID",
	"時間帯枠の識別子が YYYYMMDD_HHMM の形式でない",
)

var ErrInvalidStatus = apperr.New(
	apperr.KindInvalidArgument,
	"PRIORITY_PASS_INVALID_STATUS",
	"優先パスの状態が既知の値でない",
)

var ErrNotRequested = apperr.New(
	apperr.KindConflict,
	"PRIORITY_PASS_NOT_REQUESTED",
	"優先パスが申込中でない",
)

// 枠の生成が追いついていないだけの場合があるため、再実行で直りうる扱いにする
// 売り切れに畳むと、枠が空いているのに戻せない終端で止まる
var ErrTimeSlotNotFound = apperr.NewRetryable(
	apperr.KindNotFound,
	"PRIORITY_PASS_TIME_SLOT_NOT_FOUND",
	"申込が指す時間帯枠が見つからない",
)
