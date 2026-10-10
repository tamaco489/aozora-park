import { useState, type SubmitEvent } from "react";

import { priorityPassClient } from "../../api/prioritypass";
import { messageOf } from "../../api/errors";
import type { PriorityPass } from "../../gen/aozorapark/prioritypass/v1/prioritypass_pb";
import { PriorityPassDetail } from "./PriorityPassDetail";

// 割当は worker が非同期に行うため、状態の変化はボタンでの再取得で確かめる
export function PriorityPassView() {
  const [passId, setPassId] = useState("");
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
      const res = await priorityPassClient.getPriorityPass({ passId });
      setPriorityPass(res.priorityPass);
    } catch (err) {
      setError(messageOf(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <h3>状況を確認する</h3>
      <p className="hint">
        控えた優先パスの識別子を入れて状態を引く。
        requested のままなら、少し待ってもう一度取得すると issued か sold_out に変わる
      </p>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="prioritypass-view-passId">優先パスの識別子</label>
          <input
            id="prioritypass-view-passId"
            value={passId}
            onChange={(e) => setPassId(e.target.value)}
            placeholder="01a09793-928d-7716-9725-5bb1173c6309"
          />
        </div>

        <button type="submit" disabled={loading || passId === ""}>
          {loading ? "取得中" : "取得"}
        </button>
      </form>

      {error !== "" && <p role="alert">{error}</p>}
      {priorityPass && <PriorityPassDetail priorityPass={priorityPass} />}
    </section>
  );
}
