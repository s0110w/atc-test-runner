# atr

AtCoder 用のテストランナー。
ローカルのソースコードを、ターミナルから AtCoder のサンプルケースでテストします。
外部依存ゼロの Go 製・単一バイナリです。

## Install

```console
$ brew install s0110w/tap/atr
```

または `go build -o atr .`

## Usage

コンテストに参加するときは `new` で一括セットアップします。

```console
$ atr new abc300                  # 全問題のディレクトリ + サンプル + テンプレート
$ cd abc300/a
$ vim main.py
$ atr t                           # テスト実行(コマンドは設定から)
2 cases found

sample-1
time: 0.038798 sec
AC
...
```

単発の問題は `atr d abc086_a`(フル URL でも可)で `test/` にサンプルを保存できます。

- `-c COMMAND` テスト対象(デフォルトは設定の `command`、なければ `./a.out`)、`-t SECONDS` 時間制限(デフォルト 10 秒、0 で無制限)
- テストケースは `test/sample-N.in` / `test/sample-N.out` 配置
- 比較は末尾改行のみ許容する完全一致(手元で AC なら提出先でもサンプルは AC)
- 全ケース AC なら exit 0(`atr t && git commit` のようなシェル連携ができます)

開催中のコンテストの問題ページはログインが必要です。
ブラウザの開発者ツールで `REVEL_SESSION` クッキーの値をコピーし、環境変数に設定してください。

```console
$ export ATR_SESSION='<REVEL_SESSION の値>'
```

## Config

設定は任意です。`atr.toml` を作業ディレクトリの上位(同階層含む)のどこかに置くと、最も近いものが 1 つだけ使われます。
ホームディレクトリ直下に置けば実質グローバル設定になります。

```toml
command = "python3 main.py"   # atr t のデフォルトコマンド
template = "./template"       # atr new が各問題ディレクトリへコピーするディレクトリ(このファイルからの相対)
```

設定がなくてもすべてのコマンドは動きます(`command` は `./a.out`、テンプレートはスキップ)。

## License

MIT
