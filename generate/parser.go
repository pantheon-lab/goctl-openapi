package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

var strColon = []byte(":")

// openapiVersion is the OpenAPI Specification version the output conforms to,
// see https://spec.openapis.org/oas/v3.1.2.html
const openapiVersion = "3.1.2"

const (
	jsonTagKey      = "json"
	defaultOption   = "default"
	stringOption    = "string"
	optionalOption  = "optional"
	omitemptyOption = "omitempty"
	optionsOption   = "options"
	rangeOption     = "range"
	exampleOption   = "example"
	optionSeparator = "|"
	equalToken      = "="
	atRespDoc       = "@respdoc-"
)

var rangeRegexp = regexp.MustCompile(`^([\[(])\s*([+-]?\d+(?:\.\d+)?)?\s*:\s*([+-]?\d+(?:\.\d+)?)?\s*([\])])$`)

// applyRangeOption renders go-zero's range option, e.g. [1:10] or (0:100], as
// JSON Schema bounds. A parenthesis marks an exclusive bound, which OAS 3.1
// (JSON Schema 2020-12) expresses as a numeric exclusiveMinimum/exclusiveMaximum
// rather than a boolean modifier. Either bound may be omitted.
func applyRangeOption(schema *openapiSchemaObject, option string) {
	result := rangeRegexp.FindStringSubmatch(strings.TrimSpace(option))
	if result == nil {
		return
	}
	var bounds [2]*float64
	for i, s := range []string{result[2], result[3]} {
		if s == "" {
			continue
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return
		}
		bounds[i] = &v
	}
	min, max := bounds[0], bounds[1]
	if min != nil && max != nil && *max < *min {
		max = min
	}
	if min != nil {
		if result[1] == "(" {
			schema.ExclusiveMinimum = min
		} else {
			schema.Minimum = min
		}
	}
	if max != nil {
		if result[4] == ")" {
			schema.ExclusiveMaximum = max
		} else {
			schema.Maximum = max
		}
	}
}

// enumValues types go-zero's options=a|b|c values to match the schema, since
// JSON Schema enum compares by JSON value: "1" never matches an integer 1.
// A nullable schema must also list null for null to stay a valid value.
func enumValues(option string, schema *openapiSchemaObject) []interface{} {
	var values []interface{}
	for _, v := range strings.Split(option, optionSeparator) {
		values = append(values, typedValue(v, schema.Type))
	}
	if schema.Nullable {
		values = append(values, nil)
	}
	return values
}

func applyGenerate(p *plugin.Plugin, host string, basePath string, schemes string) (*openapiObject, error) {
	title := p.Api.Info.Properties["title"]
	if unquoted, err := strconv.Unquote(title); err == nil {
		title = unquoted
	}
	version := p.Api.Info.Properties["version"]
	if unquoted, err := strconv.Unquote(version); err == nil {
		version = unquoted
	}
	desc := p.Api.Info.Properties["desc"]
	if unquoted, err := strconv.Unquote(desc); err == nil {
		desc = unquoted
	}

	o := openapiObject{
		OpenAPI: openapiVersion,
		Paths:   make(openapiPathsObject),
		Components: openapiComponentsObject{
			Schemas: make(openapiSchemasObject),
		},
		Info: openapiInfoObject{
			Title:       title,
			Version:     version,
			Description: desc,
		},
	}

	// Build servers array
	schemeList := []string{"http", "https"}
	if len(schemes) > 0 {
		supportedSchemes := []string{"http", "https", "ws", "wss"}
		ss := strings.Split(schemes, ",")
		for i := range ss {
			scheme := strings.TrimSpace(ss[i])
			if !contains(supportedSchemes, scheme) {
				log.Fatalf("unsupport scheme: [%s], only support [http, https, ws, wss]", scheme)
			}
			ss[i] = scheme
		}
		schemeList = ss
	}

	serverURL := buildServerURL(schemeList, host, basePath)
	o.Servers = []openapiServerObject{
		{
			URL:         serverURL,
			Description: "API server",
		},
	}

	o.Components.SecuritySchemes = make(openapiSecuritySchemesObject)
	o.Components.SecuritySchemes["bearerAuth"] = openapiSecuritySchemeObject{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "Enter JWT Bearer token",
	}
	o.Components.SecuritySchemes["apiKey"] = openapiSecuritySchemeObject{
		Type:        "apiKey",
		In:          "header",
		Name:        "x-api-key",
		Description: "API key authentication via x-api-key header",
	}

	requestResponseRefs := refMap{}
	paramOnlyTypes := map[string]bool{}
	renderServiceRoutes(p.Api.Service, p.Api.Service.Groups, o.Paths, requestResponseRefs, paramOnlyTypes, &o.Components)

	renderReplyAsDefinition(o.Components.Schemas, p.Api.Types, requestResponseRefs)
	pruneUnreferencedSchemas(o.Components.Schemas, requestResponseRefs, paramOnlyTypes)

	return &o, nil
}

func buildServerURL(schemes []string, host, basePath string) string {
	scheme := schemes[0]

	url := scheme + "://"
	if len(host) > 0 {
		url += host
	} else {
		url += "localhost"
	}
	if len(basePath) > 0 {
		if basePath[0] != '/' {
			url += "/"
		}
		url += basePath
	}
	return url
}

func renderServiceRoutes(service spec.Service, groups []spec.Group, paths openapiPathsObject, requestResponseRefs refMap, paramOnlyTypes map[string]bool, components *openapiComponentsObject) {
	// Pass 1: count request type usage for shared requestBody dedup
	requestTypeCount := map[string]int{}
	for _, group := range groups {
		for _, route := range group.Routes {
			if route.RequestType != nil && len(route.RequestType.Name()) > 0 {
				requestTypeCount[route.RequestType.Name()]++
			}
		}
	}

	for _, group := range groups {
		for _, route := range group.Routes {
			path := group.GetAnnotation("prefix") + route.Path
			if len(path) == 0 || path[0] != '/' {
				path = "/" + path
			}
			parameters := openapiParametersObject{}

			if countParams(path) > 0 {
				p := strings.Split(path, "/")
				for i := range p {
					part := p[i]
					if strings.Contains(part, ":") {
						key := strings.TrimPrefix(p[i], ":")
						path = strings.Replace(path, fmt.Sprintf(":%s", key), fmt.Sprintf("{%s}", key), 1)

						spo := openapiParameterObject{
							Name:     key,
							In:       "path",
							Required: true,
							Schema: &openapiSchemaObject{
								Type: "string",
							},
						}

						// Look up the type from the request struct member. A path
						// parameter is always required: OAS 3.1.2 mandates
						// `required: true` for `in: path`, even if tagged optional.
						if defineStruct, ok := route.RequestType.(spec.DefineStruct); ok {
							for _, member := range defineStruct.Members {
								memberKey := ""
								for _, tag := range member.Tags() {
									if tag.Key == "path" {
										memberKey = tag.Name
									}
								}
								if memberKey == key {
									memberSchema := schemaOfTypeName(member.Type.Name())
									spo.Schema = &memberSchema
								}
							}
						}

						prop := route.AtDoc.Properties[key]
						if prop != "" {
							spo.Description = strings.Trim(prop, "\"")
						}

						parameters = append(parameters, spo)
					}
				}
			}

			hasBody := false
			if defineStruct, ok := route.RequestType.(spec.DefineStruct); ok {
				for _, member := range defineStruct.Members {
					if hasHeaderParameters(member) {
						parameters = parseHeader(member, parameters)
					}
				}
				if strings.ToUpper(route.Method) == http.MethodGet {
					for _, member := range defineStruct.Members {
						if hasPathParameters(member) || hasHeaderParameters(member) {
							continue
						}
						if embedStruct, isEmbed := member.Type.(spec.DefineStruct); isEmbed {
							for _, m := range embedStruct.Members {
								parameters = append(parameters, renderStruct(m))
							}
							continue
						}
						parameters = append(parameters, renderStruct(member))
					}
				} else {
					formBody := isFormBodyStruct(defineStruct)
					hasBody = formBody || hasJSONBodyMember(defineStruct)
					// go-zero's httpx.Parse sources form-tagged fields from r.Form,
					// i.e. the URL query string plus a form-encoded body. Alongside a
					// JSON body they can only come from the query string, so they
					// must still appear as query parameters here.
					for _, member := range defineStruct.Members {
						if hasPathParameters(member) || hasHeaderParameters(member) {
							continue
						}
						// in a form body the form fields are the body, not query params
						if formBody || !member.IsFormMember() {
							continue
						}
						if embedStruct, isEmbed := member.Type.(spec.DefineStruct); isEmbed {
							for _, m := range embedStruct.Members {
								if m.IsFormMember() {
									parameters = append(parameters, renderStruct(m))
								}
							}
							continue
						}
						parameters = append(parameters, renderStruct(member))
					}
				}
			}

			pathItemObject, ok := paths[path]
			if !ok {
				pathItemObject = openapiPathItemObject{}
			}

			desc := "A successful response."

			operationObject := &openapiOperationObject{
				Tags:       []string{getTag(service, group)},
				Parameters: parameters,
				Responses: openapiResponsesObject{
					"200": {
						Description: desc,
						Content:     map[string]openapiMediaTypeObject{},
					},
				},
			}

			// Set response content
			if route.ResponseType != nil && len(route.ResponseType.Name()) > 0 {
				// schemaOfTypeName keeps primitives (bool, int64, []string, ...) inline
				// instead of referencing a components schema that never exists.
				schema := schemaOfTypeName(route.ResponseType.Name())
				operationObject.Responses["200"].Content["application/json"] = openapiMediaTypeObject{Schema: &schema}
				markSchemaRefs(requestResponseRefs, schema)
			} else {
				// Empty response — a bare schema:{} matches any type, which Swagger UI
				// renders as "string"; type: object renders {} instead.
				operationObject.Responses["200"].Content["application/json"] = openapiMediaTypeObject{
					Schema: &openapiSchemaObject{Type: "object"},
				}
			}

			// a request type without a body is rendered entirely as parameters,
			// so its schema is only kept if something else references it
			if !hasBody && route.RequestType != nil && len(route.RequestType.Name()) > 0 {
				paramOnlyTypes[route.RequestType.Name()] = true
			}

			// Set request body
			if hasBody {
				typeName := route.RequestType.Name()
				if len(typeName) > 0 {
					if requestTypeCount[typeName] > 1 {
						if components.RequestBodies == nil {
							components.RequestBodies = make(openapiRequestBodiesObject)
						}
						if _, exists := components.RequestBodies[typeName]; !exists {
							reqBody := buildRequestBody(route, typeName)
							components.RequestBodies[typeName] = *reqBody
						}
						operationObject.RequestBody = &openapiRequestBodyObject{
							Ref: "#/components/requestBodies/" + typeName,
						}
					} else {
						operationObject.RequestBody = buildRequestBody(route, typeName)
					}

					requestResponseRefs["#/components/schemas/"+typeName] = struct{}{}
				}
			}

			// Process @respdoc- annotations
			for _, v := range route.Doc {
				markerIndex := strings.Index(v, atRespDoc)
				if markerIndex >= 0 {
					l := strings.Index(v, "(")
					r := strings.Index(v, ")")
					code := strings.TrimSpace(v[markerIndex+len(atRespDoc) : l])
					var comment string
					commentIndex := strings.Index(v, "//")
					if commentIndex > 0 {
						comment = strings.TrimSpace(strings.Trim(v[commentIndex+2:], "*/"))
					}
					content := strings.TrimSpace(v[l+1 : r])
					if strings.Index(v, ":") > 0 {
						lines := strings.Split(content, "\n")
						kv := make(map[string]string, len(lines))
						for _, line := range lines {
							sep := strings.Index(line, ":")
							key := strings.TrimSpace(line[:sep])
							value := strings.TrimSpace(line[sep+1:])
							kv[key] = value
						}
						kvByte, err := json.Marshal(kv)
						if err != nil {
							continue
						}
						// a 3.1 Media Type Object example, rather than the deprecated
						// Schema Object `example`
						var example interface{}
						if err := json.Unmarshal(kvByte, &example); err != nil {
							continue
						}
						operationObject.Responses[code] = openapiResponseObject{
							Description: responseDescription(code, comment),
							Content: map[string]openapiMediaTypeObject{
								"application/json": {
									Schema:  &openapiSchemaObject{Type: "object"},
									Example: example,
								},
							},
						}
					} else if len(content) > 0 {
						schema := schemaOfTypeName(content)
						markSchemaRefs(requestResponseRefs, schema)
						operationObject.Responses[code] = openapiResponseObject{
							Description: responseDescription(code, comment),
							Content: map[string]openapiMediaTypeObject{
								"application/json": {Schema: &schema},
							},
						}
					}
				}
			}

			operationObject.OperationID = route.Handler

			operationObject.Summary = strings.ReplaceAll(route.JoinedDoc(), "\"", "")

			if len(route.AtDoc.Properties) > 0 {
				operationObject.Description = route.AtDoc.Properties["description"]
				if unquoted, err := strconv.Unquote(operationObject.Description); err == nil {
					operationObject.Description = unquoted
				}
			}

			operationObject.Description = strings.ReplaceAll(operationObject.Description, "\"", "")

			if group.Annotation.Properties["jwt"] != "" {
				operationObject.Security = &[]openapiSecurityRequirementObject{
					{"bearerAuth": []string{}},
					{"apiKey": []string{}},
				}
			}

			switch strings.ToUpper(route.Method) {
			case http.MethodGet:
				pathItemObject.Get = operationObject
			case http.MethodPost:
				pathItemObject.Post = operationObject
			case http.MethodDelete:
				pathItemObject.Delete = operationObject
			case http.MethodPut:
				pathItemObject.Put = operationObject
			case http.MethodPatch:
				pathItemObject.Patch = operationObject
			}

			paths[path] = pathItemObject
		}
	}
}

// responseDescription returns a non-empty Response Object description, which
// OAS 3.1.2 requires, falling back to the status text of the code.
func responseDescription(code, comment string) string {
	if comment != "" {
		return comment
	}
	if n, err := strconv.Atoi(code); err == nil && http.StatusText(n) != "" {
		return http.StatusText(n)
	}
	return "Response " + code
}

// markSchemaRefs records the components schema a (possibly array) schema
// points at, if any.
func markSchemaRefs(refs refMap, schema openapiSchemaObject) {
	for schema.Items != nil {
		schema = *schema.Items
	}
	if schema.Ref != "" {
		refs[schema.Ref] = struct{}{}
	}
}

// pruneUnreferencedSchemas drops the schemas of request types that are only
// ever rendered as parameters (e.g. a GET request type, whose fields are all
// query parameters), unless an operation reaches them directly or through
// nested $refs.
func pruneUnreferencedSchemas(schemas openapiSchemasObject, roots refMap, candidates map[string]bool) {
	const prefix = "#/components/schemas/"
	reachable := map[string]bool{}
	var queue []string
	var visit func(s openapiSchemaObject)
	visit = func(s openapiSchemaObject) {
		if name := strings.TrimPrefix(s.Ref, prefix); s.Ref != "" && !reachable[name] {
			reachable[name] = true
			queue = append(queue, name)
		}
		for _, prop := range s.Properties {
			visit(prop)
		}
		if s.Items != nil {
			visit(*s.Items)
		}
		if s.AdditionalProperties != nil && s.AdditionalProperties.Schema != nil {
			visit(*s.AdditionalProperties.Schema)
		}
	}
	for ref := range roots {
		visit(openapiSchemaObject{Ref: ref})
	}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if s, ok := schemas[name]; ok {
			visit(s)
		}
	}
	for name := range candidates {
		if !reachable[name] {
			delete(schemas, name)
		}
	}
}

// isFileUploadMember reports whether a form-tagged member looks like a file
// upload rather than an ordinary form value. go-zero's .api DSL has no real
// file-upload type — uploads bypass struct fields via http.Request.FormFile —
// so multipart.FileHeader in the type name is the strongest signal when
// expressible; otherwise the form tag or field name must contain a whole word
// like "file" or "upload", so that e.g. profile_id/ProfileId don't match.
func isFileUploadMember(member spec.Member) bool {
	if strings.Contains(strings.ToLower(member.Type.Name()), "multipart.fileheader") {
		return true
	}
	names := []string{member.Name}
	for _, tag := range member.Tags() {
		if tag.Key == "form" {
			names = append(names, tag.Name)
		}
	}
	for _, name := range names {
		for _, word := range splitWords(name) {
			if uploadWords[word] {
				return true
			}
		}
	}
	return false
}

var uploadWords = map[string]bool{"file": true, "files": true, "upload": true, "attachment": true}

// splitWords splits snake_case, kebab-case and camelCase identifiers into
// lowercase words.
func splitWords(s string) []string {
	var words []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			words = append(words, strings.ToLower(string(cur)))
			cur = cur[:0]
		}
	}
	runes := []rune(s)
	for i, r := range runes {
		switch {
		case r == '_' || r == '-' || r == '.' || r == ' ':
			flush()
			continue
		case unicode.IsUpper(r) && i > 0 &&
			(unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))):
			flush()
		}
		cur = append(cur, r)
	}
	flush()
	return words
}

func buildRequestBody(route spec.Route, typeName string) *openapiRequestBodyObject {
	reqRef := "#/components/schemas/" + typeName

	contentType := "application/json"
	if defineStruct, ok := route.RequestType.(spec.DefineStruct); ok {
		switch {
		// a file upload is only ever sent as multipart, never as a JSON body
		case isMultipartStruct(defineStruct):
			contentType = "multipart/form-data"
		case isFormBodyStruct(defineStruct):
			contentType = "application/x-www-form-urlencoded"
		}
	}

	body := &openapiRequestBodyObject{
		Required: true,
		Content: map[string]openapiMediaTypeObject{
			contentType: {
				Schema: &openapiSchemaObject{
					Ref: reqRef,
				},
			},
		},
	}

	doc := strings.Join(route.RequestType.Documents(), ",")
	doc = strings.Replace(doc, "//", "", -1)
	if doc != "" {
		body.Description = strings.TrimSpace(doc)
	}

	return body
}

func getTag(service spec.Service, group spec.Group) string {
	tags := service.Name
	if value := group.GetAnnotation("group"); len(value) > 0 {
		tags = value
	}
	if value := group.GetAnnotation("swtags"); len(value) > 0 {
		tags = value
	}
	return tags
}

// simpleQuerySchema flattens non-primitive query parameter schemas. A $ref or
// object schema (and an array of them) under `in: query` makes Swagger UI render
// a full-height JSON editor, and a URL query value is a string anyway.
func simpleQuerySchema(s openapiSchemaObject) openapiSchemaObject {
	switch {
	case s.Ref != "" || s.Type == "object":
		return openapiSchemaObject{Type: "string"}
	case s.Type == "array" && s.Items != nil:
		items := simpleQuerySchema(*s.Items)
		s.Items = &items
	}
	return s
}

func renderStruct(member spec.Member) openapiParameterObject {
	// Delegate to schemaOfTypeName rather than a primitive-only lookup, so
	// non-primitive query members (e.g. a form-tagged []Filter) get a valid
	// schema instead of the literal string "invalid" that a bare reflect.Kind
	// zero-value would otherwise produce.
	built := simpleQuerySchema(schemaOfTypeName(member.Type.Name()))
	schema := &built
	ftype := schema.Type
	sp := openapiParameterObject{In: "query", Schema: schema}

	for _, tag := range member.Tags() {
		if !isBindingTag(tag) {
			continue
		}

		sp.Name = tag.Name
		sp.Required = isRequiredTag(tag)
		for _, option := range tag.Options {
			if strings.HasPrefix(option, optionsOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					schema.Enum = enumValues(segs[1], schema)
				}
			}

			if strings.HasPrefix(option, rangeOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					applyRangeOption(schema, segs[1])
				}
			}

			if strings.HasPrefix(option, defaultOption) {
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					schema.Default = typedValue(segs[1], ftype)
				}
			}

			if strings.HasPrefix(option, exampleOption) {
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					sp.Example = typedValue(segs[1], ftype)
				}
			}

			if option == stringOption {
				schema.Type = "string"
				schema.Format = ""
			}
		}
	}

	if len(member.Comment) > 0 {
		sp.Description = strings.TrimLeft(member.Comment, "//")
	}

	return sp
}

func renderReplyAsDefinition(d openapiSchemasObject, p []spec.Type, refs refMap) {
	for _, i2 := range p {
		schema := openapiSchemaObject{
			Type: "object",
		}
		defineStruct, _ := i2.(spec.DefineStruct)

		schema.Title = defineStruct.Name()
		// form fields are query params (see renderServiceRoutes), not JSON body
		// fields — except in a form body, where they are the body.
		skipForm := !isFormBodyStruct(defineStruct)
		multipart := isMultipartStruct(defineStruct)

		for _, member := range defineStruct.Members {
			if hasPathParameters(member) || hasHeaderParameters(member) || isJSONIgnored(member) || (skipForm && member.IsFormMember()) {
				continue
			}
			propName := member.Name
			if tag, err := member.GetPropertyName(); err == nil {
				propName = tag
			}
			if propName == "" {
				memberStruct, _ := member.Type.(spec.DefineStruct)
				for _, m := range memberStruct.Members {
					if hasHeaderParameters(m) || hasPathParameters(m) || isJSONIgnored(m) || (skipForm && m.IsFormMember()) {
						continue
					}
					subName := m.Name
					if tag, err := m.GetPropertyName(); err == nil {
						subName = tag
					}
					if schema.Properties == nil {
						schema.Properties = make(map[string]openapiSchemaObject)
					}
					schema.Properties[subName] = bodyFieldSchema(m, multipart)
					appendRequired(&schema, m)
				}
				continue
			}
			if schema.Properties == nil {
				schema.Properties = make(map[string]openapiSchemaObject)
			}
			schema.Properties[propName] = bodyFieldSchema(member, multipart)
			appendRequired(&schema, member)
		}

		d[i2.Name()] = schema
	}
}

// bodyFieldSchema is schemaOfField, except that in a multipart body a file
// upload field is a binary string, so Swagger UI renders a file picker.
func bodyFieldSchema(member spec.Member, multipart bool) openapiSchemaObject {
	schema := schemaOfField(member)
	if !multipart || !member.IsFormMember() || !isFileUploadMember(member) {
		return schema
	}
	file := openapiSchemaObject{Type: "string", Format: "binary"}
	if schema.Type == "array" {
		schema.Items = &file
		return schema
	}
	file.Description = schema.Description
	return file
}

func isMultipartStruct(s spec.DefineStruct) bool {
	for _, m := range s.Members {
		if m.IsFormMember() && isFileUploadMember(m) {
			return true
		}
	}
	return false
}

// isFormBodyStruct reports whether a non-GET request is sent as a form body:
// a file upload, or a struct whose only body fields are form-tagged.
func isFormBodyStruct(s spec.DefineStruct) bool {
	return isMultipartStruct(s) || (!hasJSONBodyMember(s) && hasFormMember(s))
}

// hasFormMember reports whether any member, including those of embedded
// structs, is form-tagged.
func hasFormMember(s spec.DefineStruct) bool {
	for _, m := range s.Members {
		if m.IsFormMember() {
			return true
		}
		if embed, ok := m.Type.(spec.DefineStruct); ok && (m.IsInline || m.Name == "" || len(m.Tags()) == 0) {
			if hasFormMember(embed) {
				return true
			}
		}
	}
	return false
}

// hasJSONBodyMember reports whether any member is actually bound from the JSON
// body, i.e. isn't a path/header/form parameter or json:"-".
func hasJSONBodyMember(s spec.DefineStruct) bool {
	for _, m := range s.Members {
		if hasPathParameters(m) || hasHeaderParameters(m) || isJSONIgnored(m) || m.IsFormMember() {
			continue
		}
		if embed, ok := m.Type.(spec.DefineStruct); ok && (m.IsInline || m.Name == "" || len(m.Tags()) == 0) {
			if hasJSONBodyMember(embed) {
				return true
			}
			continue
		}
		return true
	}
	return false
}

func hasPathParameters(member spec.Member) bool {
	for _, tag := range member.Tags() {
		if tag.Key == "path" {
			return true
		}
	}
	return false
}

func hasHeaderParameters(member spec.Member) bool {
	for _, tag := range member.Tags() {
		if tag.Key == "header" {
			return true
		}
	}
	return false
}

// isJSONIgnored reports whether a member is tagged `json:"-"`, meaning it should
// be excluded entirely from the schema (matching encoding/json semantics), rather
// than renamed and required as go-zero's Member.GetPropertyName does.
func isJSONIgnored(member spec.Member) bool {
	for _, tag := range member.Tags() {
		if tag.Key == jsonTagKey && tag.Name == "-" {
			return true
		}
	}
	return false
}

// isBindingTag reports whether a tag is one go-zero binds a field from.
// Any other tag (validate, copier, mask, ...) carries no OpenAPI meaning and
// must not leak into parameter names or required lists.
func isBindingTag(tag *spec.Tag) bool {
	switch tag.Key {
	case jsonTagKey, "form", "path", "header":
		return true
	}
	return false
}

// isRequiredTag mirrors go-zero's binding: a field is optional when tagged
// optional/omitempty, or when it has a default=... to fall back on.
func isRequiredTag(tag *spec.Tag) bool {
	for _, option := range tag.Options {
		if strings.HasPrefix(option, optionalOption) || strings.HasPrefix(option, omitemptyOption) ||
			strings.HasPrefix(option, defaultOption) {
			return false
		}
	}
	return true
}

func appendRequired(schema *openapiSchemaObject, member spec.Member) {
	for _, tag := range member.Tags() {
		if isBindingTag(tag) && isRequiredTag(tag) && !containsStr(schema.Required, tag.Name) {
			schema.Required = append(schema.Required, tag.Name)
		}
	}
}

func schemaOfField(member spec.Member) openapiSchemaObject {
	schema := schemaOfTypeName(member.Type.Name())

	comment := member.GetComment()
	comment = strings.Replace(comment, "//", "", -1)
	if comment != "" {
		schema.Description = comment
	}

	for _, tag := range member.Tags() {
		if len(tag.Options) == 0 {
			continue
		}
		for _, option := range tag.Options {
			switch {
			case strings.HasPrefix(option, defaultOption):
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					schema.Default = typedValue(segs[1], schema.Type)
				}
			case strings.HasPrefix(option, optionsOption):
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					schema.Enum = enumValues(segs[1], &schema)
				}
			case strings.HasPrefix(option, rangeOption):
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					applyRangeOption(&schema, segs[1])
				}
			case strings.HasPrefix(option, exampleOption):
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					schema.Examples = []interface{}{typedValue(segs[1], schema.Type)}
				}
			}
			if option == stringOption {
				schema.Type = "string"
				schema.Format = ""
			}
		}
	}

	return schema
}

func typedValue(val string, schemaType string) interface{} {
	switch schemaType {
	case "integer":
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			return n
		}
		return val
	case "number":
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return val
	case "boolean":
		if val == "true" {
			return true
		}
		if val == "false" {
			return false
		}
		return val
	default:
		return val
	}
}

// schemaOfTypeName recursively builds an openapiSchemaObject from a type name string.
// Handles nested arrays (e.g., [][]string), primitives, and struct references.
func schemaOfTypeName(typeName string) openapiSchemaObject {
	if strings.HasPrefix(typeName, "[]") {
		inner := schemaOfTypeName(typeName[2:])
		return openapiSchemaObject{
			Type:  "array",
			Items: &inner,
		}
	}

	nullable := strings.HasPrefix(typeName, "*")
	if nullable {
		typeName = typeName[1:]
	}

	kind := openapiTypes[typeName]
	if kind != reflect.Invalid {
		ftype, format, ok := primitiveSchema(kind, typeName)
		result := openapiSchemaObject{}
		if ok {
			result = openapiSchemaObject{Type: ftype, Format: format}
		} else {
			result = openapiSchemaObject{Type: kind.String()}
		}
		if nullable {
			result.Nullable = true
		}
		return result
	}

	// Not a primitive — could be struct, map, interface, etc.
	// Every map is a plain `type: object`: any explicit additionalProperties
	// schema makes Swagger UI render placeholder "additionalProp1" keys in examples.
	if strings.HasPrefix(typeName, "map[") {
		return openapiSchemaObject{Type: "object", Nullable: nullable}
	}

	cleanName := typeName
	cleanName = strings.Replace(cleanName, "{", "", -1)
	cleanName = strings.Replace(cleanName, "}", "", -1)

	switch cleanName {
	case "interface":
		return openapiSchemaObject{Type: "object"}
	case "mapstringstring":
		return openapiSchemaObject{Type: "object"}
	default:
		result := openapiSchemaObject{Ref: "#/components/schemas/" + cleanName}
		if nullable {
			result.Nullable = true
		}
		return result
	}
}

func primitiveSchema(kind reflect.Kind, t string) (ftype, format string, ok bool) {
	switch kind {
	case reflect.Int:
		return "integer", "int32", true
	case reflect.Uint:
		return "integer", "uint32", true
	case reflect.Int8:
		return "integer", "int8", true
	case reflect.Uint8:
		return "integer", "uint8", true
	case reflect.Int16:
		return "integer", "int16", true
	case reflect.Uint16:
		return "integer", "uint16", true
	case reflect.Int32:
		return "integer", "int32", true
	case reflect.Uint32:
		return "integer", "uint32", true
	case reflect.Int64:
		return "integer", "int64", true
	case reflect.Uint64:
		return "integer", "uint64", true
	case reflect.Bool:
		return "boolean", "", true
	case reflect.String:
		return "string", "", true
	case reflect.Float32:
		return "number", "float", true
	case reflect.Float64:
		return "number", "double", true
	case reflect.Slice:
		return strings.Replace(t, "[]", "", -1), "", true
	default:
		return "", "", false
	}
}

func stringToBytes(s string) (b []byte) {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func countParams(path string) uint16 {
	var n uint16
	s := stringToBytes(path)
	n += uint16(bytes.Count(s, strColon))
	return n
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func containsStr(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func parseHeader(m spec.Member, parameters openapiParametersObject) openapiParametersObject {
	tempKind := openapiTypes[strings.Replace(m.Type.Name(), "[]", "", -1)]
	ftype, format, ok := primitiveSchema(tempKind, m.Type.Name())

	schema := &openapiSchemaObject{}
	if ok {
		schema.Type = ftype
		schema.Format = format
	} else {
		schema.Type = tempKind.String()
	}

	if m.Name == "" {
		// an embedded struct contributes its own header members, not itself
		if memberDefineStruct, ok := m.Type.(spec.DefineStruct); ok {
			for _, cm := range memberDefineStruct.Members {
				if hasHeaderParameters(cm) {
					parameters = parseHeader(cm, parameters)
				}
			}
			return parameters
		}
	}

	sp := openapiParameterObject{In: "header", Schema: schema}

	for _, tag := range m.Tags() {
		if !isBindingTag(tag) {
			continue
		}

		sp.Name = tag.Name
		sp.Required = isRequiredTag(tag)
		for _, option := range tag.Options {
			if strings.HasPrefix(option, optionsOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					schema.Enum = enumValues(segs[1], schema)
				}
			}

			if strings.HasPrefix(option, rangeOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					applyRangeOption(schema, segs[1])
				}
			}

			if strings.HasPrefix(option, defaultOption) {
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					schema.Default = typedValue(segs[1], ftype)
				}
			}

			if strings.HasPrefix(option, exampleOption) {
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					sp.Example = typedValue(segs[1], ftype)
				}
			}

			if option == stringOption {
				schema.Type = "string"
				schema.Format = ""
			}
		}
	}
	sp.Description = strings.TrimLeft(m.Comment, "//")
	// OAS 3.1.2: a header parameter named Accept, Content-Type or Authorization
	// SHALL be ignored; those are described by content types and securitySchemes.
	switch http.CanonicalHeaderKey(sp.Name) {
	case "", "Accept", "Content-Type", "Authorization":
		return parameters
	}
	return append(parameters, sp)
}
