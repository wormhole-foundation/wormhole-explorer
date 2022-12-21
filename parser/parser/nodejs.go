package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

type NodeJS struct{}

func (n *NodeJS) Parse(parserFunc string, data []byte) (string, error) {
	ctx := goja.New()

	values := make([]string, len(data))
	for index, d := range data {
		values[index] = fmt.Sprintf("%d", d)
	}
	_, err := ctx.RunString(parserFunc)
	if err != nil {
		return "", err
	}
	val, err := ctx.RunString(fmt.Sprintf("parse([%s])", strings.Join(values, ",")))
	if err != nil {
		return "", err
	}
	if val == goja.Undefined() {
		return "", errors.New("function doesn't return an object")
	}
	result, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(result), nil

}
