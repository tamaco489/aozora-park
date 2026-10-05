// AttractionFields は登録と更新で共通の 5 項目
//
// 優先パスの条件は request では入れ子にせず平たく並ぶため、ここでも同じ形で持つ
// 入力の制約は proto の protovalidate が正、ここでは型と最小限の入力補助だけを持たせる
export type AttractionFieldValues = {
  name: string;
  enabled: boolean;
  startTime: string;
  endTime: string;
  intervalMinutes: number;
  capacityPerSlot: number;
};

type Props = {
  idPrefix: string;
  values: AttractionFieldValues;
  onChange: (values: AttractionFieldValues) => void;
};

export function AttractionFields({ idPrefix, values, onChange }: Props) {
  return (
    <>
      <div className="field">
        <label htmlFor={`${idPrefix}-name`}>表示名</label>
        <input
          id={`${idPrefix}-name`}
          value={values.name}
          onChange={(e) => onChange({ ...values, name: e.target.value })}
          placeholder="そらとびコースター"
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-enabled`}>優先パスの対象</label>
        <input
          id={`${idPrefix}-enabled`}
          type="checkbox"
          checked={values.enabled}
          onChange={(e) => onChange({ ...values, enabled: e.target.checked })}
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-startTime`}>枠の開始時刻</label>
        <input
          id={`${idPrefix}-startTime`}
          value={values.startTime}
          onChange={(e) => onChange({ ...values, startTime: e.target.value })}
          placeholder="09:00"
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-endTime`}>枠の終了時刻</label>
        <input
          id={`${idPrefix}-endTime`}
          value={values.endTime}
          onChange={(e) => onChange({ ...values, endTime: e.target.value })}
          placeholder="18:00"
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-intervalMinutes`}>枠を刻む間隔 (分)</label>
        <input
          id={`${idPrefix}-intervalMinutes`}
          type="number"
          value={values.intervalMinutes}
          onChange={(e) =>
            onChange({ ...values, intervalMinutes: Number(e.target.value) })
          }
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-capacityPerSlot`}>枠あたりの上限枚数</label>
        <input
          id={`${idPrefix}-capacityPerSlot`}
          type="number"
          value={values.capacityPerSlot}
          onChange={(e) =>
            onChange({ ...values, capacityPerSlot: Number(e.target.value) })
          }
        />
      </div>
    </>
  );
}
