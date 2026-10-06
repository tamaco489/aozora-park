import { useState, type SubmitEvent } from "react";

import { parkClient } from "../../api/park";
import { messageOf } from "../../api/errors";
import type { TicketType } from "../../gen/aozorapark/park/v1/park_pb";
import { TicketTypeDetail } from "./TicketTypeDetail";
import {
  TicketTypeFields,
  type TicketTypeFieldValues,
} from "./TicketTypeFields";

const initial: TicketTypeFieldValues = {
  name: "",
  price: 6800,
  entryTimeFrom: "09:00",
  entryTimeTo: "18:00",
};

// UpdateTicketType は部分更新ではないため全項目を送る
// 取得の RPC がまだ無く現在の値を読み込めないため、変更しない項目も手で埋める
export function TicketTypeUpdate() {
  const [parkId, setParkId] = useState("");
  const [ticketTypeId, setTicketTypeId] = useState("");
  const [values, setValues] = useState(initial);
  const [ticketType, setTicketType] = useState<TicketType | undefined>(
    undefined,
  );
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setTicketType(undefined);
    setError("");

    try {
      const res = await parkClient.updateTicketType({
        parkId,
        ticketTypeId,
        ...values,
        // price は int64 のため bigint で渡す、円に小数は無いので切り捨てる
        price: BigInt(Math.trunc(values.price)),
      });
      setTicketType(res.ticketType);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h3>更新する</h3>
      <p className="hint">
        パークと券種の識別子が要る。
        部分更新ではないため、変更しない項目も現在の値を入れて送る
      </p>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="tickettype-update-parkId">パークの識別子</label>
          <input
            id="tickettype-update-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <div className="field">
          <label htmlFor="tickettype-update-ticketTypeId">
            券種の識別子
          </label>
          <input
            id="tickettype-update-ticketTypeId"
            value={ticketTypeId}
            onChange={(e) => setTicketTypeId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <TicketTypeFields
          idPrefix="tickettype-update"
          values={values}
          onChange={setValues}
        />

        <button
          type="submit"
          disabled={loading || parkId === "" || ticketTypeId === ""}
        >
          {loading ? "更新中" : "更新"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {ticketType && <TicketTypeDetail ticketType={ticketType} />}
    </section>
  );
}
