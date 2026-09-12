# 全般ルール

## 応答

- 日本語・簡潔・直接的に書く
- 絵文字を使わない (明示的に求められた場合を除く)

## 日本語表記

- 括弧は全角 (（）) ではなく半角 `()` を使用し、前後に半角スペースを入れる
- 文中で英字を使う場合は前後に半角スペースを入れる (例: `xxx hoge xxx`)

## コメント

- 「なぜ」が自明でない場合のみ書く。「何をしているか」は書かない
- 句点 (。) を含めない

## ドキュメント作成

- デフォルトでは端的・簡潔にまとめる
- ユーザーから詳細に書くよう要望があった場合にのみ詳細に記載する

## README 作成

- `docs/` 配下に英語表記のドキュメントを作成する
- 同階層に日本語版の `README.ja.md` を作成する
- 英語版と日本語版は相互リンクで遷移できるようにする

## Markdown

- ハードタブを使わず、スペースでインデントする (MD010)
- コードブロックの前後に空行を入れる (MD031)
- 裸の URL を使わず `[テキスト](URL)` または `<URL>` の形式にする (MD034)
- 見出しの代わりに太字を使わない (MD036)
- コードブロックには必ず言語指定をつける (MD040)
- テーブルのパイプ文字を揃える (MD060, style: aligned)
- 同じ内容の見出しを複数使わない (MD024) — 繰り返しラベルは見出しにせず段落のインラインに組み込む

## 図の使い分け

図の種類で原本の形式を分ける。Mermaid を構成図に使わない。

| 図の種類               | 原本                | 用途例                                   |
| ---------------------- | ------------------- | ---------------------------------------- |
| 層や領域を面で表す図   | draw.io (`.drawio`) | 同心円のアーキテクチャ図、インフラ構成図 |
| 箱と矢印で関係を表す図 | SVG (`.svg`)        | パッケージ依存、コンポーネント間の接続   |
| 処理の順序を表す図     | Mermaid (`.mmd`)    | シーケンス図、フロー図                   |

- **draw.io** は面の重なりと矢印の自動ルーティングが効く。同心円のように帯の中へ文字を置く図はこちらにする
- **SVG** は座標を自分で決める代わりに、書き出しまでコマンドで完結する。箱と矢印が主の図はこちらにする
- **Excalidraw の MCP はビューアでの下書きに使う。** `.excalidraw` を原本として残さない。ファイルからの書き出しに使えるコマンドライン手段が無く、PNG が手作業になるため

## 図の原本と書き出し

**原本と PNG の両方をコミットする。**

```sh
# draw.io (-p はページ番号、1 から数える)
drawio --no-sandbox -x -f png -s 2 -p 1 -o {output}.png {input}.drawio

# SVG (HTML に包んで headless Chrome で撮る、--force-device-scale-factor で解像度を上げる)
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless --disable-gpu \
  --hide-scrollbars --force-device-scale-factor=2 --window-size={幅},{高さ} \
  --default-background-color=FFFFFFFF --screenshot={output}.png {input}.html

# Mermaid (--scale 3 を必ず付ける、既定では日本語が判読できない)
npx @mermaid-js/mermaid-cli -i {input}.mmd -o {output}.png --backgroundColor white --scale 3
```

SVG を図として直接埋め込まない。フォントの扱いが環境で変わるため、埋め込むのは PNG にする。
