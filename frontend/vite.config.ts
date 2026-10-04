import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// api の待ち受け先、backend の just run-api が 8080 で起動する
const apiTarget = "http://localhost:8080";

// DOC: https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  // 本番は Firebase Hosting の rewrite が転送する、ローカルも同じ形にしてプリフライトを出さない
  server: {
    proxy: {
      "^/aozorapark\\.": { target: apiTarget },
      "^/grpc\\.health\\.": { target: apiTarget },
    },
  },
});
