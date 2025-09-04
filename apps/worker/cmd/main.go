package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	awscloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/costwatchai/costwatch/internal/costwatch"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch/metric"
	"github.com/costwatchai/costwatch/internal/scheduler"
)

func main() {
	lh := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	log := slog.New(lh)

	if err := run(context.Background(), log); err != nil {
		log.Error("Failed to run CostWatch", "error", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	log.Info("Starting CostWatch...")
	fmt.Println("CostWatch started")

	// Set up cancellable context with OS signals for graceful shutdown.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ===========================================================================
	// Clickhouse Connection
	cfg := clickstore.Config{
		Host:     "localhost",
		Port:     9000,
		Username: "default",
		Password: "password",
		Database: "costwatch",
	}

	c, err := clickstore.NewClient(ctx, log, cfg)
	if err != nil {
		return fmt.Errorf("clickstore.NewClient: %w", err)
	}
	defer c.Close()

	// ===========================================================================
	// CostWatch
	tenantID := "default"
	wtc, err := costwatch.New(ctx, log, c, tenantID)
	if err != nil {
		return fmt.Errorf("costwatch.New: %w", err)
	}

	// ===========================================================================
	// AWS Provider
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
	)
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	svc := cloudwatch.NewService(awsCfg)
	wtc.RegisterService(svc)

	ibMtr := metric.NewIncomingBytes(log.WithGroup("incoming_bytes"), awscloudwatch.NewFromConfig(awsCfg))
	svc.NewMetric(ibMtr)

	// Also expose this service/metric for pricing lookups globally.
	costwatch.RegisterGlobalService(svc)

	// ===========================================================================
	// Fetch last 48 hours on startup
	end := time.Now().UTC()
	start := end.Add(-48 * time.Hour)
	log.Info("fetching metrics (backfill)", "start", start, "end", end)
	err = wtc.FetchMetrics(ctx, start, end)
	if err != nil {
		log.Error("wtc.FetchMetrics: to fetch metrics %w", err.Error())
	}

	// Scheduler: fetch last hour by convention every 30s
	interval := time.Second * 30
	tickLog := log.WithGroup("scheduler")
	lt := scheduler.NewLocalTicker(tickLog, interval, func(jctx context.Context) error {
		end = time.Now().UTC()
		start = end.Add(-1 * time.Hour)
		tickLog.Info("fetching metrics (periodic)", "start", start, "end", end)
		return wtc.FetchMetrics(jctx, start, end)
	})
	lt.Start(ctx)
	defer lt.Stop()

	// Block until shutdown signal.
	<-ctx.Done()
	log.Info("CostWatch worker shutting down", "reason", ctx.Err())
	return nil
}
