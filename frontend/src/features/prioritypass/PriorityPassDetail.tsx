import {
  PriorityPassStatus,
  type PriorityPass,
} from "../../gen/aozorapark/prioritypass/v1/prioritypass_pb";

// 生成された enum は数値のため、画面に出す文言への変換をこの機能の中に閉じる
function statusLabel(status: PriorityPassStatus): string {
  switch (status) {
    case PriorityPassStatus.REQUESTED:
      return "申込を受け付けた (requested)";
    case PriorityPassStatus.ISSUED:
      return "発行済み (issued)";
    case PriorityPassStatus.SOLD_OUT:
      return "枠が埋まっていて発行できなかった (sold_out)";
    case PriorityPassStatus.UNSPECIFIED:
      return "未設定 (unspecified)";
  }
}

// PriorityPassDetail は申込・参照の結果を同じ形で見せる
export function PriorityPassDetail({
  priorityPass,
}: {
  priorityPass: PriorityPass;
}) {
  return (
    <dl>
      <dt>優先パスの識別子</dt>
      <dd>{priorityPass.passId}</dd>
      <dt>パークの識別子</dt>
      <dd>{priorityPass.parkId}</dd>
      <dt>券の識別子</dt>
      <dd>{priorityPass.ticketId}</dd>
      <dt>アトラクションの識別子</dt>
      <dd>{priorityPass.attractionId}</dd>
      <dt>時間帯枠の識別子</dt>
      <dd>{priorityPass.timeSlotId}</dd>
      <dt>状態</dt>
      <dd>{statusLabel(priorityPass.status)}</dd>
    </dl>
  );
}
