package buffer

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

func TestBuffer(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	Enable(vm)

	if c := vm.Get("Buffer"); c == nil {
		t.Fatal("buffer not found")
	}

	tests := map[string]struct {
		Script         string
		ExpectedOutput string
	}{
		"from string": {
			Script:         `JSON.stringify(Buffer.from("test string"))`,
			ExpectedOutput: `{"type":"Buffer","data":[116,101,115,116,32,115,116,114,105,110,103]}`,
		},
		"from int array": {
			Script:         `JSON.stringify(Buffer.from([1, 2, 3]))`,
			ExpectedOutput: `{"type":"Buffer","data":[1,2,3]}`,
		},
		"from string get length": {
			Script:         `Buffer.from("test string").length`,
			ExpectedOutput: "11",
		},
		"from int array get length": {
			Script:         "Buffer.from([1, 2, 3]).length",
			ExpectedOutput: "3",
		},
	}

	for testName, test := range tests {
		test := test
		t.Run(testName, func(t *testing.T) {
			val, err := vm.RunString(test.Script)
			if err != nil {
				t.Fatalf("unexpected script error: %v", err)
			}

			if val.String() != test.ExpectedOutput {
				t.Errorf("got = %q, want = %q", val.String(), test.ExpectedOutput)
			}
		})
	}
}
