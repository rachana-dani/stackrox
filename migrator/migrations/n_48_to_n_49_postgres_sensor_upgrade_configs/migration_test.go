// Code generated by pg-bindings generator. DO NOT EDIT.

//go:build sql_integration

package n48ton49

import (
	"context"
	"testing"

	"github.com/stackrox/rox/generated/storage"
	legacy "github.com/stackrox/rox/migrator/migrations/n_48_to_n_49_postgres_sensor_upgrade_configs/legacy"
	pgStore "github.com/stackrox/rox/migrator/migrations/n_48_to_n_49_postgres_sensor_upgrade_configs/postgres"
	pghelper "github.com/stackrox/rox/migrator/migrations/postgreshelper"

	"github.com/stackrox/rox/pkg/bolthelper"
	"github.com/stackrox/rox/pkg/sac"

	"github.com/stackrox/rox/pkg/features"

	"github.com/stackrox/rox/pkg/testutils"
	"github.com/stackrox/rox/pkg/testutils/envisolator"

	"github.com/stretchr/testify/suite"

	bolt "go.etcd.io/bbolt"
)

func TestMigration(t *testing.T) {
	suite.Run(t, new(postgresMigrationSuite))
}

type postgresMigrationSuite struct {
	suite.Suite
	envIsolator *envisolator.EnvIsolator
	ctx         context.Context

	legacyDB   *bolt.DB
	postgresDB *pghelper.TestPostgres
}

var _ suite.TearDownTestSuite = (*postgresMigrationSuite)(nil)

func (s *postgresMigrationSuite) SetupTest() {
	s.envIsolator = envisolator.NewEnvIsolator(s.T())
	s.envIsolator.Setenv(features.PostgresDatastore.EnvVar(), "true")
	if !features.PostgresDatastore.Enabled() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	var err error
	s.legacyDB, err = bolthelper.NewTemp(s.T().Name() + ".db")
	s.NoError(err)

	s.Require().NoError(err)

	s.ctx = sac.WithAllAccess(context.Background())
	s.postgresDB = pghelper.ForT(s.T(), true)
}

func (s *postgresMigrationSuite) TearDownTest() {
	testutils.TearDownDB(s.legacyDB)
	s.postgresDB.Teardown(s.T())
}

func (s *postgresMigrationSuite) TestSensorUpgradeConfigMigration() {
	newStore := pgStore.New(s.ctx, s.postgresDB.Pool)
	legacyStore := legacy.New(s.legacyDB)

	// Prepare data and write to legacy DB
	sensorUpgradeConfig := &storage.SensorUpgradeConfig{}
	s.NoError(testutils.FullInit(sensorUpgradeConfig, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
	s.NoError(legacyStore.Upsert(s.ctx, sensorUpgradeConfig))

	// Move
	s.NoError(move(s.postgresDB.GetGormDB(), s.postgresDB.Pool, legacyStore))

	// Verify
	fetched, found, err := newStore.Get(s.ctx)
	s.NoError(err)
	s.True(found)
	s.Equal(sensorUpgradeConfig, fetched)
}
