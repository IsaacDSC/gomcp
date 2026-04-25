package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrToolNotFound = errors.New("tool not found")

type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]any
	Call(ctx context.Context, arguments json.RawMessage) (map[string]any, error)
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry(toolList ...Tool) *Registry {
	registry := &Registry{tools: make(map[string]Tool, len(toolList))}
	for _, tool := range toolList {
		registry.tools[tool.Name()] = tool
	}
	return registry
}

func (r *Registry) List() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

func (r *Registry) Call(ctx context.Context, name string, arguments json.RawMessage) (map[string]any, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}
	return tool.Call(ctx, arguments)
}

type EchoTool struct{}

func (EchoTool) Name() string { return "echo" }

func (EchoTool) Description() string {
	return "Echoes the provided message."
}

func (EchoTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"message": map[string]any{
				"type":        "string",
				"description": "Message to echo back.",
			},
		},
		"required": []string{"message"},
	}
}

func (EchoTool) Call(_ context.Context, arguments json.RawMessage) (map[string]any, error) {
	var input struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, fmt.Errorf("parse echo input: %w", err)
	}
	return map[string]any{"message": input.Message}, nil
}

type TimestampTool struct{}

func (TimestampTool) Name() string { return "timestamp" }

func (TimestampTool) Description() string {
	return "Returns current timestamp in RFC3339."
}

func (TimestampTool) InputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (TimestampTool) Call(_ context.Context, _ json.RawMessage) (map[string]any, error) {
	return map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}
