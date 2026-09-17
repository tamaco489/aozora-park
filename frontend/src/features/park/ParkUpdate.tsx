import { useState, type SubmitEvent } from "react";

import { parkClient } from "../../api/park";
import { messageOf } from "../../api/errors";
import type { Park } from "../../gen/aozorapark/park/v1/park_pb";
import { ParkDetail } from "./ParkDetail";
import { ParkFields, type ParkFieldValues } from "./ParkFields";

const initial: ParkFieldValues = {
  name: "",
  defaultDailyCapacity: 5000,
  inventoryDays: 60,
};

export function ParkUpdate() {
  const [parkId, setParkId] = useState("");
  const [values, setValues] = useState(initial);
  const [park, setPark] = useState<Park | undefined>(undefined);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // UpdatePark は部分更新ではないため、変更しない項目も現在の値を送る必要がある
  async function handleLoad() {
    setLoading(true);
    setPark(undefined);
    setError("");

    try {
      const res = await parkClient.getPark({ parkId });

      if (res.park) {
        setValues({
          name: res.park.name,
          defaultDailyCapacity: res.park.defaultDailyCapacity,
          inventoryDays: res.park.inventoryDays,
        });
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
    setPark(undefined);
    setError("");

    try {
      const res = await parkClient.updatePark({ parkId, ...values });
      setPark(res.park);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h2>更新する</h2>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="update-parkId">識別子</label>
          <input
            id="update-parkId"
            className="narrow"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
          <button
            type="button"
            onClick={handleLoad}
            disabled={loading || parkId === ""}
          >
            現在の値を読み込む
          </button>
        </div>

        <ParkFields idPrefix="update" values={values} onChange={setValues} />

        <button type="submit" disabled={loading || parkId === ""}>
          {loading ? "更新中" : "更新"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {park && <ParkDetail park={park} />}
    </section>
  );
}
