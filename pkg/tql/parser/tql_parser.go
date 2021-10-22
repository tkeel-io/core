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
	4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 5, 4, 33, 10, 4, 3, 4, 3,
	4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 7, 4, 44, 10, 4, 12, 4, 14,
	4, 47, 11, 4, 3, 5, 3, 5, 3, 5, 5, 5, 52, 10, 5, 3, 6, 3, 6, 3, 6, 5, 6,
	57, 10, 6, 3, 6, 2, 3, 6, 7, 2, 4, 6, 8, 10, 2, 5, 3, 2, 34, 36, 3, 2,
	37, 38, 4, 2, 9, 9, 11, 15, 2, 60, 2, 12, 3, 2, 2, 2, 4, 16, 3, 2, 2, 2,
	6, 32, 3, 2, 2, 2, 8, 48, 3, 2, 2, 2, 10, 53, 3, 2, 2, 2, 12, 13, 7, 19,
	2, 2, 13, 14, 5, 4, 3, 2, 14, 15, 7, 2, 2, 3, 15, 3, 3, 2, 2, 2, 16, 21,
	5, 6, 4, 2, 17, 18, 7, 3, 2, 2, 18, 20, 5, 6, 4, 2, 19, 17, 3, 2, 2, 2,
	20, 23, 3, 2, 2, 2, 21, 19, 3, 2, 2, 2, 21, 22, 3, 2, 2, 2, 22, 5, 3, 2,
	2, 2, 23, 21, 3, 2, 2, 2, 24, 25, 8, 4, 1, 2, 25, 26, 7, 42, 2, 2, 26,
	27, 7, 4, 2, 2, 27, 33, 7, 42, 2, 2, 28, 29, 5, 8, 5, 2, 29, 30, 7, 4,
	2, 2, 30, 31, 5, 10, 6, 2, 31, 33, 3, 2, 2, 2, 32, 24, 3, 2, 2, 2, 32,
	28, 3, 2, 2, 2, 33, 45, 3, 2, 2, 2, 34, 35, 12, 5, 2, 2, 35, 36, 9, 2,
	2, 2, 36, 44, 5, 6, 4, 6, 37, 38, 12, 4, 2, 2, 38, 39, 9, 3, 2, 2, 39,
	44, 5, 6, 4, 5, 40, 41, 12, 3, 2, 2, 41, 42, 9, 4, 2, 2, 42, 44, 5, 6,
	4, 4, 43, 34, 3, 2, 2, 2, 43, 37, 3, 2, 2, 2, 43, 40, 3, 2, 2, 2, 44, 47,
	3, 2, 2, 2, 45, 43, 3, 2, 2, 2, 45, 46, 3, 2, 2, 2, 46, 7, 3, 2, 2, 2,
	47, 45, 3, 2, 2, 2, 48, 51, 7, 42, 2, 2, 49, 50, 7, 39, 2, 2, 50, 52, 7,
	42, 2, 2, 51, 49, 3, 2, 2, 2, 51, 52, 3, 2, 2, 2, 52, 9, 3, 2, 2, 2, 53,
	56, 7, 42, 2, 2, 54, 55, 7, 39, 2, 2, 55, 57, 7, 42, 2, 2, 56, 54, 3, 2,
	2, 2, 56, 57, 3, 2, 2, 2, 57, 11, 3, 2, 2, 2, 8, 21, 32, 43, 45, 51, 56,
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

func (s *ExpressionContext) AllENTITYNAME() []antlr.TerminalNode {
	return s.GetTokens(TQLParserENTITYNAME)
}

func (s *ExpressionContext) ENTITYNAME(i int) antlr.TerminalNode {
	return s.GetToken(TQLParserENTITYNAME, i)
}

func (s *ExpressionContext) AS() antlr.TerminalNode {
	return s.GetToken(TQLParserAS, 0)
}

func (s *ExpressionContext) SourceEntity() ISourceEntityContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceEntityContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISourceEntityContext)
}

func (s *ExpressionContext) TargetEntity() ITargetEntityContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITargetEntityContext)(nil)).Elem(), 0)

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
	p.SetState(30)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext()) {
	case 1:
		localctx = NewExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(23)
			p.Match(TQLParserENTITYNAME)
		}
		{
			p.SetState(24)
			p.Match(TQLParserAS)
		}
		{
			p.SetState(25)
			p.Match(TQLParserENTITYNAME)
		}

	case 2:
		localctx = NewExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(26)
			p.SourceEntity()
		}
		{
			p.SetState(27)
			p.Match(TQLParserAS)
		}
		{
			p.SetState(28)
			p.TargetEntity()
		}

	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(43)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(41)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext()) {
			case 1:
				localctx = NewMulDivContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(32)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
				}
				p.SetState(33)

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
					p.SetState(34)
					p.expr(4)
				}

			case 2:
				localctx = NewAddSubContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(35)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
				}
				p.SetState(36)

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
					p.SetState(37)
					p.expr(3)
				}

			case 3:
				localctx = NewCompareValueContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(38)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
				}
				p.SetState(39)

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
					p.SetState(40)
					p.expr(2)
				}

			}

		}
		p.SetState(45)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext())
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

func (s *SourceEntityContext) AllENTITYNAME() []antlr.TerminalNode {
	return s.GetTokens(TQLParserENTITYNAME)
}

func (s *SourceEntityContext) ENTITYNAME(i int) antlr.TerminalNode {
	return s.GetToken(TQLParserENTITYNAME, i)
}

func (s *SourceEntityContext) DOT() antlr.TerminalNode {
	return s.GetToken(TQLParserDOT, 0)
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
		p.SetState(46)
		p.Match(TQLParserENTITYNAME)
	}
	p.SetState(49)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == TQLParserDOT {
		{
			p.SetState(47)
			p.Match(TQLParserDOT)
		}
		{
			p.SetState(48)
			p.Match(TQLParserENTITYNAME)
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

func (s *TargetEntityContext) AllENTITYNAME() []antlr.TerminalNode {
	return s.GetTokens(TQLParserENTITYNAME)
}

func (s *TargetEntityContext) ENTITYNAME(i int) antlr.TerminalNode {
	return s.GetToken(TQLParserENTITYNAME, i)
}

func (s *TargetEntityContext) DOT() antlr.TerminalNode {
	return s.GetToken(TQLParserDOT, 0)
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
		p.SetState(51)
		p.Match(TQLParserENTITYNAME)
	}
	p.SetState(54)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 5, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(52)
			p.Match(TQLParserDOT)
		}
		{
			p.SetState(53)
			p.Match(TQLParserENTITYNAME)
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
