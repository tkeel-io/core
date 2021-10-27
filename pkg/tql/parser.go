package tql

import (
	"fmt"
	"strings"
	"strconv"

	"tql/parser"
	"github.com/antlr/antlr4/runtime/Go/antlr"
)


type TQLListener struct {
	*parser.BaseTQLListener

	targetEntity []string
	sourceEntity []string
	tentacles    map[string][]string
	input        map[string]interface{}
	output       map[string]interface{}
	
	execs        []*Exec
	// computing results
	stack        []int
}

type Exec struct {
	SourceEntities      []string
	TargetProperty 		string
	// each split of ExitFields
	Field               string
	// real math expression from Field which replace key with value
	Expression          string
}

func (l *TQLListener) pushT(entity string) {
	for _, e := range l.targetEntity{
		if e == entity{
			return
		}
	}
	l.targetEntity = append(l.targetEntity, entity)
}

func (l *TQLListener) pushS(entity string) {
	for _, e := range l.sourceEntity{
		if e == entity{
			return
		}
	}
	l.sourceEntity = append(l.sourceEntity, entity)
}

func (l *TQLListener) AddTentacle(k string, v string) {
	fmt.Println("AddTentacle ", k, v)
	if _, err := l.tentacles[k]; !err{
		var vv []string
		vv = append(vv, v)
		if l.tentacles == nil{
			l.tentacles = make(map[string][]string)
		}
		l.tentacles[k] = vv
	}else{
		for _, e := range l.tentacles[k]{
			if e == v{
				return
			}
		}
		l.tentacles[k] = append(l.tentacles[k], v)
	}
	l.pushS(k)
}

// ExitSourceEntity is called when production entity is exited.
func (l *TQLListener) ExitSourceEntity(c *parser.SourceEntityContext) {
	fmt.Println("ExitSourceEntity", c.GetText())
	text := c.GetText()
	if strings.Contains(text, ".") {
		arr := strings.Split(text, ".")
		l.AddTentacle(arr[0], arr[1])
	} else {
		l.pushS(text)
	}

	//record SourceEntities
	if len(l.execs) > 0 {
		e := l.execs[len(l.execs)-1]
		if len(e.SourceEntities)-len(e.TargetProperty) == 1 {
			e.SourceEntities = append(e.SourceEntities, text)
			return
		}
	}
	var ne Exec
	ne.SourceEntities = append(ne.SourceEntities, text)
	fmt.Println("add SourceEntities:", ne)
	l.execs = append(l.execs, &ne)
}

// ExitTargetEntity is called when production entity is exited.
func (l *TQLListener) ExitTargetEntity(c *parser.TargetEntityContext) {
	fmt.Println("ExitTargetEntity",c.GetText())
	text := c.GetText()
	if strings.Contains(text, "."){
		arr := strings.Split(text, ".")
		l.pushT(arr[0])
	}else{
		l.pushT(text)
	}
}

// ExitTargeProperty is called when production entity is exited.
func (l *TQLListener) ExitTargetProperty(c *parser.TargetPropertyContext) {
	fmt.Println("TargetPropertyContext",c.GetText())
	tp := c.GetText()
	// record TargetProperty
	e := l.execs[len(l.execs)-1]
	e.TargetProperty = tp
	fmt.Println("add TargetProperty:", e)

}

// ExitExpression is called when production Expression is exited.
func (l *TQLListener) ExitExpression(c *parser.ExpressionContext) {
	//fmt.Println("ExitExpression",c.GetText())
}

// ExitRoot is called when production root is exited.
func (l *TQLListener) ExitRoot(c *parser.RootContext) {
	//fmt.Println("ExitRoot",c.GetText())
}

// ExitFields is called when production fields is exited.
func (l *TQLListener) ExitFields(c *parser.FieldsContext) {
	fmt.Println("ExitFields",c.GetText())
	//record Field
	fields := c.GetText()
	fieldArr := strings.Split(fields, ",")
	for ind, f := range fieldArr {
		field := strings.Split(f, "as")[0]
		e := l.execs[ind]
		e.Field = field
		fmt.Println("====add field:", field)
	}
}

// ExitCompareValue is called when production CompareValue is exited.
func (l *TQLListener) ExitCompareValue(c *parser.CompareValueContext) {
	fmt.Println("ExitCompareValue",c.GetText())
}

func (l *TQLListener) GetExpression(index int, in map[string]interface{}) string{
	e := l.execs[index]
	for k, v := range in{
		// convert to string, need to add space otherwise "expecting <EOF>" error
		nk := " "+ fmt.Sprintf("%v", v) + " "
		e.Field = strings.ReplaceAll(e.Field, k, nk)
	}
	e.Expression = e.Field
	return e.Expression
}

func (l *TQLListener) push(i int) {
	l.stack = append(l.stack, i)
}

func (l *TQLListener) pop() int {
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
func (l *TQLListener) ExitNumber(c *parser.NumberContext) {
	i, err := strconv.Atoi(c.GetText())
	if err != nil {
		panic(err.Error())
	}

	l.push(i)
}

// ExitMulDiv is called when exiting the MulDiv production.
func (l *TQLListener) ExitMulDiv(c *parser.MulDivContext) {
	fmt.Println("ExitMulDiv",c.GetText())
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
func (l *TQLListener) ExitAddSub(c *parser.AddSubContext) {
	fmt.Println("ExitAddSub",c.GetText())
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
func computing(input string) int{
	// Setup the input
	is := antlr.NewInputStream(input)

	// Create the Lexer
	lexer := parser.NewTQLLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the Parser
	p := parser.NewTQLParser(stream)

	// Finally parse the expression (by walking the tree)
	var listener TQLListener
	antlr.ParseTreeWalkerDefault.Walk(&listener, p.Computing())
	//fmt.Println("========results: \n", listener.pop())
	return listener.pop()
}

func (l *TQLListener)GetParseConfigs() map[string]interface{}{
	configMap := make(map[string]interface{})
	configMap["SourceEntity"] = l.sourceEntity
	configMap["TargetEntity"] = l.targetEntity
	// if tentacles is null map, it should be map["*"]["*", ]
	if l.tentacles == nil{
		l.tentacles = make(map[string][]string)
		l.tentacles["*"] = append(l.tentacles["*"], "*")
	}
	configMap["Tentacles"] = l.tentacles
	return configMap
}

func (l *TQLListener)GetComputeResults(in map[string]interface{}) map[string]interface{}{
	//
	//fmt.Println("get Expression:", l.GetExpression(0, in))

	out := make(map[string]interface{})
	for ind, e := range l.execs{
		numExpr := l.GetExpression(ind, in)
		 out[e.TargetProperty] = computing(numExpr)
	}
	return out
}

// Parse takes a tql string expression and returns a parsed dict.
func Parse(input string) TQLListener {
	// Setup the input
	is := antlr.NewInputStream(input)

	// Create the Lexer
	lexer := parser.NewTQLLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the Parser
	p := parser.NewTQLParser(stream)

	// Finally parse the expression (by walking the tree)
	var listener TQLListener
	antlr.ParseTreeWalkerDefault.Walk(&listener, p.Root())
	//fmt.Println("\n\nget sourceEntity", listener.sourceEntity)
	//fmt.Println("get targetEntity", listener.targetEntity)
	//fmt.Println("get tentacles", listener.tentacles)
	return listener
}
