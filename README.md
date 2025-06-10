# ローカルRust Playgroundサーバー

## 使用方法

```bash
# サーバー起動
go run main.go rust_bin.go

# curlでテスト
curl -X POST http://localhost:8081/execute \
  -H "Content-Type: application/json" \
  -d '{"code":"fn main() { println!(\"Hello from Go server!\"); }"}'

# サンプル mdbook 起動
cd sample-book && mdbook serve --open
```
