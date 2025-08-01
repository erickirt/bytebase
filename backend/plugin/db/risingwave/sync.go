package risingwave

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/db"
	"github.com/bytebase/bytebase/backend/plugin/db/util"
	"github.com/bytebase/bytebase/backend/plugin/parser/sql/ast"
	pgrawparser "github.com/bytebase/bytebase/backend/plugin/parser/sql/engine/pg"
)

const systemSchemas = "'rw_catalog', 'information_schema', 'pg_catalog'"

// SyncInstance syncs the instance.
func (d *Driver) SyncInstance(ctx context.Context) (*db.InstanceMetadata, error) {
	version, err := d.getVersion(ctx)
	if err != nil {
		return nil, err
	}

	instanceRoles, err := d.getInstanceRoles(ctx)
	if err != nil {
		return nil, err
	}

	// Query db info
	databases, err := d.getDatabases(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databases")
	}

	return &db.InstanceMetadata{
		Version:   version,
		Databases: databases,
		Metadata: &storepb.Instance{
			Roles: instanceRoles,
		},
	}, nil
}

// SyncDBSchema syncs a single database schema.
func (d *Driver) SyncDBSchema(ctx context.Context) (*storepb.DatabaseSchemaMetadata, error) {
	// Query db info
	databases, err := d.getDatabases(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databases")
	}

	var databaseMetadata *storepb.DatabaseSchemaMetadata
	for _, database := range databases {
		if database.Name == d.databaseName {
			databaseMetadata = database
			break
		}
	}
	if databaseMetadata == nil {
		return nil, common.Errorf(common.NotFound, "database %q not found", d.databaseName)
	}

	txn, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	schemaList, err := getSchemas(txn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get schemas from database %q", d.databaseName)
	}
	columnMap, err := getTableColumns(txn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get table columns from database %q", d.databaseName)
	}
	tableMap, err := getTables(txn, columnMap)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get tables from database %q", d.databaseName)
	}
	viewMap, err := getViews(txn, columnMap)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get views from database %q", d.databaseName)
	}
	functionMap, err := getFunctions(txn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get functions from database %q", d.databaseName)
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	for _, schemaName := range schemaList {
		databaseMetadata.Schemas = append(databaseMetadata.Schemas, &storepb.SchemaMetadata{
			Name:      schemaName,
			Tables:    tableMap[schemaName],
			Views:     viewMap[schemaName],
			Functions: functionMap[schemaName],
		})
	}
	// No extensions in RisingWave.
	databaseMetadata.Extensions = make([]*storepb.ExtensionMetadata, 0)

	return databaseMetadata, err
}

var listSchemaQuery = fmt.Sprintf(`
SELECT nspname
FROM pg_catalog.pg_namespace
WHERE nspname NOT IN (%s) ORDER BY nspname;
`, systemSchemas)

func getSchemas(txn *sql.Tx) ([]string, error) {
	rows, err := txn.Query(listSchemaQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, err
		}
		result = append(result, schemaName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

var listTableQuery = `
SELECT
	T.name as table_name,
	S.name AS schema_name
FROM rw_catalog.rw_tables AS T
	JOIN rw_catalog.rw_schemas AS S
	ON T.schema_id = S.id;`

// getTables gets all tables of a database.
func getTables(txn *sql.Tx, columnMap map[db.TableKey][]*storepb.ColumnMetadata) (map[string][]*storepb.TableMetadata, error) {
	indexMap, err := getIndexes(txn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get indices")
	}

	tableMap := make(map[string][]*storepb.TableMetadata)
	rows, err := txn.Query(listTableQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		table := &storepb.TableMetadata{}
		// var tbl tableSchema
		var schemaName string
		if err := rows.Scan(&table.Name, &schemaName); err != nil {
			return nil, err
		}
		// TODO: get table statistics of RisingWave later.
		table.DataSize = 0
		table.IndexSize = 0
		table.RowCount = 0
		// Comment is not supported in RisingWave.
		table.Comment = ""
		key := db.TableKey{Schema: schemaName, Table: table.Name}
		table.Columns = columnMap[key]
		table.Indexes = indexMap[key]
		// No foreign keys in RisingWave.
		table.ForeignKeys = make([]*storepb.ForeignKeyMetadata, 0)

		tableMap[schemaName] = append(tableMap[schemaName], table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tableMap, nil
}

var listColumnQuery = `
SELECT
	cols.table_schema,
	cols.table_name,
	cols.column_name,
	cols.data_type,
	cols.ordinal_position,
	cols.column_default,
	cols.is_nullable,
	cols.collation_name,
	cols.udt_schema,
	cols.udt_name
FROM INFORMATION_SCHEMA.COLUMNS AS cols` + fmt.Sprintf(`
WHERE cols.table_schema NOT IN (%s)
ORDER BY cols.table_schema, cols.table_name, cols.ordinal_position;`, systemSchemas)

// getTableColumns gets the columns of a table.
func getTableColumns(txn *sql.Tx) (map[db.TableKey][]*storepb.ColumnMetadata, error) {
	columnsMap := make(map[db.TableKey][]*storepb.ColumnMetadata)
	rows, err := txn.Query(listColumnQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		column := &storepb.ColumnMetadata{}
		var schemaName, tableName, nullable string
		var defaultStr, collation, udtSchema, udtName sql.NullString
		if err := rows.Scan(&schemaName, &tableName, &column.Name, &column.Type, &column.Position, &defaultStr, &nullable, &collation, &udtSchema, &udtName); err != nil {
			return nil, err
		}
		if defaultStr.Valid {
			// Store in Default field (migration from DefaultExpression to Default)
			column.Default = defaultStr.String
		}
		isNullBool, err := util.ConvertYesNo(nullable)
		if err != nil {
			return nil, err
		}
		column.Nullable = isNullBool
		switch column.Type {
		case "USER-DEFINED":
			column.Type = fmt.Sprintf("%s.%s", udtSchema.String, udtName.String)
		case "ARRAY":
			column.Type = udtName.String
		default:
			// Keep the original type for other cases
		}
		column.Collation = collation.String
		column.Comment = ""

		key := db.TableKey{Schema: schemaName, Table: tableName}
		columnsMap[key] = append(columnsMap[key], column)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columnsMap, nil
}

var listViewQuery = `
WITH all_views AS (
  SELECT
    schema_id,
    name,
    definition
  FROM
    rw_materialized_views
  UNION
  ALL
  SELECT
    schema_id,
    name,
    definition
  FROM
    rw_views
)
SELECT
  S.name AS schema_name,
  all_views.name AS view_name,
  definition
FROM
  all_views
  JOIN rw_schemas S ON all_views.schema_id = S.id
  AND S.name NOT IN ` + fmt.Sprintf("(%s)", systemSchemas) + `
ORDER BY schema_name, view_name;`

// getViews gets all views of a database.
func getViews(txn *sql.Tx, columnMap map[db.TableKey][]*storepb.ColumnMetadata) (map[string][]*storepb.ViewMetadata, error) {
	viewMap := make(map[string][]*storepb.ViewMetadata)

	rows, err := txn.Query(listViewQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		view := &storepb.ViewMetadata{}
		var schemaName string
		var def sql.NullString
		if err := rows.Scan(&schemaName, &view.Name, &def); err != nil {
			return nil, err
		}
		// Return error on NULL view definition.
		// https://github.com/bytebase/bytebase/issues/343
		if !def.Valid {
			return nil, errors.Errorf("schema %q view %q has empty definition; please check whether proper privileges have been granted to Bytebase", schemaName, view.Name)
		}
		view.Definition = def.String
		view.Comment = ""

		key := db.TableKey{Schema: schemaName, Table: view.Name}
		view.Columns = columnMap[key]

		viewMap[schemaName] = append(viewMap[schemaName], view)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// TODO: RisingWave doesn't support view dependencies analyze yet.

	return viewMap, nil
}

var listIndexQuery = `
SELECT
	I.name AS idx_name,
	T.name AS table_name,
	S.name AS schema_name,
	I.definition
FROM rw_indexes AS I
	JOIN rw_tables AS T
	ON I.primary_table_id = T.id
	JOIN rw_schemas AS S
	ON I.schema_id = S.id;
`

// getIndexes gets all indices of a database.
func getIndexes(txn *sql.Tx) (map[db.TableKey][]*storepb.IndexMetadata, error) {
	indexMap := make(map[db.TableKey][]*storepb.IndexMetadata)

	rows, err := txn.Query(listIndexQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		index := &storepb.IndexMetadata{}
		var schemaName, tableName, statement string
		if err := rows.Scan(&index.Name, &tableName, &schemaName, &statement); err != nil {
			return nil, err
		}

		nodes, err := pgrawparser.Parse(pgrawparser.ParseContext{}, statement)
		if err != nil {
			return nil, err
		}
		if len(nodes) != 1 {
			return nil, errors.Errorf("invalid number of statements %v, expecting one", len(nodes))
		}
		node, ok := nodes[0].(*ast.CreateIndexStmt)
		if !ok {
			return nil, errors.Errorf("statement %q is not index statement", statement)
		}
		deparsed, err := pgrawparser.Deparse(pgrawparser.DeparseContext{}, node)
		if err != nil {
			return nil, err
		}
		// Instead of using indexdef, we use deparsed format so that the definition has quoted identifiers.
		index.Definition = deparsed

		index.Type = getIndexMethodType(statement)
		index.Unique = node.Index.Unique
		index.Expressions = node.Index.GetKeyNameList()
		// No primary key index in RisingWave.
		index.Primary = false
		// Comment is unsupported in RisingWave in v1.0.
		index.Comment = ""

		key := db.TableKey{Schema: schemaName, Table: tableName}
		indexMap[key] = append(indexMap[key], index)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexMap, nil
}

func getIndexMethodType(stmt string) string {
	re := regexp.MustCompile(`USING (\w+) `)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) == 0 {
		return ""
	}
	return matches[1]
}

var listFunctionQuery = `SELECT
	F.name AS function_name,
	S.name AS schema_name
FROM rw_catalog.rw_functions AS F
	JOIN rw_schemas AS S
	ON F.schema_id = S.id;`

// getFunctions gets all functions of a database.
func getFunctions(txn *sql.Tx) (map[string][]*storepb.FunctionMetadata, error) {
	functionMap := make(map[string][]*storepb.FunctionMetadata)

	rows, err := txn.Query(listFunctionQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		function := &storepb.FunctionMetadata{}
		var schemaName string
		if err := rows.Scan(&function.Name, &schemaName); err != nil {
			return nil, err
		}
		// Omit the definition.
		function.Definition = "..."

		functionMap[schemaName] = append(functionMap[schemaName], function)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return functionMap, nil
}
