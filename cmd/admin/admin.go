package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/costwatchai/costwatch/internal/clickstore"
	cwapi "github.com/costwatchai/costwatch/internal/costwatch/api"
	"github.com/costwatchai/costwatch/internal/monolith"
	"github.com/costwatchai/costwatch/internal/spec"
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
	case "openapi":
		if err := printOpenAPISpec(context.Background(), log); err != nil {
			log.Error("Failed to generate OpenAPI spec", "error", err.Error())
			os.Exit(1)
		}
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
	fmt.Fprintln(os.Stderr, "  openapi\t\tPrint OpenAPI 3.1 spec to stdout")
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

func printOpenAPISpec(ctx context.Context, log *slog.Logger) error {
	// Build a Mason API using a lightweight server (no HTTP listener needed)
	srv := monolith.NewServer(log, monolith.ServerOptions{})

	// Register CostWatch API routes using a non-connecting test store
	test := clickstore.NewTestStore(log)
	client := &clickstore.Client{Conn: test}
	cw, err := cwapi.New(ctx, log, client)
	if err != nil {
		return fmt.Errorf("api.New: %w", err)
	}
	cw.SetupRoutes(srv.API)

	// Generate and print the OpenAPI document
	data, err := spec.Generate(srv.API)
	if err != nil {
		return fmt.Errorf("spec.Generate: %w", err)
	}
	os.Stdout.Write(data)
	os.Stdout.Write([]byte("\n"))
	return nil
}
