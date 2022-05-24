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

type TestGrandparentsStoreSuite struct {
	suite.Suite
	envIsolator *envisolator.EnvIsolator
	store       Store
	pool        *pgxpool.Pool
}

func TestTestGrandparentsStore(t *testing.T) {
	suite.Run(t, new(TestGrandparentsStoreSuite))
}

func (s *TestGrandparentsStoreSuite) SetupTest() {
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

func (s *TestGrandparentsStoreSuite) TearDownTest() {
	if s.pool != nil {
		s.pool.Close()
	}
	s.envIsolator.RestoreAll()
}

func (s *TestGrandparentsStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	testGrandparent := &storage.TestGrandparent{}
	s.NoError(testutils.FullInit(testGrandparent, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundTestGrandparent, exists, err := store.Get(ctx, testGrandparent.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundTestGrandparent)

	s.NoError(store.Upsert(ctx, testGrandparent))
	foundTestGrandparent, exists, err = store.Get(ctx, testGrandparent.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(testGrandparent, foundTestGrandparent)

	testGrandparentCount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, testGrandparentCount)

	testGrandparentExists, err := store.Exists(ctx, testGrandparent.GetId())
	s.NoError(err)
	s.True(testGrandparentExists)
	s.NoError(store.Upsert(ctx, testGrandparent))

	foundTestGrandparent, exists, err = store.Get(ctx, testGrandparent.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(testGrandparent, foundTestGrandparent)

	s.NoError(store.Delete(ctx, testGrandparent.GetId()))
	foundTestGrandparent, exists, err = store.Get(ctx, testGrandparent.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundTestGrandparent)

	var testGrandparents []*storage.TestGrandparent
	for i := 0; i < 200; i++ {
		testGrandparent := &storage.TestGrandparent{}
		s.NoError(testutils.FullInit(testGrandparent, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		testGrandparents = append(testGrandparents, testGrandparent)
	}

	s.NoError(store.UpsertMany(ctx, testGrandparents))

	testGrandparentCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, testGrandparentCount)
}
