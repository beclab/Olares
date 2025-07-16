---
outline: [2, 3]
---

# 密钥

在应用中，通常需要保存一些重要的用户信息，例如外部系统的“密码”和“访问令牌”。Olares 提供了一个统一的 Vault 安全存储各种密钥（基于 Infisical）。

应用只需要做简单的申请，即可获得接口访问权限。申请方式是在应用 Chart 的 [OlaresManifest.yaml](../package/manifest.md#sysdata) 中添加 `sysData` 权限，例如：

```yaml
permission:
  sysData:
    - dataType: secret
      group: secret.infisical
      version: v1
      ops:
        - RetrieveSecret?workspace=your-app # 每个应用申明自己独立的workspace
        - CreateSecret?workspace=your-app
        - DeleteSecret?workspace=your-app
        - UpdateSecret?workspace=your-app
        - ListSecret?workspace=your-app
```

## 调用接口

你可以像请求其他 Provider 一样调用 API。使用 ops 的全名（包括 workspace 参数）作为 URI。

调用接口时需要加入 header。
```http
X-Authorization: <cookie 中的 auth_token>
```

### RetrieveSecret
- **Request Body**
  ```json
  {
    "name": "string", // secret name
    "env": "string" // environment of secret, test | dev | staging | prod (default)
  }
  ```
- **Success Response**
  ```json
  {
    "code": http.StatusOK, // 200 is ok
    "message": "",
    "data":{
      "name": "string", // secret name
      "value": "string", // secret value
      "env": "string" // environment of secret, test | dev | staging | prod
    }
  }
  ```

### CreateSecret
- **Request Body**
  ```json
  {
    "name": "string", // secret name
    "value": "string", // secret value
    "env": "string" // environment of secret, test | dev | staging | prod (default)
  }
  ```

- **Success Response**
  ```json
  {
    "code": http.StatusOK, // 200 is ok
    "message": "",
    "data":""
  }
  ```


### DeleteSecret
- **Request Body**
  ```json
  {
    "name": "string", // secret name
    "env": "string" // environment of secret, test | dev | staging | prod (default)
  }
  ```

- **Success Response**
  ```json
  {
    "code": http.StatusOK, // 200 is ok
    "message": "",
    "data":""
  }
  ```


### UpdateSecret
- **Request Body**
  ```json
  {
    "name": "string", // secret name
    "value": "string", // secret value
    "env": "string" // environment of secret, test | dev | staging | prod (default)
  }
  ```

- **Success Response**
  ```json
  {
    "code": http.StatusOK, // 200 is ok
    "message": "",
    "data":""
  }
  ```

### ListSecret
- **Request Body**
  ```json
  {
    "env": "string" // environment of secret, test | dev | staging | prod (default)
  }
  ```

- **Success Response**
  ```json
  {
    "code": http.StatusOK, // 200 is ok
    "message": "",
    "data":{
      "name": "string", // secret name
      "value": "string", // secret value
      "env": "string" // environment of secret, test | dev | staging | prod
    }
  }
  ```