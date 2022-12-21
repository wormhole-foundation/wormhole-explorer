package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeJS_Parse(t *testing.T) {
	parser := &NodeJS{}

	t.Run("simple object", func(t *testing.T) {
		fn := `function parse(data) {
			return { a: data[0], b: data[1] };
		}`
		data := []byte{4, 5}
		value, err := parser.Parse(fn, data)
		assert.Nil(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, `{"a":4,"b":5}`, value)
	})

	t.Run("parse json", func(t *testing.T) {
		fn := `function parse(data) {
			const json = String.fromCharCode(...data);
			const obj = JSON.parse(json);
			const first = obj.first + 10;
			const second = obj.second + 5;
			return { first: first, second: second, type: "Test" };
		}`
		data := `{"first":10,"second":20}`
		value, err := parser.Parse(fn, []byte(data))
		assert.Nil(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, `{"first":20,"second":25,"type":"Test"}`, value)
	})
}

func TestNodeJS_Failed(t *testing.T) {
	parser := &NodeJS{}

	t.Run("parser doesn't exist", func(t *testing.T) {
		fn := `function execute(data) {
			return { a: data[0], b: data[1] };
		}`
		data := []byte{1, 2}
		_, err := parser.Parse(fn, data)
		assert.NotNil(t, err)
	})

	t.Run("syntaxis error", func(t *testing.T) {
		fn := `function parse(data) {
			return { a: data[0], b: data[1];
		}`
		data := []byte{1, 2}
		_, err := parser.Parse(fn, data)
		assert.NotNil(t, err)
	})

	t.Run("parser doesn't return an object", func(t *testing.T) {
		fn := `function parse(data) {}`
		data := []byte{1, 2}
		_, err := parser.Parse(fn, data)
		assert.NotNil(t, err)
	})

}
