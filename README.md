# atr

AtCoder 用のテストランナー。
ローカルのソースコードを、ターミナルから AtCoder のサンプルケースでテストします。
外部依存ゼロの Go 製・単一バイナリです。

## Install

```console
$ brew install s0110w/tap/atr
```

Go があればソースからも入れられます(リポジトリのルートで実行)。

```console
$ git clone https://github.com/s0110w/atc-test-runner && cd atc-test-runner
$ go build -o atr ./cmd/atr
$ mv atr /usr/local/bin/   # PATH の通った場所へ
```

## Quick start(コンテスト本番の流れ)

1. **(初回だけ)テンプレートと設定を用意** — 毎回書くひな形と実行コマンドを [Config](#config) の要領で `~/atr.toml` に置いておくと、以降のコンテストで使い回せます。省略しても動きます。
2. **(開催中のみ)ログイン用クッキーを設定** — 開催中のコンテストはサンプルの取得にログインが必要です。[Login](#login) を参照。過去問なら不要。
3. **コンテストを一括セットアップ** — 全問題のディレクトリ・サンプル・テンプレートをまとめて用意します。

   ```console
   $ atr new abc300
   ```
4. **1 問ずつ解く → テスト → 提出** — を各問題で繰り返します。

   ```console
   $ cd abc300/a
   $ vim main.cpp
   $ g++ -O2 -o a.out main.cpp   # ← ビルドは自分で。atr はビルドしません
   $ atr t
   2 cases found

   sample-1
   time: 0.003 sec
   AC

   sample-2
   time: 0.004 sec
   AC

   test success: 2 cases
   ```

   全ケース AC なら exit 0 で終わるので、`atr t && ...` のようにシェルと連携できます。AC を確認したら AtCoder に提出し、次の問題へ。

単発の問題だけ欲しいときは `atr d`(download)で `test/` にサンプルを保存できます。

```console
$ atr d abc086_a        # 問題 ID。フル URL でも可
```

## Commands

| コマンド | 別名 | 引数 | 説明 |
|---|---|---|---|
| `atr new` | `atr n` | `[-s]` コンテスト ID(例 `abc300`) | 全問題のディレクトリ + サンプル + テンプレートを一括作成。`-s` で対象の問題を TUI で選択 |
| `atr download` | `atr d` | 問題 ID(例 `abc086_a`)または URL | サンプルを `test/` に保存 |
| `atr test` | `atr t` | (下記オプション) | `test/` のサンプルでテスト実行 |

引数なしで `atr` を実行するとコマンド一覧が、`atr test -h` で `test` のオプションが出ます。

`atr new` のオプション:

- `-s` セットアップする問題を TUI で選択(↑↓/jk で移動、スペースで切り替え、a で全選択/全解除、enter で確定、q で中断)。既定は選択なしの全問題。設定の `select = true` で常に選択にできます

`atr test` のオプション:

- `-c COMMAND` テスト対象のコマンド(既定は設定の `command`、なければ `./a.out`)
- `-t SECONDS` 時間制限秒(既定 10、`0` で無制限)
- `-d DIR` テストケースの置き場所(既定 `test`)

## Judgement

**`atr t` はビルドをしません。指定されたコマンドを実行し、標準出力をサンプルと比較するだけ**です。
C++ のようにビルドが要る言語は、先に自分でビルドしてから `atr t`(既定コマンド `./a.out`)。
毎回ビルドまで任せたいなら設定のコマンドに含められます: `command = "g++ -O2 -o a.out main.cpp && ./a.out"`。

各ケースの判定は次の 4 つ:

| 表示 | 意味 |
|---|---|
| `AC` | 正解。出力がサンプルと一致 |
| `WA` | 不正解。出力が一致しない(input / output / expected を並べて表示) |
| `TLE` | 時間制限(`-t`)超過で強制終了 |
| `RE` | コマンドが非ゼロ終了(実行時エラー・クラッシュなど) |

比較は完全一致で、末尾の改行差と CRLF だけは許容します(手元で AC ならサンプルは提出先でも AC)。
空白の違いは WA になり、誤差許容が必要な問題には非対応です。

WA の例:

```console
$ atr t
1 cases found

sample-1
time: 0.003 sec
WA
--- input ---
3
--- output ---
5
--- expected ---
6

test failed: 0 AC / 1 cases
```

## Login

開催中のコンテストは問題ページがログイン必須なので、`atr new` / `atr d` がサンプルを取得できません。
ブラウザで AtCoder にログインした状態で、開発者ツールの Application → Cookies → `https://atcoder.jp` から `REVEL_SESSION` の値をコピーし、環境変数に入れてください。

```console
$ export ATR_SESSION='<REVEL_SESSION の値>'
```

コンテスト終了後の問題や過去問は不要です。セッションが切れたら取り直してください。

## Config

設定は任意です。`atr.toml` を作業ディレクトリから上位(同階層含む)のどこかに置くと、最も近いもの 1 つだけが使われます。
ホームディレクトリ直下に置けば実質グローバル設定になります。設定がなくても全コマンド動きます(`command` は `./a.out`、テンプレートはスキップ)。

```toml
command = "python3 main.py"        # atr t のデフォルトコマンド
contest_template = "./contest"     # atr new がコンテストディレクトリ直下へ中身をコピー(この atr.toml からの相対)
task_template = "./task"           # atr new が各問題ディレクトリへ中身をコピー(同上)
select = true                      # atr new で常に問題選択 TUI を出す(省略時は全問題)
```

テンプレートは 2 層あり、どちらも指定したディレクトリの**中身**がコピーされます。

- `contest_template` はコンテスト直下に 1 回だけ展開。テスト実行用のシェルスクリプトやプロジェクトの雛形など
- `task_template` は各問題ディレクトリに展開。解答ソースコードの雛形など

例えば下記のように置くと、

```
~/atr.toml
~/contest/
    test-all.sh     # コンテスト全体で使う道具
~/task/
    main.cpp        # 毎回のひな形
```

`atr new abc300` はこうなります:

```
abc300/
    test-all.sh
    a/
        main.cpp
        test/
            sample-1.in
            sample-1.out
    b/
        main.cpp
        test/
            ...
```

## License

MIT
