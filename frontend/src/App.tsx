import { Route, Routes } from "react-router";

import { Home } from "./Home";
import { Layout } from "./Layout";
import { NotFound } from "./NotFound";
import { AttractionPage } from "./features/attraction/AttractionPage";
import { InventoryPage } from "./features/inventory/InventoryPage";
import { ParkPage } from "./features/park/ParkPage";
import { TicketTypePage } from "./features/tickettype/TicketTypePage";

// 機能ごとにページを分ける、一覧の RPC ができるまで ID はパスに載せない
export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Home />} />
        <Route path="parks" element={<ParkPage />} />
        <Route path="attractions" element={<AttractionPage />} />
        <Route path="ticket-types" element={<TicketTypePage />} />
        <Route path="inventory" element={<InventoryPage />} />
        <Route path="*" element={<NotFound />} />
      </Route>
    </Routes>
  );
}
