import { useState, type SubmitEvent } from "react";

import { inventoryClient } from "../../api/inventory";
import { messageOf } from "../../api/errors";
import type { TimeSlot } from "../../gen/aozorapark/inventory/v1/inventory_pb";

// 時間帯枠はサーバが開始時刻の昇順で返すため、画面では並べ替えない
export function TimeSlotList() {
  const [parkId, setParkId] = useState("");
  const [attractionId, setAttractionId] = useState("");
  const [date, setDate] = useState("");
  const [timeSlots, setTimeSlots] = useState<TimeSlot[] | undefined>(undefined);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setTimeSlots(undefined);
    setError("");

    try {
      const res = await inventoryClient.listTimeSlots({
        parkId,
        attractionId,
        date,
      });
      setTimeSlots(res.timeSlots);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h3>時間帯枠の一覧を見る</h3>
      <p className="hint">
        優先パスの対象にしたアトラクションの枠を、開始時刻の昇順で出す
      </p>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="timeslot-list-parkId">パークの識別子</label>
          <input
            id="timeslot-list-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="timeslot-list-attractionId">
            アトラクションの識別子
          </label>
          <input
            id="timeslot-list-attractionId"
            value={attractionId}
            onChange={(e) => setAttractionId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="timeslot-list-date">日付</label>
          <input
            id="timeslot-list-date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </div>

        <button
          type="submit"
          disabled={
            loading || parkId === "" || attractionId === "" || date === ""
          }
        >
          {loading ? "取得中" : "取得"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}

      {timeSlots && timeSlots.length === 0 && <p>その日の時間帯枠はありません</p>}

      {timeSlots && timeSlots.length > 0 && (
        <table>
          <thead>
            <tr>
              <th scope="col">開始時刻</th>
              <th scope="col">上限枚数</th>
              <th scope="col">残りの枚数</th>
            </tr>
          </thead>
          <tbody>
            {timeSlots.map((timeSlot) => (
              <tr key={timeSlot.timeSlotId}>
                <td>{timeSlot.startTime}</td>
                <td>{timeSlot.capacity}</td>
                <td>{timeSlot.remaining}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
