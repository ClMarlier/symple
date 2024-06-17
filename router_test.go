package symple

import (
	"testing"
)

func TestPrefixPath(t *testing.T) {
	values := []map[string]string{}
	v1 := map[string]string{"input1": "", "input2": "GET /hello", "expected": "GET /hello"}
	v2 := map[string]string{"input1": "", "input2": "/hello", "expected": "/hello"}
	v3 := map[string]string{"input1": "/prefix", "input2": "GET /hello", "expected": "GET /prefix/hello"}
	v4 := map[string]string{"input1": "/prefix", "input2": "/hello", "expected": "/prefix/hello"}
	values = append(values, v1, v2, v3, v4)

	for _, v := range values {
		res := applyPrefix(v["input2"], v["input1"])
		if res != v["expected"] {
			t.Fatalf("Invalid path values: %s, expected %s", res, v["expected"])
		}
	}
}
