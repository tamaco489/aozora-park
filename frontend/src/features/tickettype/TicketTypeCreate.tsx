import { useState, type SubmitEvent } from "react";

import { parkClient } from "../../api/park";
import { messageOf } from "../../api/errors";
import type { TicketType } from "../../gen/aozorapark/park/v1/park_pb";
import { TicketTypeDetail } from "./TicketTypeDetail";
import {
  TicketTypeFields,
  type TicketTypeFieldValues,
} from "./TicketTypeFields";

// 識別子はサーバが UUID v7 で採番するため、入力欄を持たない
const initial: TicketTypeFieldValues = {
  name: "",
  price: 6800,
  entryTimeFrom: "09:00",
  entryTimeTo: "18:00",
};

export function TicketTypeCreate() {
  const [parkId, setParkId] = useState("");
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
      const res = await parkClient.createTicketType({
        parkId,
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
      <h3>登録する</h3>
      <p className="hint">
        パークの識別子が要る。価格は円で、無料の券種のために 0 を許す
      </p>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="tickettype-create-parkId">パークの識別子</label>
          <input
            id="tickettype-create-parkId"
            value={parkId}
            onChange={(e) => setParkId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <TicketTypeFields
          idPrefix="tickettype-create"
          values={values}
          onChange={setValues}
        />

        <button type="submit" disabled={loading || parkId === ""}>
          {loading ? "登録中" : "登録"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {ticketType && <TicketTypeDetail ticketType={ticketType} />}
    </section>
  );
}
