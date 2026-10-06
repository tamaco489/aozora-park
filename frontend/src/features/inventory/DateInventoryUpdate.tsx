import { useState, type SubmitEvent } from "react";

import { inventoryClient } from "../../api/inventory";
import { messageOf } from "../../api/errors";
import type { DateInventory } from "../../gen/aozorapark/inventory/v1/inventory_pb";
import { DateInventoryDetail } from "./DateInventoryDetail";

export function DateInventoryUpdate() {
  const [parkId, setParkId] = useState("");
  const [date, setDate] = useState("");
  const [capacity, setCapacity] = useState(0);
  const [remaining, setRemaining] = useState(0);
  const [dateInventory, setDateInventory] = useState<
    DateInventory | undefined
  >(undefined);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // UpdateDateInventory は部分更新ではないため、変更しない項目も現在の値を送る
  async function handleLoad() {
    setLoading(true);
    setDateInventory(undefined);
    setError("");

    try {
      const res = await inventoryClient.getDateInventory({ parkId, date });

      if (res.dateInventory) {
        setCapacity(res.dateInventory.capacity);
        setRemaining(res.dateInventory.remaining);
      }
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setDateInventory(undefined);
    setError("");

    try {
      const res = await inventoryClient.updateDateInventory({
        parkId,
        date,
        capacity,
        remaining,
      });
      setDateInventory(res.dateInventory);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  const identified = parkId !== "" && date !== "";

  return (
    <section>
      <h3>入場枠を上書きする</h3>
      <p className="hint">
        保存済みの枠だけを書き換える。残りの人数は 0 以上、上限人数以下にする
      </p>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="dateinventory-update-parkId">パークの識別子</label>
          <input
            id="dateinventory-update-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="dateinventory-update-date">日付</label>
          <input
            id="dateinventory-update-date"
            className="narrow"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
          <button type="button" onClick={handleLoad} disabled={loading || !identified}>
            現在の値を読み込む
          </button>
        </div>

        <div className="field">
          <label htmlFor="dateinventory-update-capacity">上限人数</label>
          <input
            id="dateinventory-update-capacity"
            type="number"
            value={capacity}
            onChange={(e) => setCapacity(Number(e.target.value))}
          />
        </div>

        <div className="field">
          <label htmlFor="dateinventory-update-remaining">残りの人数</label>
          <input
            id="dateinventory-update-remaining"
            type="number"
            value={remaining}
            onChange={(e) => setRemaining(Number(e.target.value))}
          />
        </div>

        <button type="submit" disabled={loading || !identified}>
          {loading ? "上書き中" : "上書き"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {dateInventory && <DateInventoryDetail dateInventory={dateInventory} />}
    </section>
  );
}
