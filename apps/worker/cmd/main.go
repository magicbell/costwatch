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
	"github.com/costwatchai/costwatch/internal/appconfig"
	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/costwatchai/costwatch/internal/costwatch"
	"github.com/costwatchai/costwatch/internal/health"
	"github.com/costwatchai/costwatch/internal/monolith"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch/metric"
	"github.com/costwatchai/costwatch/internal/scheduler"
)

const desc = "CostWatch - copyright MagicBell, Inc."

func main() {
	env, ok := os.LookupEnv("APP_ENV")
	if !ok {
		fmt.Println("APP_ENV not set")
		os.Exit(1)
	}

	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if err := run(log, env); err != nil {
		log.Error("Failed to run CostWatch Worker", "error", err.Error())
		os.Exit(1)
	}
}

func run(log *slog.Logger, env string) error {
	ctx := context.Background()
	cfg, err := appconfig.Init(env, desc)
	if err != nil {
		panic(fmt.Errorf("appconfig.Init: %w", err))
	}

	log.Info("Starting CostWatch...")

	// SetLastSync up cancellable context with OS signals for graceful shutdown.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Server: run in background to avoid blocking the scheduler
	srv := monolith.NewServer(log, monolith.ServerOptions{
		VersionPrefix: "v1",
		EnableCORS:    false,
		LambdaAware:   true,
		DefaultPort:   "4001",
	}, health.SetupRoutes)

	go func() {
		if err := srv.Run(); err != nil {
			log.Error("worker server exited", "error", err)
		}
	}()

	cs, err := clickstore.NewClient(ctx, log, cfg.Clickhouse)
	if err != nil {
		return fmt.Errorf("clickstore.NewClient: %w", err)
	}
	defer cs.Close()

	// ===========================================================================
	// CostWatch
	cw, err := costwatch.New(ctx, log, cs)
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
	costwatch.RegisterService(svc)

	ib := metric.NewIncomingBytes(log.WithGroup("incoming_bytes"), awscloudwatch.NewFromConfig(awsCfg))
	svc.NewMetric(ib)

	// Leading sync at startup
	if err := cw.Sync(ctx); err != nil {
		log.Error("leading sync failed", "error", err)
	}

	// Scheduler: run every 30s with CostWatch.Sync
	interval := time.Second * 30
	tickLog := log.WithGroup("scheduler")
	lt := scheduler.NewLocalTicker(tickLog, interval, cw.Sync)
	lt.Start(ctx)
	defer lt.Stop()

	// Block until shutdown signal.
	<-ctx.Done()
	log.Info("CostWatch worker shutting down", "reason", ctx.Err())
	return nil
}
