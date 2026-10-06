import type { TicketType } from "../../gen/aozorapark/park/v1/park_pb";

// TicketTypeDetail は登録・更新の結果を同じ形で見せる
export function TicketTypeDetail({ ticketType }: { ticketType: TicketType }) {
  return (
    <dl>
      <dt>パークの識別子</dt>
      <dd>{ticketType.parkId}</dd>
      <dt>券種の識別子</dt>
      <dd>{ticketType.ticketTypeId}</dd>
      <dt>表示名</dt>
      <dd>{ticketType.name}</dd>
      <dt>販売価格</dt>
      <dd>{ticketType.price.toString()} 円</dd>
      <dt>入場できる時間帯</dt>
      <dd>
        {ticketType.entryTimeFrom} - {ticketType.entryTimeTo}
      </dd>
    </dl>
  );
}
