---
outline: [2, 3]
---

# Secret

In an app, it's often necessary to save some important user information, such as `passwords` and `Access Tokens` for external systems. Olares provides a unified Vault, based on Infisical, to securely store various keys.

To retrieve this information, the app only needs a simple application for API access permission. This can be done by adding a `sysData` permission to the [OlaresManifest.yaml](../package/manifest.md#sysdata) in the application chart.

```yaml
permission:
  sysData:
    - dataType: secret
      group: secret.infisical
      version: v1
      ops:
        - RetrieveSecret?workspace=your-app # Each app should define its own workspace
        - CreateSecret?workspace=your-app
        - DeleteSecret?workspace=your-app
        - UpdateSecret?workspace=your-app
        - ListSecret?workspace=your-app
```

## Call API

You can call the API in the same way you would request other providers. Use the full name of ops (including the workspace parameter) as the URI.

Please include this **header** in all requests.
```http
X-Authorization: token          # auth_token in cookie
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