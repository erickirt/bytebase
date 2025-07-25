package pg

// Framework code is generated by the generator.

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/advisor"
	"github.com/bytebase/bytebase/backend/plugin/parser/sql/ast"
)

var (
	_ advisor.Advisor = (*InsertMustSpecifyColumnAdvisor)(nil)
	_ ast.Visitor     = (*insertMustSpecifyColumnChecker)(nil)
)

func init() {
	advisor.Register(storepb.Engine_POSTGRES, advisor.PostgreSQLInsertMustSpecifyColumn, &InsertMustSpecifyColumnAdvisor{})
}

// InsertMustSpecifyColumnAdvisor is the advisor checking for to enforce column specified.
type InsertMustSpecifyColumnAdvisor struct {
}

// Check checks for to enforce column specified.
func (*InsertMustSpecifyColumnAdvisor) Check(_ context.Context, checkCtx advisor.Context) ([]*storepb.Advice, error) {
	stmtList, ok := checkCtx.AST.([]ast.Node)
	if !ok {
		return nil, errors.Errorf("failed to convert to Node")
	}

	level, err := advisor.NewStatusBySQLReviewRuleLevel(checkCtx.Rule.Level)
	if err != nil {
		return nil, err
	}
	checker := &insertMustSpecifyColumnChecker{
		level: level,
		title: string(checkCtx.Rule.Type),
	}

	for _, stmt := range stmtList {
		checker.text = advisor.NormalizeStatement(stmt.Text())
		ast.Walk(checker, stmt)
	}

	return checker.adviceList, nil
}

type insertMustSpecifyColumnChecker struct {
	adviceList []*storepb.Advice
	level      storepb.Advice_Status
	title      string
	text       string
}

// Visit implements ast.Visitor interface.
func (checker *insertMustSpecifyColumnChecker) Visit(in ast.Node) ast.Visitor {
	if node, ok := in.(*ast.InsertStmt); ok && len(node.ColumnList) == 0 {
		checker.adviceList = append(checker.adviceList, &storepb.Advice{
			Status:        checker.level,
			Code:          advisor.InsertNotSpecifyColumn.Int32(),
			Title:         checker.title,
			Content:       fmt.Sprintf("The INSERT statement must specify columns but \"%s\" does not", checker.text),
			StartPosition: common.ConvertPGParserLineToPosition(node.LastLine()),
		})
	}

	return checker
}
