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
			fmt.Printf("unexpected class name for arg0: %q\n", className)
		}
		call.Arguments[0].ToObject(b.runtime)
	default:
		fmt.Printf("unexpected type for arg0: %q\n", t)
	}

	return b.runtime.ToValue(1)
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	b := &Buffer{
		runtime: runtime,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("from", b.From)
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("Buffer", require.Require(runtime, "Buffer"))
}

func init() {
	require.RegisterNativeModule("Buffer", Require)
}
