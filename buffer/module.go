package buffer

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

type Buffer struct {
	runtime *goja.Runtime
}

func NewBufferObject(runtime *goja.Runtime, b []byte) *goja.Object {
	object := runtime.NewObject()
	object.Set("type", "Buffer")
	object.Set("data", b)
	object.DefineDataProperty("length", runtime.ToValue(len(b)), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	return object
}

func (b *Buffer) Alloc(call goja.FunctionCall) goja.Value {
	length := call.Arguments[0].ToInteger()
	data := make([]byte, length)
	if len(call.Arguments) == 1 {
		return NewBufferObject(b.runtime, data)
	}

	fill := call.Arguments[1].String()
	if fillLen := len(fill); fillLen > 0 {
		for i := range data {
			data[i] = fill[i%fillLen]
		}
	}

	return NewBufferObject(b.runtime, data)
}

func (b *Buffer) From(call goja.FunctionCall) goja.Value {
	arg0 := call.Arguments[0]
	switch t := arg0.ExportType().String(); t {
	case "string":
		return NewBufferObject(b.runtime, []byte(arg0.String()))
	case "[]interface {}": // object type
		obj := arg0.ToObject(b.runtime)
		fmt.Println()
		switch className := obj.ClassName(); className {
		case "Array":
			data := make([]byte, len(obj.Keys()))
			for i, key := range obj.Keys() {
				data[i] = byte(obj.Get(key).ToInteger())
			}
			return NewBufferObject(b.runtime, data)
		default:
			return NewBufferObject(b.runtime, []byte{})
		}
	default:
		return NewBufferObject(b.runtime, []byte{})
	}
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	b := &Buffer{
		runtime: runtime,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("alloc", b.Alloc)
	obj.Set("from", b.From)
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("Buffer", require.Require(runtime, "Buffer"))
}

func init() {
	require.RegisterNativeModule("Buffer", Require)
}
