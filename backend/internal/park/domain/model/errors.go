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

var ErrAttractionNotFound = apperr.New(apperr.KindNotFound, "PARK_ATTRACTION_NOT_FOUND", "アトラクションが見つからない")

var ErrAttractionAlreadyExists = apperr.New(apperr.KindConflict, "PARK_ATTRACTION_ALREADY_EXISTS", "アトラクションが既に存在する")

var ErrAttractionInvalidID = apperr.New(apperr.KindInvalidArgument, "PARK_ATTRACTION_INVALID_ID", "アトラクションの識別子が空")

var ErrAttractionInvalidName = apperr.New(apperr.KindInvalidArgument, "PARK_ATTRACTION_INVALID_NAME", "アトラクションの表示名が 1 文字から 100 文字の範囲にない")

var ErrAttractionInvalidTime = apperr.New(apperr.KindInvalidArgument, "PARK_ATTRACTION_INVALID_TIME", "優先パスの時刻が HH:MM の 24 時間表記でない")

var ErrAttractionInvalidTimeRange = apperr.New(apperr.KindInvalidArgument, "PARK_ATTRACTION_INVALID_TIME_RANGE", "優先パスの終了時刻が開始時刻より後にない")

var ErrAttractionInvalidIntervalMinutes = apperr.New(apperr.KindInvalidArgument, "PARK_ATTRACTION_INVALID_INTERVAL_MINUTES", "優先パスの時間帯枠を刻む間隔が 1 分未満")

var ErrAttractionInvalidCapacityPerSlot = apperr.New(apperr.KindInvalidArgument, "PARK_ATTRACTION_INVALID_CAPACITY_PER_SLOT", "優先パスの時間帯枠 1 つあたりの上限が 1 枚未満")

var ErrTicketTypeNotFound = apperr.New(apperr.KindNotFound, "PARK_TICKET_TYPE_NOT_FOUND", "券種が見つからない")

var ErrTicketTypeAlreadyExists = apperr.New(apperr.KindConflict, "PARK_TICKET_TYPE_ALREADY_EXISTS", "券種が既に存在する")

var ErrTicketTypeInvalidID = apperr.New(apperr.KindInvalidArgument, "PARK_TICKET_TYPE_INVALID_ID", "券種の識別子が空")

var ErrTicketTypeInvalidName = apperr.New(apperr.KindInvalidArgument, "PARK_TICKET_TYPE_INVALID_NAME", "券種の表示名が 1 文字から 100 文字の範囲にない")

var ErrTicketTypeInvalidPrice = apperr.New(apperr.KindInvalidArgument, "PARK_TICKET_TYPE_INVALID_PRICE", "券種の価格が 0 円未満")

var ErrTicketTypeInvalidEntryTime = apperr.New(apperr.KindInvalidArgument, "PARK_TICKET_TYPE_INVALID_ENTRY_TIME", "入場できる時間帯の時刻が HH:MM の 24 時間表記でない")

var ErrTicketTypeInvalidEntryTimeRange = apperr.New(apperr.KindInvalidArgument, "PARK_TICKET_TYPE_INVALID_ENTRY_TIME_RANGE", "入場できる時間帯の終了時刻が開始時刻より後にない")
