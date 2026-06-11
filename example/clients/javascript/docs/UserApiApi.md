# TypeTitleHere.UserApiApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getUserInfo**](UserApiApi.md#getUserInfo) | **GET** /api/user/{id} | 获取用户信息
[**login**](UserApiApi.md#login) | **POST** /api/user/login | 登录
[**register**](UserApiApi.md#register) | **POST** /api/user/register | 注册
[**searchUser**](UserApiApi.md#searchUser) | **GET** /api/user/search | 用户搜索



## getUserInfo

> UserInfoReply getUserInfo(id)

获取用户信息

### Example

```javascript
import TypeTitleHere from 'type_title_here';

let apiInstance = new TypeTitleHere.UserApiApi();
let id = "id_example"; // String | 
apiInstance.getUserInfo(id, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**|  | 

### Return type

[**UserInfoReply**](UserInfoReply.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## login

> Object login(appId, loginReq)

登录

### Example

```javascript
import TypeTitleHere from 'type_title_here';

let apiInstance = new TypeTitleHere.UserApiApi();
let appId = "appId_example"; // String | APPID-TEST
let loginReq = new TypeTitleHere.LoginReq(); // LoginReq | 
apiInstance.login(appId, loginReq, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **appId** | **String**| APPID-TEST | 
 **loginReq** | [**LoginReq**](LoginReq.md)|  | 

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## register

> Object register(registerReq)

注册

### Example

```javascript
import TypeTitleHere from 'type_title_here';

let apiInstance = new TypeTitleHere.UserApiApi();
let registerReq = new TypeTitleHere.RegisterReq(); // RegisterReq | 注册请求结构
apiInstance.register(registerReq, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **registerReq** | [**RegisterReq**](RegisterReq.md)| 注册请求结构 | 

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## searchUser

> UserInfoReply searchUser(keyWord)

用户搜索

### Example

```javascript
import TypeTitleHere from 'type_title_here';

let apiInstance = new TypeTitleHere.UserApiApi();
let keyWord = "keyWord_example"; // String |  关键词
apiInstance.searchUser(keyWord, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **keyWord** | **String**|  关键词 | 

### Return type

[**UserInfoReply**](UserInfoReply.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

