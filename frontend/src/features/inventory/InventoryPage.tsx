import { DateInventoryUpdate } from "./DateInventoryUpdate";
import { DateInventoryView } from "./DateInventoryView";
import { TimeSlotList } from "./TimeSlotList";

export function InventoryPage() {
  return (
    <>
      <h2>枠</h2>
      <DateInventoryView />
      <DateInventoryUpdate />
      <TimeSlotList />
    </>
  );
}
