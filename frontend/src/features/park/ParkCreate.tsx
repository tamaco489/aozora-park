import { useState, type SubmitEvent } from "react";

import { parkClient } from "../../api/park";
import { messageOf } from "../../api/errors";
import type { Park } from "../../gen/aozorapark/park/v1/park_pb";
import { ParkDetail } from "./ParkDetail";
import { ParkFields, type ParkFieldValues } from "./ParkFields";

// 識別子はサーバが UUID v7 で採番するため、入力欄を持たない
const initial: ParkFieldValues = {
  name: "",
  defaultDailyCapacity: 5000,
  inventoryDays: 60,
};

export function ParkCreate() {
  const [values, setValues] = useState(initial);
  const [park, setPark] = useState<Park | undefined>(undefined);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setPark(undefined);
    setError("");

    try {
      const res = await parkClient.createPark(values);
      setPark(res.park);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h2>登録する</h2>

      <form onSubmit={handleSubmit}>
        <ParkFields idPrefix="create" values={values} onChange={setValues} />

        <button type="submit" disabled={loading}>
          {loading ? "登録中" : "登録"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {park && <ParkDetail park={park} />}
    </section>
  );
}
