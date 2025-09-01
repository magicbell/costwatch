package clickstore

import (
	"context"
	"fmt"
)

func (c *Client) Setup(ctx context.Context, dbName string) error {
	// Implementation for creating a new schema in Clickhouse

	if err := c.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)); err != nil {
		return fmt.Errorf("setup.CreateDB: %w", err)
	}

	if err := c.Exec(ctx, fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
		service String,
		label String,  
		value Float64,  
		timestamp DateTime64(3, 'UTC'),  
		) 
		ENGINE = MergeTree()
				TTL toDateTime(timestamp) + toIntervalDay(90)
				ORDER BY (
					service,
					label,
					toStartOfDay(timestamp), 
					timestamp
				)
				PRIMARY KEY (
					service,
					label,
					toStartOfDay(timestamp), 
				)`,
		dbName+".metrics"),
	); err != nil {
		return fmt.Errorf("setup.CreateTable: %w", err)
	}

	return nil
}
