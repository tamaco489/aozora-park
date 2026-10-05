import { useState, type SubmitEvent } from "react";

import { inventoryClient } from "../../api/inventory";
import { messageOf } from "../../api/errors";
import type { DateInventory } from "../../gen/aozorapark/inventory/v1/inventory_pb";
import { DateInventoryDetail } from "./DateInventoryDetail";

// 枠は job generate が先に作る、ここでは作らず参照するだけにする
export function DateInventoryView() {
  const [parkId, setParkId] = useState("");
  const [date, setDate] = useState("");
  const [dateInventory, setDateInventory] = useState<
    DateInventory | undefined
  >(undefined);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setDateInventory(undefined);
    setError("");

    try {
      const res = await inventoryClient.getDateInventory({ parkId, date });
      setDateInventory(res.dateInventory);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h3>入場枠を参照する</h3>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="dateinventory-view-parkId">パークの識別子</label>
          <input
            id="dateinventory-view-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="dateinventory-view-date">日付</label>
          <input
            id="dateinventory-view-date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </div>

        <button type="submit" disabled={loading || parkId === "" || date === ""}>
          {loading ? "取得中" : "取得"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {dateInventory && <DateInventoryDetail dateInventory={dateInventory} />}
    </section>
  );
}
