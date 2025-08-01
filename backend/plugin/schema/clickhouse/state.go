package clickhouse

import (
	"fmt"
	"io"
	"slices"
	"strings"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/db/util"
)

type databaseState struct {
	name    string
	schemas map[string]*schemaState
}

func newDatabaseState() *databaseState {
	return &databaseState{
		schemas: make(map[string]*schemaState),
	}
}

func convertToDatabaseState(database *storepb.DatabaseSchemaMetadata) *databaseState {
	state := newDatabaseState()
	state.name = database.Name
	for _, schema := range database.Schemas {
		state.schemas[schema.Name] = convertToSchemaState(schema)
	}
	return state
}

type schemaState struct {
	name   string
	tables map[string]*tableState
	views  map[string]*viewState
}

func newSchemaState() *schemaState {
	return &schemaState{
		tables: make(map[string]*tableState),
		views:  make(map[string]*viewState),
	}
}

func convertToSchemaState(schema *storepb.SchemaMetadata) *schemaState {
	state := newSchemaState()
	state.name = schema.Name
	for i, table := range schema.Tables {
		state.tables[table.Name] = convertToTableState(i, table)
	}
	for i, view := range schema.Views {
		state.views[view.Name] = convertToViewState(i, view)
	}
	return state
}

type tableState struct {
	id          int
	name        string
	columns     map[string]*columnState
	sortingKeys []string
	primaryKeys []string
	comment     string
	engine      string
}

func (t *tableState) toString(buf *strings.Builder) error {
	if _, err := fmt.Fprintf(buf, "CREATE TABLE %s (\n  ", t.name); err != nil {
		return err
	}
	columns := []*columnState{}
	for _, column := range t.columns {
		columns = append(columns, column)
	}
	slices.SortFunc(columns, func(x, y *columnState) int {
		if x.id < y.id {
			return -1
		} else if x.id > y.id {
			return 1
		}
		return 0
	})
	for i, column := range columns {
		if i > 0 {
			if _, err := fmt.Fprint(buf, ",\n  "); err != nil {
				return err
			}
		}
		if err := column.toString(buf); err != nil {
			return err
		}
	}
	if _, err := buf.WriteString("\n)"); err != nil {
		return err
	}
	if t.engine != "" {
		if _, err := fmt.Fprintf(buf, "\nENGINE = %s", t.engine); err != nil {
			return err
		}
	}
	if len(t.sortingKeys) > 0 {
		if _, err := fmt.Fprintf(buf, "\nORDER BY (%s)", strings.Join(t.sortingKeys, ", ")); err != nil {
			return err
		}
	} else {
		// For merge tree table, we need to specify ORDER BY tuple() to make it work.
		// Reference: https://clickhouse.com/docs/en/engines/table-engines/mergetree-family/mergetree#order_by
		// TODO: handle other engines.
		if strings.ToLower(t.engine) == "mergetree" {
			if _, err := buf.WriteString("\nORDER BY tuple()"); err != nil {
				return err
			}
		}
	}
	if len(t.primaryKeys) > 0 {
		if _, err := fmt.Fprintf(buf, "\nPRIMARY KEY (%s)", strings.Join(t.primaryKeys, ", ")); err != nil {
			return err
		}
	}
	if t.comment != "" {
		if _, err := fmt.Fprintf(buf, "\nCOMMENT '%s'", t.comment); err != nil {
			return err
		}
	}
	if _, err := buf.WriteString(";\n"); err != nil {
		return err
	}
	return nil
}

func newTableState(id int, name string) *tableState {
	return &tableState{
		id:      id,
		name:    name,
		columns: make(map[string]*columnState),
	}
}

func convertToTableState(id int, table *storepb.TableMetadata) *tableState {
	state := newTableState(id, table.Name)
	state.comment = table.Comment
	state.engine = table.Engine
	state.sortingKeys = table.SortingKeys
	for i, column := range table.Columns {
		state.columns[column.Name] = convertToColumnState(i, column)
	}
	for _, index := range table.Indexes {
		if index.Primary {
			state.primaryKeys = index.Expressions
			break
		}
	}
	return state
}

type defaultValue interface {
	toString() string
}

type defaultValueNull struct {
}

func (*defaultValueNull) toString() string {
	return "NULL"
}

type defaultValueString struct {
	value string
}

func (d *defaultValueString) toString() string {
	return fmt.Sprintf("'%s'", d.value)
}

type defaultValueExpression struct {
	value string
}

func (d *defaultValueExpression) toString() string {
	return d.value
}

type columnState struct {
	id           int
	name         string
	tp           string
	defaultValue defaultValue
	comment      string
	nullable     bool
}

func (c *columnState) toString(buf *strings.Builder) error {
	if _, err := fmt.Fprintf(buf, "%s ", c.name); err != nil {
		return err
	}
	if _, err := buf.WriteString(c.tp); err != nil {
		return err
	}
	if !c.nullable {
		if _, err := buf.WriteString(" NOT NULL"); err != nil {
			return err
		}
	}
	if c.defaultValue != nil {
		if _, err := fmt.Fprintf(buf, " DEFAULT %s", c.defaultValue.toString()); err != nil {
			return err
		}
	}
	if c.comment != "" {
		if _, err := fmt.Fprintf(buf, " COMMENT '%s'", c.comment); err != nil {
			return err
		}
	}
	return nil
}

func convertToColumnState(id int, column *storepb.ColumnMetadata) *columnState {
	result := &columnState{
		id:       id,
		name:     column.Name,
		tp:       column.Type,
		nullable: column.Nullable,
		comment:  column.Comment,
	}
	// Handle default values using the unified Default field
	if column.Default == "NULL" {
		result.defaultValue = &defaultValueNull{}
	} else if column.Default != "" {
		// Check if it's an expression or a literal value
		// Simple heuristic: if it contains parentheses, operators, or functions, treat as expression
		if isExpression(column.Default) {
			result.defaultValue = &defaultValueExpression{value: column.Default}
		} else {
			result.defaultValue = &defaultValueString{value: column.Default}
		}
	}
	return result
}

// isExpression checks if a default value should be treated as an expression
func isExpression(value string) bool {
	// Simple heuristic: if it contains parentheses, operators, or common functions, treat as expression
	if strings.Contains(value, "(") || strings.Contains(value, ")") ||
		strings.Contains(value, "NOW") || strings.Contains(value, "today") ||
		strings.Contains(value, "CURRENT_") || strings.Contains(value, "UUID") ||
		strings.Contains(value, "TIMESTAMP") || strings.Contains(value, "SEQUENCE") ||
		strings.Contains(value, "generateUUIDv4") || strings.Contains(value, "rand") {
		return true
	}
	return false
}

type viewState struct {
	id         int
	name       string
	definition string
	comment    string
}

func convertToViewState(id int, view *storepb.ViewMetadata) *viewState {
	return &viewState{
		id:         id,
		name:       view.Name,
		definition: view.Definition,
		comment:    view.Comment,
	}
}

func (v *viewState) toString(buf io.StringWriter) error {
	stmt := fmt.Sprintf("CREATE OR REPLACE VIEW %s AS (%s)", v.name, util.TrimStatement(v.definition))
	if v.comment != "" {
		stmt += fmt.Sprintf(" COMMENT '%s'", v.comment)
	}
	stmt += ";\n"
	if _, err := buf.WriteString(stmt); err != nil {
		return err
	}
	return nil
}
