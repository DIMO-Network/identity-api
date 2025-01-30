package helpers

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"testing"
	"time"

	"github.com/DIMO-Network/shared/db"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var DBSettings = db.Settings{
	Name:     "identity_api",
	Host:     "localhost",
	User:     "dimo",
	Password: "dimo",
	// MaxOpenConnections: 2,
	// MaxIdleConnections: 2,
}

// StartContainerDatabase starts postgres container with default test settings, and migrates the db. Caller must terminate container.
func StartContainerDatabase(ctx context.Context, t *testing.T, migrationsDirRelPath string) (dbs db.Store, container *postgres.PostgresContainer) {
	settings := DBSettings // Copy.
	pgPort := "5432/tcp"

	var err error

	container, err = postgres.Run(
		ctx,
		"docker.io/postgres:16.6-alpine",
		postgres.WithDatabase("identity_api"),
		postgres.WithUsername("dimo"),
		postgres.WithPassword("dimo"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	defer func() {
		if err != nil {
			container.Terminate(ctx) //nolint:errcheck
			t.Fatalf("Failed to set up Postgres container: %v", err)
		}
	}()

	mappedPort, err := container.MappedPort(ctx, nat.Port(pgPort))
	if err != nil {
		err = fmt.Errorf("couldn't get Postgres port: %w", err)
		return
	}

	settings.Port = mappedPort.Port()
	dbs = db.NewDbConnectionForTest(ctx, &settings, false)
	for !dbs.IsReady() {
		time.Sleep(500 * time.Millisecond)
	}

	_, err = dbs.DBS().Writer.Exec("CREATE SCHEMA IF NOT EXISTS identity_api")
	if err != nil {
		err = fmt.Errorf("error creating schema: %w", err)
		return
	}

	goose.SetTableName("identity_api.migrations")
	err = goose.RunContext(ctx, "up", dbs.DBS().Writer.DB, migrationsDirRelPath)
	if err != nil {
		err = fmt.Errorf("error running migrations: %w", err)
		return
	}

	err = container.Snapshot(ctx, postgres.WithSnapshotName("dimo_snapshot"))
	if err != nil {
		err = fmt.Errorf("error taking blank snapshot: %w", err)
	}

	return
}

func GenerateWallet() (*ecdsa.PrivateKey, *common.Address, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	userAddr := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, &userAddr, nil
}
