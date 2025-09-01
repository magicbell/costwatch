package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/costwatchai/costwatch/internal/clickstore"
)

func main() {
	lh := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	log := slog.New(lh)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]

	switch cmd {
	case "seed-clickhouse":
		if err := seedClickhouse(context.Background(), log); err != nil {
			log.Error("Failed to seed ClickHouse", "error", err.Error())
			os.Exit(1)
		}
		log.Info("ClickHouse schema setup complete")
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: go run cmd/admin/admin.go <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Available commands:")
	fmt.Fprintln(os.Stderr, "  seed-clickhouse\tInitialize ClickHouse schema")
	fmt.Fprintln(os.Stderr, "")
}

func seedClickhouse(ctx context.Context, log *slog.Logger) error {
	// Connect to an existing database for setup
	cfg := clickstore.Config{
		Host:     "localhost",
		Port:     9000,
		Username: "default",
		Password: "password",
		Database: "default",
	}

	c, err := clickstore.NewClient(ctx, log, cfg)
	if err != nil {
		return fmt.Errorf("clickstore.NewClient: %w", err)
	}
	defer c.Close()

	// Create target database and tables
	targetDB := "costwatch"
	if err := c.Setup(ctx, targetDB); err != nil {
		return fmt.Errorf("clickstore.Setup: %w", err)
	}

	return nil
}
