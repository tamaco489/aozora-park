import { useState } from "react";

import { parkClient } from "../../api/park";
import { messageOf } from "../../api/errors";
import type { Park } from "../../gen/aozorapark/park/v1/park_pb";
import { ParkDetail } from "./ParkDetail";

// 一覧の RPC が無いため、識別子を入力して 1 件だけ引く
export function ParkView() {
  const [parkId, setParkId] = useState("");
  const [park, setPark] = useState<Park | undefined>(undefined);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setPark(undefined);
    setError("");

    try {
      const res = await parkClient.getPark({ parkId });
      setPark(res.park);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h2>参照する</h2>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="view-parkId">識別子</label>
          <input
            id="view-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <button type="submit" disabled={loading || parkId === ""}>
          {loading ? "取得中" : "取得"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {park && <ParkDetail park={park} />}
    </section>
  );
}
