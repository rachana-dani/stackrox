// Code generated by pg-bindings generator. DO NOT EDIT.

//go:build sql_integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/postgres/pgtest"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/testutils"
	"github.com/stackrox/rox/pkg/testutils/envisolator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AlertsStoreSuite struct {
	suite.Suite
	envIsolator *envisolator.EnvIsolator
	store       Store
	testDB      *pgtest.TestPostgres
}

func TestAlertsStore(t *testing.T) {
	suite.Run(t, new(AlertsStoreSuite))
}

func (s *AlertsStoreSuite) SetupSuite() {
	s.envIsolator = envisolator.NewEnvIsolator(s.T())
	s.envIsolator.Setenv(env.PostgresDatastoreEnabled.EnvVar(), "true")

	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	s.testDB = pgtest.ForT(s.T())
	s.store = New(s.testDB.Pool)
}

func (s *AlertsStoreSuite) SetupTest() {
	ctx := sac.WithAllAccess(context.Background())
	tag, err := s.testDB.Exec(ctx, "TRUNCATE alerts CASCADE")
	s.T().Log("alerts", tag)
	s.NoError(err)
}

func (s *AlertsStoreSuite) TearDownSuite() {
	s.testDB.Teardown(s.T())
	s.envIsolator.RestoreAll()
}

func (s *AlertsStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	alert := &storage.Alert{}
	s.NoError(testutils.FullInit(alert, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundAlert, exists, err := store.Get(ctx, alert.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundAlert)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, alert))
	foundAlert, exists, err = store.Get(ctx, alert.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(alert, foundAlert)

	alertCount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, alertCount)
	alertCount, err = store.Count(withNoAccessCtx)
	s.NoError(err)
	s.Zero(alertCount)

	alertExists, err := store.Exists(ctx, alert.GetId())
	s.NoError(err)
	s.True(alertExists)
	s.NoError(store.Upsert(ctx, alert))
	s.ErrorIs(store.Upsert(withNoAccessCtx, alert), sac.ErrResourceAccessDenied)

	foundAlert, exists, err = store.Get(ctx, alert.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(alert, foundAlert)

	s.NoError(store.Delete(ctx, alert.GetId()))
	foundAlert, exists, err = store.Get(ctx, alert.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundAlert)
	s.NoError(store.Delete(withNoAccessCtx, alert.GetId()))

	var alerts []*storage.Alert
	for i := 0; i < 200; i++ {
		alert := &storage.Alert{}
		s.NoError(testutils.FullInit(alert, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		alerts = append(alerts, alert)
	}

	s.NoError(store.UpsertMany(ctx, alerts))

	alertCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, alertCount)
}

func (s *AlertsStoreSuite) TestSACUpsert() {
	obj := &storage.Alert{}
	s.NoError(testutils.FullInit(obj, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	ctxs := getSACContexts(obj, storage.Access_READ_WRITE_ACCESS)
	for name, expectedErr := range map[string]error{
		withAllAccess:           nil,
		withNoAccess:            sac.ErrResourceAccessDenied,
		withNoAccessToCluster:   sac.ErrResourceAccessDenied,
		withAccessToDifferentNs: sac.ErrResourceAccessDenied,
		withAccess:              nil,
		withAccessToCluster:     nil,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			assert.ErrorIs(t, s.store.Upsert(ctxs[name], obj), expectedErr)
		})
	}
}

func (s *AlertsStoreSuite) TestSACUpsertMany() {
	obj := &storage.Alert{}
	s.NoError(testutils.FullInit(obj, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	ctxs := getSACContexts(obj, storage.Access_READ_WRITE_ACCESS)
	for name, expectedErr := range map[string]error{
		withAllAccess:           nil,
		withNoAccess:            sac.ErrResourceAccessDenied,
		withNoAccessToCluster:   sac.ErrResourceAccessDenied,
		withAccessToDifferentNs: sac.ErrResourceAccessDenied,
		withAccess:              nil,
		withAccessToCluster:     nil,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			assert.ErrorIs(t, s.store.UpsertMany(ctxs[name], []*storage.Alert{obj}), expectedErr)
		})
	}
}

func (s *AlertsStoreSuite) TestSACCount() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	objB := &storage.Alert{}
	s.NoError(testutils.FullInit(objB, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	withAllAccessCtx := sac.WithAllAccess(context.Background())
	s.store.Upsert(withAllAccessCtx, objA)
	s.store.Upsert(withAllAccessCtx, objB)

	ctxs := getSACContexts(objA, storage.Access_READ_ACCESS)
	for name, expectedCount := range map[string]int{
		withAllAccess:           2,
		withNoAccess:            0,
		withNoAccessToCluster:   0,
		withAccessToDifferentNs: 0,
		withAccess:              1,
		withAccessToCluster:     1,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			count, err := s.store.Count(ctxs[name])
			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
		})
	}
}

func (s *AlertsStoreSuite) TestSACWalk() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	objB := &storage.Alert{}
	s.NoError(testutils.FullInit(objB, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	withAllAccessCtx := sac.WithAllAccess(context.Background())
	s.store.Upsert(withAllAccessCtx, objA)
	s.store.Upsert(withAllAccessCtx, objB)

	ctxs := getSACContexts(objA, storage.Access_READ_ACCESS)
	for name, expectedIds := range map[string][]string{
		withAllAccess:           []string{objA.GetId(), objB.GetId()},
		withNoAccess:            []string{},
		withNoAccessToCluster:   []string{},
		withAccessToDifferentNs: []string{},
		withAccess:              []string{objA.GetId()},
		withAccessToCluster:     []string{objA.GetId()},
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			ids := []string{}
			getIds := func(obj *storage.Alert) error {
				ids = append(ids, obj.GetId())
				return nil
			}
			err := s.store.Walk(ctxs[name], getIds)
			assert.NoError(t, err)
			assert.ElementsMatch(t, expectedIds, ids)
		})
	}
}

func (s *AlertsStoreSuite) TestSACGetIDs() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	objB := &storage.Alert{}
	s.NoError(testutils.FullInit(objB, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	withAllAccessCtx := sac.WithAllAccess(context.Background())
	s.store.Upsert(withAllAccessCtx, objA)
	s.store.Upsert(withAllAccessCtx, objB)

	ctxs := getSACContexts(objA, storage.Access_READ_ACCESS)
	for name, expectedIds := range map[string][]string{
		withAllAccess:           []string{objA.GetId(), objB.GetId()},
		withNoAccess:            []string{},
		withNoAccessToCluster:   []string{},
		withAccessToDifferentNs: []string{},
		withAccess:              []string{objA.GetId()},
		withAccessToCluster:     []string{objA.GetId()},
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			ids, err := s.store.GetIDs(ctxs[name])
			assert.NoError(t, err)
			assert.EqualValues(t, expectedIds, ids)
		})
	}
}

func (s *AlertsStoreSuite) TestSACExists() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	withAllAccessCtx := sac.WithAllAccess(context.Background())
	s.store.Upsert(withAllAccessCtx, objA)

	ctxs := getSACContexts(objA, storage.Access_READ_ACCESS)
	for name, expected := range map[string]bool{
		withAllAccess:           true,
		withNoAccess:            false,
		withNoAccessToCluster:   false,
		withAccessToDifferentNs: false,
		withAccess:              true,
		withAccessToCluster:     true,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			exists, err := s.store.Exists(ctxs[name], objA.GetId())
			assert.NoError(t, err)
			assert.Equal(t, expected, exists)
		})
	}
}

func (s *AlertsStoreSuite) TestSACGet() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	withAllAccessCtx := sac.WithAllAccess(context.Background())
	s.store.Upsert(withAllAccessCtx, objA)

	ctxs := getSACContexts(objA, storage.Access_READ_ACCESS)
	for name, expected := range map[string]bool{
		withAllAccess:           true,
		withNoAccess:            false,
		withNoAccessToCluster:   false,
		withAccessToDifferentNs: false,
		withAccess:              true,
		withAccessToCluster:     true,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			actual, exists, err := s.store.Get(ctxs[name], objA.GetId())
			assert.NoError(t, err)
			assert.Equal(t, expected, exists)
			if expected == true {
				assert.Equal(t, objA, actual)
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}

func (s *AlertsStoreSuite) TestSACDelete() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	objB := &storage.Alert{}
	s.NoError(testutils.FullInit(objB, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
	withAllAccessCtx := sac.WithAllAccess(context.Background())

	ctxs := getSACContexts(objA, storage.Access_READ_WRITE_ACCESS)
	for name, expectedCount := range map[string]int{
		withAllAccess:           0,
		withNoAccess:            2,
		withNoAccessToCluster:   2,
		withAccessToDifferentNs: 2,
		withAccess:              1,
		withAccessToCluster:     1,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			s.SetupTest()

			s.NoError(s.store.Upsert(withAllAccessCtx, objA))
			s.NoError(s.store.Upsert(withAllAccessCtx, objB))

			assert.NoError(t, s.store.Delete(ctxs[name], objA.GetId()))
			assert.NoError(t, s.store.Delete(ctxs[name], objB.GetId()))

			count, err := s.store.Count(withAllAccessCtx)
			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
		})
	}
}

func (s *AlertsStoreSuite) TestSACDeleteMany() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	objB := &storage.Alert{}
	s.NoError(testutils.FullInit(objB, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
	withAllAccessCtx := sac.WithAllAccess(context.Background())

	ctxs := getSACContexts(objA, storage.Access_READ_WRITE_ACCESS)
	for name, expectedCount := range map[string]int{
		withAllAccess:           0,
		withNoAccess:            2,
		withNoAccessToCluster:   2,
		withAccessToDifferentNs: 2,
		withAccess:              1,
		withAccessToCluster:     1,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			s.SetupTest()

			s.NoError(s.store.Upsert(withAllAccessCtx, objA))
			s.NoError(s.store.Upsert(withAllAccessCtx, objB))

			assert.NoError(t, s.store.DeleteMany(ctxs[name], []string{
				objA.GetId(),
				objB.GetId(),
			}))

			count, err := s.store.Count(withAllAccessCtx)
			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
		})
	}
}

func (s *AlertsStoreSuite) TestSACGetMany() {
	objA := &storage.Alert{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	objB := &storage.Alert{}
	s.NoError(testutils.FullInit(objB, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	withAllAccessCtx := sac.WithAllAccess(context.Background())
	s.store.Upsert(withAllAccessCtx, objA)
	s.store.Upsert(withAllAccessCtx, objB)

	ctxs := getSACContexts(objA, storage.Access_READ_ACCESS)
	for name, expected := range map[string]struct {
		elems          []*storage.Alert
		missingIndices []int
	}{
		withAllAccess:           {elems: []*storage.Alert{objA, objB}, missingIndices: []int{}},
		withNoAccess:            {elems: []*storage.Alert{}, missingIndices: []int{0, 1}},
		withNoAccessToCluster:   {elems: []*storage.Alert{}, missingIndices: []int{0, 1}},
		withAccessToDifferentNs: {elems: []*storage.Alert{}, missingIndices: []int{0, 1}},
		withAccess:              {elems: []*storage.Alert{objA}, missingIndices: []int{1}},
		withAccessToCluster:     {elems: []*storage.Alert{objA}, missingIndices: []int{1}},
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			actual, missingIndices, err := s.store.GetMany(ctxs[name], []string{objA.GetId(), objB.GetId()})
			assert.NoError(t, err)
			assert.Equal(t, expected.elems, actual)
			assert.Equal(t, expected.missingIndices, missingIndices)
		})
	}

	s.T().Run("with no ids", func(t *testing.T) {
		actual, missingIndices, err := s.store.GetMany(withAllAccessCtx, []string{})
		assert.Nil(t, err)
		assert.Nil(t, actual)
		assert.Nil(t, missingIndices)
	})
}

const (
	withAllAccess           = "AllAccess"
	withNoAccess            = "NoAccess"
	withAccessToDifferentNs = "AccessToDifferentNs"
	withAccess              = "Access"
	withAccessToCluster     = "AccessToCluster"
	withNoAccessToCluster   = "NoAccessToCluster"
)

func getSACContexts(obj *storage.Alert, access storage.Access) map[string]context.Context {
	return map[string]context.Context{
		withAllAccess: sac.WithAllAccess(context.Background()),
		withNoAccess:  sac.WithNoAccess(context.Background()),
		withAccessToDifferentNs: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys(obj.GetClusterId()),
				sac.NamespaceScopeKeys("unknown ns"),
			)),
		withAccess: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys(obj.GetClusterId()),
				sac.NamespaceScopeKeys(obj.GetNamespace()),
			)),
		withAccessToCluster: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys(obj.GetClusterId()),
			)),
		withNoAccessToCluster: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys("unknown cluster"),
			)),
	}
}
