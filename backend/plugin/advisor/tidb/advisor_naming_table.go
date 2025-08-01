package tidb

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/advisor"
)

var (
	_ advisor.Advisor = (*NamingTableConventionAdvisor)(nil)
	_ ast.Visitor     = (*namingTableConventionChecker)(nil)
)

func init() {
	advisor.Register(storepb.Engine_TIDB, advisor.MySQLNamingTableConvention, &NamingTableConventionAdvisor{})
}

// NamingTableConventionAdvisor is the advisor checking for table naming convention.
type NamingTableConventionAdvisor struct {
}

// Check checks for table naming convention.
func (*NamingTableConventionAdvisor) Check(_ context.Context, checkCtx advisor.Context) ([]*storepb.Advice, error) {
	root, ok := checkCtx.AST.([]ast.StmtNode)
	if !ok {
		return nil, errors.Errorf("failed to convert to StmtNode")
	}

	level, err := advisor.NewStatusBySQLReviewRuleLevel(checkCtx.Rule.Level)
	if err != nil {
		return nil, err
	}
	format, maxLength, err := advisor.UnmarshalNamingRulePayloadAsRegexp(checkCtx.Rule.Payload)
	if err != nil {
		return nil, err
	}
	checker := &namingTableConventionChecker{
		level:     level,
		title:     string(checkCtx.Rule.Type),
		format:    format,
		maxLength: maxLength,
	}
	for _, stmtNode := range root {
		(stmtNode).Accept(checker)
	}

	return checker.adviceList, nil
}

type namingTableConventionChecker struct {
	adviceList []*storepb.Advice
	level      storepb.Advice_Status
	title      string
	format     *regexp.Regexp
	maxLength  int
}

// Enter implements the ast.Visitor interface.
func (v *namingTableConventionChecker) Enter(in ast.Node) (ast.Node, bool) {
	var tableNames []string
	switch node := in.(type) {
	// CREATE TABLE
	case *ast.CreateTableStmt:
		// Original string
		tableNames = append(tableNames, node.Table.Name.O)
	// ALTER TABLE
	case *ast.AlterTableStmt:
		for _, spec := range node.Specs {
			// RENAME TABLE
			if spec.Tp == ast.AlterTableRenameTable {
				tableNames = append(tableNames, spec.NewTable.Name.O)
			}
		}
	// RENAME TABLE
	case *ast.RenameTableStmt:
		for _, table2Table := range node.TableToTables {
			tableNames = append(tableNames, table2Table.NewTable.Name.O)
		}
	}

	for _, tableName := range tableNames {
		if !v.format.MatchString(tableName) {
			v.adviceList = append(v.adviceList, &storepb.Advice{
				Status:        v.level,
				Code:          advisor.NamingTableConventionMismatch.Int32(),
				Title:         v.title,
				Content:       fmt.Sprintf("`%s` mismatches table naming convention, naming format should be %q", tableName, v.format),
				StartPosition: common.ConvertANTLRLineToPosition(in.OriginTextPosition()),
			})
		}
		if v.maxLength > 0 && len(tableName) > v.maxLength {
			v.adviceList = append(v.adviceList, &storepb.Advice{
				Status:        v.level,
				Code:          advisor.NamingTableConventionMismatch.Int32(),
				Title:         v.title,
				Content:       fmt.Sprintf("`%s` mismatches table naming convention, its length should be within %d characters", tableName, v.maxLength),
				StartPosition: common.ConvertANTLRLineToPosition(in.OriginTextPosition()),
			})
		}
	}
	return in, false
}

// Leave implements the ast.Visitor interface.
func (*namingTableConventionChecker) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
