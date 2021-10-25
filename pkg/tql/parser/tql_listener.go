// Generated from TQL.g4 by ANTLR 4.7.

package parser // TQL

import "github.com/antlr/antlr4/runtime/Go/antlr"

// TQLListener is a complete listener for a parse tree produced by TQLParser.
type TQLListener interface {
	antlr.ParseTreeListener

	// EnterRoot is called when entering the root production.
	EnterRoot(c *RootContext)

	// EnterFields is called when entering the fields production.
	EnterFields(c *FieldsContext)

	// EnterTargetEntity is called when entering the targetEntity production.
	EnterTargetEntity(c *TargetEntityContext)

	// EnterCompareValue is called when entering the CompareValue production.
	EnterCompareValue(c *CompareValueContext)

	// EnterExpression is called when entering the Expression production.
	EnterExpression(c *ExpressionContext)

	// EnterMulDiv is called when entering the MulDiv production.
	EnterMulDiv(c *MulDivContext)

	// EnterAddSub is called when entering the AddSub production.
	EnterAddSub(c *AddSubContext)

	// EnterSourceEntity is called when entering the sourceEntity production.
	EnterSourceEntity(c *SourceEntityContext)

	// EnterTargetProperty is called when entering the targetProperty production.
	EnterTargetProperty(c *TargetPropertyContext)

	// ExitRoot is called when exiting the root production.
	ExitRoot(c *RootContext)

	// ExitFields is called when exiting the fields production.
	ExitFields(c *FieldsContext)

	// ExitTargetEntity is called when exiting the targetEntity production.
	ExitTargetEntity(c *TargetEntityContext)

	// ExitCompareValue is called when exiting the CompareValue production.
	ExitCompareValue(c *CompareValueContext)

	// ExitExpression is called when exiting the Expression production.
	ExitExpression(c *ExpressionContext)

	// ExitMulDiv is called when exiting the MulDiv production.
	ExitMulDiv(c *MulDivContext)

	// ExitAddSub is called when exiting the AddSub production.
	ExitAddSub(c *AddSubContext)

	// ExitSourceEntity is called when exiting the sourceEntity production.
	ExitSourceEntity(c *SourceEntityContext)

	// ExitTargetProperty is called when exiting the targetProperty production.
	ExitTargetProperty(c *TargetPropertyContext)
}
