package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/isaacdsc/mcp-server/internal/tools"
)

type Server struct {
	registry *tools.Registry
	logger   *slog.Logger
}

func NewServer(registry *tools.Registry, logger *slog.Logger) *Server {
	return &Server{registry: registry, logger: logger}
}

func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	encoder := json.NewEncoder(out)

	for scanner.Scan() {
		line := scanner.Bytes()
		resp := s.handleLine(ctx, line)
		if resp == nil {
			continue
		}
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	return nil
}

func (s *Server) handleLine(ctx context.Context, line []byte) *Response {
	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		return &Response{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: errCodeParse, Message: "invalid JSON"},
		}
	}

	if req.JSONRPC != "2.0" || req.Method == "" {
		return &Response{
			JSONRPC: "2.0",
			ID:      decodeID(req.ID),
			Error:   &RPCError{Code: errCodeInvalid, Message: "invalid request"},
		}
	}

	result, rpcErr := s.handleMethod(ctx, req.Method, req.Params)
	if rpcErr != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      decodeID(req.ID),
			Error:   rpcErr,
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      decodeID(req.ID),
		Result:  result,
	}
}

func (s *Server) handleMethod(ctx context.Context, method string, params json.RawMessage) (any, *RPCError) {
	switch method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]any{
				"name":    "mcp-go-server",
				"version": "0.1.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		}, nil
	case "notifications/initialized":
		return map[string]any{}, nil
	case "tools/list":
		toolList := s.registry.List()
		sort.Slice(toolList, func(i, j int) bool {
			return toolList[i].Name() < toolList[j].Name()
		})

		items := make([]map[string]any, 0, len(toolList))
		for _, tool := range toolList {
			items = append(items, map[string]any{
				"name":        tool.Name(),
				"description": tool.Description(),
				"inputSchema": tool.InputSchema(),
			})
		}
		return map[string]any{"tools": items}, nil
	case "tools/call":
		var payload struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(params, &payload); err != nil {
			return nil, &RPCError{Code: errCodeBadParams, Message: "invalid tools/call params"}
		}
		if payload.Name == "" {
			return nil, &RPCError{Code: errCodeBadParams, Message: "tool name is required"}
		}
		if len(payload.Arguments) == 0 {
			payload.Arguments = json.RawMessage(`{}`)
		}

		content, err := s.registry.Call(ctx, payload.Name, payload.Arguments)
		if err != nil {
			if errors.Is(err, tools.ErrToolNotFound) {
				return nil, &RPCError{Code: errCodeNotFound, Message: err.Error()}
			}
			return nil, &RPCError{Code: errCodeInternal, Message: "tool execution failed"}
		}

		serialized, err := json.Marshal(content)
		if err != nil {
			return nil, &RPCError{Code: errCodeInternal, Message: "failed to serialize tool result"}
		}

		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": string(serialized),
				},
			},
		}, nil
	default:
		return nil, &RPCError{Code: errCodeNotFound, Message: "method not found"}
	}
}

func decodeID(idRaw json.RawMessage) interface{} {
	if len(idRaw) == 0 {
		return nil
	}

	var id interface{}
	if err := json.Unmarshal(idRaw, &id); err != nil {
		return nil
	}
	return id
}
