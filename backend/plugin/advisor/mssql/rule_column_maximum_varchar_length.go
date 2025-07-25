package mssql

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/tsql-parser"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	"github.com/bytebase/bytebase/backend/common/log"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/advisor"
	tsqlparser "github.com/bytebase/bytebase/backend/plugin/parser/tsql"
)

var (
	_ advisor.Advisor = (*ColumnMaximumVarcharLengthAdvisor)(nil)
)

func init() {
	advisor.Register(storepb.Engine_MSSQL, advisor.MSSQLColumnMaximumVarcharLength, &ColumnMaximumVarcharLengthAdvisor{})
}

// ColumnMaximumVarcharLengthAdvisor is the advisor checking for maximum varchar length..
type ColumnMaximumVarcharLengthAdvisor struct {
}

// Check checks for maximum varchar length.
func (*ColumnMaximumVarcharLengthAdvisor) Check(_ context.Context, checkCtx advisor.Context) ([]*storepb.Advice, error) {
	tree, ok := checkCtx.AST.(antlr.Tree)
	if !ok {
		return nil, errors.Errorf("failed to convert to Tree")
	}

	level, err := advisor.NewStatusBySQLReviewRuleLevel(checkCtx.Rule.Level)
	if err != nil {
		return nil, err
	}
	payload, err := advisor.UnmarshalNumberTypeRulePayload(checkCtx.Rule.Payload)
	if err != nil {
		return nil, err
	}

	if payload.Number <= 0 {
		return nil, nil
	}

	// Create the rule
	rule := NewColumnMaximumVarcharLengthRule(level, string(checkCtx.Rule.Type), payload.Number)

	// Create the generic checker with the rule
	checker := NewGenericChecker([]Rule{rule})

	antlr.ParseTreeWalkerDefault.Walk(checker, tree)

	return checker.GetAdviceList(), nil
}

// ColumnMaximumVarcharLengthRule checks for maximum varchar length.
type ColumnMaximumVarcharLengthRule struct {
	BaseRule

	checkTypeString map[string]any
	maximum         int
}

// NewColumnMaximumVarcharLengthRule creates a new ColumnMaximumVarcharLengthRule.
func NewColumnMaximumVarcharLengthRule(level storepb.Advice_Status, title string, maximum int) *ColumnMaximumVarcharLengthRule {
	return &ColumnMaximumVarcharLengthRule{
		BaseRule: BaseRule{
			level: level,
			title: title,
		},
		checkTypeString: map[string]any{
			"varchar":  nil,
			"nvarchar": nil,
			"char":     nil,
			"nchar":    nil,
		},
		maximum: maximum,
	}
}

// Name returns the rule name.
func (*ColumnMaximumVarcharLengthRule) Name() string {
	return "ColumnMaximumVarcharLengthRule"
}

// OnEnter is called when entering a parse tree node.
func (r *ColumnMaximumVarcharLengthRule) OnEnter(ctx antlr.ParserRuleContext, nodeType string) error {
	if nodeType == "Data_type" {
		r.enterDataType(ctx.(*parser.Data_typeContext))
	}
	return nil
}

// OnExit is called when exiting a parse tree node.
func (*ColumnMaximumVarcharLengthRule) OnExit(_ antlr.ParserRuleContext, _ string) error {
	// This rule doesn't need exit processing
	return nil
}

func (r *ColumnMaximumVarcharLengthRule) enterDataType(ctx *parser.Data_typeContext) {
	currentLength := 0
	line := ctx.GetStart().GetLine()
	if ctx.MAX() != nil && (ctx.VARCHAR() != nil || ctx.NVARCHAR() != nil) {
		// https://learn.microsoft.com/en-us/sql/t-sql/data-types/data-types-transact-sql?view=sql-server-ver16&redirectedfrom=MSDN
		currentLength = math.MaxInt32 // 2 ^ 31 - 1
		line = ctx.MAX().GetSymbol().GetLine()
	} else if ctx.GetExt_type() != nil && ctx.GetScale() != nil && ctx.GetPrec() == nil && ctx.GetInc() == nil {
		_, normalizedTypeString := tsqlparser.NormalizeTSQLIdentifier(ctx.GetExt_type())
		if _, ok := r.checkTypeString[normalizedTypeString]; !ok {
			return
		}
		length, err := strconv.Atoi(ctx.GetScale().GetText())
		if err != nil {
			slog.Error("failed to convert scale to int", log.BBError(err))
		}
		currentLength = length
		line = ctx.GetScale().GetLine()
	} else if ctx.GetUnscaled_type() != nil {
		_, normalizedTypeString := tsqlparser.NormalizeTSQLIdentifier(ctx.GetUnscaled_type())
		if _, ok := r.checkTypeString[normalizedTypeString]; !ok {
			return
		}
		line = ctx.GetUnscaled_type().GetStart().GetLine()
	}
	if currentLength > r.maximum {
		r.AddAdvice(&storepb.Advice{
			Status:        r.level,
			Code:          advisor.VarcharLengthExceedsLimit.Int32(),
			Title:         r.title,
			Content:       fmt.Sprintf("The maximum varchar length is %d.", r.maximum),
			StartPosition: common.ConvertANTLRLineToPosition(line),
		})
	}
}
