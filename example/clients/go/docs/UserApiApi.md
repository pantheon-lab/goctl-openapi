# \UserApiAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetUserInfo**](UserApiAPI.md#GetUserInfo) | **Get** /api/user/{id} | 获取用户信息
[**Login**](UserApiAPI.md#Login) | **Post** /api/user/login | 登录
[**Register**](UserApiAPI.md#Register) | **Post** /api/user/register | 注册
[**SearchUser**](UserApiAPI.md#SearchUser) | **Get** /api/user/search | 用户搜索



## GetUserInfo

> UserInfoReply GetUserInfo(ctx, id).Execute()

获取用户信息

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	id := "id_example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserApiAPI.GetUserInfo(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserApiAPI.GetUserInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUserInfo`: UserInfoReply
	fmt.Fprintf(os.Stdout, "Response from `UserApiAPI.GetUserInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**UserInfoReply**](UserInfoReply.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Login

> interface{} Login(ctx).AppId(appId).LoginReq(loginReq).Execute()

登录

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	appId := "appId_example" // string | APPID-TEST
	loginReq := *openapiclient.NewLoginReq("Password_example", "Username_example") // LoginReq | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserApiAPI.Login(context.Background()).AppId(appId).LoginReq(loginReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserApiAPI.Login``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Login`: interface{}
	fmt.Fprintf(os.Stdout, "Response from `UserApiAPI.Login`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiLoginRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **appId** | **string** | APPID-TEST | 
 **loginReq** | [**LoginReq**](LoginReq.md) |  | 

### Return type

**interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Register

> interface{} Register(ctx).RegisterReq(registerReq).Execute()

注册

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	registerReq := *openapiclient.NewRegisterReq("Mobile_example", "Password_example", "Username_example") // RegisterReq | 注册请求结构

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserApiAPI.Register(context.Background()).RegisterReq(registerReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserApiAPI.Register``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Register`: interface{}
	fmt.Fprintf(os.Stdout, "Response from `UserApiAPI.Register`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRegisterRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **registerReq** | [**RegisterReq**](RegisterReq.md) | 注册请求结构 | 

### Return type

**interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SearchUser

> UserInfoReply SearchUser(ctx).KeyWord(keyWord).Execute()

用户搜索

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	keyWord := "keyWord_example" // string |  关键词

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserApiAPI.SearchUser(context.Background()).KeyWord(keyWord).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserApiAPI.SearchUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchUser`: UserInfoReply
	fmt.Fprintf(os.Stdout, "Response from `UserApiAPI.SearchUser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSearchUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **keyWord** | **string** |  关键词 | 

### Return type

[**UserInfoReply**](UserInfoReply.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

