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
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 51, 70, 4,
	2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7, 3,
	2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 7, 3, 25, 10,
	3, 12, 3, 14, 3, 28, 11, 3, 3, 4, 3, 4, 3, 5, 3, 5, 3, 5, 6, 5, 35, 10,
	5, 13, 5, 14, 5, 36, 3, 5, 3, 5, 6, 5, 41, 10, 5, 13, 5, 14, 5, 42, 5,
	5, 45, 10, 5, 3, 5, 3, 5, 3, 5, 3, 5, 3, 5, 3, 5, 3, 5, 3, 5, 3, 5, 7,
	5, 56, 10, 5, 12, 5, 14, 5, 59, 11, 5, 3, 6, 3, 6, 3, 6, 5, 6, 64, 10,
	6, 5, 6, 66, 10, 6, 3, 7, 3, 7, 3, 7, 2, 3, 8, 8, 2, 4, 6, 8, 10, 12, 2,
	5, 3, 2, 36, 38, 3, 2, 39, 40, 4, 2, 11, 11, 13, 17, 2, 72, 2, 14, 3, 2,
	2, 2, 4, 21, 3, 2, 2, 2, 6, 29, 3, 2, 2, 2, 8, 44, 3, 2, 2, 2, 10, 65,
	3, 2, 2, 2, 12, 67, 3, 2, 2, 2, 14, 15, 7, 4, 2, 2, 15, 16, 7, 5, 2, 2,
	16, 17, 5, 6, 4, 2, 17, 18, 7, 21, 2, 2, 18, 19, 5, 4, 3, 2, 19, 20, 7,
	2, 2, 3, 20, 3, 3, 2, 2, 2, 21, 26, 5, 8, 5, 2, 22, 23, 7, 3, 2, 2, 23,
	25, 5, 8, 5, 2, 24, 22, 3, 2, 2, 2, 25, 28, 3, 2, 2, 2, 26, 24, 3, 2, 2,
	2, 26, 27, 3, 2, 2, 2, 27, 5, 3, 2, 2, 2, 28, 26, 3, 2, 2, 2, 29, 30, 7,
	46, 2, 2, 30, 7, 3, 2, 2, 2, 31, 32, 8, 5, 1, 2, 32, 45, 5, 10, 6, 2, 33,
	35, 5, 10, 6, 2, 34, 33, 3, 2, 2, 2, 35, 36, 3, 2, 2, 2, 36, 34, 3, 2,
	2, 2, 36, 37, 3, 2, 2, 2, 37, 38, 3, 2, 2, 2, 38, 40, 7, 6, 2, 2, 39, 41,
	5, 12, 7, 2, 40, 39, 3, 2, 2, 2, 41, 42, 3, 2, 2, 2, 42, 40, 3, 2, 2, 2,
	42, 43, 3, 2, 2, 2, 43, 45, 3, 2, 2, 2, 44, 31, 3, 2, 2, 2, 44, 34, 3,
	2, 2, 2, 45, 57, 3, 2, 2, 2, 46, 47, 12, 5, 2, 2, 47, 48, 9, 2, 2, 2, 48,
	56, 5, 8, 5, 6, 49, 50, 12, 4, 2, 2, 50, 51, 9, 3, 2, 2, 51, 56, 5, 8,
	5, 5, 52, 53, 12, 3, 2, 2, 53, 54, 9, 4, 2, 2, 54, 56, 5, 8, 5, 4, 55,
	46, 3, 2, 2, 2, 55, 49, 3, 2, 2, 2, 55, 52, 3, 2, 2, 2, 56, 59, 3, 2, 2,
	2, 57, 55, 3, 2, 2, 2, 57, 58, 3, 2, 2, 2, 58, 9, 3, 2, 2, 2, 59, 57, 3,
	2, 2, 2, 60, 66, 7, 36, 2, 2, 61, 63, 7, 44, 2, 2, 62, 64, 7, 45, 2, 2,
	63, 62, 3, 2, 2, 2, 63, 64, 3, 2, 2, 2, 64, 66, 3, 2, 2, 2, 65, 60, 3,
	2, 2, 2, 65, 61, 3, 2, 2, 2, 66, 11, 3, 2, 2, 2, 67, 68, 7, 44, 2, 2, 68,
	13, 3, 2, 2, 2, 10, 26, 36, 42, 44, 55, 57, 63, 65,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "','", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "'*'",
	"'/'", "'%'", "'+'", "'-'", "'.'",
}
var symbolicNames = []string{
	"", "", "INSERT", "INTO", "AS", "AND", "CASE", "ELSE", "END", "EQ", "FROM",
	"GT", "GTE", "LT", "LTE", "NE", "NOT", "NULL", "OR", "SELECT", "THEN",
	"WHERE", "WHEN", "GROUP", "BY", "TUMBLINGWINDOW", "HOPPINGWINDOW", "SLIDINGWINDOW",
	"SESSIONWINDOW", "DD", "HH", "MI", "SS", "MS", "MUL", "DIV", "MOD", "ADD",
	"SUB", "DOT", "TRUE", "FALSE", "ENTITYNAME", "PROPERTYNAME", "TARGETENTITY",
	"NUMBER", "INTEGER", "FLOAT", "STRING", "WHITESPACE",
}

var ruleNames = []string{
	"root", "fields", "targetEntity", "expr", "sourceEntity", "targetProperty",
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
	TQLParserINSERT         = 2
	TQLParserINTO           = 3
	TQLParserAS             = 4
	TQLParserAND            = 5
	TQLParserCASE           = 6
	TQLParserELSE           = 7
	TQLParserEND            = 8
	TQLParserEQ             = 9
	TQLParserFROM           = 10
	TQLParserGT             = 11
	TQLParserGTE            = 12
	TQLParserLT             = 13
	TQLParserLTE            = 14
	TQLParserNE             = 15
	TQLParserNOT            = 16
	TQLParserNULL           = 17
	TQLParserOR             = 18
	TQLParserSELECT         = 19
	TQLParserTHEN           = 20
	TQLParserWHERE          = 21
	TQLParserWHEN           = 22
	TQLParserGROUP          = 23
	TQLParserBY             = 24
	TQLParserTUMBLINGWINDOW = 25
	TQLParserHOPPINGWINDOW  = 26
	TQLParserSLIDINGWINDOW  = 27
	TQLParserSESSIONWINDOW  = 28
	TQLParserDD             = 29
	TQLParserHH             = 30
	TQLParserMI             = 31
	TQLParserSS             = 32
	TQLParserMS             = 33
	TQLParserMUL            = 34
	TQLParserDIV            = 35
	TQLParserMOD            = 36
	TQLParserADD            = 37
	TQLParserSUB            = 38
	TQLParserDOT            = 39
	TQLParserTRUE           = 40
	TQLParserFALSE          = 41
	TQLParserENTITYNAME     = 42
	TQLParserPROPERTYNAME   = 43
	TQLParserTARGETENTITY   = 44
	TQLParserNUMBER         = 45
	TQLParserINTEGER        = 46
	TQLParserFLOAT          = 47
	TQLParserSTRING         = 48
	TQLParserWHITESPACE     = 49
)

// TQLParser rules.
const (
	TQLParserRULE_root           = 0
	TQLParserRULE_fields         = 1
	TQLParserRULE_targetEntity   = 2
	TQLParserRULE_expr           = 3
	TQLParserRULE_sourceEntity   = 4
	TQLParserRULE_targetProperty = 5
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

func (s *RootContext) INSERT() antlr.TerminalNode {
	return s.GetToken(TQLParserINSERT, 0)
}

func (s *RootContext) INTO() antlr.TerminalNode {
	return s.GetToken(TQLParserINTO, 0)
}

func (s *RootContext) TargetEntity() ITargetEntityContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITargetEntityContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITargetEntityContext)
}

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
		p.SetState(12)
		p.Match(TQLParserINSERT)
	}
	{
		p.SetState(13)
		p.Match(TQLParserINTO)
	}
	{
		p.SetState(14)
		p.TargetEntity()
	}
	{
		p.SetState(15)
		p.Match(TQLParserSELECT)
	}
	{
		p.SetState(16)
		p.Fields()
	}
	{
		p.SetState(17)
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
		p.SetState(19)
		p.expr(0)
	}
	p.SetState(24)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == TQLParserT__0 {
		{
			p.SetState(20)
			p.Match(TQLParserT__0)
		}
		{
			p.SetState(21)
			p.expr(0)
		}

		p.SetState(26)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
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

func (s *TargetEntityContext) TARGETENTITY() antlr.TerminalNode {
	return s.GetToken(TQLParserTARGETENTITY, 0)
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
	p.EnterRule(localctx, 4, TQLParserRULE_targetEntity)

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
		p.SetState(27)
		p.Match(TQLParserTARGETENTITY)
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

func (s *ExpressionContext) AS() antlr.TerminalNode {
	return s.GetToken(TQLParserAS, 0)
}

func (s *ExpressionContext) AllTargetProperty() []ITargetPropertyContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ITargetPropertyContext)(nil)).Elem())
	var tst = make([]ITargetPropertyContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ITargetPropertyContext)
		}
	}

	return tst
}

func (s *ExpressionContext) TargetProperty(i int) ITargetPropertyContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITargetPropertyContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ITargetPropertyContext)
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
	_startState := 6
	p.EnterRecursionRule(localctx, 6, TQLParserRULE_expr, _p)
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
	p.SetState(42)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext()) {
	case 1:
		localctx = NewExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(30)
			p.SourceEntity()
		}

	case 2:
		localctx = NewExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		p.SetState(32)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == TQLParserMUL || _la == TQLParserENTITYNAME {
			{
				p.SetState(31)
				p.SourceEntity()
			}

			p.SetState(34)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(36)
			p.Match(TQLParserAS)
		}
		p.SetState(38)
		p.GetErrorHandler().Sync(p)
		_alt = 1
		for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1:
				{
					p.SetState(37)
					p.TargetProperty()
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(40)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())
		}

	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(55)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 5, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(53)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 4, p.GetParserRuleContext()) {
			case 1:
				localctx = NewMulDivContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(44)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
				}
				p.SetState(45)

				var _lt = p.GetTokenStream().LT(1)

				localctx.(*MulDivContext).op = _lt

				_la = p.GetTokenStream().LA(1)

				if !(((_la-34)&-(0x1f+1)) == 0 && ((1<<uint((_la-34)))&((1<<(TQLParserMUL-34))|(1<<(TQLParserDIV-34))|(1<<(TQLParserMOD-34)))) != 0) {
					var _ri = p.GetErrorHandler().RecoverInline(p)

					localctx.(*MulDivContext).op = _ri
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
				{
					p.SetState(46)
					p.expr(4)
				}

			case 2:
				localctx = NewAddSubContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(47)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
				}
				p.SetState(48)

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
					p.SetState(49)
					p.expr(3)
				}

			case 3:
				localctx = NewCompareValueContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TQLParserRULE_expr)
				p.SetState(50)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
				}
				p.SetState(51)

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
					p.SetState(52)
					p.expr(2)
				}

			}

		}
		p.SetState(57)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 5, p.GetParserRuleContext())
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
	p.EnterRule(localctx, 8, TQLParserRULE_sourceEntity)

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

	p.SetState(63)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case TQLParserMUL:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(58)
			p.Match(TQLParserMUL)
		}

	case TQLParserENTITYNAME:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(59)
			p.Match(TQLParserENTITYNAME)
		}
		p.SetState(61)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 6, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(60)
				p.Match(TQLParserPROPERTYNAME)
			}

		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// ITargetPropertyContext is an interface to support dynamic dispatch.
type ITargetPropertyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsTargetPropertyContext differentiates from other interfaces.
	IsTargetPropertyContext()
}

type TargetPropertyContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTargetPropertyContext() *TargetPropertyContext {
	var p = new(TargetPropertyContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = TQLParserRULE_targetProperty
	return p
}

func (*TargetPropertyContext) IsTargetPropertyContext() {}

func NewTargetPropertyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TargetPropertyContext {
	var p = new(TargetPropertyContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = TQLParserRULE_targetProperty

	return p
}

func (s *TargetPropertyContext) GetParser() antlr.Parser { return s.parser }

func (s *TargetPropertyContext) ENTITYNAME() antlr.TerminalNode {
	return s.GetToken(TQLParserENTITYNAME, 0)
}

func (s *TargetPropertyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TargetPropertyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TargetPropertyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.EnterTargetProperty(s)
	}
}

func (s *TargetPropertyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TQLListener); ok {
		listenerT.ExitTargetProperty(s)
	}
}

func (p *TQLParser) TargetProperty() (localctx ITargetPropertyContext) {
	localctx = NewTargetPropertyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, TQLParserRULE_targetProperty)

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
		p.SetState(65)
		p.Match(TQLParserENTITYNAME)
	}

	return localctx
}

func (p *TQLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 3:
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
