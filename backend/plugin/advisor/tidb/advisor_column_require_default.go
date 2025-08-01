package tidb

// Framework code is generated by the generator.

import (
	"context"
	"fmt"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/advisor"
)

var (
	_ advisor.Advisor = (*ColumRequireDefaultAdvisor)(nil)
	_ ast.Visitor     = (*columRequireDefaultChecker)(nil)
)

func init() {
	advisor.Register(storepb.Engine_TIDB, advisor.MySQLRequireColumnDefault, &ColumRequireDefaultAdvisor{})
}

// ColumRequireDefaultAdvisor is the advisor checking for column default requirement.
type ColumRequireDefaultAdvisor struct {
}

// Check checks for column default requirement.
func (*ColumRequireDefaultAdvisor) Check(_ context.Context, checkCtx advisor.Context) ([]*storepb.Advice, error) {
	stmtList, ok := checkCtx.AST.([]ast.StmtNode)
	if !ok {
		return nil, errors.Errorf("failed to convert to StmtNode")
	}

	level, err := advisor.NewStatusBySQLReviewRuleLevel(checkCtx.Rule.Level)
	if err != nil {
		return nil, err
	}
	checker := &columRequireDefaultChecker{
		level: level,
		title: string(checkCtx.Rule.Type),
	}

	for _, stmt := range stmtList {
		checker.text = stmt.Text()
		checker.line = stmt.OriginTextPosition()
		(stmt).Accept(checker)
	}

	return checker.adviceList, nil
}

type columRequireDefaultChecker struct {
	adviceList []*storepb.Advice
	level      storepb.Advice_Status
	title      string
	text       string
	line       int
}

func getPkColumnsFromConstraints(node *ast.CreateTableStmt) map[string]bool {
	pkColumns := make(map[string]bool)
	for _, constraint := range node.Constraints {
		if constraint.Tp != ast.ConstraintPrimaryKey {
			continue
		}
		for _, key := range constraint.Keys {
			if key.Column == nil {
				continue
			}
			pkColumns[key.Column.Name.L] = true
		}
	}
	return pkColumns
}

// Enter implements the ast.Visitor interface.
func (checker *columRequireDefaultChecker) Enter(in ast.Node) (ast.Node, bool) {
	var columnList []columnData
	switch node := in.(type) {
	case *ast.CreateTableStmt:
		tableName := node.Table.Name.O
		pkColumns := getPkColumnsFromConstraints(node)
		for _, column := range node.Cols {
			if !hasDefault(column) && needDefault(column) && !pkColumns[column.Name.Name.L] {
				columnList = append(columnList, columnData{
					table:  tableName,
					column: column.Name.Name.O,
					line:   column.OriginTextPosition(),
				})
			}
		}
	case *ast.AlterTableStmt:
		tableName := node.Table.Name.O
		for _, spec := range node.Specs {
			switch spec.Tp {
			case ast.AlterTableAddColumns:
				for _, column := range spec.NewColumns {
					if !hasDefault(column) && needDefault(column) {
						columnList = append(columnList, columnData{
							table:  tableName,
							column: column.Name.Name.O,
							line:   node.OriginTextPosition(),
						})
					}
				}
			case ast.AlterTableChangeColumn, ast.AlterTableModifyColumn:
				column := spec.NewColumns[0]
				if !hasDefault(column) && needDefault(column) {
					columnList = append(columnList, columnData{
						table:  tableName,
						column: column.Name.Name.O,
						line:   node.OriginTextPosition(),
					})
				}
			default:
			}
		}
	}

	for _, column := range columnList {
		checker.adviceList = append(checker.adviceList, &storepb.Advice{
			Status:        checker.level,
			Code:          advisor.NoDefault.Int32(),
			Title:         checker.title,
			Content:       fmt.Sprintf("Column `%s`.`%s` doesn't have DEFAULT.", column.table, column.column),
			StartPosition: common.ConvertANTLRLineToPosition(column.line),
		})
	}

	return in, false
}

// Leave implements the ast.Visitor interface.
func (*columRequireDefaultChecker) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func hasDefault(column *ast.ColumnDef) bool {
	for _, option := range column.Options {
		if option.Tp == ast.ColumnOptionDefaultValue {
			return true
		}
	}
	return false
}
