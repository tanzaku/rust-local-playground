# Local Rust Playground Server

A web server that allows you to compile and execute Rust code locally by sending requests to this local server instead of calling the official Rust Playground API from mdbook.

## Features

- **Easy Setup**: Automatically downloads and sets up Rust toolchain on first startup, even if Rust is not installed on the system
- **Isolated Environment**: Places Rust toolchain in a dedicated directory without affecting the system environment
- **Cross-Platform**: Supports Linux, macOS, and Windows

## Installation

### 1. Download Binary (Recommended)

Download the binary for your platform from [Releases](https://github.com/tanzaku/rust-local-playground/releases):

- `rust-local-playground-linux-amd64` - Linux x64
- `rust-local-playground-linux-arm64` - Linux ARM64
- `rust-local-playground-darwin-amd64` - macOS Intel
- `rust-local-playground-darwin-arm64` - macOS Apple Silicon
- `rust-local-playground-windows-amd64.exe` - Windows x64

### 2. Run

```bash
# Linux/macOS (add executable permission first)
chmod +x rust-local-playground-linux-amd64
./rust-local-playground-linux-amd64

# Windows
rust-local-playground-windows-amd64.exe
```

The server will start at `http://localhost:8081`. The first startup may take some time to download the Rust toolchain.

## Usage

### Using with mdbook

To use the local server with mdbook, the following configuration is required:

1. Add JavaScript file to `book.toml`
2. Change the Playground API endpoint to the local server

For detailed configuration examples, see `sample-book/book.toml` and `sample-book/local-playground.js`.

### Direct API Usage

```bash
curl -X POST http://localhost:8081/execute \
  -H "Content-Type: application/json" \
  -d '{"code":"fn main() { println!(\"Hello, Rust!\"); }"}'
```

**Response Example:**
```json
{
  "success": true,
  "result": "Hello, Rust!\n",
  "error": ""
}
```

## For Developers

```bash
# Run from source (for development)
go run main.go rust_bin.go

# Using Make
make run-server

# Launch sample mdbook
make serve-mdbook
```