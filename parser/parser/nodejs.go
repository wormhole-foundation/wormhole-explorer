package parser

import (
	"fmt"
	"strings"

	v8 "rogchap.com/v8go"
)

type NodeJS struct{}

func (n *NodeJS) Parse(parserFunc string, data []byte) (interface{}, error) {
	ctx := v8.NewContext()
	values := make([]string, len(data))
	for index, d := range data {
		values[index] = fmt.Sprintf("%d", d)
	}
	_, err := ctx.RunScript(parserFunc, "parser.js")
	if err != nil {
		return nil, err
	}
	_, err = ctx.RunScript(fmt.Sprintf("const result = parse([%s])", strings.Join(values, ",")), "main.js")
	if err != nil {
		return nil, err
	}
	val, err := ctx.RunScript("result", "")
	if err != nil {
		return nil, err
	}
	obj, err := val.AsObject()
	if err != nil {
		return nil, err
	}
	return obj, nil

}
