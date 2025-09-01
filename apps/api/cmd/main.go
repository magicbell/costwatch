package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ardanlabs/conf/v3"
	"github.com/costwatchai/costwatch/apps/api"
	"github.com/costwatchai/costwatch/internal/appconfig"
	"github.com/magicbell/magicbell/src/gofoundation/logger"
)

const desc = "CostWatch - copyright MagicBell, Inc."

func main() {
	env, ok := os.LookupEnv("APP_ENV")
	if !ok {
		fmt.Println("APP_ENV not set")
		os.Exit(1)
	}

	log, err := logger.New("costwatch.api", env)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := run(log, env); err != nil {
		log.Error("run", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, env string) error {
	cfg, err := appconfig.Init(env, desc)
	if err != nil {
		panic(fmt.Errorf("appconfig.Init: %w", err))
	}

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Debug("startup", "config", out)

	// Monolith
	mono := api.NewMonolith(log, cfg)

	// ===========================================================================
	// Run
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "4000"
	}

	if err := mono.Run(port); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
