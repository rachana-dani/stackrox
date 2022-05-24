// Code generated by pg-bindings generator. DO NOT EDIT.

//go:build sql_integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/features"
	"github.com/stackrox/rox/pkg/postgres/pgtest"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/testutils"
	"github.com/stackrox/rox/pkg/testutils/envisolator"
	"github.com/stretchr/testify/suite"
)

type TestChild1StoreSuite struct {
	suite.Suite
	envIsolator *envisolator.EnvIsolator
	store       Store
	pool        *pgxpool.Pool
}

func TestTestChild1Store(t *testing.T) {
	suite.Run(t, new(TestChild1StoreSuite))
}

func (s *TestChild1StoreSuite) SetupTest() {
	s.envIsolator = envisolator.NewEnvIsolator(s.T())
	s.envIsolator.Setenv(features.PostgresDatastore.EnvVar(), "true")

	if !features.PostgresDatastore.Enabled() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	ctx := sac.WithAllAccess(context.Background())

	source := pgtest.GetConnectionString(s.T())
	config, err := pgxpool.ParseConfig(source)
	s.Require().NoError(err)
	pool, err := pgxpool.ConnectConfig(ctx, config)
	s.Require().NoError(err)

	Destroy(ctx, pool)

	s.pool = pool
	s.store = New(ctx, pool)
}

func (s *TestChild1StoreSuite) TearDownTest() {
	if s.pool != nil {
		s.pool.Close()
	}
	s.envIsolator.RestoreAll()
}

func (s *TestChild1StoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	testChild1 := &storage.TestChild1{}
	s.NoError(testutils.FullInit(testChild1, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundTestChild1, exists, err := store.Get(ctx, testChild1.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundTestChild1)

	s.NoError(store.Upsert(ctx, testChild1))
	foundTestChild1, exists, err = store.Get(ctx, testChild1.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(testChild1, foundTestChild1)

	testChild1Count, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, testChild1Count)

	testChild1Exists, err := store.Exists(ctx, testChild1.GetId())
	s.NoError(err)
	s.True(testChild1Exists)
	s.NoError(store.Upsert(ctx, testChild1))

	foundTestChild1, exists, err = store.Get(ctx, testChild1.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(testChild1, foundTestChild1)

	s.NoError(store.Delete(ctx, testChild1.GetId()))
	foundTestChild1, exists, err = store.Get(ctx, testChild1.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundTestChild1)

	var testChild1s []*storage.TestChild1
	for i := 0; i < 200; i++ {
		testChild1 := &storage.TestChild1{}
		s.NoError(testutils.FullInit(testChild1, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		testChild1s = append(testChild1s, testChild1)
	}

	s.NoError(store.UpsertMany(ctx, testChild1s))

	testChild1Count, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, testChild1Count)
}
