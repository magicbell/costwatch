package clickstore

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Config struct {
	Host     string `conf:"default:localhost,help:ClickHouse host"`
	Port     int    `conf:"default:9000,help:ClickHouse port"`
	Username string `conf:"default:default,help:ClickHouse username"`
	Password string `conf:"default:password,help:ClickHouse password,noprint"`
	Database string `conf:"default:costwatch,help:ClickHouse database"`
	SSL      bool   `conf:"help:Use secure connection to ClickHouse"`
	Conn     struct {
		Name    string `conf:"default:costwatch,help:ClickHouse connection name"`
		Version string `conf:"default:1.0.0,help:ClickHouse connection version"`
	}
}

type Client struct {
	driver.Conn
}

func NewClient(ctx context.Context, log *slog.Logger, cfg Config) (*Client, error) {
	var ssl *tls.Config
	if cfg.SSL {
		ssl = &tls.Config{}
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{
					Name:    cfg.Conn.Name,
					Version: cfg.Conn.Version,
				},
			},
		},
		Debugf: func(format string, v ...interface{}) {
			log.Debug(fmt.Sprintf(format, v...))
		},
		TLS: ssl,
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse.Open: %w", err)
	}

	c := &Client{
		Conn: conn,
	}

	return c, nil
}
