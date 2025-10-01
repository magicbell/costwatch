package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/tailbits/costwatch/internal/clickstore"
	cwapi "github.com/tailbits/costwatch/internal/costwatch/api"
	"github.com/tailbits/costwatch/internal/monolith"
	"github.com/tailbits/costwatch/internal/spec"
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

// getenv returns the value of the environment variable key or def if not set.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getenvInt returns the integer value of the environment variable key or def if not set/invalid.
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func seedClickhouse(ctx context.Context, log *slog.Logger) error {
	// Read ClickHouse connection details from environment with sensible defaults.
	// This allows docker-compose to pass CLICKHOUSE_HOST=clickhouse so we don't
	// accidentally connect to localhost/::1 inside the container.
	host := getenv("CLICKHOUSE_HOST", "localhost")
	port := getenvInt("CLICKHOUSE_PORT", 9000)
	user := getenv("CLICKHOUSE_USERNAME", "default")
	pass := getenv("CLICKHOUSE_PASSWORD", "password")
	// Use the built-in "default" DB for seeding so we can create the target DB if missing.
	db := getenv("CLICKHOUSE_DATABASE", "default")

	cfg := clickstore.Config{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		Database: db,
	}

	c, err := clickstore.NewClient(ctx, log, cfg)
	if err != nil {
		return fmt.Errorf("clickstore.NewClient: %w", err)
	}
	defer c.Close()

	// Create target database and tables
	targetDB := getenv("CLICKHOUSE_TARGET_DATABASE", "costwatch")
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
