package clickstore

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed sql/schema.sql
var schemaSQL string

func (c *Client) Setup(ctx context.Context, dbName string) error {
	if dbName == "" {
		return fmt.Errorf("setup: empty database name")
	}

	quotedDB := fmt.Sprintf("`%s`", dbName)
	tableFQN := fmt.Sprintf("%s.`metrics`", quotedDB)

	if err := c.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", quotedDB)); err != nil {
		return fmt.Errorf("setup.CreateDB: %w", err)
	}

	if err := c.Exec(ctx, fmt.Sprintf(schemaSQL, tableFQN)); err != nil {
		return fmt.Errorf("setup.CreateTable: %w", err)
	}

	return nil
}
