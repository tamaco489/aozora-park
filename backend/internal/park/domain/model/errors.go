package model

import (
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

var ErrNotFound = apperr.New(apperr.KindNotFound, "PARK_NOT_FOUND", "パークが見つからない")

var ErrAlreadyExists = apperr.New(apperr.KindConflict, "PARK_ALREADY_EXISTS", "パークが既に存在する")

var ErrInvalidID = apperr.New(apperr.KindInvalidArgument, "PARK_INVALID_ID", "パークの識別子が空")

var ErrInvalidName = apperr.New(apperr.KindInvalidArgument, "PARK_INVALID_NAME", "パークの表示名が 1 文字から 100 文字の範囲にない")

var ErrInvalidDailyCapacity = apperr.New(apperr.KindInvalidArgument, "PARK_INVALID_DAILY_CAPACITY", "1 日あたりの上限人数が 1 人未満")

var ErrInvalidInventoryDays = apperr.New(apperr.KindInvalidArgument, "PARK_INVALID_INVENTORY_DAYS", "枠を生成する日数が 1 日から 90 日の範囲にない")
