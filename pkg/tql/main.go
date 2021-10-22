// example.go
package main

import (
	"fmt"
	"strings"
	//"strconv"

	"tql/parser"
	"github.com/antlr/antlr4/runtime/Go/antlr"
)


type TQLListener struct {
	*parser.BaseTQLListener

	targetEntity []string
	sourceEntity []string
	tentacles    map[string][]string
}

func (l *TQLListener) pushT(entity string) {
	l.targetEntity = append(l.targetEntity, entity)
}

func (l *TQLListener) popT() string {
	if len(l.targetEntity) < 1 {
		panic("stack is empty unable to pop")
	}

	// Get the last value from the stack.
	result := l.targetEntity[len(l.targetEntity)-1]

	// Pop the last element from the stack.
	l.targetEntity = l.targetEntity[:len(l.targetEntity)-1]

	return result
}
func (l *TQLListener) pushS(entity string) {
	l.sourceEntity = append(l.sourceEntity, entity)
}

func (l *TQLListener) popS() string {
	if len(l.sourceEntity) < 1 {
		panic("stack is empty unable to pop")
	}

	// Get the last value from the stack.
	result := l.sourceEntity[len(l.sourceEntity)-1]

	// Pop the last element from the stack.
	l.sourceEntity = l.sourceEntity[:len(l.sourceEntity)-1]

	return result
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
		l.tentacles[k] = append(l.tentacles[k], v)
	}
	l.pushS(k)
}

// ExitSourceEntity is called when production entity is exited.
func (l *TQLListener) ExitSourceEntity(c *parser.SourceEntityContext) {
	fmt.Println("ExitSourceEntity",c.GetText())
	text := c.GetText()
	if strings.Contains(text, "."){
		arr := strings.Split(text, ".")
		l.AddTentacle(arr[0], arr[1])
	}else{
		l.pushT(text)
	}
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

// ExitExpression is called when production Expression is exited.
func (l *TQLListener) ExitExpression(c *parser.ExpressionContext) {
	fmt.Println("ExitExpression",c.GetText())
}

// ExitRoot is called when production root is exited.
func (l *TQLListener) ExitRoot(c *parser.RootContext) {
	fmt.Println("ExitRoot",c.GetText())
}

// ExitFields is called when production fields is exited.
func (l *TQLListener) ExitFields(c *parser.FieldsContext) {
	fmt.Println("ExitFields",c.GetText())
}

// ExitCompareValue is called when production CompareValue is exited.
func (l *TQLListener) ExitCompareValue(c *parser.CompareValueContext) {
	fmt.Println("ExitCompareValue",c.GetText())
}

// Parse takes a tql string expression and returns a parsed dict.
func Parse(input string) {
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
	fmt.Println("\n\nget sourceEntity", listener.sourceEntity)
	fmt.Println("get targetEntity", listener.targetEntity)
	fmt.Println("get tentacles", listener.tentacles)
}

func main() {
	tql := `select src_entityA.property1 AS tar_entityB.property2`
	//tql := `select
	//		entity AS entityD`
	//tql := `select entityA.property1 AS property2
    //               //entityB.property2 AS property3`
	fmt.Println("parse tql: ", tql)
	Parse(tql)
}


