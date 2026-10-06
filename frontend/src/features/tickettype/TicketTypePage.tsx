import { TicketTypeCreate } from "./TicketTypeCreate";
import { TicketTypeUpdate } from "./TicketTypeUpdate";

export function TicketTypePage() {
  return (
    <>
      <h2>券種</h2>
      <TicketTypeCreate />
      <TicketTypeUpdate />
    </>
  );
}
