// Code generated by pg-bindings generator. DO NOT EDIT.

package schema

import (
	"fmt"
	"reflect"

	"github.com/stackrox/rox/central/globaldb"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/postgres"
	"github.com/stackrox/rox/pkg/postgres/walker"
	"github.com/stackrox/rox/pkg/search"
)

var (
	// CreateTablePodsStmt holds the create statement for table `pods`.
	CreateTablePodsStmt = &postgres.CreateStmts{
		Table: `
               create table if not exists pods (
                   Id varchar,
                   Name varchar,
                   DeploymentId varchar,
                   Namespace varchar,
                   ClusterId varchar,
                   serialized bytea,
                   PRIMARY KEY(Id)
               )
               `,
		Indexes: []string{},
		Children: []*postgres.CreateStmts{
			&postgres.CreateStmts{
				Table: `
               create table if not exists pods_live_instances (
                   pods_Id varchar,
                   idx integer,
                   ImageDigest varchar,
                   PRIMARY KEY(pods_Id, idx),
                   CONSTRAINT fk_parent_table_0 FOREIGN KEY (pods_Id) REFERENCES pods(Id) ON DELETE CASCADE
               )
               `,
				Indexes: []string{
					"create index if not exists podsLiveInstances_idx on pods_live_instances using btree(idx)",
				},
				Children: []*postgres.CreateStmts{},
			},
		},
	}

	// PodsSchema is the go schema for table `pods`.
	PodsSchema = func() *walker.Schema {
		schema := globaldb.GetSchemaForTable("pods")
		if schema != nil {
			return schema
		}
		schema = walker.Walk(reflect.TypeOf((*storage.Pod)(nil)), "pods")
		referencedSchemas := map[string]*walker.Schema{
			"storage.Deployment": DeploymentsSchema,
		}

		schema.ResolveReferences(func(messageTypeName string) *walker.Schema {
			return referencedSchemas[fmt.Sprintf("storage.%s", messageTypeName)]
		})
		schema.SetOptionsMap(search.Walk(v1.SearchCategory_PODS, "pod", (*storage.Pod)(nil)))
		globaldb.RegisterTable(schema)
		return schema
	}()
)
