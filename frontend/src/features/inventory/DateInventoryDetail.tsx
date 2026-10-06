import type { DateInventory } from "../../gen/aozorapark/inventory/v1/inventory_pb";

// DateInventoryDetail は参照・上書きの結果を同じ形で見せる
export function DateInventoryDetail({
  dateInventory,
}: {
  dateInventory: DateInventory;
}) {
  return (
    <dl>
      <dt>パークの識別子</dt>
      <dd>{dateInventory.parkId}</dd>
      <dt>日付</dt>
      <dd>{dateInventory.date}</dd>
      <dt>上限人数</dt>
      <dd>{dateInventory.capacity} 人</dd>
      <dt>残りの人数</dt>
      <dd>{dateInventory.remaining} 人</dd>
    </dl>
  );
}
