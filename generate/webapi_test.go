package generate_test

import (
	"testing"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

func TestWebAPIPatterns(t *testing.T) {
	// This test validates patterns found in the real-world example/web.api file.
	// It constructs a spec mirroring key patterns and validates the OpenAPI output.

	types := []spec.Type{
		// Nested struct with multiple field types
		defineStruct("NewRenderRequest",
			member("RefId", primitiveType("string"), `json:"ref_id,optional"`, ""),
			member("TextContent", primitiveType("string"), `json:"text_content"`, ""),
			member("LanguageCode", primitiveType("string"), `json:"language_code"`, ""),
			member("SkipAudioRender", primitiveType("bool"), `json:"skipAudioRender,default=false"`, ""),
			member("Gesture", primitiveType("string"), `json:"gesture,optional"`, ""),
		),
		// Struct with path, form, and header params
		defineStruct("GetProjectKioskConfigReq",
			member("Project", primitiveType("string"), `path:"project"`, ""),
			member("KioskId", primitiveType("string"), `path:"kioskId,optional"`, ""),
		),
		// GET with form query params + default
		defineStruct("ListModelActionsReq",
			member("UseEmbedded", primitiveType("bool"), `form:"use_embedded,default=false"`, ""),
		),
		// Struct with map field
		defineStruct("SessionInfo",
			member("Session", primitiveType("string"), `json:"session"`, ""),
			member("Parameters", spec.PrimitiveType{RawName: "map[string]interface{}"}, `json:"parameters"`, ""),
		),
		// Deeply nested struct
		defineStruct("DialogFlowReq",
			member("SessionInfo", defineStruct("SessionInfo",
				member("Session", primitiveType("string"), `json:"session"`, ""),
				member("Parameters", spec.PrimitiveType{RawName: "map[string]interface{}"}, `json:"parameters"`, ""),
			), `json:"sessionInfo"`, ""),
			member("LanguageCode", primitiveType("string"), `json:"languageCode"`, ""),
		),
		// Response with array of structs
		defineStruct("GetCatalogListResp",
			member("Nodes", spec.ArrayType{RawName: "[]Catalog",
				Value: defineStruct("Catalog",
					member("Id", primitiveType("string"), `json:"id"`, ""),
					member("Title", primitiveType("string"), `json:"title"`, ""),
				),
			}, `json:"nodes"`, ""),
		),
		// Pointer types + optional
		defineStruct("VideoLookup",
			member("Id", primitiveType("string"), `json:"id"`, ""),
			member("ModelConfigHash", primitiveType("*string"), `json:"model_config_hash,optional"`, ""),
			member("Available", primitiveType("*bool"), `json:"available,optional"`, ""),
			member("Ttl", primitiveType("*int"), `json:"ttl,optional"`, ""),
		),
		// Inline struct composition
		defineStruct("SingpostKioskUpdateStatusReq",
			member("", defineStruct("SingpostGetPaymentStatusReq",
				member("MerchantReferenceId", primitiveType("string"), `path:"merchant_reference_id"`, ""),
			), "", ""),
			member("KioskReceiveStatus", primitiveType("string"), `json:"kioskReceiveStatus"`, ""),
		),
		// GET with form params + header
		defineStruct("UserAuthMagicLinkStreamReq",
			member("ChallengeId", primitiveType("string"), `form:"challengeId"`, ""),
			member("Token", primitiveType("string"), `header:"x-api-key"`, ""),
		),
		// WeatherResp with map result
		defineStruct("WeatherResp",
			member("Success", primitiveType("bool"), `json:"success"`, ""),
			member("Result", spec.PrimitiveType{RawName: "map[string]interface{}"}, `json:"result"`, ""),
		),
	}

	routes := []spec.Route{
		// POST with body, nested struct
		{
			Method:  "post",
			Path:    "/new_request",
			Handler: "NewRequestHandler",
			RequestType: defineStruct("NewRenderRequest",
				member("RefId", primitiveType("string"), `json:"ref_id,optional"`, ""),
				member("TextContent", primitiveType("string"), `json:"text_content"`, ""),
				member("LanguageCode", primitiveType("string"), `json:"language_code"`, ""),
			),
			ResponseType: defineStruct("RenderRequest",
				member("Id", primitiveType("string"), `json:"id"`, ""),
				member("DisplayText", primitiveType("string"), `json:"display_text,optional"`, ""),
			),
		},
		// GET with path params
		{
			Method:  "get",
			Path:    "/config/:project",
			Handler: "GetProjectKioskConfigHandler",
			RequestType: defineStruct("GetProjectKioskConfigReq",
				member("Project", primitiveType("string"), `path:"project"`, ""),
			),
			ResponseType: defineStruct("GetProjectKioskConfigResp",
				member("Nodes", spec.ArrayType{RawName: "[]ProjectKioskConfig",
					Value: defineStruct("ProjectKioskConfig",
						member("Id", primitiveType("string"), `json:"id"`, ""),
						member("Project", primitiveType("string"), `json:"project"`, ""),
					),
				}, `json:"nodes"`, ""),
				member("Success", primitiveType("bool"), `json:"success"`, ""),
			),
		},
		// GET with form query params (JWT protected)
		{
			Method:  "get",
			Path:    "/list_model_actions",
			Handler: "ListModelActionsHandler",
			RequestType: defineStruct("ListModelActionsReq",
				member("UseEmbedded", primitiveType("bool"), `form:"use_embedded,default=false"`, ""),
			),
			ResponseType: defineStruct("ListModelActionsResp",
				member("Nodes", spec.ArrayType{RawName: "[]ModelAction",
					Value: defineStruct("ModelAction",
						member("ModelId", primitiveType("string"), `json:"model_id"`, ""),
					),
				}, `json:"data"`, ""),
			),
		},
		// DELETE with path param
		{
			Method:  "delete",
			Path:    "/:hash",
			Handler: "VideoLookupDeleteHandler",
			RequestType: defineStruct("VideoLookupDeleteReq",
				member("Hash", primitiveType("string"), `path:"hash"`, ""),
			),
			ResponseType: defineStruct("VideoLookupDeleteResp",
				member("ReturnCode", primitiveType("int"), `json:"returnCode"`, ""),
				member("Message", primitiveType("string"), `json:"message"`, ""),
			),
		},
		// GET with both form and header params
		{
			Method:  "get",
			Path:    "/magic-link",
			Handler: "UserAuthMagicLinkStreamHandler",
			RequestType: defineStruct("UserAuthMagicLinkStreamReq",
				member("ChallengeId", primitiveType("string"), `form:"challengeId"`, ""),
				member("Token", primitiveType("string"), `header:"x-api-key"`, ""),
			),
		},
		// POST with form data (UploadReq)
		{
			Method:  "post",
			Path:    "/",
			Handler: "UploadHandler",
			RequestType: defineStruct("UploadReq",
				member("File", primitiveType("string"), `form:"file,optional"`, ""),
				member("Path", primitiveType("string"), `form:"path,optional"`, ""),
			),
			ResponseType: defineStruct("UploadResp",
				member("Url", primitiveType("string"), `json:"url"`, ""),
				member("FullUrl", primitiveType("string"), `json:"full_url"`, ""),
			),
		},
		// GET with no request type, just response
		{
			Method:  "get",
			Path:    "/health",
			Handler: "PingHandler",
		},
		// POST with a non-file form field: should not trigger multipart/form-data
		{
			Method:  "post",
			Path:    "/comment",
			Handler: "CommentHandler",
			RequestType: defineStruct("CommentReq",
				member("Comment", primitiveType("string"), `form:"comment"`, ""),
			),
			ResponseType: defineStruct("CommentResp",
				member("Success", primitiveType("bool"), `json:"success"`, ""),
			),
		},
	}

	groups := []spec.Group{
		{
			Annotation: spec.Annotation{
				Properties: map[string]string{
					"prefix": "/v1/render",
					"group":  "render",
				},
			},
			Routes: []spec.Route{routes[0]},
		},
		{
			Annotation: spec.Annotation{
				Properties: map[string]string{
					"prefix": "/v1/kiosk",
					"group":  "kiosk",
				},
			},
			Routes: []spec.Route{routes[1]},
		},
		{
			Annotation: spec.Annotation{
				Properties: map[string]string{
					"prefix": "/v1/profile",
					"group":  "profile",
					"jwt":    "Auth",
				},
			},
			Routes: []spec.Route{routes[2]},
		},
		{
			Annotation: spec.Annotation{
				Properties: map[string]string{
					"prefix":     "/v1/video-lookup",
					"group":      "video_lookup",
					"middleware": "XApiKeyHeaderMiddleware",
				},
			},
			Routes: []spec.Route{routes[3]},
		},
		{
			Annotation: spec.Annotation{
				Properties: map[string]string{
					"prefix":  "/v1/userauth",
					"group":   "userauth",
					"timeout": "5m",
					"sse":     "true",
				},
			},
			Routes: []spec.Route{routes[4]},
		},
		{
			Annotation: spec.Annotation{
				Properties: map[string]string{
					"prefix": "/v1/upload",
					"group":  "upload",
				},
			},
			Routes: []spec.Route{routes[5]},
		},
		{
			Routes: []spec.Route{routes[6]},
		},
		{
			Routes: []spec.Route{routes[7]},
		},
	}

	p := &plugin.Plugin{
		Api: &spec.ApiSpec{
			Types: types,
			Info: spec.Info{
				Properties: map[string]string{
					"title":   `"Web API"`,
					"version": `"1.0.0"`,
					"desc":    `"Comprehensive web API examples"`,
				},
			},
			Service: spec.Service{
				Name:   "web-api",
				Groups: groups,
			},
		},
	}

	result := runGenerate(t, p, "api.example.com", "/v1", "https")
	t.Logf("Generated OpenAPI spec for web.api patterns")

	// 1. Verify root fields
	openapi, _ := result["openapi"].(string)
	if openapi != "3.1.0" {
		t.Errorf("expected openapi 3.1.0, got %v", openapi)
	}

	info, _ := result["info"].(map[string]interface{})
	if info["title"] != "Web API" {
		t.Errorf("expected title 'Web API', got %v", info["title"])
	}

	// 2. Verify server URL
	servers, _ := result["servers"].([]interface{})
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	server, _ := servers[0].(map[string]interface{})
	if server["url"] != "https://api.example.com/v1" {
		t.Errorf("expected URL 'https://api.example.com/v1', got %v", server["url"])
	}

	// 3. Verify paths exist
	paths, _ := result["paths"].(map[string]interface{})

	// POST with request body
	checkPathMethod(t, paths, "/v1/render/new_request", "post", "NewRequestHandler")
	postOp := paths["/v1/render/new_request"].(map[string]interface{})["post"].(map[string]interface{})

	// Should have requestBody, not body parameter
	_, hasParams := postOp["parameters"]
	if hasParams {
		t.Errorf("POST /v1/render/new_request should not have parameters")
	}
	reqBody, hasBody := postOp["requestBody"]
	if !hasBody {
		t.Fatal("POST /v1/render/new_request should have requestBody")
	}
	reqBodyMap, _ := reqBody.(map[string]interface{})
	content, _ := reqBodyMap["content"].(map[string]interface{})
	if _, ok := content["application/json"]; !ok {
		t.Errorf("requestBody should have application/json content")
	}

	// Response should have content
	resp200 := postOp["responses"].(map[string]interface{})["200"].(map[string]interface{})
	if _, ok := resp200["content"]; !ok {
		t.Errorf("200 response should have content")
	}

	// GET with path parameter — verify {project} and parameter
	checkPathMethod(t, paths, "/v1/kiosk/config/{project}", "get", "GetProjectKioskConfigHandler")
	getOp := paths["/v1/kiosk/config/{project}"].(map[string]interface{})["get"].(map[string]interface{})
	getParams, _ := getOp["parameters"].([]interface{})
	if len(getParams) == 0 {
		t.Errorf("expected path parameter for /v1/kiosk/config/{project}")
	} else {
		param, _ := getParams[0].(map[string]interface{})
		if param["in"] != "path" {
			t.Errorf("expected path parameter, got in=%v", param["in"])
		}
		if param["required"] != true {
			t.Errorf("path parameter should be required")
		}
	}

	// GET with form query params
	checkPathMethod(t, paths, "/v1/profile/list_model_actions", "get", "ListModelActionsHandler")
	listOp := paths["/v1/profile/list_model_actions"].(map[string]interface{})["get"].(map[string]interface{})
	listParams, _ := listOp["parameters"].([]interface{})
	if len(listParams) == 0 {
		t.Errorf("expected query parameters for GET /v1/profile/list_model_actions")
	} else {
		param, _ := listParams[0].(map[string]interface{})
		if param["in"] != "query" {
			t.Errorf("expected query parameter, got in=%v", param["in"])
		}
		schema, _ := param["schema"].(map[string]interface{})
		if schema["type"] != "boolean" {
			t.Errorf("expected boolean type for use_embedded, got %v", schema["type"])
		}
		if _, ok := schema["format"]; ok {
			t.Errorf("boolean schema should not have format, got %v", schema["format"])
		}
		// GET should NOT have requestBody
		if _, has := listOp["requestBody"]; has {
			t.Errorf("GET /v1/profile/list_model_actions should not have requestBody")
		}
	}

	// DELETE with path parameter
	checkPathMethod(t, paths, "/v1/video-lookup/{hash}", "delete", "VideoLookupDeleteHandler")

	// GET with both form and header params
	checkPathMethod(t, paths, "/v1/userauth/magic-link", "get", "UserAuthMagicLinkStreamHandler")
	magicOp := paths["/v1/userauth/magic-link"].(map[string]interface{})["get"].(map[string]interface{})
	magicParams, _ := magicOp["parameters"].([]interface{})
	hasForm := false
	hasHeader := false
	for _, p := range magicParams {
		pm, _ := p.(map[string]interface{})
		if pm["in"] == "form" || pm["in"] == "query" {
			hasForm = true
		}
		if pm["in"] == "header" {
			hasHeader = true
		}
	}
	if !hasForm {
		t.Errorf("expected query parameter for challengeId")
	}
	if !hasHeader {
		t.Errorf("expected header parameter for x-api-key")
	}

	// POST with form data (multipart)
	checkPathMethod(t, paths, "/v1/upload/", "post", "UploadHandler")
	uploadOp := paths["/v1/upload/"].(map[string]interface{})["post"].(map[string]interface{})
	uploadBody, _ := uploadOp["requestBody"].(map[string]interface{})
	uploadContent, _ := uploadBody["content"].(map[string]interface{})
	if _, ok := uploadContent["multipart/form-data"]; !ok {
		t.Errorf("upload should have multipart/form-data content")
	}

	// GET with no request type — no parameters or request body
	checkPathMethod(t, paths, "/health", "get", "PingHandler")
	pingOp := paths["/health"].(map[string]interface{})["get"].(map[string]interface{})
	if _, has := pingOp["parameters"]; has {
		t.Errorf("health check should not have parameters")
	}

	// POST with only a non-file form field: it's a query param, so no body at all
	checkPathMethod(t, paths, "/comment", "post", "CommentHandler")
	commentOp := paths["/comment"].(map[string]interface{})["post"].(map[string]interface{})
	if rb, has := commentOp["requestBody"]; has {
		t.Errorf("comment endpoint should have no requestBody, got %v", rb)
	}
	if _, has := pingOp["requestBody"]; has {
		t.Errorf("health check should not have requestBody")
	}

	// 4. Verify JWT security on profile group
	if listOp["security"] == nil {
		t.Errorf("list_model_actions should have JWT security")
	} else {
		sec, _ := listOp["security"].([]interface{})
		if len(sec) == 2 {
			sec0, _ := sec[0].(map[string]interface{})
			sec1, _ := sec[1].(map[string]interface{})
			if _, ok := sec0["bearerAuth"]; !ok {
				t.Errorf("expected bearerAuth security in first entry, got %v", sec)
			}
			if _, ok := sec1["apiKey"]; !ok {
				t.Errorf("expected apiKey security in second entry, got %v", sec)
			}
		} else {
			t.Errorf("expected 2 security entries (bearerAuth + apiKey), got %d: %v", len(sec), sec)
		}
	}

	// 5. Verify tags
	for _, pathKey := range []string{"/v1/render/new_request", "/v1/kiosk/config/{project}", "/v1/upload/"} {
		pi, ok := paths[pathKey].(map[string]interface{})
		if !ok {
			continue
		}
		for _, method := range []string{"get", "post", "delete"} {
			op, ok := pi[method].(map[string]interface{})
			if !ok {
				continue
			}
			tags, _ := op["tags"].([]interface{})
			if len(tags) == 0 {
				t.Errorf("operation %s %s should have tags", method, pathKey)
			}
		}
	}

	// 6. Verify components schemas
	comp, _ := result["components"].(map[string]interface{})
	schemas, _ := comp["schemas"].(map[string]interface{})

	// Check NewRenderRequest has properties (nested structs)
	renderReq, ok := schemas["NewRenderRequest"]
	if !ok {
		t.Errorf("expected NewRenderRequest in schemas")
	} else {
		rr, _ := renderReq.(map[string]interface{})
		props, _ := rr["properties"].(map[string]interface{})
		if props["text_content"] == nil {
			t.Errorf("expected text_content property in NewRenderRequest")
		}
		// Check default value (typed as bool)
		skipAudio, _ := props["skipAudioRender"].(map[string]interface{})
		if skipAudio["default"] != false {
			t.Errorf("expected default=false for skipAudioRender, got %v (type: %T)", skipAudio["default"], skipAudio["default"])
		}
	}

	// Check pointer types in VideoLookup
	videoLookup, ok := schemas["VideoLookup"]
	if !ok {
		t.Errorf("expected VideoLookup in schemas")
	} else {
		vl, _ := videoLookup.(map[string]interface{})
		props, _ := vl["properties"].(map[string]interface{})
		for _, propName := range []string{"id", "model_config_hash", "available", "ttl"} {
			if props[propName] == nil {
				t.Errorf("expected %s property in VideoLookup", propName)
			}
		}
		// *string should have nullable: true
		if hash, ok := props["model_config_hash"].(map[string]interface{}); ok {
			if hash["nullable"] != true {
				t.Errorf("expected nullable=true for model_config_hash (*string), got %v", hash["nullable"])
			}
		}
		// *bool should have nullable: true and no format
		if avail, ok := props["available"].(map[string]interface{}); ok {
			if avail["nullable"] != true {
				t.Errorf("expected nullable=true for available (*bool), got %v", avail["nullable"])
			}
			if _, ok := avail["format"]; ok {
				t.Errorf("boolean should not have format")
			}
		}
		// *int should have nullable: true
		if ttl, ok := props["ttl"].(map[string]interface{}); ok {
			if ttl["nullable"] != true {
				t.Errorf("expected nullable=true for ttl (*int), got %v", ttl["nullable"])
			}
		}
		// Non-pointer string should NOT have nullable
		if id, ok := props["id"].(map[string]interface{}); ok {
			if id["nullable"] == true {
				t.Errorf("expected no nullable for id (string), got nullable=true")
			}
		}
	}

	// Check inline struct in SingpostKioskUpdateStatusReq
	// Path-parameter fields from embedded struct are not included in schema properties
	skus, ok := schemas["SingpostKioskUpdateStatusReq"]
	if !ok {
		t.Errorf("expected SingpostKioskUpdateStatusReq in schemas")
	} else {
		sk, _ := skus.(map[string]interface{})
		props, _ := sk["properties"].(map[string]interface{})
		if props["merchant_reference_id"] != nil {
			t.Errorf("merchant_reference_id is a path param, should not be in body schema")
		}
		if props["kioskReceiveStatus"] == nil {
			t.Errorf("expected kioskReceiveStatus as body property, got: %v", keys(props))
		}
	}

	// Check map[string]interface{} type in SessionInfo
	sessionInfo, ok := schemas["SessionInfo"]
	if !ok {
		t.Errorf("expected SessionInfo in schemas")
	} else {
		si, _ := sessionInfo.(map[string]interface{})
		props, _ := si["properties"].(map[string]interface{})
		params, _ := props["parameters"].(map[string]interface{})
		if params["type"] != "object" {
			t.Errorf("expected map[string]interface{} to be type 'object', got %v", params["type"])
		}
		if _, has := params["additionalProperties"]; has {
			t.Errorf("expected map[string]interface{} to omit additionalProperties, got %v", params["additionalProperties"])
		}
	}

	// Check array of structs in GetCatalogListResp
	catalogResp, ok := schemas["GetCatalogListResp"]
	if !ok {
		t.Errorf("expected GetCatalogListResp in schemas")
	} else {
		cr, _ := catalogResp.(map[string]interface{})
		props, _ := cr["properties"].(map[string]interface{})
		nodes, _ := props["nodes"].(map[string]interface{})
		if nodes["type"] != "array" {
			t.Errorf("expected nodes type 'array', got %v", nodes["type"])
		}
	}

	// 7. Verify securitySchemes in components
	secSchemes, _ := comp["securitySchemes"].(map[string]interface{})

	// Check bearerAuth scheme
	bearerAuth, ok := secSchemes["bearerAuth"]
	if !ok {
		t.Errorf("expected bearerAuth in securitySchemes")
	} else {
		ba, _ := bearerAuth.(map[string]interface{})
		if ba["type"] != "http" {
			t.Errorf("expected bearerAuth type 'http', got %v", ba["type"])
		}
		if ba["scheme"] != "bearer" {
			t.Errorf("expected bearerAuth scheme 'bearer', got %v", ba["scheme"])
		}
	}

	// Check apiKey scheme
	apiKey, ok := secSchemes["apiKey"]
	if !ok {
		t.Errorf("expected apiKey in securitySchemes")
	} else {
		ak, _ := apiKey.(map[string]interface{})
		if ak["type"] != "apiKey" {
			t.Errorf("expected apiKey type 'apiKey', got %v", ak["type"])
		}
		if ak["in"] != "header" {
			t.Errorf("expected apiKey in 'header', got %v", ak["in"])
		}
	}

	// 8. Verify health check has no security and default tag
	if _, has := pingOp["security"]; has {
		t.Errorf("health check should not have security")
	}
	tags, _ := pingOp["tags"].([]interface{})
	if len(tags) == 0 {
		t.Errorf("health check should have default service name tag")
	}
}

func checkPathMethod(t *testing.T, paths map[string]interface{}, path, method, handler string) {
	t.Helper()
	pi, ok := paths[path]
	if !ok {
		t.Fatalf("expected path %s, got paths: %v", path, keys(paths))
	}
	op, ok := pi.(map[string]interface{})[method].(map[string]interface{})
	if !ok {
		t.Fatalf("expected %s method at path %s", method, path)
	}
	if op["operationId"] != handler {
		t.Errorf("expected operationId '%s' at %s %s, got %v", handler, method, path, op["operationId"])
	}
}
