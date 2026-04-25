package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoTool_Call(t *testing.T) {
	t.Parallel()

	tool := EchoTool{}
	out, err := tool.Call(context.Background(), json.RawMessage(`{"message":"hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := out["message"]; got != "hello" {
		t.Fatalf("unexpected message: %v", got)
	}
}

func TestRegistry_CallToolNotFound(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(EchoTool{})
	_, err := registry.Call(context.Background(), "missing", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGolangProjectContextTool_Call(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/test\n\ngo 1.26.1\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "cmd"), 0o755); err != nil {
		t.Fatalf("mkdir cmd: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "internal"), 0o755); err != nil {
		t.Fatalf("mkdir internal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	tool := NewGolangProjectContextTool(tempDir)
	out, err := tool.Call(context.Background(), json.RawMessage(`{"max_depth":4}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := out["is_golang_project"]; got != true {
		t.Fatalf("expected is_golang_project true, got %v", got)
	}
	if got := out["confidence"]; got != "high" {
		t.Fatalf("expected confidence high, got %v", got)
	}
}

func TestGolangProjectContextTool_CallInvalidInput(t *testing.T) {
	t.Parallel()

	tool := NewGolangProjectContextTool(t.TempDir())
	_, err := tool.Call(context.Background(), json.RawMessage(`{"max_depth":99}`))
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGolangProjectContextTool_CallNonGoProject(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# sample\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	tool := NewGolangProjectContextTool(tempDir)
	out, err := tool.Call(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := out["is_golang_project"]; got != false {
		t.Fatalf("expected is_golang_project false, got %v", got)
	}
}
