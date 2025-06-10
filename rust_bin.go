package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// RUSTUP_HOME, CARGO_HOME のEnvを自動で持ったCommandをラップした構造体
type LocalRustCommand struct {
	name    string
	args    []string
	Env     []string
	Dir     string
	Stdout  *strings.Builder
	Stderr  *strings.Builder
	cmd     *exec.Cmd
	Process *os.Process
}

func NewLocalRustCommand(name string, args ...string) *LocalRustCommand {
	env := append(os.Environ(),
		"RUSTUP_HOME="+getRustHome(),
		"CARGO_HOME="+getCargoHome(),
		"CARGO_TARGET_DIR=",
	)
	filteredEnv := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "CARGO_TARGET_DIR=") {
			filteredEnv = append(filteredEnv, e)
		}
	}

	return &LocalRustCommand{name, args, filteredEnv, "", nil, nil, nil, nil}
}

func (localCmd *LocalRustCommand) CombinedOutput() ([]byte, error) {
	cmd := exec.Command(localCmd.name, localCmd.args...)
	cmd.Env = localCmd.Env
	cmd.Dir = localCmd.Dir
	return cmd.CombinedOutput()
}

func (localCmd *LocalRustCommand) Start() error {
	cmd := exec.Command(localCmd.name, localCmd.args...)
	cmd.Env = localCmd.Env
	cmd.Dir = localCmd.Dir
	cmd.Stdout = localCmd.Stdout
	cmd.Stderr = localCmd.Stderr
	localCmd.cmd = cmd
	err := cmd.Start()
	if err == nil {
		localCmd.Process = cmd.Process
	}
	return err
}

func (localCmd *LocalRustCommand) Wait() error {
	return localCmd.cmd.Wait()
}

// 作業ディレクトリのパス管理
func getWorkspaceDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "local-playground-workspace")
}

func getRustHome() string {
	return filepath.Join(getWorkspaceDir(), "rust-toolchain")
}

func getCargoHome() string {
	return filepath.Join(getRustHome(), "cargo")
}

func getCargoPath() string {
	rustTarget := getRustTarget()
	toolchainName := "stable-" + rustTarget
	rustHome := getRustHome()

	var cargoPath string
	if runtime.GOOS == "windows" {
		cargoPath = filepath.Join(rustHome, "toolchains", toolchainName, "bin", "cargo.exe")
	} else {
		cargoPath = filepath.Join(rustHome, "toolchains", toolchainName, "bin", "cargo")
	}

	return cargoPath
}

func getRustupInitPath() string {
	filename := "rustup-init"
	if runtime.GOOS == "windows" {
		filename = "rustup-init.exe"
	}
	return filepath.Join(getWorkspaceDir(), filename)
}

func getTempProjectDir() string {
	return filepath.Join(getWorkspaceDir(), "temp-projects")
}

// rustup-initのダウンロード
func downloadRustupInit() (string, error) {
	var url string

	switch runtime.GOOS {
	case "windows":
		// Windows 64bit GNU
		url = "https://static.rust-lang.org/rustup/dist/x86_64-pc-windows-gnu/rustup-init.exe"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			url = "https://static.rust-lang.org/rustup/dist/aarch64-apple-darwin/rustup-init"
		} else {
			url = "https://static.rust-lang.org/rustup/dist/x86_64-apple-darwin/rustup-init"
		}
	default: // linux
		if runtime.GOARCH == "arm64" {
			url = "https://static.rust-lang.org/rustup/dist/aarch64-unknown-linux-gnu/rustup-init"
		} else {
			url = "https://static.rust-lang.org/rustup/dist/x86_64-unknown-linux-gnu/rustup-init"
		}
	}

	rustupPath := getRustupInitPath()

	// 既に存在する場合はそれを使用
	if stat, err := os.Stat(rustupPath); err == nil {
		log.Printf("rustup-init already exists at: %s (size: %d bytes)", rustupPath, stat.Size())
		return rustupPath, nil
	}

	log.Printf("Downloading rustup-init from: %s", url)
	log.Printf("Target path: %s", rustupPath)

	// Create workspace directory first
	workspaceDir := getWorkspaceDir()
	err := os.MkdirAll(workspaceDir, 0o755)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %v", err)
	}
	log.Printf("Created workspace directory: %s", workspaceDir)

	// HTTP GETリクエスト (60秒タイムアウト)
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download rustup-init: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download rustup-init: HTTP %d", resp.StatusCode)
	}
	log.Printf("HTTP response: %d, Content-Length: %s", resp.StatusCode, resp.Header.Get("Content-Length"))

	// ファイルに書き込み
	file, err := os.Create(rustupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}
	log.Printf("Downloaded %d bytes to: %s", written, rustupPath)

	// Unix系では実行権限を付与
	if runtime.GOOS != "windows" {
		err = os.Chmod(rustupPath, 0o755)
		if err != nil {
			return "", fmt.Errorf("failed to set executable permission: %v", err)
		}
		log.Printf("Set executable permission on: %s", rustupPath)
	}

	// ファイルが正常に作成されたか確認
	if stat, err := os.Stat(rustupPath); err != nil {
		return "", fmt.Errorf("downloaded file verification failed: %v", err)
	} else {
		log.Printf("File verification successful: %s (size: %d bytes)", rustupPath, stat.Size())
	}

	return rustupPath, nil
}

// プラットフォーム対応のRustターゲットを取得
func getRustTarget() string {
	switch runtime.GOOS {
	case "windows":
		return "x86_64-pc-windows-gnu"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "aarch64-apple-darwin"
		} else {
			return "x86_64-apple-darwin"
		}
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "aarch64-unknown-linux-gnu"
		} else {
			return "x86_64-unknown-linux-gnu"
		}
	}
	return "x86_64-unknown-linux-gnu"
}

// Rustツールチェーンの初期化
func initializeRustToolchain() error {
	rustHome := getRustHome()
	cargoPath := getCargoPath()

	// 既にインストール済みかチェック
	if _, err := os.Stat(cargoPath); err == nil {
		log.Printf("Rust toolchain already exists at: %s", rustHome)
		return nil
	}

	log.Printf("Installing Rust toolchain to: %s", rustHome)

	// ディレクトリを作成
	err := os.MkdirAll(rustHome, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create rust home directory: %v", err)
	}

	err = os.MkdirAll(getCargoHome(), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create cargo home directory: %v", err)
	}

	// rustup-initをダウンロード
	rustupInitPath, err := downloadRustupInit()
	if err != nil {
		return fmt.Errorf("failed to download rustup-init: %v", err)
	}

	// rustupでRustをインストール
	// プラットフォーム固有のツールチェーンを既定として使用
	rustTarget := getRustTarget()
	defaultToolchain := "stable-" + rustTarget
	log.Printf("Using platform-specific toolchain as default: %s", defaultToolchain)

	cmd := NewLocalRustCommand(rustupInitPath,
		"--default-toolchain", defaultToolchain,
		"--profile", "minimal",
		"--no-modify-path",
		"-y",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install Rust: %v\nOutput: %s", err, string(output))
	}

	log.Println("Rust toolchain installed successfully!")

	// Log installed toolchain paths
	log.Printf("Installed Rust toolchain paths:")
	log.Printf("  RUSTUP_HOME: %s", rustHome)
	log.Printf("  CARGO_HOME: %s", getCargoHome())
	log.Printf("  cargo: %s", getCargoPath())

	return nil
}
