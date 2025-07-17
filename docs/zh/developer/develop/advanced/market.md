# 应用市场

开发者可利用应用市场提供的 Provider 接口，在自己开发的应用（比如，三方 应用市场扩展）中调用安装、卸载接口来安装或卸载应用。

Provider 的申请和调用方法可以详细阅读 [Service Provider](./provider.md)

应用市场提供的 Provider

| Group            | version | dataType | ops                           |
| ---------------- | ------- | -------- | ----------------------------- |
| service.appstore | v1      | app      | InstallDevApp UninstallDevApp |

## 安装接口
- **Request**
    - **URL**: <br>`http://$OS_SYSTEM_SERVER/system-server/v1alpha1/app/service.appstore/v1/InstallDevApp`

    - **Method**: `POST`

    - **Header**
        ```http
        X-Authorization: token          # cookie 中的 auth_token
        X-Access-Token: access_token    # provider 授权接口获取的 access token
        ```

    - **Body**（以 Golang struct 为例）
        ```go
        type InstallOptions struct {
            App string `json:"appName"` //必须
            RepoUrl string `json:"repoUrl"` //必须
            CfgUrl string `json:"cfgUrl"` //可选
            Version string `json:"version"` //升级时需要
            Source string `json:"source"` //必须
        }
        ```

- **请求返回**
    ```go
    type InstallationResponse struct {
        Code int `json:"code"`
        Msg string `json:"message,omitempty"`
        Data InstallationResponseData `json:"data"`
    }

    type InstallationResponseData struct {
        UID string `json:"uid"`
    }
    ```

## 卸载接口
- **Request**
    - **URL**: <br>`http://$OS_SYSTEM_SERVER/system-server/v1alpha1/app/service.appstore/v1/UninstallDevApp`

    - **Method**: `POST`

    - **Header**
        ```http
        X-Authorization: token          # cookie 中的 auth_token
        X-Access-Token: access_token    # provider 授权接口获取的 access token
        ```

    - **Body**（以 Golang struct 为例）
        ```go
        type UninstallData struct {
            Name string `json:"name"` //required
        }
        ```

- **请求返回**
    ```go
    type InstallationResponse struct {
        Code int `json:"code"`
        Msg string `json:"message,omitempty"`
        Data InstallationResponseData `json:"data"`
    }

    type InstallationResponseData struct {
        UID string `json:"uid"`
    }
    ```