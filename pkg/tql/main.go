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
	input        map[string]interface{}
	output       map[string]interface{}
}

func (l *TQLListener) pushT(entity string) {
	for _, e := range l.targetEntity{
		if e == entity{
			return
		}
	}
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
	for _, e := range l.sourceEntity{
		if e == entity{
			return
		}
	}
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
	fmt.Println("ExitSourceEntity",c.GetText())
	text := c.GetText()
	if strings.Contains(text, "."){
		arr := strings.Split(text, ".")
		l.AddTentacle(arr[0], arr[1])
	}else{
		l.pushS(text)
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

// ExitTargeProperty is called when production entity is exited.
func (l *TQLListener) ExitTargeProperty(c *parser.TargetPropertyContext) {
	fmt.Println("TargetPropertyContext",c.GetText())
	//text := c.GetText()
	//if strings.Contains(text, "."){
	//	arr := strings.Split(text, ".")
	//	l.pushT(arr[0])
	//}else{
	//	l.pushT(text)
	//}
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
	l.output = make(map[string]interface{})
	for _, v := range in {
		l.output["property2"] = v
	}
	return l.output
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

func main() {
	//tql := `insert into entity3 select
	//	entity1.property1 as property1,
	//	entity2.property2.name as property2,
	//	entity1.property1 + entity2.property3 as property3`

	tql := `insert into target_entity select *`

	fmt.Println("parse tql: ", tql)
	l := Parse(tql)
	cfg := l.GetParseConfigs()
	fmt.Println("========\n ", cfg)

	// 如果是entity, 则为 entity 的 properties map; 如果是property 则为单独的 property map
	in := make(map[string]interface{})
	in["property1"] = 1
	fmt.Println("in: ", in)
	out := l.GetComputeResults(in)
	fmt.Println("out: ", out)
}
