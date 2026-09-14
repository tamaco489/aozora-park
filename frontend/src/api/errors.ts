import { Code, ConnectError } from "@connectrpc/connect";

// messageOf は connect のエラーを画面に出す文言に変える
//
// サーバは apperr の分類を connect のコードに変換して返すため、コードで振り分ける
export function messageOf(err: unknown): string {
  const connectErr = ConnectError.from(err);

  switch (connectErr.code) {
    case Code.NotFound:
      return "そのパークは見つかりません";
    case Code.AlreadyExists:
      return "そのパークは既に登録されています";
    case Code.InvalidArgument:
      return `入力が正しくありません: ${connectErr.rawMessage}`;
    case Code.Unavailable:
      return "api に繋がりません。just run-api で起動しているか確認してください";
    // ネットワークに届かない場合と CORS で遮断された場合は、どちらも fetch の失敗として Unknown になる
    // ブラウザは遮断の理由を JavaScript に渡さないため、コードからは区別できない
    case Code.Unknown:
      return "api に繋がらないか CORS で遮断されています。api が起動しているか、CORS_ALLOWED_ORIGINS がこの画面のオリジンを含むかを確認してください";
    default:
      return `失敗しました: ${connectErr.rawMessage}`;
  }
}
