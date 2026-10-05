import type { Attraction } from "../../gen/aozorapark/park/v1/park_pb";

// AttractionDetail は登録・更新の結果を同じ形で見せる
export function AttractionDetail({ attraction }: { attraction: Attraction }) {
  const config = attraction.priorityPassConfig;

  return (
    <dl>
      <dt>識別子</dt>
      <dd>{attraction.attractionId}</dd>
      <dt>パークの識別子</dt>
      <dd>{attraction.parkId}</dd>
      <dt>表示名</dt>
      <dd>{attraction.name}</dd>
      <dt>優先パスの対象</dt>
      <dd>{config?.enabled === true ? "対象" : "対象外"}</dd>
      <dt>枠の時間帯</dt>
      <dd>
        {config?.startTime} - {config?.endTime}
      </dd>
      <dt>枠を刻む間隔</dt>
      <dd>{config?.intervalMinutes} 分</dd>
      <dt>枠あたりの上限枚数</dt>
      <dd>{config?.capacityPerSlot} 枚</dd>
    </dl>
  );
}
