// TicketTypeFields は登録と更新で共通の 4 項目
//
// 入力の制約は proto の protovalidate が正、ここでは型と最小限の入力補助だけを持たせる
export type TicketTypeFieldValues = {
  name: string;
  price: number;
  entryTimeFrom: string;
  entryTimeTo: string;
};

type Props = {
  idPrefix: string;
  values: TicketTypeFieldValues;
  onChange: (values: TicketTypeFieldValues) => void;
};

export function TicketTypeFields({ idPrefix, values, onChange }: Props) {
  return (
    <>
      <div className="field">
        <label htmlFor={`${idPrefix}-name`}>表示名</label>
        <input
          id={`${idPrefix}-name`}
          value={values.name}
          onChange={(e) => onChange({ ...values, name: e.target.value })}
          placeholder="1 日券 おとな"
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-price`}>販売価格 (円)</label>
        <input
          id={`${idPrefix}-price`}
          type="number"
          step="1"
          value={values.price}
          onChange={(e) => onChange({ ...values, price: Number(e.target.value) })}
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-entryTimeFrom`}>入場できる開始時刻</label>
        <input
          id={`${idPrefix}-entryTimeFrom`}
          value={values.entryTimeFrom}
          onChange={(e) =>
            onChange({ ...values, entryTimeFrom: e.target.value })
          }
          placeholder="09:00"
        />
      </div>

      <div className="field">
        <label htmlFor={`${idPrefix}-entryTimeTo`}>入場できる終了時刻</label>
        <input
          id={`${idPrefix}-entryTimeTo`}
          value={values.entryTimeTo}
          onChange={(e) => onChange({ ...values, entryTimeTo: e.target.value })}
          placeholder="18:00"
        />
      </div>
    </>
  );
}
