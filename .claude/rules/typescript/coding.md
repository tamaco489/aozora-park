# TypeScript のコーディング規約

ディレクトリは `frontend/`。Vite と React で画面を作り、connect-web で api を呼ぶ。

この規約のどれが機械の検査に載っているかは `frontend/.oxlintrc.json` と `frontend/tsconfig.app.json` を見る。
**どの検査が見ているかを規約側に書かない。** 検査を外したときに規約が嘘になるため。

## 構成

### ディレクトリ構成

```text
frontend/
├── src/
│   ├── main.tsx          # React の起動だけを書く
│   ├── App.tsx           # 画面の部品を並べるだけ
│   ├── index.css         # 全体の CSS
│   ├── api/              # api との接続。画面から切り離す
│   │   ├── transport.ts  # 接続先とトランスポート
│   │   ├── park.ts       # サービスごとのクライアント
│   │   └── errors.ts     # connect のエラーを画面の文言に変える
│   ├── features/         # 業務機能ごとの画面
│   │   └── park/
│   └── gen/              # buf generate の出力。手で編集しない
├── .env.development      # ローカルの接続先。コミットする
└── .oxlintrc.json
```

- `src/features/` は業務機能で分け、`backend/internal/<機能>` と名前を揃える
- **画面の部品から connect のクライアントを作らない。** `src/api/` で作ったものを import する
- `src/api/` はサービス 1 つにつき 1 ファイルにする (`park.ts`)。トランスポートは全サービスで共有する

## api との接続

### 生成した型を使う

**`src/gen/` の型とサービス定義をそのまま使い、手で型を書かない。** proto が正で、手で書いた型は proto の変更に追従しない。

```ts
import { createClient } from "@connectrpc/connect";

import { ParkService } from "../gen/aozorapark/park/v1/park_service_pb";
import { transport } from "./transport";

export const parkClient = createClient(ParkService, transport);
```

- 型だけを使う import は `import type` か `type` 修飾子を付ける (`import { useState, type SubmitEvent } from "react"`)
- 画面が持つ入力の値など、proto に対応する型が無いものだけを自分で定義する

### 接続先

- 接続先は `VITE_API_BASE_URL` から読む。Vite は `VITE_` で始まる変数だけを `import.meta.env` に出す
- **既定値を置かない。** 未設定なら起動時に例外を投げ、接続先を誤ったまま動く余地を残さない
- `.env.development` はローカルの接続先だけを持つため、コミットする。ルートの `.gitignore` が `.env.*` を除外しているので、`frontend/.gitignore` で戻している
- **秘匿値を `VITE_` の変数に入れない。** ビルドした JavaScript にそのまま埋め込まれる
- 個人の上書きは `.env.development.local` に置く (`*.local` は追跡しない)

### エラーの扱い

**connect のエラーを文言に変える処理は `src/api/errors.ts` の 1 か所だけに置く。** 画面は `messageOf(err)` を呼ぶだけにする。

- `ConnectError.from(err)` で変換し、`Code` で振り分ける。エラーを文字列で比較しない
- サーバは `apperr` の分類を connect のコードに変換して返すため、画面で扱う分類はコードで足りる
- ネットワークに届かない場合と CORS で遮断された場合は、どちらも `Code.Unknown` になり区別できない。文言は両方の可能性を示す
- 失敗の表示には `role="alert"` を付け、色だけに頼らない

### 入力の検証

**入力の制約は proto の protovalidate が正。** 画面で同じ検証を重ねない。

- 画面が持つのは型と、`type="number"` のような最小限の入力補助だけにする
- 弾かれた理由は `Code.InvalidArgument` のメッセージとして返るので、それを表示する

### 更新の RPC

connect の更新 RPC は部分更新ではない。**変えない項目も現在の値を送る。** フォームの初期値は取得の RPC の結果で埋める。

## 画面の部品

### 部品の書き方

- 部品は `function` 宣言の名前付きエクスポートにする。default export は `App` だけにする (`main.tsx` が読む)
- props の型はファイル内で `Props` と名付ける。1 つしか受け取らない小さな部品は引数に直接書いてよい
- 登録・参照・更新で同じ入力欄や表示を使う場合は、部品に切り出して共有する (`ParkFields` `ParkDetail`)
- 同じ部品を 1 画面に複数置くときは、`id` が重ならないよう接頭辞を props で受け取る

### RPC を呼ぶ画面の状態

値・結果・エラー・読み込み中の 4 つを `useState` で持ち、`try` / `catch` / `finally` で切り替える。

```tsx
async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
  e.preventDefault();
  setLoading(true);
  setPark(undefined);
  setError("");

  try {
    const res = await parkClient.getPark({ parkId });
    setPark(res.park);
  } catch (err) {
    setError(messageOf(err));
  } finally {
    setLoading(false);
  }
}
```

- 送信の前に、前回の結果とエラーを消す
- 読み込み中はボタンを `disabled` にし、二重送信を防ぐ
- フォームは `<form onSubmit>` で受け、`preventDefault` を呼ぶ

### 非推奨の API を使わない

`@types/react` などで `@deprecated` が付いた型や関数を使わない。

- `@deprecated` は TypeScript の提案 (suggestion) の診断で、`tsc` はエラーにしない。検査は型情報を使う lint で行う
- 例: フォームのイベントは `React.FormEvent` ではなく `SubmitEvent<HTMLFormElement>` を使う

### CSS

- 全体の CSS は `src/index.css` の 1 枚に置く。CSS フレームワークは使わない
- 要素とクラスのセレクタで書き、見た目のためだけのクラスは必要になった箇所にだけ付ける (`.field` `.narrow`)

## 書き方

### コメント

- 「なぜ」が自明でないときだけ書く。「何をしているか」は書かない
- 日本語で書き、句点 (。) を含めない。1 行で書き、長くなるなら 1 行 1 論点で分ける
- 公開する関数と部品の説明は、名前から始める (`// messageOf は connect のエラーを画面に出す文言に変える`)

### import の並び

外部のパッケージ → 空行 → リポジトリ内のファイル の順に並べる。

```ts
import { useState, type SubmitEvent } from "react";

import { parkClient } from "../../api/park";
import { ParkDetail } from "./ParkDetail";
```

## 実行と依存

### 実行

- `just install` `just dev` `just lint` `just build` `just preview` を `frontend/` で実行する
- **PR を出す前に `just lint` と `just build` を通す。** `build` は `tsc -b` を含むため、型の検査もここで落ちる
- `just dev` の前に、backend で `just run-api` を起動しておく。Vite の既定のオリジン (`http://localhost:5173`) を api の CORS が許可している
- 5173 が塞がっていると Vite は別のポートで起動し、CORS で遮断される。先に塞いでいるプロセスを止める

### 依存

- 依存は `npm ci` で `package-lock.json` のとおりに入れる。`package-lock.json` はコミットする
- Node のバージョンは `.tool-versions` から読む
- 依存は必要になった時点で足す。**追加したら理由を PR に書く**
