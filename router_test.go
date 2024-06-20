package symple

import (
	"testing"
)

func TestPrefixPath(t *testing.T) {
	values := []struct {
		name          string
		path          string
		prefix        string
		expectedValue string
	}{
		{name: "empty prefix with method", path: "GET /hello", prefix: "", expectedValue: "GET /hello"},
		{name: "empty prefix without method", path: "/hello", prefix: "", expectedValue: "/hello"},
		{name: "with prefix with method", path: "GET /hello", prefix: "/prefix", expectedValue: "GET /prefix/hello"},
		{name: "with prefix without method", path: "/hello", prefix: "/prefix", expectedValue: "/prefix/hello"},
	}

	for _, val := range values {
		t.Run(val.name, func(t *testing.T) {
			res := applyPrefix(val.path, val.prefix)
			if res != val.expectedValue {
				t.Fatalf("Invalid path values: %s, expected %s", res, val.expectedValue)
			}
		})

	}
}
