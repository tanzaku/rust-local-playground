# Go Rust Playground Server - Cross Build

APP_NAME = local-playground-server
BUILD_DIR = dist

# デフォルトターゲット
.PHONY: all
all: clean build-all

# ディレクトリ作成
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# 全プラットフォーム用ビルド
.PHONY: build-all
build-all: $(BUILD_DIR)
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go rust_bin.go
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 main.go rust_bin.go
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 main.go rust_bin.go
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 main.go rust_bin.go
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe main.go rust_bin.go
	@echo "Build completed. Binaries are in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# 個別プラットフォーム用ビルド
.PHONY: linux-amd64
linux-amd64: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go rust_bin.go

.PHONY: linux-arm64
linux-arm64: $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 main.go rust_bin.go

.PHONY: darwin-amd64
darwin-amd64: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 main.go rust_bin.go

.PHONY: darwin-arm64
darwin-arm64: $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 main.go rust_bin.go

.PHONY: windows-amd64
windows-amd64: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe main.go rust_bin.go

.PHONY: run-sample-book
serve-sample-book:
	cd sample-book && mdbook serve --open

.PHONY: run-server
run-server:
	go run main.go rust_bin.go

# 掃除
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
