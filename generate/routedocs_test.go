package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/parser"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

// goctl's analyzer drops comments written between @handler and the route
// method, which is where @respdoc- annotations usually go; they must be
// recovered from the .api source, including imported files.
func TestRespDocBetweenHandlerAndMethod(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) string {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	write("sub/types.api", `syntax = "v1"

type ErrorResponse {
	Code int `+"`json:\"code\"`"+`
}

service demo-api {
	@handler ping
	/*
	@respdoc-400 (
		100101: out of authority
		100102: user not exist
	) // Error code list
	*/
	/* @respdoc-500 (ErrorResponse) // Server Error */
	get /ping
}
`)
	mainFile := write("main.api", `syntax = "v1"

import "sub/types.api"

service demo-api {
	@handler pong
	// @respdoc-418 (no closing paren // malformed, must not panic
	get /pong
}
`)

	api, err := parser.Parse(mainFile, "")
	if err != nil {
		t.Fatal(err)
	}
	result := runGenerate(t, &plugin.Plugin{Api: api, ApiFilePath: mainFile}, "", "", "")

	paths := result["paths"].(map[string]interface{})
	resps := paths["/ping"].(map[string]interface{})["get"].(map[string]interface{})["responses"].(map[string]interface{})

	resp400, ok := resps["400"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 400 response from @respdoc-400, got %v", resps)
	}
	if resp400["description"] != "Error code list" {
		t.Errorf("expected description 'Error code list', got %v", resp400["description"])
	}
	example := resp400["content"].(map[string]interface{})["application/json"].(map[string]interface{})["example"].(map[string]interface{})
	if example["100101"] != "out of authority" || example["100102"] != "user not exist" {
		t.Errorf("expected error code example, got %v", example)
	}

	resp500, ok := resps["500"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 500 response from @respdoc-500, got %v", resps)
	}
	schema := resp500["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
	if schema["$ref"] != "#/components/schemas/ErrorResponse" {
		t.Errorf("expected ErrorResponse ref, got %v", schema)
	}

	pong := paths["/pong"].(map[string]interface{})["get"].(map[string]interface{})["responses"].(map[string]interface{})
	if _, ok := pong["418"]; ok {
		t.Errorf("malformed @respdoc must be skipped, got %v", pong)
	}
}

// Without an .api file to read, the docs goctl does provide are still used.
func TestRespDocFromHandlerDoc(t *testing.T) {
	route := spec.Route{
		Method:     "get",
		Path:       "/doc",
		Handler:    "doc",
		HandlerDoc: []string{"// @respdoc-404 (ErrorResponse) // Not here"},
	}
	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types:   []spec.Type{defineStruct("ErrorResponse", member("Code", primitiveType("int"), `json:"code"`, ""))},
			Service: spec.Service{Name: "test-api", Groups: []spec.Group{{Routes: []spec.Route{route}}}},
		},
	}

	result := runGenerate(t, p, "", "", "")
	resps := result["paths"].(map[string]interface{})["/doc"].(map[string]interface{})["get"].(map[string]interface{})["responses"].(map[string]interface{})
	if resp, ok := resps["404"].(map[string]interface{}); !ok || resp["description"] != "Not here" {
		t.Errorf("expected 404 response from HandlerDoc, got %v", resps)
	}
}
