package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/isaacdsc/mcp-server/internal/tools"
)

func TestServer_HandleToolsList(t *testing.T) {
	t.Parallel()

	server := NewServer(
		tools.NewRegistry(tools.EchoTool{}, tools.TimestampTool{}),
		slog.Default(),
	)

	resp := server.handleLine(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	if resp == nil || resp.Error != nil {
		t.Fatalf("unexpected error response: %#v", resp)
	}

	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}
	toolList, ok := resultMap["tools"].([]map[string]any)
	if ok {
		if len(toolList) != 2 {
			t.Fatalf("expected 2 tools, got %d", len(toolList))
		}
		return
	}

	// Fallback for interface conversions from map[string]any.
	raw, err := json.Marshal(resultMap["tools"])
	if err != nil {
		t.Fatalf("marshal tools: %v", err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal tools: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(decoded))
	}
}

func TestServer_HandleToolsCall(t *testing.T) {
	t.Parallel()

	server := NewServer(
		tools.NewRegistry(tools.EchoTool{}),
		slog.Default(),
	)

	req := `{"jsonrpc":"2.0","id":"abc","method":"tools/call","params":{"name":"echo","arguments":{"message":"ok"}}}`
	resp := server.handleLine(context.Background(), []byte(req))
	if resp == nil || resp.Error != nil {
		t.Fatalf("unexpected error response: %#v", resp)
	}

	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}
	contentAny, ok := resultMap["content"]
	if !ok {
		t.Fatalf("missing content field")
	}
	contentRaw, err := json.Marshal(contentAny)
	if err != nil {
		t.Fatalf("marshal content: %v", err)
	}
	var content []map[string]any
	if err := json.Unmarshal(contentRaw, &content); err != nil {
		t.Fatalf("unmarshal content: %v", err)
	}
	if len(content) != 1 {
		t.Fatalf("expected one content item, got %d", len(content))
	}
	if content[0]["type"] != "text" {
		t.Fatalf("unexpected content type: %v", content[0]["type"])
	}
}
