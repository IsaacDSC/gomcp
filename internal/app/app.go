package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/isaacdsc/mcp-server/internal/mcp"
	"github.com/isaacdsc/mcp-server/internal/tools"
)

func Run(ctx context.Context, in io.Reader, out io.Writer, logger *slog.Logger) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	registry := tools.NewRegistry(
		tools.EchoTool{},
		tools.TimestampTool{},
		tools.NewGolangProjectContextTool(workingDir),
	)

	server := mcp.NewServer(registry, logger)
	return server.Serve(ctx, in, out)
}
