// Generated from TQL.g4 by ANTLR 4.7.

package parser // TQL

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 48, 59, 4,
	2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 3, 2, 3, 2, 3,
	2, 3, 2, 3, 3, 3, 3, 3, 3, 7, 3, 20, 10, 3, 12, 3, 14, 3, 23, 11, 3, 3,
	4, 3, 4, 6, 4, 27, 10, 4, 13, 4, 14, 4, 28, 3, 4, 3, 4, 6, 4, 33, 10, 4,
	13, 4, 14, 4, 34, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
	7, 4, 46, 10, 4, 12, 4, 14, 4, 49, 11, 4, 3, 5, 3, 5, 5, 5, 53, 10, 5,
	3, 6, 3, 6, 5, 6, 57, 10, 6, 3, 6, 2, 3, 6, 7, 2, 4, 6, 8, 10, 2, 5, 3,
	2, 34, 36, 3, 2, 37, 38, 4, 2, 9, 9, 11, 15, 2, 61, 2, 12, 3, 2, 2, 2,
	4, 16, 3, 2, 2, 2, 6, 24, 3, 2, 2, 2, 8, 50, 3, 2, 2, 2, 10, 54, 3, 2,
	2, 2, 12, 13, 7, 19, 2, 2, 13, 14, 5, 4, 3, 2, 14, 15, 7, 2, 2, 3, 15,
	3, 3, 2, 2, 2, 16, 21, 5, 6, 4, 2, 17, 18, 7, 3, 2, 2, 18, 20, 5, 6, 4,
	2, 19, 17, 3, 2, 2, 2, 20, 23, 3, 2, 2, 2, 21, 19, 3, 2, 2, 2, 21, 22,
	3, 2, 2, 2, 22, 5, 3, 2, 2, 2, 23, 21, 3, 2, 2, 2, 24, 26, 8, 4, 1, 2,
	25, 27, 5, 8, 5, 2, 26, 25, 3, 2, 2, 2, 27, 28, 3, 2, 2, 2, 28, 26, 3,
	2, 2, 2, 28, 29, 3, 2, 2, 2, 29, 30, 3, 2, 2, 2, 30, 32, 7, 4, 2, 2, 31,
	33, 5, 10, 6, 2, 32, 31, 3, 2, 2, 2, 33, 34, 3, 2, 2, 2, 34, 32, 3, 2,
	2, 2, 34, 35, 3, 2, 2, 2, 35, 47, 3, 2, 2, 2, 36, 37, 12, 5, 2, 2, 37,
	38, 9, 2, 2, 2, 38, 46, 5, 6, 4, 6, 39, 40, 12, 4, 2, 2, 40, 41, 9, 3,
	2, 2, 41, 46, 5, 6, 4, 5, 42, 43, 12, 3, 2, 2, 43, 44, 9, 4, 2, 2, 44,
	46, 5, 6, 4, 4, 45, 36, 3, 2, 2, 2, 45, 39, 3, 2, 2, 2, 45, 42, 3, 2, 2,
	2, 46, 49, 3, 2, 2, 2, 47, 45, 3, 2, 2, 2, 47, 48, 3, 2, 2, 2, 48, 7, 3,
	2, 2, 2, 49, 47, 3, 2, 2, 2, 50, 52, 7, 42, 2, 2, 51, 53, 7, 43, 2, 2,
	52, 51, 3, 2, 2, 2, 52, 53, 3, 2, 2, 2, 53, 9, 3, 2, 2, 2, 54, 56, 7, 42,
	2, 2, 55, 57, 7, 43, 2, 2, 56, 55, 3, 2, 2, 2, 56, 57, 3, 2, 2, 2, 57,
	11, 3, 2, 2, 2, 9, 21, 28, 34, 45, 47, 52, 56,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "','", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "'*'", "'/'",
	"'%'", "'+'", "'-'", "'.'",
}
var symbolicNames = []string{
	"", "", "AS", "AND", "CASE", "ELSE", "END", "EQ", "FROM", "GT", "GTE",
	"LT", "LTE", "NE", "NOT", "NULL", "OR", "SELECT", "THEN", "WHERE", "WHEN",
	"GROUP", "BY", "TUMBLINGWINDOW", "HOPPINGWINDOW", "SLIDINGWINDOW", "SESSIONWINDOW",
	"DD", "HH", "MI", "SS", "MS", "MUL", "DIV", "MOD", "ADD", "SUB", "DOT",
	"TRUE", "FALSE", "ENTITYNAME", "PROPERTYNAME", "NUMBER", "INTEGER", "FLOAT",
	"STRING", "WHITESPACE",
}

var ruleNames = []string{
	"root", "fields", "expr", "sourceEntity", "targetEntity",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type TQLParser struct {
	*antlr.BaseParser
}

func NewTQLParser(input antlr.TokenStream) *TQLParser {
	this := new(TQLParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "TQL.g4"

	return this
}

// TQLParser tokens.
const (
	TQLParserEOF            = antlr.TokenEOF
	TQLParserT__0           = 1
	TQLParserAS             = 2
	TQLParserAND            = 3
	TQLParserCASE           = 4
	TQLParserELSE           = 5
	TQLParserEND            = 6
	TQLParserEQ             = 7
	TQLParserFROM           = 8
	TQLParserGT             = 9
	TQLParserGTE            = 10
	TQLParserLT             = 11
	TQLParserLTE            = 12
	TQLParserNE             = 13
	TQLParserNOT            = 14
	TQLParserNULL           = 15
	TQLParserOR             = 16
	TQLParserSELECT         = 17
	TQLParserTHEN           = 18
	TQLParserWHERE          = 19
	TQLParserWHEN           = 20
	TQLParserGROUP          = 21
	TQLParserBY             = 22
	TQLParserTUMBLINGWINDOW = 23
	TQLParserHOPPINGWINDOW  = 24
	TQLParserSLIDINGWINDOW  = 25
	TQLParserSESSIONWINDOW  = 26
	TQLParserDD             = 27
	TQLParserHH             = 28
	TQLParserMI             = 29
	TQLParserSS             = 30
	TQLParserMS             = 31
	TQLParserMUL            = 32
	TQLParserDIV            = 33
	TQLParserMOD            = 34
	TQLParserADD            = 35
	TQLParserSUB            = 36
	TQLParserDOT            = 37
	TQLParserTRUE           = 38
	TQLParserFALSE          = 39
	TQLParserENTITYNAME     = 40
	TQLParserPROPERTYNAME   = 41
	TQLParserNUMBER         = 42
	TQLParserINTEGER        = 43
	TQLParserFLOAT          = 44
	TQLParserSTRING         = 45
	TQLParserWHITESPACE     = 46
)

// TQLParser rules.
const (
	TQLParserRULE_root         = 0
	TQLParserRULE_fields       = 1
	TQLParserRULE_expr         = 2
	TQLParserRULE_sourceEntity = 3
	TQLParserRULE_targetEntity = 4
)

// IRootContext is an interface to support dynamic dispatch.
type IRootContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsRootContext differentiates from other interfaces.
	IsRootContext()
}

type RootContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRootContext() *RootContext {
	var p = new(RootContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = TQLParserRULE_root
	return p
}

func (*RootContext) IsRootContext() {}

func NewRootContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RootContext {
	var p = new(RootContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = TQLParserRULE_root

	return p
}

func (s *RootContext) GetParser() antlr.Parser { return s.parser }

func (s *RootContext) SELECT() antlr.TerminalNode {
	return s.GetToken(TQLParserSELECT, 0)
}

func (s *RootContext) Fields() IFieldsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFieldsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IFieldsContext)
}

func (s *RootContext) EOF() antlr.TerminalNode {
	return s.GetToken(TQLParserEOF, 0)
}

func (s *RootContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RootContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RootContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterRoot(s)
	}
}

func (s *RootContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitRoot(s)
	}
}

func (p *TQLParser) Root() (localctx IRootContext) {
	localctx = NewRootContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, TQLParserRULE_root)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(10)
		p.Match(TQLParserSELECT)
	}
	{
		p.SetState(11)
		p.Fields()
	}
	{
		p.SetState(12)
		p.Match(TQLParserEOF)
	}

	return localctx
}

// IFieldsContext is an interface to support dynamic dispatch.
type IFieldsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsFieldsContext differentiates from other interfaces.
	IsFieldsContext()
}

type FieldsContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldsContext() *FieldsContext {
	var p = new(FieldsContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = TQLParserRULE_fields
	return p
}

func (*FieldsContext) IsFieldsContext() {}

func NewFieldsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldsContext {
	var p = new(FieldsContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = TQLParserRULE_fields

	return p
}

func (s *FieldsContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldsContext) AllExpr() []IExprContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExprContext)(nil)).Elem())
	var tst = make([]IExprContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExprContext)
		}
	}

	return tst
}

func (s *FieldsContext) Expr(i int) IExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExprContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *FieldsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterFields(s)
	}
}

func (s *FieldsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitFields(s)
	}
}

func (p *TQLParser) Fields() (localctx IFieldsContext) {
	localctx = NewFieldsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, TQLParserRULE_fields)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(14)
		p.expr(0)
	}
	p.SetState(19)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == TQLParserT__0 {
		{
			p.SetState(15)
			p.Match(TQLParserT__0)
		}
		{
			p.SetState(16)
			p.expr(0)
		}

		p.SetState(21)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}

	return localctx
}

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = TQLParserRULE_expr
	return p
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = TQLParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) CopyFrom(ctx *ExprContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type CompareValueContext struct {
	*ExprContext
	op antlr.Token
}

func NewCompareValueContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CompareValueContext {
	var p = new(CompareValueContext)

	p.ExprContext = NewEmptyExprContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExprContext))

	return p
}

func (s *CompareValueContext) GetOp() antlr.Token { return s.op }

func (s *CompareValueContext) SetOp(v antlr.Token) { s.op = v }

func (s *CompareValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CompareValueContext) AllExpr() []IExprContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExprContext)(nil)).Elem())
	var tst = make([]IExprContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExprContext)
		}
	}

	return tst
}

func (s *CompareValueContext) Expr(i int) IExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExprContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *CompareValueContext) EQ() antlr.TerminalNode {
	return s.GetToken(TQLParserEQ, 0)
}

func (s *CompareValueContext) GT() antlr.TerminalNode {
	return s.GetToken(TQLParserGT, 0)
}

func (s *CompareValueContext) LT() antlr.TerminalNode {
	return s.GetToken(TQLParserLT, 0)
}

func (s *CompareValueContext) GTE() antlr.TerminalNode {
	return s.GetToken(TQLParserGTE, 0)
}

func (s *CompareValueContext) LTE() antlr.TerminalNode {
	return s.GetToken(TQLParserLTE, 0)
}

func (s *CompareValueContext) NE() antlr.TerminalNode {
	return s.GetToken(TQLParserNE, 0)
}

func (s *CompareValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterCompareValue(s)
	}
}

func (s *CompareValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitCompareValue(s)
	}
}

type ExpressionContext struct {
	*ExprContext
}

func NewExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ExpressionContext {
	var p = new(ExpressionContext)

	p.ExprContext = NewEmptyExprContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExprContext))

	return p
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) AS() antlr.TerminalNode {
	return s.GetToken(TQLParserAS, 0)
}

func (s *ExpressionContext) AllSourceEntity() []ISourceEntityContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISourceEntityContext)(nil)).Elem())
	var tst = make([]ISourceEntityContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISourceEntityContext)
		}
	}

	return tst
}

func (s *ExpressionContext) SourceEntity(i int) ISourceEntityContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceEntityContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISourceEntityContext)
}

func (s *ExpressionContext) AllTargetEntity() []ITargetEntityContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ITargetEntityContext)(nil)).Elem())
	var tst = make([]ITargetEntityContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ITargetEntityContext)
		}
	}

	return tst
}

func (s *ExpressionContext) TargetEntity(i int) ITargetEntityContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITargetEntityContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ITargetEntityContext)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitExpression(s)
	}
}

type MulDivContext struct {
	*ExprContext
	op antlr.Token
}

func NewMulDivContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MulDivContext {
	var p = new(MulDivContext)

	p.ExprContext = NewEmptyExprContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExprContext))

	return p
}

func (s *MulDivContext) GetOp() antlr.Token { return s.op }

func (s *MulDivContext) SetOp(v antlr.Token) { s.op = v }

func (s *MulDivContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MulDivContext) AllExpr() []IExprContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExprContext)(nil)).Elem())
	var tst = make([]IExprContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExprContext)
		}
	}

	return tst
}

func (s *MulDivContext) Expr(i int) IExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExprContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *MulDivContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterMulDiv(s)
	}
}

func (s *MulDivContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitMulDiv(s)
	}
}

type AddSubContext struct {
	*ExprContext
	op antlr.Token
}

func NewAddSubContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AddSubContext {
	var p = new(AddSubContext)

	p.ExprContext = NewEmptyExprContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExprContext))

	return p
}

func (s *AddSubContext) GetOp() antlr.Token { return s.op }

func (s *AddSubContext) SetOp(v antlr.Token) { s.op = v }

func (s *AddSubContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AddSubContext) AllExpr() []IExprContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExprContext)(nil)).Elem())
	var tst = make([]IExprContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExprContext)
		}
	}

	return tst
}

func (s *AddSubContext) Expr(i int) IExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExprContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *AddSubContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterAddSub(s)
	}
}

func (s *AddSubContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitAddSub(s)
	}
}

func (p *TQLParser) Expr() (localctx IExprContext) {
	return p.expr(0)
}

func (p *TQLParser) expr(_p int) (localctx IExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()
	_parentState := p.GetState()
	localctx = NewExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 4
	p.EnterRecursionRule(localctx, 4, TQLParserRULE_expr, _p)
	var _la int

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	localctx = NewExpressionContext(p, localctx)
	p.SetParserRuleContext(localctx)
	_prevctx = localctx

	p.SetState(24)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == TQLParserENTITYNAME {
		{
			p.SetState(23)
			p.SourceEntity()
		}

		p.SetState(26)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(28)
		p.Match(TQLParserAS)
	}
	p.SetState(30)
	p.GetErrorHandler().Sync(p)
	_alt = 1
	for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		switch _alt {
		case 1:
			{
				p.SetState(29)
				p.TargetEntity()
			}

		default:
			panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		}

		p.SetState(32)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())
	}

	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(45)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 4, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(43)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext()) {
			case 1:
				localctx = NewMulDivContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(34)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
				}
				p.SetState(35)

				var _lt = p.GetTokenStream().LT(1)

				localctx.(*MulDivContext).op = _lt

				_la = p.GetTokenStream().LA(1)

				if !(((_la-32)&-(0x1f+1)) == 0 && ((1<<uint((_la-32)))&((1<<(TQLParserMUL-32))|(1<<(TQLParserDIV-32))|(1<<(TQLParserMOD-32)))) != 0) {
					var _ri = p.GetErrorHandler().RecoverInline(p)

					localctx.(*MulDivContext).op = _ri
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
				{
					p.SetState(36)
					p.expr(4)
				}

			case 2:
				localctx = NewAddSubContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(37)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
				}
				p.SetState(38)

				var _lt = p.GetTokenStream().LT(1)

				localctx.(*AddSubContext).op = _lt

				_la = p.GetTokenStream().LA(1)

				if !(_la == TQLParserADD || _la == TQLParserSUB) {
					var _ri = p.GetErrorHandler().RecoverInline(p)

					localctx.(*AddSubContext).op = _ri
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
				{
					p.SetState(39)
					p.expr(3)
				}

			case 3:
				localctx = NewCompareValueContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(40)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
				}
				p.SetState(41)

				var _lt = p.GetTokenStream().LT(1)

				localctx.(*CompareValueContext).op = _lt

				_la = p.GetTokenStream().LA(1)

				if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<TQLParserEQ)|(1<<TQLParserGT)|(1<<TQLParserGTE)|(1<<TQLParserLT)|(1<<TQLParserLTE)|(1<<TQLParserNE))) != 0) {
					var _ri = p.GetErrorHandler().RecoverInline(p)

					localctx.(*CompareValueContext).op = _ri
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
				{
					p.SetState(42)
					p.expr(2)
				}

			}

		}
		p.SetState(47)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 4, p.GetParserRuleContext())
	}

	return localctx
}

// ISourceEntityContext is an interface to support dynamic dispatch.
type ISourceEntityContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSourceEntityContext differentiates from other interfaces.
	IsSourceEntityContext()
}

type SourceEntityContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySourceEntityContext() *SourceEntityContext {
	var p = new(SourceEntityContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = TQLParserRULE_sourceEntity
	return p
}

func (*SourceEntityContext) IsSourceEntityContext() {}

func NewSourceEntityContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SourceEntityContext {
	var p = new(SourceEntityContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = TQLParserRULE_sourceEntity

	return p
}

func (s *SourceEntityContext) GetParser() antlr.Parser { return s.parser }

func (s *SourceEntityContext) ENTITYNAME() antlr.TerminalNode {
	return s.GetToken(TQLParserENTITYNAME, 0)
}

func (s *SourceEntityContext) PROPERTYNAME() antlr.TerminalNode {
	return s.GetToken(TQLParserPROPERTYNAME, 0)
}

func (s *SourceEntityContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SourceEntityContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SourceEntityContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterSourceEntity(s)
	}
}

func (s *SourceEntityContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitSourceEntity(s)
	}
}

func (p *TQLParser) SourceEntity() (localctx ISourceEntityContext) {
	localctx = NewSourceEntityContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, TQLParserRULE_sourceEntity)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(48)
		p.Match(TQLParserENTITYNAME)
	}
	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == TQLParserPROPERTYNAME {
		{
			p.SetState(49)
			p.Match(TQLParserPROPERTYNAME)
		}

	}

	return localctx
}

// ITargetEntityContext is an interface to support dynamic dispatch.
type ITargetEntityContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsTargetEntityContext differentiates from other interfaces.
	IsTargetEntityContext()
}

type TargetEntityContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTargetEntityContext() *TargetEntityContext {
	var p = new(TargetEntityContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = TQLParserRULE_targetEntity
	return p
}

func (*TargetEntityContext) IsTargetEntityContext() {}

func NewTargetEntityContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TargetEntityContext {
	var p = new(TargetEntityContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = TQLParserRULE_targetEntity

	return p
}

func (s *TargetEntityContext) GetParser() antlr.Parser { return s.parser }

func (s *TargetEntityContext) ENTITYNAME() antlr.TerminalNode {
	return s.GetToken(TQLParserENTITYNAME, 0)
}

func (s *TargetEntityContext) PROPERTYNAME() antlr.TerminalNode {
	return s.GetToken(TQLParserPROPERTYNAME, 0)
}

func (s *TargetEntityContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TargetEntityContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TargetEntityContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterTargetEntity(s)
	}
}

func (s *TargetEntityContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitTargetEntity(s)
	}
}

func (p *TQLParser) TargetEntity() (localctx ITargetEntityContext) {
	localctx = NewTargetEntityContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, TQLParserRULE_targetEntity)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(52)
		p.Match(TQLParserENTITYNAME)
	}
	p.SetState(54)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 6, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(53)
			p.Match(TQLParserPROPERTYNAME)
		}

	}

	return localctx
}

func (p *TQLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 2:
		var t *ExprContext = nil
		if localctx != nil {
			t = localctx.(*ExprContext)
		}
		return p.Expr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *TQLParser) Expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 3)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 2)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 1)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
