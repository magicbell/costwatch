package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	awscloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/costwatchai/costwatch"
	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch/metric"
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

	// ===========================================================================
	// Watch Costs
	start := time.Now().UTC().Add(-2 * 24 * time.Hour).Truncate(24 * time.Hour)
	end := time.Now().UTC()
	log.Debug("fetching metrics", "start", start, "end", end)

	if err := wtc.FetchMetrics(ctx, start, end); err != nil {
		return fmt.Errorf("wtc.FetchMetrics: %w", err)
	}

	usg, err := wtc.ServiceUsage(ctx, svc.Label(), time.Now().Add(-1*time.Hour), time.Now())
	if err != nil {
		return fmt.Errorf("wtc.ServiceUsage: %w", err)
	}

	for l, m := range usg {
		log.Debug("Usage", "label", l, "cost", m.Cost(), "usage", m.Units())
	}

	return nil
}
