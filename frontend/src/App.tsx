import { AttractionCreate } from "./features/attraction/AttractionCreate";
import { AttractionUpdate } from "./features/attraction/AttractionUpdate";
import { DateInventoryUpdate } from "./features/inventory/DateInventoryUpdate";
import { DateInventoryView } from "./features/inventory/DateInventoryView";
import { TimeSlotList } from "./features/inventory/TimeSlotList";
import { ParkCreate } from "./features/park/ParkCreate";
import { ParkUpdate } from "./features/park/ParkUpdate";
import { ParkView } from "./features/park/ParkView";
import { TicketTypeCreate } from "./features/tickettype/TicketTypeCreate";
import { TicketTypeUpdate } from "./features/tickettype/TicketTypeUpdate";

// ルーターは入れず、機能ごとの見出しで区切って 1 画面に並べる
export default function App() {
  return (
    <main>
      <h1>Aozora Park</h1>

      <h2>パーク</h2>
      <ParkCreate />
      <ParkView />
      <ParkUpdate />

      <h2>アトラクション</h2>
      <AttractionCreate />
      <AttractionUpdate />

      <h2>券種</h2>
      <TicketTypeCreate />
      <TicketTypeUpdate />

      <h2>枠</h2>
      <DateInventoryView />
      <DateInventoryUpdate />
      <TimeSlotList />
    </main>
  );
}
