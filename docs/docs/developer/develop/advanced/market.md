# Market

To install or uninstall apps in developing applications (for example, third-party Market extensions), developers can use the `install` and `uninstall` API provided by the **Market**.

For how to define and call provider, please refer to [Service Provider](./provider.md)

**Provider from Market**

| Group            | version | dataType | ops                           |
| ---------------- | ------- | -------- | ----------------------------- |
| service.appstore | v1      | app      | `InstallDevApp`<br>`UninstallDevApp` |

## Install
- **Request**
    - **URL**: <br>`http://$OS_SYSTEM_SERVER/system-server/v1alpha1/app/service.appstore/v1/InstallDevApp`

    - **Method**: `POST`

    - **Header**
        ```http
        X-Authorization: token          # auth_token in cookie
        X-Access-Token: access_token    # access token get from authorization Provider
        ```

    - **Body** (Golang struct as an example)
        ```go
        type InstallOptions struct {
            App string `json:"appName"` //required
            RepoUrl string `json:"repoUrl"` //required
            CfgUrl string `json:"cfgUrl"` //optional
            Version string `json:"version"` //required when upgrading
            Source string `json:"source"` //required
        }
        ```

- **Success Responses**
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

## Uninstall
- **Request**
    - **URL**: <br>`http://$OS_SYSTEM_SERVER/system-server/v1alpha1/app/service.appstore/v1/UninstallDevApp`

    - **Method**: `POST`

    - **Header**
        ```http
        X-Authorization: token          # auth_token in cookie
        X-Access-Token: access_token    # access token get from authorization Provider
        ```

    - **Body** (Golang struct as an example)
        ```go
        type UninstallData struct {
            Name string `json:"name"` //required
        }
        ```

- **Success Responses**
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