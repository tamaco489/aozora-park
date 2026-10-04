import { Code, ConnectError } from "@connectrpc/connect";

// messageOf は connect のエラーを画面に出す文言に変換する
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
    // api に届かなかった場合は fetch の失敗として Unknown になる
    // ブラウザは届かなかった理由を JavaScript に渡さないため、コードからは区別できない
    case Code.Unknown:
      return "api に繋がりません。api が起動しているか、プロキシや rewrites の設定が崩れていないかを確認してください";
    default:
      return `失敗しました: ${connectErr.rawMessage}`;
  }
}
