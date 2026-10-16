# goctl-openapi

A plugin for [goctl](https://github.com/zeromicro/go-zero/tree/master/tools/goctl) that generates [OpenAPI 3.1.2](https://spec.openapis.org/oas/v3.1.2.html) YAML from go-zero `.api` files.

- [Install](#install)
- [Usage](#usage)
- [Conversion reference](#conversion-reference)
  - [Document and servers](#document-and-servers)
  - [Routes and operations](#routes-and-operations)
  - [Where a field goes](#where-a-field-goes)
  - [Go types](#go-types)
  - [Tag options](#tag-options)
  - [Error responses with `@respdoc`](#error-responses-with-respdoc)
  - [JWT security](#jwt-security)
- [Viewing and client generation](#viewing-and-client-generation)
- [Limitations](#limitations)

## Install

```shell
go install github.com/pantheon-lab/goctl-openapi@latest
```

Make sure `$GOPATH/bin` is in your `PATH`.

## Usage

As a goctl plugin:

```shell
goctl api plugin -plugin goctl-openapi="openapi -filename user.yaml" -api user.api -dir .
```

Standalone, without goctl:

```shell
goctl-openapi openapi -api user.api -filename user.yaml -dir .
```

| Flag | Default | Meaning |
|---|---|---|
| `-filename` | `<api name>.yaml` | Output file name |
| `-host` | `localhost` | Host of the server URL |
| `-basepath` | | Path prefix of the server URL |
| `-schemes` | `http` | `http`, `https`, `ws` or `wss`; with a comma-separated list, the first one is used |
| `-api` | | `.api` file to parse (standalone mode) |
| `-dir` | directory of `-api` | Output directory (standalone mode) |

## Conversion reference

Every example below is real generator output, trimmed to the relevant part.

### Document and servers

```go
info (
	title:   "Demo API"
	desc:    "Conversion reference"
	version: "1.0.0"
)
```

with `-host api.example.com -basepath /v1 -schemes https` becomes

```yaml
openapi: 3.1.2
info:
  title: Demo API
  description: Conversion reference
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
    description: API server
```

### Routes and operations

```go
@server (
	group:  users
	prefix: /api
)
service demo-api {
	@doc (
		summary:     "Create a user"
		description: "Creates a user account"
	)
	@handler createUser
	post /users (CreateUserReq) returns (User)

	@doc "Get a user"
	@handler getUser
	get /users/:id (GetUserReq) returns (User)
}
```

```yaml
paths:
  /api/users:
    post:
      summary: Create a user
      description: Creates a user account
      operationId: createUser
      responses:
        "200":
          description: A successful response.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserReq'
      tags:
        - users
  /api/users/{id}:
    get:
      summary: Get a user
      operationId: getUser
      # ...
```

| `.api` | OpenAPI |
|---|---|
| `prefix: /api` + `/users/:id` | path `/api/users/{id}` |
| `@handler createUser` | `operationId: createUser` |
| `@doc "text"` or `@doc(summary: ...)` | `summary` |
| `@doc(description: ...)` | `description` |
| `@server(swtags: x)`, else `group: x`, else the service name | `tags: [x]` |
| `returns (User)` | `200` response with a `User` `$ref` |
| no `returns` | `200` response with `schema: {type: object}` |
| request type used by several routes | shared `components/requestBodies/<Type>` |

Supported methods are `get`, `post`, `put`, `patch` and `delete`.

### Where a field goes

A request field's tag decides its location:

| Tag | GET | POST / PUT / PATCH / DELETE |
|---|---|---|
| `path:"id"` | path parameter | path parameter |
| `header:"X-Trace-Id"` | header parameter | header parameter |
| `form:"page"` | query parameter | query parameter, or the body if the struct only has form fields (see below) |
| `json:"name"` | query parameter | JSON body |
| `json:"-"` | omitted | omitted |

The request body's content type follows from the fields:

| Request struct | Body content type |
|---|---|
| has `json` fields | `application/json`; any `form` fields stay query parameters |
| only `form` fields | `application/x-www-form-urlencoded` |
| a `form` field that looks like a file (named or tagged `file`, `files`, `upload`, `attachment`, or typed `multipart.FileHeader`) | `multipart/form-data` |

```go
GetUserReq {
	Id      string `path:"id"`
	Verbose bool   `form:"verbose,default=false"`
	Page    int    `form:"page,range=[1:100]"`
}
```

```yaml
parameters:
  - name: id
    in: path
    required: true
    schema:
      type: string
  - name: verbose
    in: query
    required: false
    schema:
      type: boolean
      default: false
  - name: page
    in: query
    required: true
    schema:
      type: integer
      format: int32
      minimum: 1
      maximum: 100
```

```go
LoginReq {
	Username string `form:"username"`
	Password string `form:"password"`
}

UploadReq {
	File  string `form:"file"`
	Title string `form:"title,optional"`
}
```

```yaml
/login:
  post:
    requestBody:
      required: true
      content:
        application/x-www-form-urlencoded:
          schema:
            $ref: '#/components/schemas/LoginReq'
/upload:
  post:
    requestBody:
      required: true
      content:
        multipart/form-data:
          schema:
            $ref: '#/components/schemas/UploadReq'
```

Fields of an embedded struct are treated as if declared in the outer struct:

```go
Base {
	TraceId string `header:"X-Trace-Id,optional"`
}
CreateUserReq {
	Base
	Name string `json:"name"`
}
```

```yaml
parameters:
  - name: X-Trace-Id
    in: header
    required: false
    schema:
      type: string
```

Path parameters are always `required: true`, even when tagged `optional`, because OpenAPI 3.1.2 requires it. Header parameters named `Accept`, `Content-Type` or `Authorization` are left out, because the spec says tools must ignore them. Document authentication with [`jwt`](#jwt-security) instead.

### Go types

| Go type | Schema |
|---|---|
| `string` | `type: string` |
| `bool` | `type: boolean` |
| `int`, `int32` | `type: integer, format: int32` |
| `int64`, `uint64` | `type: integer, format: int64` |
| `int8`, `int16`, `uint8`, `uint16`, `uint32` | `type: integer, format: <same name>` |
| `uint` | `type: integer, format: uint32` |
| `float32` | `type: number, format: float` |
| `float64` | `type: number, format: double` |
| `[]T` | `type: array, items: <T>` |
| `map[K]V`, `interface{}` | `type: object` |
| `Struct` | `$ref: '#/components/schemas/Struct'` |
| `*T` (primitive) | `type: [T, "null"]` |
| `*Struct` | `anyOf: [{$ref: Struct}, {type: "null"}]` |

```go
Bio  *string           `json:"bio,optional"`
Home *Address          `json:"home,optional"` // home address
Tags []string          `json:"tags"`
Meta map[string]string `json:"meta,optional"`
```

```yaml
bio:
  type:
    - string
    - "null"
home:
  anyOf:
    - $ref: '#/components/schemas/Address'
    - type: "null"
  description: home address
tags:
  type: array
  items:
    type: string
meta:
  type: object
```

A trailing `// comment` on a field becomes its `description`.

### Tag options

go-zero tag options map to JSON Schema keywords:

| Option | Example | Schema |
|---|---|---|
| none | `json:"name"` | listed in `required` |
| `optional` / `omitempty` | `json:"nickname,optional"` | not in `required` |
| `default=` | `json:"score,default=0.5"` | `default: 0.5`; not in `required` |
| `options=` | `json:"level,options=1\|2\|3"` | `enum: [1, 2, 3]`, typed like the field |
| `range=[a:b]` | `json:"age,range=[0:150]"` | `minimum: 0`, `maximum: 150` |
| `range=(a:b)` | `json:"score,range=(0:1]"` | `exclusiveMinimum: 0`, `maximum: 1` |
| `example=` | `json:"role,example=member"` | `examples: [member]` |
| `string` | `json:"id64,string"` | `type: string` |

In `range`, `[`/`]` make a bound inclusive and `(`/`)` make it exclusive. Either bound may be left out, e.g. `range=[1:]`.

```go
CreateUserReq {
	Name     string  `json:"name"`
	Nickname string  `json:"nickname,optional"`
	Age      int     `json:"age,range=[0:150]"`
	Score    float64 `json:"score,range=(0:1],default=0.5"`
	Role     string  `json:"role,options=admin|member,example=member"`
	Level    int     `json:"level,options=1|2|3"`
	Id64     int64   `json:"id64,string"`
	Secret   string  `json:"-"`
}
```

```yaml
CreateUserReq:
  type: object
  title: CreateUserReq
  properties:
    age:
      type: integer
      format: int32
      minimum: 0
      maximum: 150
    id64:
      type: string
    level:
      type: integer
      format: int32
      enum:
        - 1
        - 2
        - 3
    name:
      type: string
    nickname:
      type: string
    role:
      type: string
      enum:
        - admin
        - member
      examples:
        - member
    score:
      type: number
      format: double
      default: 0.5
      maximum: 1
      exclusiveMinimum: 0
  required:
    - name
    - age
    - role
    - level
    - id64
```

On a nullable field, `options=` also allows `null`: `*string` with `options=a|b` gives `enum: [a, b, null]`.

### Error responses with `@respdoc`

Add more responses with a `@respdoc-<status>` comment, written between `@handler` and the route (or above `@doc`/`@handler`). The text after the closing `)` becomes the response description. Without that text, the HTTP status text is used.

```go
@handler getUser
/*
@respdoc-404 (
	100404: user not found
) // Not found
*/
/* @respdoc-500 (ErrorResponse) // Server error */
get /users/:id (GetUserReq) returns (User)
```

- `(code: message` lines`)` becomes an example map of error codes. Put each code on its own line.
- `(TypeName)` becomes a response with that type's schema.

```yaml
responses:
  "404":
    description: Not found
    content:
      application/json:
        schema:
          type: object
        example:
          "100404": user not found
  "500":
    description: Server error
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/ErrorResponse'
```

### JWT security

A group with `jwt` in its `@server` block gets two security schemes, and its operations accept either one:

```go
@server (
	jwt: Auth
)
```

```yaml
paths:
  /api/users:
    post:
      security:
        - bearerAuth: []
        - apiKey: []
components:
  securitySchemes:
    apiKey:
      type: apiKey
      description: API key authentication via x-api-key header
      name: x-api-key
      in: header
    bearerAuth:
      type: http
      description: Enter JWT Bearer token
      scheme: bearer
      bearerFormat: JWT
```

## Viewing and client generation

View in Swagger UI:

```shell
docker run --rm -p 8083:8080 -e SWAGGER_JSON=/foo/user.yaml -v $PWD:/foo swaggerapi/swagger-ui
```

Generate clients with [OpenAPI Generator](https://github.com/OpenAPITools/openapi-generator), as [`example/generate-client.sh`](example/generate-client.sh) does:

```shell
for l in go javascript php; do
  docker run --rm -v "$(pwd):/local" openapitools/openapi-generator-cli generate \
    -i "/local/user.yaml" \
    -g "$l" \
    -o "/local/clients/$l"
done
```

## Limitations

- Every type in the `.api` file appears under `components/schemas`, except request types that only produce parameters and that nothing references.
- File uploads are found by field name or tag (`file`, `upload`, ...), because the `.api` language has no file type. The field's schema is a plain `string`.
- In a `@respdoc` error-code list, each `code: message` pair must be on its own line.
- `operationId` comes from the handler name, so handler names must be unique across the file and its imports.
