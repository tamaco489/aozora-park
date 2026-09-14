import { createConnectTransport } from "@connectrpc/connect-web";

// Vite は VITE_ で始まる変数だけを import.meta.env に出す
const baseUrl = import.meta.env.VITE_API_BASE_URL;

// 既定値を置かない、接続先を誤ったまま動く余地を残さないため
if (!baseUrl) {
  throw new Error(
    "VITE_API_BASE_URL が設定されていません。frontend/.env.development を確認してください",
  );
}

export const transport = createConnectTransport({ baseUrl });
