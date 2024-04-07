package devroach

import (
	"context"
	"fmt"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"io/fs"
	"testing"
)

// NewPoolT creates a new pgxpool.Pool instance for testing.
//
// It is backed by a new in-memory cockroachdb instance and migrations are applied automatically.
// When the test is finished, the cockroachdb instance is stopped and connections are closed automatically.
func NewPoolT(t *testing.T, migrationsFS fs.FS, globs ...string) *pgxpool.Pool {
	t.Helper()
	ts, err := testserver.NewTestServer()
	require.NoError(t, err)

	t.Cleanup(ts.Stop)

	pool, err := pgxpool.New(context.Background(), ts.PGURL().String())
	require.NoError(t, err)

	t.Cleanup(pool.Close)

	MigrateT(t, pool, migrationsFS, globs...)

	return pool
}

const testPort = 26257

// NewPool creates a new pgxpool.Pool instance for testing.
// 1. Check if the server is already running.
// 2. If not, start a new test server.
// 3. Create a new pool.
// 4. Apply migrations.
func NewPool(ctx context.Context, migrationsFS fs.FS, globs ...string) (pool *pgxpool.Pool, clean func(), err error) {
	log := logr.FromContextOrDiscard(ctx)

	// check if server is already running
	url := fmt.Sprintf("postgresql://root@localhost:%d/defaultdb?sslmode=disable", testPort)
	pool, err = pgxpool.New(ctx, url)
	if err == nil && pool.Ping(ctx) == nil {
		log.Info("using existing test db", "url", url)
		return pool, pool.Close, nil
	}

	// start new test server
	ts, err := StartTestServer(ctx)
	if err != nil {
		return nil, nil, err
	}
	cleanup := ts.Stop

	dbURL := ts.PGURL().String()
	log.Info("using test db", "url", dbURL)

	// create new pool
	pool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, cleanup, err
	}
	cleanup2 := func() { cleanup(); pool.Close() }

	// apply migrations
	if err = Migrate(ctx, pool, migrationsFS, globs...); err != nil {
		return nil, cleanup2, err
	}

	return pool, cleanup2, nil
}

// StartTestServer starts a new test server if not already running.
func StartTestServer(ctx context.Context) (testserver.TestServer, error) {
	// start new test server
	ts, err := testserver.NewTestServer(testserver.AddListenAddrPortOpt(testPort))
	if err != nil {
		return nil, err
	}
	log := logr.FromContextOrDiscard(ctx)
	log.Info("started test server", "url", ts.PGURL().String())
	return ts, nil
}
