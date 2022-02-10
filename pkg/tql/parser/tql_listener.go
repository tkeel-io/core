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

	// EnterExpression is called when entering the Expression production.
	EnterExpression(c *ExpressionContext)

	// EnterDummyAddSub is called when entering the DummyAddSub production.
	EnterDummyAddSub(c *DummyAddSubContext)

	// EnterDummyMulDiv is called when entering the DummyMulDiv production.
	EnterDummyMulDiv(c *DummyMulDivContext)

	// EnterDummyCompareValue is called when entering the DummyCompareValue production.
	EnterDummyCompareValue(c *DummyCompareValueContext)

	// EnterSourceEntity is called when entering the sourceEntity production.
	EnterSourceEntity(c *SourceEntityContext)

	// EnterTargetProperty is called when entering the targetProperty production.
	EnterTargetProperty(c *TargetPropertyContext)

	// EnterComputing is called when entering the computing production.
	EnterComputing(c *ComputingContext)

	// EnterNumber is called when entering the Number production.
	EnterNumber(c *NumberContext)

	// EnterCompareValue is called when entering the CompareValue production.
	EnterCompareValue(c *CompareValueContext)

	// EnterMulDiv is called when entering the MulDiv production.
	EnterMulDiv(c *MulDivContext)

	// EnterAddSub is called when entering the AddSub production.
	EnterAddSub(c *AddSubContext)

	// EnterString is called when entering the String production.
	EnterString(c *StringContext)

	// ExitRoot is called when exiting the root production.
	ExitRoot(c *RootContext)

	// ExitFields is called when exiting the fields production.
	ExitFields(c *FieldsContext)

	// ExitTargetEntity is called when exiting the targetEntity production.
	ExitTargetEntity(c *TargetEntityContext)

	// ExitExpression is called when exiting the Expression production.
	ExitExpression(c *ExpressionContext)

	// ExitDummyAddSub is called when exiting the DummyAddSub production.
	ExitDummyAddSub(c *DummyAddSubContext)

	// ExitDummyMulDiv is called when exiting the DummyMulDiv production.
	ExitDummyMulDiv(c *DummyMulDivContext)

	// ExitDummyCompareValue is called when exiting the DummyCompareValue production.
	ExitDummyCompareValue(c *DummyCompareValueContext)

	// ExitSourceEntity is called when exiting the sourceEntity production.
	ExitSourceEntity(c *SourceEntityContext)

	// ExitTargetProperty is called when exiting the targetProperty production.
	ExitTargetProperty(c *TargetPropertyContext)

	// ExitComputing is called when exiting the computing production.
	ExitComputing(c *ComputingContext)

	// ExitNumber is called when exiting the Number production.
	ExitNumber(c *NumberContext)

	// ExitCompareValue is called when exiting the CompareValue production.
	ExitCompareValue(c *CompareValueContext)

	// ExitMulDiv is called when exiting the MulDiv production.
	ExitMulDiv(c *MulDivContext)

	// ExitAddSub is called when exiting the AddSub production.
	ExitAddSub(c *AddSubContext)

	// ExitString is called when exiting the String production.
	ExitString(c *StringContext)
}
