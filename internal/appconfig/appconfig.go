package appconfig

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/ardanlabs/conf/v3"
	"github.com/costwatchai/costwatch/internal/clickstore"
)

type Config struct {
	conf.Version
	Args conf.Args `json:"args"`
	App  struct {
		Env     string `conf:"required, help:app env determines the app config (development, review, production)"`
		Stage   string `conf:"default:development, help:pipeline stage (development, review, production)"`
		Version string `conf:"default:0.0.1,       help:code version used for deployment and error tracking"`
	}
	Clickhouse clickstore.Config `conf:"help:ClickHouse connection config"`
}

func Init(env, desc string) (Config, error) {
	version := conf.Version{
		Build: env,
		Desc:  desc,
	}

	cfg := Config{
		Version: version,
	}

	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return Config{}, fmt.Errorf("conf.Parse: %s", help)
		}

		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}
