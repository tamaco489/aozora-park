import { createClient } from "@connectrpc/connect";

import { InventoryService } from "../gen/aozorapark/inventory/v1/inventory_service_pb";
import { transport } from "./transport";

// 生成した型をそのまま使う、手で型を書かない
export const inventoryClient = createClient(InventoryService, transport);
