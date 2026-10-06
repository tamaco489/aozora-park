import { ParkCreate } from "./ParkCreate";
import { ParkUpdate } from "./ParkUpdate";
import { ParkView } from "./ParkView";

export function ParkPage() {
  return (
    <>
      <h2>パーク</h2>
      <ParkCreate />
      <ParkView />
      <ParkUpdate />
    </>
  );
}
