package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/isaacdsc/mcp-server/internal/app"
)

func TestMCPFlow_ToolsListAndCall(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString(
		`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}` + "\n" +
			`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"message":"integration"}}}` + "\n" +
			`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"golang_project_context","arguments":{"max_depth":3}}}` + "\n",
	)
	var output bytes.Buffer

	if err := app.Run(context.Background(), input, &output, slog.Default()); err != nil {
		t.Fatalf("run app: %v", err)
	}

	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	if len(lines) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(lines))
	}

	var listResp map[string]any
	if err := json.Unmarshal(lines[0], &listResp); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if listResp["error"] != nil {
		t.Fatalf("unexpected list error: %v", listResp["error"])
	}

	var callResp map[string]any
	if err := json.Unmarshal(lines[1], &callResp); err != nil {
		t.Fatalf("unmarshal call response: %v", err)
	}
	if callResp["error"] != nil {
		t.Fatalf("unexpected call error: %v", callResp["error"])
	}

	var goCtxResp map[string]any
	if err := json.Unmarshal(lines[2], &goCtxResp); err != nil {
		t.Fatalf("unmarshal golang context response: %v", err)
	}
	if goCtxResp["error"] != nil {
		t.Fatalf("unexpected golang context error: %v", goCtxResp["error"])
	}
}
