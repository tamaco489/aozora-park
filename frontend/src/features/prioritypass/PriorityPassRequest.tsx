import { useState, type SubmitEvent } from "react";

import { priorityPassClient } from "../../api/prioritypass";
import { messageOf } from "../../api/errors";
import type { PriorityPass } from "../../gen/aozorapark/prioritypass/v1/prioritypass_pb";
import { PriorityPassDetail } from "./PriorityPassDetail";

// 券との突き合わせは行わないため、券の識別子は控えた値をそのまま入れる
export function PriorityPassRequest() {
  const [parkId, setParkId] = useState("");
  const [ticketId, setTicketId] = useState("");
  const [attractionId, setAttractionId] = useState("");
  const [timeSlotId, setTimeSlotId] = useState("");
  const [priorityPass, setPriorityPass] = useState<PriorityPass | undefined>(
    undefined,
  );
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setPriorityPass(undefined);
    setError("");

    try {
      const res = await priorityPassClient.requestPriorityPass({
        parkId,
        ticketId,
        attractionId,
        timeSlotId,
      });
      setPriorityPass(res.priorityPass);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h3>申し込む</h3>
      <p className="hint">
        申込は受け付けた時点で requested を返す。
        時間帯枠の割当は非同期に行うため、結果は下の状況の確認で見る。
        表示される優先パスの識別子を控える
      </p>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="prioritypass-request-parkId">パークの識別子</label>
          <input
            id="prioritypass-request-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="prioritypass-request-ticketId">券の識別子</label>
          <input
            id="prioritypass-request-ticketId"
            value={ticketId}
            onChange={(e) => setTicketId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="prioritypass-request-attractionId">
            アトラクションの識別子
          </label>
          <input
            id="prioritypass-request-attractionId"
            value={attractionId}
            onChange={(e) => setAttractionId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="prioritypass-request-timeSlotId">
            時間帯枠の識別子
          </label>
          <input
            id="prioritypass-request-timeSlotId"
            value={timeSlotId}
            onChange={(e) => setTimeSlotId(e.target.value)}
            placeholder="20260401_1030"
          />
        </div>

        <button
          type="submit"
          disabled={
            loading ||
            parkId === "" ||
            ticketId === "" ||
            attractionId === "" ||
            timeSlotId === ""
          }
        >
          {loading ? "申込中" : "申し込む"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {priorityPass && <PriorityPassDetail priorityPass={priorityPass} />}
    </section>
  );
}
