package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ExecuteRequest struct {
	Code    string `json:"code"`
	Version string `json:"version"`
	Edition string `json:"edition"`
}

type ExecuteResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Error   string `json:"error"`
}

func executeRustCode(code string) ExecuteResponse {
	cargoPath := getCargoPath()

	// Create workspace directory if it doesn't exist
	workspaceDir := getWorkspaceDir()
	err := os.MkdirAll(workspaceDir, 0o755)
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Result:  fmt.Sprintf("Failed to create workspace directory: %v", err),
			Error:   fmt.Sprintf("Failed to create workspace directory: %v", err),
		}
	}

	// Create temporary project directory
	tempProjectsDir := getTempProjectDir()
	err = os.MkdirAll(tempProjectsDir, 0o755)
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Result:  fmt.Sprintf("Failed to create temp projects directory: %v", err),
			Error:   fmt.Sprintf("Failed to create temp projects directory: %v", err),
		}
	}

	tempDir, err := os.MkdirTemp(tempProjectsDir, "rust-project-*")
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Result:  fmt.Sprintf("Failed to create temp directory: %v", err),
			Error:   fmt.Sprintf("Failed to create temp directory: %v", err),
		}
	}
	defer os.RemoveAll(tempDir)

	// Initialize Cargo project
	cargoInitCmd := NewLocalRustCommand(cargoPath, "init", "--name", "temp_project", tempDir)
	_, err = cargoInitCmd.CombinedOutput()
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Result:  fmt.Sprintf("Failed to initialize Cargo project: %v", err),
			Error:   fmt.Sprintf("Failed to initialize Cargo project: %v", err),
		}
	}

	// Write code to src/main.rs
	srcDir := filepath.Join(tempDir, "src")
	mainFile := filepath.Join(srcDir, "main.rs")
	if err := os.WriteFile(mainFile, []byte(code), 0o644); err != nil {
		return ExecuteResponse{
			Success: false,
			Result:  fmt.Sprintf("Failed to write code file: %v", err),
			Error:   fmt.Sprintf("Failed to write code file: %v", err),
		}
	}

	// Compile with Cargo
	rustTarget := getRustTarget()
	execCmd := NewLocalRustCommand(cargoPath, "run", "--target", rustTarget)
	execCmd.Dir = tempDir

	// Windowsではrust-lldを使わず、その他ではrust-lldを使用
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		// その他のプラットフォームではrust-lldを使用（self-contained）
		execCmd.Env = append(os.Environ(), "RUSTFLAGS=-C linker=rust-lld")
	}

	// Create channels for timeout handling
	errChan := make(chan error, 1)
	var stdout, stderr strings.Builder
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err = execCmd.Start()
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Result:  fmt.Sprintf("Failed to start program: %v", err),
			Error:   fmt.Sprintf("Failed to start program: %v", err),
		}
	}

	go func() {
		err := execCmd.Wait()
		errChan <- err
	}()

	select {
	case err := <-errChan:
		result := stdout.String()
		errOutput := stderr.String()
		success := true
		if err != nil {
			success = false
			if len(result) != 0 {
				result += "\n"
			}
			result += errOutput
		}

		return ExecuteResponse{
			Success: success,
			Result:  result,
			Error:   errOutput,
		}
	case <-time.After(10 * time.Second):
		if execCmd.Process != nil {
			execCmd.Process.Kill()
		}
		return ExecuteResponse{
			Success: false,
			Result:  "Execution timeout (10 seconds)",
			Error:   "Execution timeout (10 seconds)",
		}
	}
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== Received %s request to /execute ===", r.Method)

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Request parsed - Code length: %d, Version: %s, Edition: %s",
		len(req.Code), req.Version, req.Edition)

	if strings.TrimSpace(req.Code) == "" {
		log.Printf("Empty code provided")
		http.Error(w, "Code cannot be empty", http.StatusBadRequest)
		return
	}

	log.Printf("Calling executeRustCode...")
	result := executeRustCode(req.Code)
	log.Printf("executeRustCode returned: success=%t", result.Success)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(result)
}

func optionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Initialize Rust toolchain on startup
	log.Println("Initializing Rust toolchain...")
	if err := initializeRustToolchain(); err != nil {
		log.Printf("Error: Failed to initialize Rust toolchain: %v", err)
		panic(err)
	}

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			optionsHandler(w, r)
		} else {
			executeHandler(w, r)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		status := "Local Rust Playground Server\n"
		status += "POST /execute to run Rust code\n\n"

		// Show toolchain status
		if _, err := os.Stat(getCargoPath()); err == nil {
			status += fmt.Sprintf("Using local Rust toolchain at: %s\n", getRustHome())
		} else {
			panic(err)
		}

		fmt.Fprint(w, status)
	})

	port := "8081"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Go server starting on port %s", port)
	if _, err := os.Stat(getCargoPath()); err == nil {
		log.Printf("Using local Rust toolchain at: %s", getRustHome())
	} else {
		panic(err)
	}

	log.Fatal(http.ListenAndServe("localhost:"+port, nil))
}
