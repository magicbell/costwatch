package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/costwatchai/costwatch/internal/appconfig"
	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/costwatchai/costwatch/internal/costwatch"
	costwatchapi "github.com/costwatchai/costwatch/internal/costwatch/api"
	"github.com/costwatchai/costwatch/internal/health"
	"github.com/costwatchai/costwatch/internal/monolith"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch/metric"
	cg "github.com/costwatchai/costwatch/internal/provider/coingecko"
	cgm "github.com/costwatchai/costwatch/internal/provider/coingecko/metric"
	"github.com/costwatchai/costwatch/internal/spec"
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

	// Register services/metrics for pricing-only usage (no AWS calls needed here).
	svc := cloudwatch.NewService(aws.Config{})
	ib := metric.NewIncomingBytes(log.WithGroup("incoming_bytes"), nil)
	svc.NewMetric(ib)
	costwatch.RegisterService(svc)

	// CoinGecko provider with BTC/USD metrics (disabled when DEMO=false)
	cgSvc := cg.NewService()
	btc := cgm.NewPriceMetric(log.WithGroup("btc_usd"), nil)
	cgSvc.NewMetric(btc)
	costwatch.RegisterService(cgSvc)

	cs, err := clickstore.NewClient(ctx, log, cfg.Clickhouse)
	if err != nil {
		return fmt.Errorf("clickstore.NewClient: %w", err)
	}

	cw, err := costwatchapi.New(ctx, log, cs)
	if err != nil {
		return fmt.Errorf("costwatch.New: %w", err)
	}

	// Server
	srv := monolith.NewServer(log, monolith.ServerOptions{
		VersionPrefix: "v1",
		EnableCORS:    true,
		LambdaAware:   true,
		DefaultPort:   "4000",
	}, health.SetupRoutes, spec.SetupRoutes, cw.SetupRoutes)

	if err := srv.Run(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
