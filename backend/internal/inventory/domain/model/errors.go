package model

import (
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

var ErrDateInventoryNotFound = apperr.New(
	apperr.KindNotFound,
	"INVENTORY_DATE_INVENTORY_NOT_FOUND",
	"その日の入場枠が見つからない",
)

var ErrInvalidParkID = apperr.New(
	apperr.KindInvalidArgument,
	"INVENTORY_INVALID_PARK_ID",
	"パークの識別子が空",
)

var ErrInvalidAttractionID = apperr.New(
	apperr.KindInvalidArgument,
	"INVENTORY_INVALID_ATTRACTION_ID",
	"アトラクションの識別子が空",
)

var ErrInvalidTimeSlotID = apperr.New(
	apperr.KindInvalidArgument,
	"INVENTORY_INVALID_TIME_SLOT_ID",
	"時間帯枠の識別子が空",
)

var ErrInvalidDate = apperr.New(
	apperr.KindInvalidArgument,
	"INVENTORY_INVALID_DATE",
	"枠の日付が YYYY-MM-DD の形式でない",
)

var ErrInvalidStartTime = apperr.New(
	apperr.KindInvalidArgument,
	"INVENTORY_INVALID_START_TIME",
	"枠の開始時刻が HH:MM の形式でない",
)

var ErrInvalidCapacity = apperr.New(
	apperr.KindInvalidArgument,
	"INVENTORY_INVALID_CAPACITY",
	"枠の上限が 1 未満",
)

var ErrInvalidRemaining = apperr.New(
	apperr.KindInvalidArgument,
	"INVENTORY_INVALID_REMAINING",
	"枠の残りが 0 未満か上限を超えている",
)
