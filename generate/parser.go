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
	"unsafe"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

var strColon = []byte(":")

const (
	validateKey     = "validate"
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

func parseRangeOption(option string) (float64, float64, bool) {
	const str = "\\[([+-]?\\d+(\\.\\d+)?):([+-]?\\d+(\\.\\d+)?)\\]"
	result := regexp.MustCompile(str).FindStringSubmatch(option)
	if len(result) != 5 {
		return 0, 0, false
	}

	min, err := strconv.ParseFloat(result[1], 64)
	if err != nil {
		return 0, 0, false
	}

	max, err := strconv.ParseFloat(result[3], 64)
	if err != nil {
		return 0, 0, false
	}

	if max < min {
		return min, min, true
	}
	return min, max, true
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
		OpenAPI: "3.1.0",
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
	renderServiceRoutes(p.Api.Service, p.Api.Service.Groups, o.Paths, requestResponseRefs, &o.Components)

	renderReplyAsDefinition(o.Components.Schemas, p.Api.Types, requestResponseRefs)

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

func renderServiceRoutes(service spec.Service, groups []spec.Group, paths openapiPathsObject, requestResponseRefs refMap, components *openapiComponentsObject) {
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

						// Look up type and optional from request struct member
						if defineStruct, ok := route.RequestType.(spec.DefineStruct); ok {
							for _, member := range defineStruct.Members {
								memberKey := ""
								for _, tag := range member.Tags() {
									if tag.Key == "path" {
										memberKey = tag.Name
										for _, option := range tag.Options {
											if option == optionalOption || option == omitemptyOption {
												spo.Required = false
											}
										}
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
					hasBody = true
					// Check for form members — only for non-GET
					for _, member := range defineStruct.Members {
						if member.IsFormMember() {
							hasBody = true
							break
						}
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
				mediaType := openapiMediaTypeObject{}
				if strings.HasPrefix(route.ResponseType.Name(), "[]") {
					refTypeName := strings.Replace(route.ResponseType.Name(), "[", "", 1)
					refTypeName = strings.Replace(refTypeName, "]", "", 1)

					mediaType.Schema = &openapiSchemaObject{
						Type: "array",
						Items: &openapiSchemaObject{
							Ref: "#/components/schemas/" + refTypeName,
						},
					}
				} else {
					mediaType.Schema = &openapiSchemaObject{
						Ref: "#/components/schemas/" + route.ResponseType.Name(),
					}
				}
				operationObject.Responses["200"].Content["application/json"] = mediaType

				if strings.HasPrefix(route.ResponseType.Name(), "[]") {
					refTypeName := strings.Replace(route.ResponseType.Name(), "[", "", 1)
					refTypeName = strings.Replace(refTypeName, "]", "", 1)
					requestResponseRefs["#/components/schemas/"+refTypeName] = struct{}{}
				} else {
					requestResponseRefs["#/components/schemas/"+route.ResponseType.Name()] = struct{}{}
				}
			} else {
				// Empty response — explicit schema:{} for tooling compatibility
				operationObject.Responses["200"].Content["application/json"] = openapiMediaTypeObject{
					Schema: &openapiSchemaObject{},
				}
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
						operationObject.Responses[code] = openapiResponseObject{
							Description: comment,
							Content: map[string]openapiMediaTypeObject{
								"application/json": {
									Schema: &openapiSchemaObject{
										Example: string(kvByte),
									},
								},
							},
						}
					} else if len(content) > 0 {
						operationObject.Responses[code] = openapiResponseObject{
							Description: comment,
							Content: map[string]openapiMediaTypeObject{
								"application/json": {
									Schema: &openapiSchemaObject{
										Ref: "#/components/schemas/" + content,
									},
								},
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

func buildRequestBody(route spec.Route, typeName string) *openapiRequestBodyObject {
	reqRef := "#/components/schemas/" + typeName

	body := &openapiRequestBodyObject{
		Required: true,
		Content: map[string]openapiMediaTypeObject{
			"application/json": {
				Schema: &openapiSchemaObject{
					Ref: reqRef,
				},
			},
		},
	}

	if defineStruct, ok := route.RequestType.(spec.DefineStruct); ok {
		for _, member := range defineStruct.Members {
			if member.IsFormMember() {
				body.Content["multipart/form-data"] = openapiMediaTypeObject{
					Schema: &openapiSchemaObject{
						Ref: reqRef,
					},
				}
				break
			}
		}
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

func renderStruct(member spec.Member) openapiParameterObject {
	tempKind := openapiTypes[strings.Replace(member.Type.Name(), "[]", "", -1)]

	ftype, format, ok := primitiveSchema(tempKind, member.Type.Name())
	schema := &openapiSchemaObject{}
	if ok {
		schema.Type = ftype
		schema.Format = format
	} else {
		schema.Type = tempKind.String()
	}
	sp := openapiParameterObject{In: "query", Schema: schema}

	for _, tag := range member.Tags() {
		if tag.Key == validateKey {
			continue
		}

		sp.Name = tag.Name
		if len(tag.Options) == 0 {
			sp.Required = true
			continue
		}

		required := true
		for _, option := range tag.Options {
			if strings.HasPrefix(option, optionsOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					schema.Enum = strings.Split(segs[1], optionSeparator)
				}
			}

			if strings.HasPrefix(option, rangeOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					min, max, ok := parseRangeOption(segs[1])
					if ok {
						schema.Minimum = min
						schema.Maximum = max
					}
				}
			}

			if strings.HasPrefix(option, defaultOption) {
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					schema.Default = typedValue(segs[1], ftype)
				}
			} else if strings.HasPrefix(option, optionalOption) || strings.HasPrefix(option, omitemptyOption) {
				required = false
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
		sp.Required = required
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

		for _, member := range defineStruct.Members {
			if hasPathParameters(member) || hasHeaderParameters(member) {
				continue
			}
			propName := member.Name
			if tag, err := member.GetPropertyName(); err == nil {
				propName = tag
			}
			if propName == "" {
				memberStruct, _ := member.Type.(spec.DefineStruct)
				for _, m := range memberStruct.Members {
					if hasHeaderParameters(m) || hasPathParameters(m) {
						continue
					}
					subName := m.Name
					if tag, err := m.GetPropertyName(); err == nil {
						subName = tag
					}
					if schema.Properties == nil {
						schema.Properties = make(map[string]openapiSchemaObject)
					}
					schema.Properties[subName] = schemaOfField(m)

					for _, tag := range m.Tags() {
						if tag.Key == validateKey {
							continue
						}
						if len(tag.Options) == 0 {
							if !containsStr(schema.Required, tag.Name) && tag.Name != "required" {
								schema.Required = append(schema.Required, tag.Name)
							}
							continue
						}
						required := true
						for _, option := range tag.Options {
							if strings.HasPrefix(option, optionalOption) || strings.HasPrefix(option, omitemptyOption) {
								required = false
							}
						}
						if required && !containsStr(schema.Required, tag.Name) {
							schema.Required = append(schema.Required, tag.Name)
						}
					}
				}
				continue
			}
			if schema.Properties == nil {
				schema.Properties = make(map[string]openapiSchemaObject)
			}
			schema.Properties[propName] = schemaOfField(member)

			for _, tag := range member.Tags() {
				if tag.Key == validateKey {
					continue
				}
				if len(tag.Options) == 0 {
					if !containsStr(schema.Required, tag.Name) && tag.Name != "required" {
						schema.Required = append(schema.Required, tag.Name)
					}
					continue
				}

				required := true
				for _, option := range tag.Options {
					if strings.HasPrefix(option, optionalOption) || strings.HasPrefix(option, omitemptyOption) {
						required = false
					}
				}

				if required && !containsStr(schema.Required, tag.Name) {
					schema.Required = append(schema.Required, tag.Name)
				}
			}
		}

		d[i2.Name()] = schema
	}
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
					schema.Enum = strings.Split(segs[1], optionSeparator)
				}
			case strings.HasPrefix(option, rangeOption):
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					min, max, ok := parseRangeOption(segs[1])
					if ok {
						schema.Minimum = min
						schema.Maximum = max
					}
				}
			case strings.HasPrefix(option, exampleOption):
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					schema.Example = typedValue(segs[1], schema.Type)
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
	if strings.HasPrefix(typeName, "map[") {
		// map[K]V pattern — extract value type
		bracketIdx := strings.Index(typeName, "]")
		if bracketIdx > 0 {
			valueType := typeName[bracketIdx+1:]
			valueType = strings.TrimPrefix(valueType, "*")
			valueType = strings.TrimPrefix(valueType, "interface{}")
			if valueType == "" || valueType == "interface{}" {
				return openapiSchemaObject{Type: "object"}
			}
			// Handle maps of specific types like map[string]string
			if valueType == "string" {
				return openapiSchemaObject{
					Type:                 "object",
					AdditionalProperties: &openapiSchemaObject{Type: "string"},
				}
			}
			// For other map value types, reference the schema
			return openapiSchemaObject{
				Type:                 "object",
				AdditionalProperties: &openapiSchemaObject{Ref: "#/components/schemas/" + valueType},
			}
		}
	}

	cleanName := typeName
	cleanName = strings.Replace(cleanName, "{", "", -1)
	cleanName = strings.Replace(cleanName, "}", "", -1)

	switch cleanName {
	case "interface":
		return openapiSchemaObject{Type: "object"}
	case "mapstringstring":
		return openapiSchemaObject{
			Type:                 "object",
			AdditionalProperties: &openapiSchemaObject{Type: "string"},
		}
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

	sp := openapiParameterObject{In: "header", Schema: schema}

	for _, tag := range m.Tags() {
		sp.Name = tag.Name
		if len(tag.Options) == 0 {
			sp.Required = true
			continue
		}

		required := true
		for _, option := range tag.Options {
			if strings.HasPrefix(option, optionsOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					schema.Enum = strings.Split(segs[1], optionSeparator)
				}
			}

			if strings.HasPrefix(option, rangeOption) {
				segs := strings.SplitN(option, equalToken, 2)
				if len(segs) == 2 {
					min, max, ok := parseRangeOption(segs[1])
					if ok {
						schema.Minimum = min
						schema.Maximum = max
					}
				}
			}

			if strings.HasPrefix(option, defaultOption) {
				segs := strings.Split(option, equalToken)
				if len(segs) == 2 {
					schema.Default = typedValue(segs[1], ftype)
				}
			} else if strings.HasPrefix(option, optionalOption) || strings.HasPrefix(option, omitemptyOption) {
				required = false
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
		sp.Required = required
	}
	sp.Description = strings.TrimLeft(m.Comment, "//")
	if m.Name == "" {
		memberDefineStruct, ok := m.Type.(spec.DefineStruct)
		if !ok {
			return parameters
		}
		for _, cm := range memberDefineStruct.Members {
			if hasHeaderParameters(cm) {
				parameters = parseHeader(cm, parameters)
			}
		}
	}
	return append(parameters, sp)
}
