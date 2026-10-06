import { AttractionCreate } from "./AttractionCreate";
import { AttractionUpdate } from "./AttractionUpdate";

export function AttractionPage() {
  return (
    <>
      <h2>アトラクション</h2>
      <AttractionCreate />
      <AttractionUpdate />
    </>
  );
}
