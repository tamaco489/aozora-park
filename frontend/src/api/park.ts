import { createClient } from "@connectrpc/connect";

import { ParkService } from "../gen/aozorapark/park/v1/park_service_pb";
import { transport } from "./transport";

// 生成した型をそのまま使う、手で型を書かない
export const parkClient = createClient(ParkService, transport);
