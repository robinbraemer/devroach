package devroach

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"io/fs"
	"testing"
)

// MigrateT runs all migrations against the given connection.
func MigrateT(t *testing.T, pool *pgxpool.Pool, migrationsFS fs.FS, globs ...string) {
	t.Helper()
	err := Migrate(context.TODO(), pool, migrationsFS, globs...)
	require.NoError(t, err)
}

// Migrate runs all migrations against the given connection.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsFS fs.FS, globs ...string) error {
	all, err := allContents(migrationsFS, globs...)
	if err != nil {
		return err
	}
	if len(all) != 0 {
		logr.FromContextOrDiscard(ctx).Info("applying migrations", "count", len(all))
	}
	for _, sql := range all {
		_, err = pool.Exec(ctx, sql)
		if err != nil {
			return err
		}
	}
	return nil
}

func allContents(migrationsFS fs.FS, globs ...string) ([]string, error) {
	var contents []string
	for _, glob := range globs {
		matches, err := fs.Glob(migrationsFS, glob)
		if err != nil {
			return nil, err
		}
		for _, file := range matches {
			c, err := fs.ReadFile(migrationsFS, file)
			if err != nil {
				return nil, err
			}
			contents = append(contents, string(c))
		}
	}
	return contents, nil
}
