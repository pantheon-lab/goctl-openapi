# OpenAPI\Client\UserApiApi



All URIs are relative to http://localhost, except if the operation defines another base path.

| Method | HTTP request | Description |
| ------------- | ------------- | ------------- |
| [**getUserInfo()**](UserApiApi.md#getUserInfo) | **GET** /api/user/{id} | 获取用户信息 |
| [**login()**](UserApiApi.md#login) | **POST** /api/user/login | 登录 |
| [**register()**](UserApiApi.md#register) | **POST** /api/user/register | 注册 |
| [**searchUser()**](UserApiApi.md#searchUser) | **GET** /api/user/search | 用户搜索 |


## `getUserInfo()`

```php
getUserInfo($id): \OpenAPI\Client\Model\UserInfoReply
```

获取用户信息

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');



$apiInstance = new OpenAPI\Client\Api\UserApiApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client()
);
$id = 'id_example'; // string

try {
    $result = $apiInstance->getUserInfo($id);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling UserApiApi->getUserInfo: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **id** | **string**|  | |

### Return type

[**\OpenAPI\Client\Model\UserInfoReply**](../Model/UserInfoReply.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `login()`

```php
login($app_id, $login_req): mixed
```

登录

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');



$apiInstance = new OpenAPI\Client\Api\UserApiApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client()
);
$app_id = 'app_id_example'; // string | APPID-TEST
$login_req = new \OpenAPI\Client\Model\LoginReq(); // \OpenAPI\Client\Model\LoginReq

try {
    $result = $apiInstance->login($app_id, $login_req);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling UserApiApi->login: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **app_id** | **string**| APPID-TEST | |
| **login_req** | [**\OpenAPI\Client\Model\LoginReq**](../Model/LoginReq.md)|  | |

### Return type

**mixed**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `register()`

```php
register($register_req): mixed
```

注册

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');



$apiInstance = new OpenAPI\Client\Api\UserApiApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client()
);
$register_req = new \OpenAPI\Client\Model\RegisterReq(); // \OpenAPI\Client\Model\RegisterReq | 注册请求结构

try {
    $result = $apiInstance->register($register_req);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling UserApiApi->register: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **register_req** | [**\OpenAPI\Client\Model\RegisterReq**](../Model/RegisterReq.md)| 注册请求结构 | |

### Return type

**mixed**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `searchUser()`

```php
searchUser($key_word): \OpenAPI\Client\Model\UserInfoReply
```

用户搜索

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');



$apiInstance = new OpenAPI\Client\Api\UserApiApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client()
);
$key_word = 'key_word_example'; // string | 关键词

try {
    $result = $apiInstance->searchUser($key_word);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling UserApiApi->searchUser: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **key_word** | **string**| 关键词 | |

### Return type

[**\OpenAPI\Client\Model\UserInfoReply**](../Model/UserInfoReply.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)
