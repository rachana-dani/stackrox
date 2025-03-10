// Code generated by pg-bindings generator. DO NOT EDIT.

//go:build sql_integration

package postgres

import (
	"context"
	"testing"

	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/postgres/pgtest"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/testutils"
	"github.com/stackrox/rox/pkg/testutils/envisolator"
	"github.com/stretchr/testify/suite"
)

type SignatureIntegrationsStoreSuite struct {
	suite.Suite
	envIsolator *envisolator.EnvIsolator
	store       Store
	testDB      *pgtest.TestPostgres
}

func TestSignatureIntegrationsStore(t *testing.T) {
	suite.Run(t, new(SignatureIntegrationsStoreSuite))
}

func (s *SignatureIntegrationsStoreSuite) SetupSuite() {
	s.envIsolator = envisolator.NewEnvIsolator(s.T())
	s.envIsolator.Setenv(env.PostgresDatastoreEnabled.EnvVar(), "true")

	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	s.testDB = pgtest.ForT(s.T())
	s.store = New(s.testDB.Pool)
}

func (s *SignatureIntegrationsStoreSuite) SetupTest() {
	ctx := sac.WithAllAccess(context.Background())
	tag, err := s.testDB.Exec(ctx, "TRUNCATE signature_integrations CASCADE")
	s.T().Log("signature_integrations", tag)
	s.NoError(err)
}

func (s *SignatureIntegrationsStoreSuite) TearDownSuite() {
	s.testDB.Teardown(s.T())
	s.envIsolator.RestoreAll()
}

func (s *SignatureIntegrationsStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	signatureIntegration := &storage.SignatureIntegration{}
	s.NoError(testutils.FullInit(signatureIntegration, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundSignatureIntegration, exists, err := store.Get(ctx, signatureIntegration.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundSignatureIntegration)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, signatureIntegration))
	foundSignatureIntegration, exists, err = store.Get(ctx, signatureIntegration.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(signatureIntegration, foundSignatureIntegration)

	signatureIntegrationCount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, signatureIntegrationCount)
	signatureIntegrationCount, err = store.Count(withNoAccessCtx)
	s.NoError(err)
	s.Zero(signatureIntegrationCount)

	signatureIntegrationExists, err := store.Exists(ctx, signatureIntegration.GetId())
	s.NoError(err)
	s.True(signatureIntegrationExists)
	s.NoError(store.Upsert(ctx, signatureIntegration))
	s.ErrorIs(store.Upsert(withNoAccessCtx, signatureIntegration), sac.ErrResourceAccessDenied)

	foundSignatureIntegration, exists, err = store.Get(ctx, signatureIntegration.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(signatureIntegration, foundSignatureIntegration)

	s.NoError(store.Delete(ctx, signatureIntegration.GetId()))
	foundSignatureIntegration, exists, err = store.Get(ctx, signatureIntegration.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundSignatureIntegration)
	s.ErrorIs(store.Delete(withNoAccessCtx, signatureIntegration.GetId()), sac.ErrResourceAccessDenied)

	var signatureIntegrations []*storage.SignatureIntegration
	for i := 0; i < 200; i++ {
		signatureIntegration := &storage.SignatureIntegration{}
		s.NoError(testutils.FullInit(signatureIntegration, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		signatureIntegrations = append(signatureIntegrations, signatureIntegration)
	}

	s.NoError(store.UpsertMany(ctx, signatureIntegrations))

	signatureIntegrationCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, signatureIntegrationCount)
}
