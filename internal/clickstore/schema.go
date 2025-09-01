package clickstore

import (
	"context"
	"fmt"
)

func (c *Client) Setup(ctx context.Context, dbName string) error {
	if dbName == "" {
		return fmt.Errorf("setup: empty database name")
	}

	quotedDB := fmt.Sprintf("`%s`", dbName)
	tableFQN := fmt.Sprintf("%s.`metrics`", quotedDB)

	if err := c.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", quotedDB)); err != nil {
		return fmt.Errorf("setup.CreateDB: %w", err)
	}

	if err := c.Exec(ctx, fmt.Sprintf(
		`create table if not exists %s (
			tenant_id String,
			service String,
			metric String,
			value Float64,
			timestamp DateTime64(3, 'UTC')
		)
		ENGINE = ReplacingMergeTree()
		TTL toDateTime(timestamp) + toIntervalDay(90)
		order by (
			tenant_id,
			service,
			metric,
			timestamp
		)
		primary key (
			tenant_id,
			service,
			metric,
			timestamp
		)`,
		tableFQN),
	); err != nil {
		return fmt.Errorf("setup.CreateTable: %w", err)
	}

	return nil
}
