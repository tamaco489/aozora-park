import { createClient } from "@connectrpc/connect";

import { PriorityPassService } from "../gen/aozorapark/prioritypass/v1/prioritypass_service_pb";
import { transport } from "./transport";

// 生成した型をそのまま使う、手で型を書かない
export const priorityPassClient = createClient(PriorityPassService, transport);
