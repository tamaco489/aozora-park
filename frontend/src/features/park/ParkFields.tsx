// ParkFields は登録と更新で共通の 3 項目
//
// 入力の制約は proto の protovalidate が正、ここでは型と最小限の入力補助だけを持たせる
export type ParkFieldValues = {
  name: string;
  defaultDailyCapacity: number;
  inventoryDays: number;
};

type Props = {
  idPrefix: string;
  values: ParkFieldValues;
  onChange: (values: ParkFieldValues) => void;
};

export function ParkFields({ idPrefix, values, onChange }: Props) {
  return (
    <>
      <div className="field">
        <label htmlFor={`${idPrefix}-name`}>表示名</label>
        <input
          id={`${idPrefix}-name`}
          value={values.name}
          onChange={(e) => onChange({ ...values, name: e.target.value })}
          placeholder="あおぞらパーク 本園"
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-capacity`}>1 日あたりの上限人数</label>
        <input
          id={`${idPrefix}-capacity`}
          type="number"
          value={values.defaultDailyCapacity}
          onChange={(e) =>
            onChange({
              ...values,
              defaultDailyCapacity: Number(e.target.value),
            })
          }
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-days`}>枠を生成する日数</label>
        <input
          id={`${idPrefix}-days`}
          type="number"
          value={values.inventoryDays}
          onChange={(e) =>
            onChange({ ...values, inventoryDays: Number(e.target.value) })
          }
        />
      </div>
    </>
  );
}
