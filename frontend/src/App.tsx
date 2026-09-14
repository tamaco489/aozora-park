import { ParkCreate } from "./features/park/ParkCreate";
import { ParkUpdate } from "./features/park/ParkUpdate";
import { ParkView } from "./features/park/ParkView";

export default function App() {
  return (
    <main>
      <h1>Aozora Park</h1>
      <ParkCreate />
      <ParkView />
      <ParkUpdate />
    </main>
  );
}
