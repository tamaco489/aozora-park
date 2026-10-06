import { PriorityPassRequest } from "./PriorityPassRequest";
import { PriorityPassView } from "./PriorityPassView";

export function PriorityPassPage() {
  return (
    <>
      <h2>優先パス</h2>
      <PriorityPassRequest />
      <PriorityPassView />
    </>
  );
}
