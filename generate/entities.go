package generate

import (
	"encoding/json"
	"reflect"
)

var openapiTypes = map[string]reflect.Kind{
	"string":   reflect.String,
	"*string":  reflect.String,
	"int":      reflect.Int,
	"*int":     reflect.Int,
	"uint":     reflect.Uint,
	"*uint":    reflect.Uint,
	"int8":     reflect.Int8,
	"*int8":    reflect.Int8,
	"uint8":    reflect.Uint8,
	"*uint8":   reflect.Uint8,
	"int16":    reflect.Int16,
	"*int16":   reflect.Int16,
	"uint16":   reflect.Uint16,
	"*uint16":  reflect.Uint16,
	"int32":    reflect.Int,
	"*int32":   reflect.Int,
	"uint32":   reflect.Uint32,
	"*uint32":  reflect.Uint32,
	"uint64":   reflect.Int64,
	"*uint64":  reflect.Int64,
	"int64":    reflect.Int64,
	"*int64":   reflect.Int64,
	"[]string": reflect.Slice,
	"[]int":    reflect.Slice,
	"[]int64":  reflect.Slice,
	"[]int32":  reflect.Slice,
	"[]uint32": reflect.Slice,
	"[]uint64": reflect.Slice,
	"bool":     reflect.Bool,
	"*bool":    reflect.Bool,
	"struct":   reflect.Struct,
	"*struct":  reflect.Struct,
	"float32":  reflect.Float32,
	"*float32": reflect.Float32,
	"float64":  reflect.Float64,
	"*float64": reflect.Float64,
}

type openapiInfoObject struct {
	Title          string                `json:"title" yaml:"title"`
	Description    string                `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string                `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Version        string                `json:"version" yaml:"version"`
	Contact        *openapiContactObject `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *openapiLicenseObject `json:"license,omitempty" yaml:"license,omitempty"`
}

type openapiContactObject struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type openapiLicenseObject struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

type openapiServerObject struct {
	URL         string                                 `json:"url" yaml:"url"`
	Description string                                 `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]openapiServerVariableObject `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type openapiServerVariableObject struct {
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
}

type openapiComponentsObject struct {
	Schemas         openapiSchemasObject         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes openapiSecuritySchemesObject `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	RequestBodies   openapiRequestBodiesObject   `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
}

type openapiObject struct {
	OpenAPI      string                              `json:"openapi" yaml:"openapi"`
	Info         openapiInfoObject                   `json:"info" yaml:"info"`
	Servers      []openapiServerObject               `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths        openapiPathsObject                  `json:"paths" yaml:"paths"`
	Components   openapiComponentsObject             `json:"components,omitempty" yaml:"components,omitempty"`
	Security     []openapiSecurityRequirementObject  `json:"security,omitempty" yaml:"security,omitempty"`
	ExternalDocs *openapiExternalDocumentationObject `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

type openapiPathsObject map[string]openapiPathItemObject

type openapiPathItemObject struct {
	Get    *openapiOperationObject `json:"get,omitempty" yaml:"get,omitempty"`
	Delete *openapiOperationObject `json:"delete,omitempty" yaml:"delete,omitempty"`
	Post   *openapiOperationObject `json:"post,omitempty" yaml:"post,omitempty"`
	Put    *openapiOperationObject `json:"put,omitempty" yaml:"put,omitempty"`
	Patch  *openapiOperationObject `json:"patch,omitempty" yaml:"patch,omitempty"`
}

type openapiOperationObject struct {
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID  string                              `json:"operationId" yaml:"operationId"`
	Responses    openapiResponsesObject              `json:"responses" yaml:"responses"`
	Parameters   openapiParametersObject             `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *openapiRequestBodyObject           `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Tags         []string                            `json:"tags,omitempty" yaml:"tags,omitempty"`
	Deprecated   bool                                `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security     *[]openapiSecurityRequirementObject `json:"security,omitempty" yaml:"security,omitempty"`
	ExternalDocs *openapiExternalDocumentationObject `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

type openapiParametersObject []openapiParameterObject

type openapiParameterObject struct {
	Name        string                          `json:"name" yaml:"name"`
	Description string                          `json:"description,omitempty" yaml:"description,omitempty"`
	In          string                          `json:"in" yaml:"in"`
	Required    bool                            `json:"required" yaml:"required"`
	Schema      *openapiSchemaObject            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example     interface{}                     `json:"example,omitempty" yaml:"example,omitempty"`
	Examples    map[string]openapiExampleObject `json:"examples,omitempty" yaml:"examples,omitempty"`
	Deprecated  bool                            `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

type openapiExampleObject struct {
	Summary string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Value   interface{} `json:"value,omitempty" yaml:"value,omitempty"`
}

type openapiRequestBodyObject struct {
	Description string                            `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                              `json:"required" yaml:"required"`
	Content     map[string]openapiMediaTypeObject `json:"content,omitempty" yaml:"content,omitempty"`
	Ref         string                            `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type openapiRequestBodiesObject map[string]openapiRequestBodyObject

type openapiMediaTypeObject struct {
	Schema   *openapiSchemaObject            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  interface{}                     `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]openapiExampleObject `json:"examples,omitempty" yaml:"examples,omitempty"`
}

type openapiResponsesObject map[string]openapiResponseObject

type openapiResponseObject struct {
	Description string                            `json:"description" yaml:"description"`
	Content     map[string]openapiMediaTypeObject `json:"content,omitempty" yaml:"content,omitempty"`
}

type openapiSchemasObject map[string]openapiSchemaObject

type openapiSchemaObject struct {
	Type                 string                            `json:"type,omitempty" yaml:"-"` // marshaled by MarshalYAML
	Format               string                            `json:"format,omitempty" yaml:"format,omitempty"`
	Ref                  string                            `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Description          string                            `json:"description,omitempty" yaml:"description,omitempty"`
	Title                string                            `json:"title,omitempty" yaml:"title,omitempty"`
	Properties           map[string]openapiSchemaObject    `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *openapiAdditionalPropertiesValue `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Items                *openapiSchemaObject              `json:"items,omitempty" yaml:"items,omitempty"`
	Required             []string                          `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []string                          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default              interface{}                       `json:"default,omitempty" yaml:"default,omitempty"`
	Example              interface{}                       `json:"example,omitempty" yaml:"example,omitempty"`
	Nullable             bool                              `json:"nullable,omitempty" yaml:"-"` // marshaled by MarshalYAML
	ReadOnly             bool                              `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool                              `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Minimum              float64                           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              float64                           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMinimum     bool                              `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     bool                              `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	MinLength            uint64                            `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength            uint64                            `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern              string                            `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinItems             uint64                            `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems             uint64                            `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	UniqueItems          bool                              `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MinProperties        uint64                            `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	MaxProperties        uint64                            `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MultipleOf           float64                           `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
}

// openapiAdditionalPropertiesValue represents the additionalProperties keyword,
// which in OpenAPI/JSON Schema can be either a boolean or a nested schema.
type openapiAdditionalPropertiesValue struct {
	Allowed *bool
	Schema  *openapiSchemaObject
}

func schemaAdditionalProperties(s openapiSchemaObject) *openapiAdditionalPropertiesValue {
	return &openapiAdditionalPropertiesValue{Schema: &s}
}

func (a openapiAdditionalPropertiesValue) MarshalYAML() (interface{}, error) {
	if a.Schema != nil {
		return *a.Schema, nil
	}
	if a.Allowed != nil {
		return *a.Allowed, nil
	}
	return nil, nil
}

func (a openapiAdditionalPropertiesValue) MarshalJSON() ([]byte, error) {
	if a.Schema != nil {
		return json.Marshal(a.Schema)
	}
	if a.Allowed != nil {
		return json.Marshal(*a.Allowed)
	}
	return []byte("null"), nil
}

type openapiSecuritySchemesObject map[string]openapiSecuritySchemeObject

type openapiSecuritySchemeObject struct {
	Type             string              `json:"type" yaml:"type"`
	Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string              `json:"name,omitempty" yaml:"name,omitempty"`
	In               string              `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string              `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string              `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flow             string              `json:"flow,omitempty" yaml:"flow,omitempty"`
	AuthorizationURL string              `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string              `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	Scopes           openapiScopesObject `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

type openapiScopesObject map[string]string

type openapiSecurityRequirementObject map[string][]string

type openapiExternalDocumentationObject struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
}

// MarshalYAML renders nullability the OpenAPI 3.1 way: `nullable` was removed
// in 3.1, so a nullable type becomes `type: [T, "null"]` and a nullable $ref
// becomes `anyOf: [{$ref}, {type: "null"}]`.
func (s openapiSchemaObject) MarshalYAML() (interface{}, error) {
	type refSchema struct {
		Ref string `yaml:"$ref"`
	}
	if s.Ref != "" {
		if !s.Nullable {
			return refSchema{Ref: s.Ref}, nil
		}
		return map[string][]interface{}{
			"anyOf": {refSchema{Ref: s.Ref}, map[string]string{"type": "null"}},
		}, nil
	}

	type alias openapiSchemaObject
	var typ interface{}
	if s.Type != "" {
		typ = s.Type
		if s.Nullable {
			typ = []string{s.Type, "null"}
		}
	}
	return struct {
		Type  interface{} `yaml:"type,omitempty"`
		alias `yaml:",inline"`
	}{typ, alias(s)}, nil
}

type refMap map[string]struct{}

func (r openapiRequestBodyObject) MarshalYAML() (interface{}, error) {
	if r.Ref != "" {
		type refBody struct {
			Ref string `yaml:"$ref"`
		}
		return refBody{Ref: r.Ref}, nil
	}
	type alias openapiRequestBodyObject
	return alias(r), nil
}
