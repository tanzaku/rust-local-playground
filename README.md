# ローカルRust Playgroundサーバー

mdbookからRust Playgroundの公式APIを呼び出す代わりに、このローカルサーバーにリクエストを送ってRustコードをコンパイル・実行できるWebサーバーです。

## 特徴

- **簡単セットアップ**: システムにRustがインストールされていなくても、初回起動時に自動でRustツールチェーンをダウンロード・セットアップ
- **隔離された環境**: 専用ディレクトリにRustツールチェーンを配置し、システム環境に影響しない
- **クロスプラットフォーム**: Linux、macOS、Windows対応

## インストール

### 1. バイナリをダウンロード（推奨）

[Releases](https://github.com/tanzaku/rust-local-playground/releases)から対応するプラットフォーム用のバイナリをダウンロードしてください：

- `rust-local-playground-linux-amd64` - Linux x64
- `rust-local-playground-linux-arm64` - Linux ARM64
- `rust-local-playground-darwin-amd64` - macOS Intel
- `rust-local-playground-darwin-arm64` - macOS Apple Silicon
- `rust-local-playground-windows-amd64.exe` - Windows x64

### 2. 実行

```bash
# Linux/macOS
./rust-local-playground-linux-amd64

# Windows
rust-local-playground-windows-amd64.exe
```

サーバーが `http://localhost:8081` で起動します。初回起動時はRustツールチェーンのダウンロードに時間がかかる場合があります。

## 使用方法

### mdbookでの利用

mdbookでローカルサーバーを使用するには、以下の設定が必要です：

1. `book.toml` にJavaScriptファイルを追加
2. PlaygroundのAPIエンドポイントをローカルサーバーに変更

詳細な設定例は `sample-book/book.toml` と `sample-book/local-playground.js` を参照してください。

### API直接利用

```bash
curl -X POST http://localhost:8081/execute \
  -H "Content-Type: application/json" \
  -d '{"code":"fn main() { println!(\"Hello, Rust!\"); }"}'
```

**レスポンス例:**
```json
{
  "success": true,
  "result": "Hello, Rust!\n",
  "error": ""
}
```

## 開発者向け

```bash
# ソースから起動（開発用）
go run main.go rust_bin.go

# Makeを使用
make run-server

# サンプルmdbook起動
make serve-mdbook
```
