//go:build integration

package integration

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage = "postgres:16-alpine"
	testDB       = "grbac_test"
	testUser     = "postgres"
	testPassword = "postgres"
)

// SetupTestDB spins up a PostgreSQL container, runs migrations, and returns
// the DATABASE_URL and a cleanup function. Requires Docker.
func SetupTestDB(t *testing.T) (dbURL string, cleanup func()) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, postgresImage,
		postgres.WithDatabase(testDB),
		postgres.WithUsername(testUser),
		postgres.WithPassword(testPassword),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err, "failed to start postgres container")

	cleanup = func() {
		require.NoError(t, pgContainer.Terminate(ctx), "failed to terminate postgres container")
	}

	dbURL, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	// Resolve migrations path relative to module root (parent of tests/)
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "db", "migrations")
	migrationsPath, err = filepath.Abs(migrationsPath)
	require.NoError(t, err, "failed to resolve migrations path")

	m, err := migrate.New("file://"+migrationsPath, dbURL)
	require.NoError(t, err, "failed to create migrate instance")
	defer m.Close()

	require.NoError(t, m.Up(), "failed to run migrations")

	return dbURL, cleanup
}
