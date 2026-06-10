# goctl-openapi

A plugin for [goctl](https://github.com/zeromicro/go-zero/tree/master/tools/goctl) to generate OpenAPI 3.1.0 YAML documentation from go-zero `.api` files.

### 1. Install

```shell
go install github.com/pantheon-lab/goctl-openapi@latest
```

### 2. Add to PATH

Ensure `$GOPATH/bin` is in your `PATH`.

### 3. Usage

Create an `.api` file, e.g. `user.api`:

```go
info(
 title: "type title here"
 desc: "type desc here"
 author: "type author here"
 email: "type email here"
 version: "type version here"
)

type (
 RegisterReq {
  Username string `json:"username"`
  Password string `json:"password"`
  Mobile string `json:"mobile"`
 }

 LoginReq {
  Username string `json:"username"`
  Password string `json:"password"`
 }

 UserInfoReq {
  Id string `path:"id"`
 }

 UserInfoReply {
  Name string `json:"name"`
  Age int `json:"age"`
  Birthday string `json:"birthday"`
  Description string `json:"description"`
  Tag []string `json:"tag"`
 }

 UserSearchReq {
  KeyWord string `form:"keyWord"`
 }
)

service user-api {
 @doc(
  summary: "注册"
 )
 @handler register
 post /api/user/register (RegisterReq)

 @doc(
  summary: "登录"
 )
 @handler login
 post /api/user/login (LoginReq)

 @doc(
  summary: "获取用户信息"
 )
 @handler getUserInfo
 get /api/user/:id (UserInfoReq) returns (UserInfoReply)

 @doc(
  summary: "用户搜索"
 )
 @handler searchUser
 get /api/user/search (UserSearchReq) returns (UserInfoReply)
}
```

Generate OpenAPI YAML:

```shell
goctl api plugin -plugin goctl-openapi="openapi -filename user.yaml" -api user.api -dir .
```

Specify host and base path:

```shell
goctl api plugin -plugin goctl-openapi="openapi -filename user.yaml -host 127.0.0.2 -basepath /api -schemes https,wss" -api user.api -dir .
```

View in Swagger UI:

```shell
docker run --rm -p 8083:8080 -e SWAGGER_JSON=/foo/user.yaml -v $PWD:/foo swaggerapi/swagger-ui
```

### JWT Security

go-zero groups annotated with `@jwt` produce both `bearerAuth` (http/bearer, via `Authorization` header) and `apiKey` (apiKey/header, via `x-api-key` header) security schemes. Operations accept either:

```yaml
security:
  - bearerAuth: []
  - apiKey: []
```

Generate client code (OpenAPI 3.0+ tools):

```shell
for l in go javascript php; do
  docker run --rm -v "$(pwd):/go-work" swaggerapi/swagger-codegen-cli-v3 generate \
    -i "/go-work/rest.openapi.yaml" \
    -l "$l" \
    -o "/go-work/clients/$l"
done
```
