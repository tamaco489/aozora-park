package model

// DateInventory はパークの特定の日の入場枠
type DateInventory struct {
	parkID    ParkID
	date      Date
	capacity  int32
	remaining int32
}

// RestoreDateInventory は保存済みの入場枠を組み立てる
//
// 枠の作成は専用のジョブが行うため、この機能パッケージは新規生成を持たない
// infrastructure が読み出した値を入れる、保存されている値も検証を通す
func RestoreDateInventory(
	parkID ParkID,
	date Date,
	capacity int32,
	remaining int32,
) (*DateInventory, error) {
	if parkID == "" {
		return nil, ErrInvalidParkID
	}

	if err := validateDate(date); err != nil {
		return nil, err
	}

	d := &DateInventory{
		parkID: parkID,
		date:   date,
	}
	if err := d.Overwrite(capacity, remaining); err != nil {
		return nil, err
	}

	return d, nil
}

// Overwrite は上限人数と残りの人数を運営の指定した値に差し替える
//
// 検証に失敗したときは元の値を保つ
func (d *DateInventory) Overwrite(capacity, remaining int32) error {
	if err := validateQuantity(capacity, remaining); err != nil {
		return err
	}

	d.capacity = capacity
	d.remaining = remaining

	return nil
}

func (d *DateInventory) ParkID() ParkID { return d.parkID }

func (d *DateInventory) Date() Date { return d.date }

func (d *DateInventory) Capacity() int32 { return d.capacity }

func (d *DateInventory) Remaining() int32 { return d.remaining }
