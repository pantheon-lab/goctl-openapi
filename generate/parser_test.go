package generate_test

import (
	"os"
	"testing"

	"github.com/pantheon-lab/goctl-openapi/generate"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
	"gopkg.in/yaml.v3"
)

func runGenerate(t *testing.T, p *plugin.Plugin, host, basePath, schemes string) map[string]interface{} {
	t.Helper()
	dir := t.TempDir()
	p.Dir = dir
	err := generate.Do("out.yaml", host, basePath, schemes, p)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dir + "/out.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func strPointer(s string) *string { return &s }

// primitiveType returns a spec.PrimitiveType with the given name.
func primitiveType(name string) spec.PrimitiveType {
	return spec.PrimitiveType{RawName: name}
}

// defineStruct creates a spec.DefineStruct with the given name and members.
func defineStruct(name string, members ...spec.Member) spec.DefineStruct {
	return spec.DefineStruct{
		RawName: name,
		Members: members,
	}
}

// member creates a spec.Member with the given name, type, tag, and comment.
func member(name string, typ spec.Type, tag string, comment string) spec.Member {
	return spec.Member{
		Name:    name,
		Type:    typ,
		Tag:     tag,
		Comment: comment,
	}
}

func TestInfo(t *testing.T) {
	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Info: spec.Info{
				Properties: map[string]string{
					"title":   `"My API"`,
					"version": `"1.0.0"`,
					"desc":    `"My API description"`,
				},
			},
			Service: spec.Service{
				Name: "test-api",
			},
		},
	}
	result := runGenerate(t, p, "", "", "")
	info := result["info"].(map[string]interface{})
	if info["title"] != "My API" {
		t.Errorf("expected title 'My API', got %v", info["title"])
	}
	if info["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", info["version"])
	}
	if info["description"] != "My API description" {
		t.Errorf("expected description 'My API description', got %v", info["description"])
	}
}

func TestOpenAPIVersion(t *testing.T) {
	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Service: spec.Service{Name: "test"},
		},
	}
	result := runGenerate(t, p, "", "", "")
	version := result["openapi"]
	if version != "3.1.0" {
		t.Errorf("expected openapi '3.1.0', got %v", version)
	}
}

func TestServerURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		basePath string
		schemes  string
		wantURL  string
	}{
		{"defaults", "", "", "", "http://localhost"},
		{"with host", "api.example.com", "", "", "http://api.example.com"},
		{"with basepath", "", "/v2", "", "http://localhost/v2"},
		{"with host and basepath", "api.example.com", "/v2", "", "http://api.example.com/v2"},
		{"with https", "api.example.com", "/api", "https", "https://api.example.com/api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &plugin.Plugin{
				Api: &spec.ApiSpec{
					Service: spec.Service{Name: "test"},
				},
			}
			result := runGenerate(t, p, tt.host, tt.basePath, tt.schemes)
			servers := result["servers"].([]interface{})
			server := servers[0].(map[string]interface{})
			url := server["url"].(string)
			if url != tt.wantURL {
				t.Errorf("expected server URL %q, got %q", tt.wantURL, url)
			}
		})
	}
}

func TestGetEndpointWithQueryParams(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/user/search",
		Handler: "searchUser",
		RequestType: defineStruct("UserSearchReq",
			member("KeyWord", primitiveType("string"), `form:"keyWord"`, " 关键词"),
		),
		ResponseType: defineStruct("UserInfoReply",
			member("Name", primitiveType("string"), `json:"name"`, ""),
			member("Age", primitiveType("int"), `json:"age"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("UserSearchReq",
					member("KeyWord", primitiveType("string"), `form:"keyWord"`, " 关键词"),
				),
				defineStruct("UserInfoReply",
					member("Name", primitiveType("string"), `json:"name"`, ""),
					member("Age", primitiveType("int"), `json:"age"`, ""),
				),
			},
			Service: spec.Service{
				Name: "user-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	pathItem := paths["/user/search"].(map[string]interface{})
	getOp := pathItem["get"].(map[string]interface{})

	if getOp["operationId"] != "searchUser" {
		t.Errorf("expected operationId 'searchUser', got %v", getOp["operationId"])
	}

	params := getOp["parameters"].([]interface{})
	if len(params) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(params))
	}
	param := params[0].(map[string]interface{})
	if param["name"] != "keyWord" {
		t.Errorf("expected parameter name 'keyWord', got %v", param["name"])
	}
	if param["in"] != "query" {
		t.Errorf("expected 'in'='query', got %v", param["in"])
	}
	if param["required"] != true {
		t.Errorf("expected required=true, got %v", param["required"])
	}
	schema := param["schema"].(map[string]interface{})
	if schema["type"] != "string" {
		t.Errorf("expected schema type 'string', got %v", schema["type"])
	}

	// Verify response
	resp := getOp["responses"].(map[string]interface{})
	resp200 := resp["200"].(map[string]interface{})
	content := resp200["content"].(map[string]interface{})
	jsonMedia := content["application/json"].(map[string]interface{})
	respSchema := jsonMedia["schema"].(map[string]interface{})
	if respSchema["$ref"] != "#/components/schemas/UserInfoReply" {
		t.Errorf("expected $ref to UserInfoReply, got %v", respSchema["$ref"])
	}

	// Verify schema in components
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	userSchema := schemas["UserInfoReply"].(map[string]interface{})
	props := userSchema["properties"].(map[string]interface{})
	if props["name"] == nil {
		t.Errorf("expected 'name' property in UserInfoReply")
	}
	if props["age"] == nil {
		t.Errorf("expected 'age' property in UserInfoReply")
	}
}

func TestPostEndpointWithRequestBody(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/user/register",
		Handler: "register",
		RequestType: defineStruct("RegisterReq",
			member("Username", primitiveType("string"), `json:"username"`, ""),
			member("Password", primitiveType("string"), `json:"password"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("RegisterReq",
					member("Username", primitiveType("string"), `json:"username"`, ""),
					member("Password", primitiveType("string"), `json:"password"`, ""),
				),
			},
			Service: spec.Service{
				Name: "user-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	pathItem := paths["/user/register"].(map[string]interface{})
	postOp := pathItem["post"].(map[string]interface{})

	// Should have requestBody, not parameters with in:body
	if _, hasParams := postOp["parameters"]; hasParams {
		t.Errorf("POST should not have parameters with body, use requestBody instead")
	}

	reqBody := postOp["requestBody"].(map[string]interface{})
	if reqBody["required"] != true {
		t.Errorf("expected requestBody.required=true")
	}
	content := reqBody["content"].(map[string]interface{})
	jsonMedia := content["application/json"].(map[string]interface{})
	schema := jsonMedia["schema"].(map[string]interface{})
	if schema["$ref"] != "#/components/schemas/RegisterReq" {
		t.Errorf("expected $ref to RegisterReq, got %v", schema["$ref"])
	}
}

func TestPathParameters(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/user/:id",
		Handler: "getUser",
		RequestType: defineStruct("UserReq",
			member("Id", primitiveType("string"), `path:"id"`, ""),
		),
		ResponseType: defineStruct("UserReply",
			member("Name", primitiveType("string"), `json:"name"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("UserReq", member("Id", primitiveType("string"), `path:"id"`, "")),
				defineStruct("UserReply", member("Name", primitiveType("string"), `json:"name"`, "")),
			},
			Service: spec.Service{
				Name: "user-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	// Check path uses {id} not :id
	if paths["/user/{id}"] == nil {
		t.Errorf("expected path /user/{id}, got keys: %v", keys(paths))
	}

	getOp := paths["/user/{id}"].(map[string]interface{})["get"].(map[string]interface{})
	params := getOp["parameters"].([]interface{})
	if len(params) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(params))
	}
	param := params[0].(map[string]interface{})
	if param["name"] != "id" {
		t.Errorf("expected param name 'id', got %v", param["name"])
	}
	if param["in"] != "path" {
		t.Errorf("expected in='path', got %v", param["in"])
	}
	if param["required"] != true {
		t.Errorf("expected required=true")
	}
}

func keys(m map[string]interface{}) []string {
	var k []string
	for kk := range m {
		k = append(k, kk)
	}
	return k
}

func TestArrayResponse(t *testing.T) {
	// Response type name like "[]UserInfoReply"
	route := spec.Route{
		Method:  "get",
		Path:    "/users",
		Handler: "listUsers",
		ResponseType: spec.ArrayType{
			RawName: "[]UserInfoReply",
			Value:   defineStruct("UserInfoReply", member("Name", primitiveType("string"), `json:"name"`, "")),
		},
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("UserInfoReply", member("Name", primitiveType("string"), `json:"name"`, "")),
			},
			Service: spec.Service{
				Name: "user-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/users"].(map[string]interface{})["get"].(map[string]interface{})
	resp200 := getOp["responses"].(map[string]interface{})["200"].(map[string]interface{})
	content := resp200["content"].(map[string]interface{})
	jsonMedia := content["application/json"].(map[string]interface{})
	schema := jsonMedia["schema"].(map[string]interface{})

	if schema["type"] != "array" {
		t.Errorf("expected response type 'array', got %v", schema["type"])
	}
	items := schema["items"].(map[string]interface{})
	if items["$ref"] != "#/components/schemas/UserInfoReply" {
		t.Errorf("expected items $ref to UserInfoReply, got %v", items["$ref"])
	}
}

func TestHeaderParameters(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/user/login",
		Handler: "login",
		RequestType: defineStruct("LoginReq",
			member("Username", primitiveType("string"), `json:"username"`, ""),
			member("AppId", primitiveType("string"), `header:"appId"`, "APP-ID header"),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("LoginReq",
					member("Username", primitiveType("string"), `json:"username"`, ""),
					member("AppId", primitiveType("string"), `header:"appId"`, "APP-ID header"),
				),
			},
			Service: spec.Service{
				Name: "user-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	postOp := paths["/user/login"].(map[string]interface{})["post"].(map[string]interface{})

	params := postOp["parameters"].([]interface{})
	found := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "appId" {
			found = true
			if param["in"] != "header" {
				t.Errorf("expected in='header', got %v", param["in"])
			}
			break
		}
	}
	if !found {
		t.Errorf("expected header parameter 'appId' not found")
	}
}

func TestJWTSecurity(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/secure",
		Handler: "secureEndpoint",
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Service: spec.Service{
				Name: "secure-api",
				Groups: []spec.Group{
					{
						Annotation: spec.Annotation{
							Properties: map[string]string{
								"jwt": "Auth",
							},
						},
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")

	// Check components.securitySchemes
	comp := result["components"].(map[string]interface{})
	secSchemes := comp["securitySchemes"].(map[string]interface{})

	// Check bearerAuth scheme
	bearerAuth := secSchemes["bearerAuth"].(map[string]interface{})
	if bearerAuth["type"] != "http" {
		t.Errorf("expected bearerAuth type 'http', got %v", bearerAuth["type"])
	}
	if bearerAuth["scheme"] != "bearer" {
		t.Errorf("expected bearerAuth scheme 'bearer', got %v", bearerAuth["scheme"])
	}

	// Check apiKey scheme
	apiKey := secSchemes["apiKey"].(map[string]interface{})
	if apiKey["type"] != "apiKey" {
		t.Errorf("expected apiKey type 'apiKey', got %v", apiKey["type"])
	}
	if apiKey["in"] != "header" {
		t.Errorf("expected apiKey in 'header', got %v", apiKey["in"])
	}

	// Check operation has both security requirements
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/secure"].(map[string]interface{})["get"].(map[string]interface{})
	sec := getOp["security"].([]interface{})
	if len(sec) != 2 {
		t.Fatalf("expected 2 security requirements, got %d", len(sec))
	}
	req0 := sec[0].(map[string]interface{})
	req1 := sec[1].(map[string]interface{})
	if req0["bearerAuth"] == nil {
		t.Errorf("expected bearerAuth security requirement in first entry")
	}
	if req1["apiKey"] == nil {
		t.Errorf("expected apiKey security requirement in second entry")
	}
}

func TestRoutTags(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/ping",
		Handler: "ping",
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Service: spec.Service{
				Name: "default-service",
				Groups: []spec.Group{
					{
						Annotation: spec.Annotation{
							Properties: map[string]string{
								"swtags": "custom-tag",
							},
						},
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/ping"].(map[string]interface{})["get"].(map[string]interface{})
	tags := getOp["tags"].([]interface{})
	if len(tags) != 1 || tags[0] != "custom-tag" {
		t.Errorf("expected tag 'custom-tag', got %v", tags)
	}
}

func TestRespDoc(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/doc-test",
		Handler: "docTest",
		Doc: []string{
			"@respdoc-400 (100101: out of authority 100102: user not exist) // Error code list",
		},
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/doc-test"].(map[string]interface{})["get"].(map[string]interface{})
	resps := getOp["responses"].(map[string]interface{})

	resp400 := resps["400"].(map[string]interface{})
	if desc, ok := resp400["description"]; !ok || desc.(string) != "Error code list" {
		t.Errorf("expected description 'Error code list', got %v", desc)
	}
	content := resp400["content"].(map[string]interface{})
	jsonMedia := content["application/json"].(map[string]interface{})
	schema := jsonMedia["schema"].(map[string]interface{})
	example, ok := schema["example"]
	if !ok || example == "" {
		t.Errorf("expected example in 400 response")
	}
}

func TestNestedArrays(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/test",
		Handler: "testNested",
		ResponseType: defineStruct("NestedReply",
			member("Tags", spec.ArrayType{RawName: "[][]string"}, `json:"tags"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("NestedReply",
					member("Tags", spec.ArrayType{RawName: "[][]string"}, `json:"tags"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	nested := schemas["NestedReply"].(map[string]interface{})
	props := nested["properties"].(map[string]interface{})
	tags := props["tags"].(map[string]interface{})

	if tags["type"] != "array" {
		t.Errorf("expected outer type 'array', got %v", tags["type"])
	}
	inner := tags["items"].(map[string]interface{})
	if inner["type"] != "array" {
		t.Errorf("expected inner type 'array', got %v", inner["type"])
	}
	innermost := inner["items"].(map[string]interface{})
	if innermost["type"] != "string" {
		t.Errorf("expected innermost type 'string', got %v", innermost["type"])
	}
}

func TestEnumDefaultExampleOptions(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/options-test",
		Handler: "optionsTest",
		RequestType: defineStruct("OptionsReq",
			member("Status", primitiveType("string"), `json:"status"`, ""),
		),
	}

	// We test schema options via the definition rendering
	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("OptionsReq",
					member("Status", primitiveType("string"),
						`json:"status,options=active|inactive|pending"`,
						""),
					member("Count", primitiveType("int"),
						`json:"count,default=10"`,
						""),
					member("Label", primitiveType("string"),
						`json:"label,example=test-label"`,
						""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["OptionsReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})

	// Check enum
	status := props["status"].(map[string]interface{})
	enum := status["enum"].([]interface{})
	if len(enum) != 3 || enum[0] != "active" || enum[1] != "inactive" || enum[2] != "pending" {
		t.Errorf("expected enum [active inactive pending], got %v", enum)
	}

	// Check default (should be typed)
	count := props["count"].(map[string]interface{})
	if count["default"] != int(10) {
		t.Errorf("expected default 10, got %v (type: %T)", count["default"], count["default"])
	}

	// Check example
	label := props["label"].(map[string]interface{})
	if label["example"] != "test-label" {
		t.Errorf("expected example 'test-label', got %v", label["example"])
	}
}

func TestRequiredFields(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/required-test",
		Handler: "requiredTest",
		RequestType: defineStruct("RequiredReq",
			member("Name", primitiveType("string"), `json:"name"`, ""),
			member("Nickname", primitiveType("string"), `json:"nickname,optional"`, ""),
			member("Email", primitiveType("string"), `json:"email,omitempty"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("RequiredReq",
					member("Name", primitiveType("string"), `json:"name"`, ""),
					member("Nickname", primitiveType("string"), `json:"nickname,optional"`, ""),
					member("Email", primitiveType("string"), `json:"email,omitempty"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["RequiredReq"].(map[string]interface{})
	required := req["required"].([]interface{})
	requiredSet := make(map[string]bool)
	for _, r := range required {
		requiredSet[r.(string)] = true
	}
	if !requiredSet["name"] {
		t.Errorf("expected 'name' to be required")
	}
	if requiredSet["nickname"] {
		t.Errorf("expected 'nickname' to NOT be required")
	}
	if requiredSet["email"] {
		t.Errorf("expected 'email' to NOT be required")
	}
}

func TestFormData(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/upload",
		Handler: "upload",
		RequestType: defineStruct("UploadReq",
			member("File", primitiveType("string"), `form:"file"`, ""),
			member("Name", primitiveType("string"), `form:"name"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("UploadReq",
					member("File", primitiveType("string"), `form:"file"`, ""),
					member("Name", primitiveType("string"), `form:"name"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	postOp := paths["/upload"].(map[string]interface{})["post"].(map[string]interface{})

	reqBody := postOp["requestBody"].(map[string]interface{})
	content := reqBody["content"].(map[string]interface{})

	if _, hasForm := content["multipart/form-data"]; !hasForm {
		t.Errorf("expected multipart/form-data content for form fields")
	}
	if _, hasJSON := content["application/json"]; hasJSON {
		t.Errorf("a multipart upload must not advertise application/json")
	}
	if params, ok := postOp["parameters"]; ok {
		t.Errorf("expected no query parameters for a multipart body, got %v", params)
	}
}

func TestFormDataWithoutFileNotMultipart(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/submit",
		Handler: "submit",
		RequestType: defineStruct("SubmitReq",
			member("Name", primitiveType("string"), `form:"name"`, ""),
			member("Comment", primitiveType("string"), `form:"comment"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("SubmitReq",
					member("Name", primitiveType("string"), `form:"name"`, ""),
					member("Comment", primitiveType("string"), `form:"comment"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	postOp := paths["/submit"].(map[string]interface{})["post"].(map[string]interface{})

	if rb, ok := postOp["requestBody"]; ok {
		t.Errorf("expected no requestBody when every field is a form/query param, got %v", rb)
	}
	params, _ := postOp["parameters"].([]interface{})
	if len(params) != 2 {
		t.Fatalf("expected name and comment as query params, got %v", params)
	}
	for _, p := range params {
		if p.(map[string]interface{})["in"] != "query" {
			t.Errorf("expected query param, got %v", p)
		}
	}
}

// TestFormFieldsExcludedFromJSONBody mirrors chatbot-ms's UpdateStationIndoorMapReq:
// form fields are query params on a POST and must not be duplicated into the
// JSON body alongside the real json field.
func TestFormFieldsExcludedFromJSONBody(t *testing.T) {
	reqType := defineStruct("UpdateMapReq",
		member("Location", primitiveType("string"), `form:"location"`, ""),
		member("PaidArea", primitiveType("bool"), `form:"paid_area"`, ""),
		member("MapData", primitiveType("map[string]interface{}"), `json:"mapData"`, ""),
	)
	route := spec.Route{Method: "post", Path: "/indoor-map", Handler: "updateMap", RequestType: reqType}
	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types:   []spec.Type{reqType},
			Service: spec.Service{Name: "test-api", Groups: []spec.Group{{Routes: []spec.Route{route}}}},
		},
	}

	result := runGenerate(t, p, "", "", "")
	schemas := result["components"].(map[string]interface{})["schemas"].(map[string]interface{})
	req := schemas["UpdateMapReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})
	if len(props) != 1 || props["mapData"] == nil {
		t.Errorf("expected body schema to contain only mapData, got %v", props)
	}
	for _, r := range req["required"].([]interface{}) {
		if r != "mapData" {
			t.Errorf("expected only mapData in required, got %v", req["required"])
		}
	}

	postOp := result["paths"].(map[string]interface{})["/indoor-map"].(map[string]interface{})["post"].(map[string]interface{})
	if params, _ := postOp["parameters"].([]interface{}); len(params) != 2 {
		t.Errorf("expected location and paid_area as query params, got %v", params)
	}
}

func TestJSONIgnoredFieldExcluded(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/ignored-test",
		Handler: "ignoredTest",
		RequestType: defineStruct("IgnoredReq",
			member("Name", primitiveType("string"), `json:"name"`, ""),
			member("Secret", primitiveType("string"), `json:"-"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("IgnoredReq",
					member("Name", primitiveType("string"), `json:"name"`, ""),
					member("Secret", primitiveType("string"), `json:"-"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["IgnoredReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})

	if _, ok := props["name"]; !ok {
		t.Errorf("expected 'name' property to be present")
	}
	if _, ok := props["secret"]; ok {
		t.Errorf("expected 'secret' property to be excluded")
	}
	if _, ok := props["Secret"]; ok {
		t.Errorf("expected 'Secret' property to be excluded")
	}

	if required, ok := req["required"]; ok {
		for _, r := range required.([]interface{}) {
			if r == "-" {
				t.Errorf("expected literal '-' to never appear in required, got %v", required)
			}
			if r == "secret" || r == "Secret" {
				t.Errorf("expected ignored field to not appear in required, got %v", required)
			}
		}
	}
}

func TestJSONIgnoredFieldExcludedInNestedStruct(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/ignored-nested-test",
		Handler: "ignoredNestedTest",
		RequestType: defineStruct("IgnoredNestedReq",
			member("", defineStruct("",
				member("Name", primitiveType("string"), `json:"name"`, ""),
				member("Secret", primitiveType("string"), `json:"-"`, ""),
			), `json:",inline"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("IgnoredNestedReq",
					member("", defineStruct("",
						member("Name", primitiveType("string"), `json:"name"`, ""),
						member("Secret", primitiveType("string"), `json:"-"`, ""),
					), `json:",inline"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["IgnoredNestedReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})

	if _, ok := props["name"]; !ok {
		t.Errorf("expected 'name' property to be present")
	}
	if _, ok := props["secret"]; ok {
		t.Errorf("expected 'secret' property to be excluded")
	}
	if _, ok := props["Secret"]; ok {
		t.Errorf("expected 'Secret' property to be excluded")
	}
}

func TestAdditionalPropertiesForFreeFormMap(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/free-form",
		Handler: "freeForm",
		RequestType: defineStruct("FreeFormReq",
			member("Raw", primitiveType("interface{}"), `json:"raw"`, ""),
			member("Params", primitiveType("map[string]interface{}"), `json:"params"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("FreeFormReq",
					member("Raw", primitiveType("interface{}"), `json:"raw"`, ""),
					member("Params", primitiveType("map[string]interface{}"), `json:"params"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["FreeFormReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})

	for _, name := range []string{"raw", "params"} {
		field := props[name].(map[string]interface{})
		if field["type"] != "object" {
			t.Errorf("expected %s to be type object, got %v", name, field["type"])
		}
		if _, has := field["additionalProperties"]; has {
			t.Errorf("expected %s to omit additionalProperties (free-form object), got %v", name, field["additionalProperties"])
		}
	}
}

func TestTypedMapIsPlainObject(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/typed-map",
		Handler: "typedMap",
		RequestType: defineStruct("TypedMapReq",
			member("Labels", primitiveType("map[string]string"), `json:"labels"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("TypedMapReq",
					member("Labels", primitiveType("map[string]string"), `json:"labels"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["TypedMapReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})

	labels := props["labels"].(map[string]interface{})
	if labels["type"] != "object" {
		t.Errorf("expected type object, got %v", labels["type"])
	}
	if _, has := labels["additionalProperties"]; has {
		t.Errorf("expected plain object without additionalProperties, got %v", labels)
	}
}

func TestMapOfStructIsPlainObject(t *testing.T) {
	req := defineStruct("LocaleReq",
		member("Locales", primitiveType("map[string]CatalogLocale"), `json:"locales"`, ""),
		member("Extra", primitiveType("map[string]interface{}"), `json:"extra"`, ""),
	)
	route := spec.Route{Method: "post", Path: "/locales", Handler: "locales", RequestType: req}
	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types:   []spec.Type{req},
			Service: spec.Service{Name: "test-api", Groups: []spec.Group{{Routes: []spec.Route{route}}}},
		},
	}

	result := runGenerate(t, p, "", "", "")
	props := result["components"].(map[string]interface{})["schemas"].(map[string]interface{})["LocaleReq"].(map[string]interface{})["properties"].(map[string]interface{})
	for _, name := range []string{"locales", "extra"} {
		prop := props[name].(map[string]interface{})
		if len(prop) != 1 || prop["type"] != "object" {
			t.Errorf("%s: expected exactly {type: object}, got %v", name, prop)
		}
	}
}

func TestFormFieldsAsQueryParamsOnNonGetRoutes(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/mixed",
		Handler: "mixed",
		RequestType: defineStruct("MixedReq",
			member("Query", primitiveType("string"), `form:"query"`, ""),
			member("Name", primitiveType("string"), `json:"name"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("MixedReq",
					member("Query", primitiveType("string"), `form:"query"`, ""),
					member("Name", primitiveType("string"), `json:"name"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	postOp := paths["/mixed"].(map[string]interface{})["post"].(map[string]interface{})

	params, ok := postOp["parameters"].([]interface{})
	if !ok || len(params) != 1 {
		t.Fatalf("expected 1 query parameter for form field, got %v", postOp["parameters"])
	}
	param := params[0].(map[string]interface{})
	if param["name"] != "query" {
		t.Errorf("expected parameter name 'query', got %v", param["name"])
	}
	if param["in"] != "query" {
		t.Errorf("expected 'in'='query', got %v", param["in"])
	}

	reqBody := postOp["requestBody"].(map[string]interface{})
	content := reqBody["content"].(map[string]interface{})
	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Errorf("expected application/json content for json-tagged field")
	}

	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["MixedReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})
	if _, ok := props["name"]; !ok {
		t.Errorf("expected 'name' property to be present in body schema")
	}
}

func TestGroupPrefix(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/users",
		Handler: "listUsers",
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Service: spec.Service{
				Name: "user-api",
				Groups: []spec.Group{
					{
						Annotation: spec.Annotation{
							Properties: map[string]string{
								"prefix": "/api/v1",
							},
						},
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	if paths["/api/v1/users"] == nil {
		t.Errorf("expected path /api/v1/users, got keys: %v", keys(paths))
	}
}

func TestNoResponseType(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/no-response",
		Handler: "noResp",
		RequestType: defineStruct("SomeReq",
			member("Name", primitiveType("string"), `json:"name"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("SomeReq", member("Name", primitiveType("string"), `json:"name"`, "")),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	postOp := paths["/no-response"].(map[string]interface{})["post"].(map[string]interface{})

	resps := postOp["responses"].(map[string]interface{})
	resp200 := resps["200"].(map[string]interface{})
	if resp200["description"] != "A successful response." {
		t.Errorf("expected default response description")
	}
	// Should have content with schema:{} for empty response
	content, hasContent := resp200["content"].(map[string]interface{})
	if !hasContent {
		t.Fatalf("expected content for empty response")
	}
	jsonMedia, ok := content["application/json"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected application/json content")
	}
	if jsonMedia["schema"] == nil {
		t.Errorf("expected schema:{} for empty response")
	}
}

func TestMarshalYAMLWithRef(t *testing.T) {
	// This tests that $ref serialization is correct through JSON->YAML conversion
	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Service: spec.Service{Name: "test"},
		},
	}
	result := runGenerate(t, p, "", "", "")

	// Round-trip through YAML to verify $ref serialization is correct
	raw, err := yaml.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var rawMap map[string]interface{}
	if err := yaml.Unmarshal(raw, &rawMap); err != nil {
		t.Fatal(err)
	}
	// Verify the root level field is properly serialized
	if _, ok := rawMap["openapi"]; !ok {
		t.Errorf("expected openapi field at root")
	}
}

func TestInlineStruct(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/inline",
		Handler: "inlineTest",
		ResponseType: defineStruct("OuterReply",
			member("", defineStruct("Embedded",
				member("Field1", primitiveType("string"), `json:"field1"`, ""),
			), "", ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("OuterReply",
					member("", defineStruct("Embedded",
						member("Field1", primitiveType("string"), `json:"field1"`, ""),
					), "", ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	outer := schemas["OuterReply"].(map[string]interface{})
	props := outer["properties"].(map[string]interface{})
	if props["field1"] == nil {
		t.Errorf("expected inline struct fields to be flattened, got properties: %v", keys(props))
	}
}

func TestAllHTTPMethods(t *testing.T) {
	methods := []string{"get", "post", "put", "delete", "patch"}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			route := spec.Route{
				Method:  m,
				Path:    "/method-" + m,
				Handler: "handler" + m,
			}
			p := &plugin.Plugin{
				Api: &spec.ApiSpec{
					Service: spec.Service{
						Name: "test-api",
						Groups: []spec.Group{
							{
								Routes: []spec.Route{route},
							},
						},
					},
				},
			}
			result := runGenerate(t, p, "", "", "")
			paths := result["paths"].(map[string]interface{})
			pathItem := paths["/method-"+m].(map[string]interface{})
			if pathItem[m] == nil {
				t.Errorf("expected %s operation in path item", m)
			}
		})
	}
}

func TestSchemaOptionsInQueryParam(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/search",
		Handler: "search",
		RequestType: defineStruct("SearchReq",
			member("Status", primitiveType("string"), `form:"status,options=active|inactive"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("SearchReq",
					member("Status", primitiveType("string"), `form:"status,options=active|inactive"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/search"].(map[string]interface{})["get"].(map[string]interface{})
	params := getOp["parameters"].([]interface{})
	param := params[0].(map[string]interface{})
	schema := param["schema"].(map[string]interface{})
	enum := schema["enum"].([]interface{})
	if len(enum) != 2 || enum[0] != "active" || enum[1] != "inactive" {
		t.Errorf("expected enum [active inactive], got %v", enum)
	}
}

func TestBoolNoFormat(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/bool-test",
		Handler: "boolTest",
		RequestType: defineStruct("BoolReq",
			member("Active", primitiveType("bool"), `form:"active"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("BoolReq",
					member("Active", primitiveType("bool"), `form:"active"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	param := paths["/bool-test"].(map[string]interface{})["get"].(map[string]interface{})["parameters"].([]interface{})[0].(map[string]interface{})
	schema := param["schema"].(map[string]interface{})
	if schema["type"] != "boolean" {
		t.Errorf("expected type 'boolean', got %v", schema["type"])
	}
	if _, ok := schema["format"]; ok {
		t.Errorf("boolean should not have format, got %v", schema["format"])
	}
}

func TestGetEndpointNoRequestBody(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/search",
		Handler: "search",
		RequestType: defineStruct("SearchReq",
			member("Query", primitiveType("string"), `form:"query"`, ""),
			member("Page", primitiveType("int"), `form:"page,default=1"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("SearchReq",
					member("Query", primitiveType("string"), `form:"query"`, ""),
					member("Page", primitiveType("int"), `form:"page,default=1"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/search"].(map[string]interface{})["get"].(map[string]interface{})
	if _, has := getOp["requestBody"]; has {
		t.Errorf("GET endpoint should not have requestBody even with form fields")
	}
	// Verify query params are present
	params, ok := getOp["parameters"].([]interface{})
	if !ok || len(params) != 2 {
		t.Errorf("expected 2 query parameters, got %d", len(params))
	}
}

func TestPointerTypeNullable(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/nullable-test",
		Handler: "nullableTest",
		RequestType: defineStruct("NullableReq",
			member("Name", primitiveType("*string"), `json:"name"`, ""),
			member("Count", primitiveType("*int"), `json:"count,optional"`, ""),
			member("Active", primitiveType("*bool"), `json:"active"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("NullableReq",
					member("Name", primitiveType("*string"), `json:"name"`, ""),
					member("Count", primitiveType("*int"), `json:"count,optional"`, ""),
					member("Active", primitiveType("*bool"), `json:"active"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["NullableReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})

	for _, propName := range []string{"name", "count", "active"} {
		prop := props[propName].(map[string]interface{})
		if prop["nullable"] != true {
			t.Errorf("expected nullable=true for %s, got %v", propName, prop["nullable"])
		}
	}
}

func TestOptionalPathParam(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/item/:id",
		Handler: "getItem",
		RequestType: defineStruct("ItemReq",
			member("Id", primitiveType("string"), `path:"id,optional"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("ItemReq", member("Id", primitiveType("string"), `path:"id,optional"`, "")),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	param := paths["/item/{id}"].(map[string]interface{})["get"].(map[string]interface{})["parameters"].([]interface{})[0].(map[string]interface{})
	if param["required"] != false {
		t.Errorf("expected optional path param to have required=false, got %v", param["required"])
	}
}

func TestTypedDefaultValues(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/typed-defaults",
		Handler: "typedDefaults",
		RequestType: defineStruct("TypedDefaultsReq",
			member("Count", primitiveType("int"), `json:"count,default=42"`, ""),
			member("Ratio", primitiveType("float64"), `json:"ratio,default=3.14"`, ""),
			member("Active", primitiveType("bool"), `json:"active,default=true"`, ""),
			member("Name", primitiveType("string"), `json:"name,default=hello"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("TypedDefaultsReq",
					member("Count", primitiveType("int"), `json:"count,default=42"`, ""),
					member("Ratio", primitiveType("float64"), `json:"ratio,default=3.14"`, ""),
					member("Active", primitiveType("bool"), `json:"active,default=true"`, ""),
					member("Name", primitiveType("string"), `json:"name,default=hello"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["TypedDefaultsReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})

	count := props["count"].(map[string]interface{})
	if count["default"] != int(42) {
		t.Errorf("expected default 42, got %v (type: %T)", count["default"], count["default"])
	}
	ratio := props["ratio"].(map[string]interface{})
	if ratio["default"] != float64(3.14) {
		t.Errorf("expected default 3.14, got %v (type: %T)", ratio["default"], ratio["default"])
	}
	active := props["active"].(map[string]interface{})
	if active["default"] != true {
		t.Errorf("expected default true, got %v (type: %T)", active["default"], active["default"])
	}
	name := props["name"].(map[string]interface{})
	if name["default"] != "hello" {
		t.Errorf("expected default 'hello', got %v", name["default"])
	}
}

func TestStringOption(t *testing.T) {
	route := spec.Route{
		Method:  "post",
		Path:    "/string-option",
		Handler: "stringOptionHandler",
		RequestType: defineStruct("StringOptionReq",
			member("Count", primitiveType("int64"), `json:"count,string"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				defineStruct("StringOptionReq",
					member("Count", primitiveType("int64"), `json:"count,string"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	comp := result["components"].(map[string]interface{})
	schemas := comp["schemas"].(map[string]interface{})
	req := schemas["StringOptionReq"].(map[string]interface{})
	props := req["properties"].(map[string]interface{})
	count := props["count"].(map[string]interface{})
	if count["type"] != "string" {
		t.Errorf("expected type 'string' with json string option, got %v", count["type"])
	}
	if _, ok := count["format"]; ok {
		t.Errorf("string option field should not have format, got %v", count["format"])
	}
}

func TestDescriptionFromAtDoc(t *testing.T) {
	route := spec.Route{
		Method:  "get",
		Path:    "/doc",
		Handler: "docHandler",
		AtDoc: spec.AtDoc{
			Properties: map[string]string{
				"description": `"A test endpoint"`,
			},
		},
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/doc"].(map[string]interface{})["get"].(map[string]interface{})
	desc := getOp["description"].(string)
	if desc != "A test endpoint" {
		t.Errorf("expected description 'A test endpoint', got %q", desc)
	}
}

// TestArrayOfStructQueryParam mirrors chatbot-ms's LogReq: a form-tagged
// []Filter/[]Sort field used as a GET query parameter. renderStruct used to
// look up the type via a primitive-only map, so a non-primitive element type
// fell through to a bare reflect.Kind zero value and rendered as the literal
// string "invalid" instead of a real array schema.
func TestArrayOfStructQueryParam(t *testing.T) {
	filterType := defineStruct("Filter",
		member("Field", primitiveType("string"), `json:"field"`, ""),
	)
	route := spec.Route{
		Method:  "get",
		Path:    "/logs",
		Handler: "listLogs",
		RequestType: defineStruct("ListLogReq",
			member("Filters", primitiveType("[]Filter"), `form:"filters"`, ""),
		),
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				filterType,
				defineStruct("ListLogReq",
					member("Filters", primitiveType("[]Filter"), `form:"filters"`, ""),
				),
			},
			Service: spec.Service{
				Name: "test-api",
				Groups: []spec.Group{
					{
						Routes: []spec.Route{route},
					},
				},
			},
		},
	}

	result := runGenerate(t, p, "", "", "")
	paths := result["paths"].(map[string]interface{})
	getOp := paths["/logs"].(map[string]interface{})["get"].(map[string]interface{})
	params := getOp["parameters"].([]interface{})
	if len(params) != 1 {
		t.Fatalf("expected 1 query parameter, got %v", params)
	}
	param := params[0].(map[string]interface{})
	schema := param["schema"].(map[string]interface{})
	if schema["type"] != "array" {
		t.Errorf("expected array schema for []Filter query param, got %v", schema)
	}
	// no $ref/object items: Swagger UI would render a full-height JSON editor
	items, ok := schema["items"].(map[string]interface{})
	if !ok || items["$ref"] != nil || items["type"] != "string" {
		t.Errorf("expected plain string items, got %v", schema["items"])
	}
}
