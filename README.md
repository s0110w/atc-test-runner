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

問題ごとにディレクトリを掘って使います。

```console
$ mkdir abc086_a && cd abc086_a
$ atr d abc086_a                  # サンプルを test/ に保存(フル URL でも可)
saved: test/sample-1.in
saved: test/sample-1.out
...
$ atr t -c "python3 main.py"      # テスト実行
2 cases found

sample-1
time: 0.038798 sec
AC
...
```

- `-c COMMAND` テスト対象(デフォルト `./a.out`)、`-t SECONDS` 時間制限(デフォルト 10 秒、0 で無制限)
- テストケースは `test/sample-N.in` / `test/sample-N.out` 配置
- 比較は末尾改行のみ許容する完全一致(手元で AC なら提出先でもサンプルは AC)
- 全ケース AC なら exit 0(`atr t && git commit` のようなシェル連携ができます)

開催中のコンテストの問題ページはログインが必要です。
ブラウザの開発者ツールで `REVEL_SESSION` クッキーの値をコピーし、環境変数に設定してください。

```console
$ export ATR_SESSION='<REVEL_SESSION の値>'
```

## License

MIT
