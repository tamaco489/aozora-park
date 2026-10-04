import { createConnectTransport } from "@connectrpc/connect-web";

// 本番は Firebase Hosting の rewrite、ローカルは vite の server.proxy が api へ転送するため常に同一オリジンになる
export const transport = createConnectTransport({ baseUrl: "/" });
