import type { Park } from "../../gen/aozorapark/park/v1/park_pb";

// ParkDetail は取得・登録・更新の結果を同じ形で見せる
export function ParkDetail({ park }: { park: Park }) {
  return (
    <dl>
      <dt>識別子</dt>
      <dd>{park.parkId}</dd>
      <dt>表示名</dt>
      <dd>{park.name}</dd>
      <dt>1 日あたりの上限人数</dt>
      <dd>{park.defaultDailyCapacity}</dd>
      <dt>枠を生成する日数</dt>
      <dd>{park.inventoryDays}</dd>
    </dl>
  );
}
