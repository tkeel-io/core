package tql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/tql/parser"
)

var log = logger.NewLogger("core.entities")

type Listener struct {
	*parser.BaseTQLListener

	targetEntity []string
	sourceEntity []string
	tentacles    map[string][]string
	// input        map[string]interface{}
	// output       map[string]interface{}

	execs []*Exec
	// computing results
	stack []int
}

type Exec struct {
	SourceEntities []string
	TargetProperty string
	// each split of ExitFields
	Field string
	// real math expression from Field which replace key with value
	Expression string
}

func (l *Listener) pushT(entity string) {
	for _, e := range l.targetEntity {
		if e == entity {
			return
		}
	}
	l.targetEntity = append(l.targetEntity, entity)
}

func (l *Listener) pushS(entity string) {
	for _, e := range l.sourceEntity {
		if e == entity {
			return
		}
	}
	l.sourceEntity = append(l.sourceEntity, entity)
}

func (l *Listener) AddTentacle(k string, v string) {
	log.Infof("AddTentacle %v, %v", k, v)
	if _, err := l.tentacles[k]; !err {
		var vv []string
		vv = append(vv, v)
		if l.tentacles == nil {
			l.tentacles = make(map[string][]string)
		}
		l.tentacles[k] = vv
	} else {
		for _, e := range l.tentacles[k] {
			if e == v {
				return
			}
		}
		l.tentacles[k] = append(l.tentacles[k], v)
	}
	l.pushS(k)
}

// ExitSourceEntity is called when production entity is exited.
func (l *Listener) ExitSourceEntity(c *parser.SourceEntityContext) {
	log.Infof("ExitSourceEntity %v", c.GetText())
	text := c.GetText()
	if strings.Contains(text, ".") {
		arr := strings.Split(text, ".")
		l.AddTentacle(arr[0], arr[1])
	} else {
		l.pushS(text)
	}

	// record SourceEntities
	if len(l.execs) > 0 {
		e := l.execs[len(l.execs)-1]
		if len(e.SourceEntities)-len(e.TargetProperty) == 1 {
			e.SourceEntities = append(e.SourceEntities, text)
			return
		}
	}
	var ne Exec
	ne.SourceEntities = append(ne.SourceEntities, text)
	log.Infof("add SourceEntities: %v", ne)
	l.execs = append(l.execs, &ne)
}

// ExitTargetEntity is called when production entity is exited.
func (l *Listener) ExitTargetEntity(c *parser.TargetEntityContext) {
	log.Info("ExitTargetEntity", c.GetText())
	if text := c.GetText(); strings.Contains(text, ".") {
		arr := strings.Split(text, ".")
		l.pushT(arr[0])
	} else {
		l.pushT(text)
	}
}

// ExitTargeProperty is called when production entity is exited.
func (l *Listener) ExitTargetProperty(c *parser.TargetPropertyContext) {
	log.Info("TargetPropertyContext", c.GetText())
	tp := c.GetText()
	// record TargetProperty
	e := l.execs[len(l.execs)-1]
	e.TargetProperty = tp
	log.Info("add TargetProperty:", e)
}

// ExitExpression is called when production Expression is exited.
func (l *Listener) ExitExpression(c *parser.ExpressionContext) {
	// log.Info("ExitExpression",c.GetText())
}

// ExitRoot is called when production root is exited.
func (l *Listener) ExitRoot(c *parser.RootContext) {
	// log.Info("ExitRoot",c.GetText())
}

// ExitFields is called when production fields is exited.
func (l *Listener) ExitFields(c *parser.FieldsContext) {
	log.Info("ExitFields", c.GetText())
	// record Field
	fields := c.GetText()
	fieldArr := strings.Split(fields, ",")
	for ind, f := range fieldArr {
		field := strings.Split(f, "as")[0]
		e := l.execs[ind]
		e.Field = field
		log.Info("====add field:", field)
	}
}

// ExitCompareValue is called when production CompareValue is exited.
func (l *Listener) ExitCompareValue(c *parser.CompareValueContext) {
	log.Info("ExitCompareValue", c.GetText())
}

func (l *Listener) GetExpression(index int, in map[string]interface{}) string {
	e := l.execs[index]
	field := e.Field
	for k, v := range in {
		// convert to string, need to add space otherwise "expecting <EOF>" error
		nk := " " + fmt.Sprintf("%v", v) + " "
		field = strings.ReplaceAll(field, k, nk)
	}
	e.Expression = field
	return e.Expression
}

func (l *Listener) push(i int) {
	l.stack = append(l.stack, i)
}

func (l *Listener) pop() int {
	if len(l.stack) < 1 {
		panic("stack is empty unable to pop")
	}

	// Get the last value from the stack.
	result := l.stack[len(l.stack)-1]

	// Pop the last element from the stack.
	l.stack = l.stack[:len(l.stack)-1]

	return result
}

// ExitNumber is called when exiting the Number production.
func (l *Listener) ExitNumber(c *parser.NumberContext) {
	i, err := strconv.Atoi(c.GetText())
	if err != nil {
		panic(err.Error())
	}

	l.push(i)
}

// ExitMulDiv is called when exiting the MulDiv production.
func (l *Listener) ExitMulDiv(c *parser.MulDivContext) {
	log.Info("ExitMulDiv", c.GetText())
	right, left := l.pop(), l.pop()

	switch c.GetOp().GetTokenType() {
	case parser.TQLParserMUL:
		l.push(left * right)
	case parser.TQLParserDIV:
		l.push(left / right)
	default:
		panic(fmt.Sprintf("unexpected operation: %s", c.GetOp().GetText()))
	}
}

// ExitAddSub is called when exiting the AddSub production.
func (l *Listener) ExitAddSub(c *parser.AddSubContext) {
	log.Info("ExitAddSub", c.GetText())
	right, left := l.pop(), l.pop()

	switch c.GetOp().GetTokenType() {
	case parser.TQLParserADD:
		l.push(left + right)
	case parser.TQLParserSUB:
		l.push(left - right)
	default:
		panic(fmt.Sprintf("unexpected operation: %s", c.GetOp().GetText()))
	}
}

// Computing takes a number expression and returns results.
func computing(input string) int {
	// Setup the input
	is := antlr.NewInputStream(input)

	// Create the Lexer
	lexer := parser.NewTQLLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the Parser
	p := parser.NewTQLParser(stream)

	// Finally parse the expression (by walking the tree)
	var listener Listener
	antlr.ParseTreeWalkerDefault.Walk(&listener, p.Computing())
	// log.Info("========results: \n", listener.pop())
	return listener.pop()
}

func (l *Listener) GetParseConfigs() TQLConfig {
	tqlConfig := TQLConfig{
		SourceEntities: l.sourceEntity,
	}

	if len(l.targetEntity) > 0 {
		tqlConfig.TargetEntity = l.targetEntity[0]
	}

	// if tentacles is null map, it should be map["*"]["*", ]
	if l.tentacles == nil {
		tqlConfig.Tentacles = []TentacleConfig{
			{SourceEntity: "*", PropertyKeys: []string{"*"}},
		}
	}

	for entityID, propertyKeys := range l.tentacles {
		tqlConfig.Tentacles = append(tqlConfig.Tentacles,
			TentacleConfig{SourceEntity: entityID, PropertyKeys: propertyKeys})
	}

	return tqlConfig
}

func (l *Listener) GetComputeResults(in map[string]interface{}) map[string]interface{} {
	//
	// log.Info("get Expression:", l.GetExpression(0, in))

	out := make(map[string]interface{})
	for ind, e := range l.execs {
		numExpr := l.GetExpression(ind, in)
		log.Infof("get number expression %s", numExpr)
		out[e.TargetProperty] = computing(numExpr)
	}
	return out
}

// Parse takes a tql string expression and returns a parsed dict.
func Parse(input string) Listener {
	// Setup the input
	is := antlr.NewInputStream(input)

	// Create the Lexer
	lexer := parser.NewTQLLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the Parser
	p := parser.NewTQLParser(stream)

	// Finally parse the expression (by walking the tree)
	var listener Listener
	antlr.ParseTreeWalkerDefault.Walk(&listener, p.Root())
	return listener
}
