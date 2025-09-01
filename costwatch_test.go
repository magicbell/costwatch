package costwatch_test

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/costwatchai/costwatch"
	"github.com/costwatchai/costwatch/internal/clickstore"
	"github.com/costwatchai/costwatch/internal/provider/aws/cloudwatch"
	"github.com/costwatchai/costwatch/internal/provider/testprovider"
	"gotest.tools/v3/assert"
)

func TestCostwatch(t *testing.T) {
	lh := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	log := slog.New(lh)
	ctx := context.Background()

	// ===========================================================================
	// Billable Service
	svc := cloudwatch.NewService(aws.Config{})
	assert.Equal(t, len(svc.Metrics()), 0)

	// ===========================================================================
	// Billable Metric
	mtr := testprovider.NewTestMetric(t)
	svc.NewMetric(mtr)

	assert.Equal(t, len(svc.Metrics()), 1)
	assert.Equal(t, svc.Metrics()[0].Label(), "IncomingBytes")

	startTS, err := time.Parse(time.RFC3339, "2025-08-22T11:11:00Z")
	assert.NilError(t, err)

	endTS, err := time.Parse(time.RFC3339, "2025-08-30T00:11:00Z")
	assert.NilError(t, err)

	dps, err := mtr.Datapoints(ctx, mtr.Label(), startTS, endTS)
	assert.NilError(t, err)
	assert.Equal(t, len(dps), 168)

	// ===========================================================================
	// Clickhouse Connection
	cfg := clickstore.Config{
		Host:     "localhost",
		Port:     9000,
		Username: "default",
		Password: "password",
	}

	c, err := clickstore.NewClient(ctx, log, cfg)
	assert.NilError(t, err)

	// Setup the database
	dbName := fmt.Sprintf("test_%d", rand.IntN(9000)+1000)

	err = c.Setup(ctx, dbName)
	assert.NilError(t, err)

	err = c.Close()
	assert.NilError(t, err)

	// switch database
	cfg.Database = dbName
	c, err = clickstore.NewClient(ctx, log, cfg)
	assert.NilError(t, err)

	defer c.Exec(ctx, "DROP DATABASE IF EXISTS %s", dbName)

	// ===========================================================================
	// Costwatch
	wtc, err := costwatch.New(ctx, log, c, "test-tenant")
	assert.NilError(t, err)

	// ===========================================================================
	// Register Service
	err = wtc.RegisterService(svc)
	assert.NilError(t, err)

	err = wtc.RegisterService(svc)

	// ===========================================================================
	// Run Costwatch Data Fetch
	err = wtc.FetchMetrics(ctx, startTS, endTS)
	assert.NilError(t, err)

	usg, err := wtc.ServiceUsage(ctx, "aws.CloudWatch", startTS, endTS)
	assert.NilError(t, err)
	assert.Equal(t, len(usg), 1)

	assert.Assert(t, usg["IncomingBytes"] != nil)
	assert.Equal(t, usg["IncomingBytes"].Units(), 9.170550853037e+12)
	assert.Equal(t, usg["IncomingBytes"].Cost(), 458527.54265184986)
}
