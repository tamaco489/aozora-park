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
