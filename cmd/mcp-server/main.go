package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/isaacdsc/mcp-server/internal/app"
	"github.com/isaacdsc/mcp-server/internal/config"
	"github.com/isaacdsc/mcp-server/internal/observability"
)

func main() {
	var healthcheckOnly bool
	flag.BoolVar(&healthcheckOnly, "healthcheck", false, "run a one-shot healthcheck and exit")
	flag.Parse()

	if healthcheckOnly {
		fmt.Println("ok")
		return
	}

	cfg := config.Load()
	logger := observability.NewLogger(cfg.LogLevel)

	if err := app.Run(context.Background(), os.Stdin, os.Stdout, logger); err != nil {
		logger.Error("server stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
