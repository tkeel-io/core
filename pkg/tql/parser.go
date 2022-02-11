/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/dop251/goja"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/tql/parser"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const Sep = "."

type OpKind string

const (
	OpKindString OpKind = "string"
	OpKindNumber OpKind = "number"
	OpKindIgnore OpKind = "ignore"
)

type Listener struct {
	*parser.BaseTQLListener

	targetEntity []string
	sourceEntity []string
	tentacles    map[string][]string
	// input        map[string]interface{}
	// output       map[string]interface{}

	opKind OpKind
	// computing results
	stack    []int
	strStack []string

	evalContexts map[string]EvalContext
}

func newListener() *Listener {
	return &Listener{
		evalContexts: make(map[string]EvalContext),
	}
}

type EvalContext struct {
	Field             string
	ParamKeys         []string
	TargetPropertyKey string
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

func (l *Listener) ExitSource(c *parser.SourceContext) {
	text := c.GetText()
	log.Info("ExitSource", text)
	SourceEntity := c.SourceEntity().GetText()
	PropertyEntity := c.PropertyEntity().GetText()
	l.pushS(SourceEntity)
	l.AddTentacle(SourceEntity, strings.TrimPrefix(PropertyEntity, Sep))
}

// ExitSourceEntity is called when production entity is exited.
func (l *Listener) ExitSourceEntity(c *parser.SourceEntityContext) {
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
	log.Info("ExitTargetProperty", c.GetText())
}

func (l *Listener) ExitField(c *parser.FieldContext) {
	evalCtx := EvalContext{Field: c.Expr().GetText()}
	if c.TargetProperty() != nil {
		evalCtx.TargetPropertyKey = c.TargetProperty().GetText()
	}

	evalCtx.ParamKeys = getParamKeys(c.Expr().GetChildren())
	l.evalContexts[c.GetText()] = evalCtx
}

func getParamKeys(nodes []antlr.Tree) []string {
	var sources []string
	for _, node := range nodes {
		if sourceContext, ok := node.(*parser.SourceContext); ok {
			sources = append(sources, sourceContext.GetText())
			continue
		}
		sources = append(sources, getParamKeys(node.GetChildren())...)
	}

	return sources
}

// ExitRoot is called when production root is exited.
func (l *Listener) ExitRoot(c *parser.RootContext) {
	log.Info("ExitRoot", c.GetText())
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

// EnterString is called when entering the String production.
func (l *Listener) EnterString(c *parser.StringContext) {
	l.opKind = OpKindString
	l.strStack = append(l.strStack, c.GetText())
}

// ExitMulDiv is called when exiting the MulDiv production.
func (l *Listener) ExitMulDiv(c *parser.MulDivContext) {
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
	switch l.opKind {
	case OpKindNumber:
		right, left := l.pop(), l.pop()
		switch c.GetOp().GetTokenType() {
		case parser.TQLParserADD:
			l.push(left + right)
		case parser.TQLParserSUB:
			l.push(left - right)
		default:
			panic(fmt.Sprintf("unexpected operation: %s", c.GetOp().GetText()))
		}
	case OpKindString:
		right, left := l.strStack[1], l.strStack[0]
		l.strStack = []string{}
		switch c.GetOp().GetTokenType() {
		case parser.TQLParserADD:
			val := (left[1:len(left)-1] + right[1:len(right)-1])
			l.strStack = append(l.strStack, "'"+val+"'")
		case parser.TQLParserSUB:
			panic("not support string sub")
		default:
			panic(fmt.Sprintf("unexpected operation: %s", c.GetOp().GetText()))
		}
	}
}

func (l *Listener) GetParseConfigs() (TQLConfig, error) {
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

	log.Debug("result of TQL", zap.Any("result", tqlConfig))
	return tqlConfig, nil
}

func (l *Listener) GetComputeResults(in map[string][]byte) map[string]constraint.Node {
	out := make(map[string]constraint.Node)
	for _, evalCtx := range l.evalContexts {
		params := make(map[string][]byte)
		for _, propertyKey := range evalCtx.ParamKeys {
			if val, ok := in[propertyKey]; ok {
				params[propertyKey] = val
				continue
			}
			break
		}

		if len(params) == len(evalCtx.ParamKeys) {
			evalExpr := evalCtx.Field
			for key, val := range params {
				evalExpr = strings.ReplaceAll(
					evalExpr, key, " "+string(val)+" ")
			}
			val, _ := goja.New().RunString(evalExpr)
			out[evalCtx.TargetPropertyKey] = constraint.NewNode(val.Export())
		}
	}

	return out
}

func (l *Listener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	log.Error("SyntaxError", recognizer, offendingSymbol, line, column, msg, e)
}

// Parse takes a tql string expression and returns a parsed dict.
func Parse(input string) (*Listener, error) {
	// Setup the input
	is := antlr.NewInputStream(input)

	// Create the Lexer
	lexer := parser.NewTQLLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the Parser
	p := parser.NewTQLParser(stream)

	// Finally parse the expression (by walking the tree)
	var listener = newListener()
	antlr.ParseTreeWalkerDefault.Walk(listener, p.Root())
	return listener, nil
}
