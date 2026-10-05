import { useState, type SubmitEvent } from "react";

import { parkClient } from "../../api/park";
import { messageOf } from "../../api/errors";
import type { Attraction } from "../../gen/aozorapark/park/v1/park_pb";
import { AttractionDetail } from "./AttractionDetail";
import {
  AttractionFields,
  type AttractionFieldValues,
} from "./AttractionFields";

const initial: AttractionFieldValues = {
  name: "",
  enabled: true,
  startTime: "09:00",
  endTime: "18:00",
  intervalMinutes: 30,
  capacityPerSlot: 50,
};

// UpdateAttraction は部分更新ではないため全項目を送る
// 取得の RPC がまだ無く現在の値を読み込めないため、変更しない項目も手で埋める
export function AttractionUpdate() {
  const [parkId, setParkId] = useState("");
  const [attractionId, setAttractionId] = useState("");
  const [values, setValues] = useState(initial);
  const [attraction, setAttraction] = useState<Attraction | undefined>(
    undefined,
  );
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setAttraction(undefined);
    setError("");

    try {
      const res = await parkClient.updateAttraction({
        parkId,
        attractionId,
        ...values,
      });
      setAttraction(res.attraction);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h3>更新する</h3>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="attraction-update-parkId">パークの識別子</label>
          <input
            id="attraction-update-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="attraction-update-attractionId">識別子</label>
          <input
            id="attraction-update-attractionId"
            value={attractionId}
            onChange={(e) => setAttractionId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <AttractionFields
          idPrefix="attraction-update"
          values={values}
          onChange={setValues}
        />

        <button
          type="submit"
          disabled={loading || parkId === "" || attractionId === ""}
        >
          {loading ? "更新中" : "更新"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {attraction && <AttractionDetail attraction={attraction} />}
    </section>
  );
}
