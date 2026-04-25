package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	var (
		image      = flag.String("image", "mcp-go-server:local", "Docker image with MCP server")
		workspace  = flag.String("workspace", ".", "Host workspace path to mount at /workspace")
		tool       = flag.String("tool", "timestamp", "MCP tool to call: echo|timestamp|golang_project_context")
		message    = flag.String("message", "ola", "Message for echo tool")
		maxDepth   = flag.Int("max-depth", 4, "max_depth for golang_project_context")
		withFiles  = flag.Bool("include-files", false, "include_files for golang_project_context")
		logLevel   = flag.String("log-level", "INFO", "LOG_LEVEL passed to container")
		showPretty = flag.Bool("pretty", true, "Pretty-print tool result")
	)
	flag.Parse()

	absWorkspace, err := filepath.Abs(*workspace)
	if err != nil {
		exitErr(fmt.Errorf("resolve workspace path: %w", err))
	}

	toolArgs, err := buildToolArgs(*tool, *message, *maxDepth, *withFiles)
	if err != nil {
		exitErr(err)
	}

	if err := run(*image, absWorkspace, *logLevel, *tool, toolArgs, *showPretty); err != nil {
		exitErr(err)
	}
}

func buildToolArgs(tool, message string, maxDepth int, includeFiles bool) (map[string]any, error) {
	switch tool {
	case "echo":
		return map[string]any{"message": message}, nil
	case "timestamp":
		return map[string]any{}, nil
	case "golang_project_context":
		return map[string]any{
			"workspace_path": "/workspace",
			"max_depth":      maxDepth,
			"include_files":  includeFiles,
		}, nil
	default:
		return nil, fmt.Errorf("tool invalida %q (use: echo|timestamp|golang_project_context)", tool)
	}
}

func run(image, workspace, logLevel, tool string, args map[string]any, pretty bool) error {
	cmd := exec.Command(
		"docker", "run", "--rm", "-i",
		"-v", fmt.Sprintf("%s:/workspace:ro", workspace),
		"-w", "/workspace",
		"-e", fmt.Sprintf("LOG_LEVEL=%s", logLevel),
		image,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("open stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start docker process: %w", err)
	}

	reader := bufio.NewReader(stdout)

	if err := sendAndWait(stdin, reader, rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "example-cli",
				"version": "1.0.0",
			},
		},
	}, 1, nil); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("initialize failed: %w", err)
	}

	if err := writeRequest(stdin, rpcRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]any{},
	}); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("send initialized notification: %w", err)
	}

	var result json.RawMessage
	if err := sendAndWait(stdin, reader, rpcRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      tool,
			"arguments": args,
		},
	}, 2, &result); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("tools/call failed: %w", err)
	}

	if err := stdin.Close(); err != nil {
		return fmt.Errorf("close stdin: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wait docker process: %w", err)
	}

	if pretty {
		var out any
		if err := json.Unmarshal(result, &out); err != nil {
			return fmt.Errorf("decode tool result: %w", err)
		}
		prettyJSON, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("format result json: %w", err)
		}
		fmt.Println(string(prettyJSON))
		return nil
	}

	fmt.Println(string(result))
	return nil
}

func sendAndWait(stdin io.Writer, reader *bufio.Reader, req rpcRequest, expectedID int, resultOut *json.RawMessage) error {
	if err := writeRequest(stdin, req); err != nil {
		return err
	}

	for {
		resp, err := readResponse(reader)
		if err != nil {
			return err
		}

		if resp.ID == nil || *resp.ID != expectedID {
			continue
		}

		if resp.Error != nil {
			return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		if resultOut != nil {
			*resultOut = resp.Result
		}
		return nil
	}
}

func writeRequest(w io.Writer, req rpcRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	data = append(data, '\n')

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write request: %w", err)
	}
	return nil
}

func readResponse(reader *bufio.Reader) (*rpcResponse, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("unexpected EOF reading rpc response")
		}
		return nil, fmt.Errorf("read rpc response: %w", err)
	}

	var resp rpcResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("decode rpc response: %w", err)
	}
	return &resp, nil
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "erro: %v\n", err)
	os.Exit(1)
}
