import { useState, type SubmitEvent } from "react";

import { parkClient } from "../../api/park";
import { messageOf } from "../../api/errors";
import type { Attraction } from "../../gen/aozorapark/park/v1/park_pb";
import { AttractionDetail } from "./AttractionDetail";
import {
  AttractionFields,
  type AttractionFieldValues,
} from "./AttractionFields";

// 識別子はサーバが UUID v7 で採番するため、入力欄を持たない
const initial: AttractionFieldValues = {
  name: "",
  enabled: true,
  startTime: "09:00",
  endTime: "18:00",
  intervalMinutes: 30,
  capacityPerSlot: 50,
};

export function AttractionCreate() {
  const [parkId, setParkId] = useState("");
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
      const res = await parkClient.createAttraction({ parkId, ...values });
      setAttraction(res.attraction);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h3>登録する</h3>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="attraction-create-parkId">パークの識別子</label>
          <input
            id="attraction-create-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <AttractionFields
          idPrefix="attraction-create"
          values={values}
          onChange={setValues}
        />

        <button type="submit" disabled={loading || parkId === ""}>
          {loading ? "登録中" : "登録"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {attraction && <AttractionDetail attraction={attraction} />}
    </section>
  );
}
