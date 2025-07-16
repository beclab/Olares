# Service Provider

APP 的开发者可在应用的 Chart 中配置 `ProviderRegistry` 或者请求调用其他 Provider 的权限。

## 申明 Provider

```yaml
apiVersion: sys.bytetrade.io/v1alpha1
kind: ProviderRegistry
metadata:
  # provider 名称，需要加入namespace，避免重复
  name: provider-{{ .Release.Namespace }}

  # provider registry需要安装到user-system下面
  namespace: user-system-{{ .Values.bfl.username}}
spec:
  version: v2   # 最新版本是 v2，系统同时兼容 v1 版本

  # provider 的 dataType。建议加上 app name 避免重复
  dataType: legacy_{{ .Release.Name }}
  deployment: {{ .Release.Name }}
  description: {{ .Release.Name }} legacy api v2

  # provider 可访问的服务。格式通常为 <appServiceName>.<appNameSpace>:<servicePort>
  endpoint: {{ .Release.Name }}-svc.{{ .Release.Namespace }}:1234

  # provider 的组名。建议加上 app name 避免重复
  group: api.{{ .Release.Name }}
  kind: provider
  namespace: "{{ .Release.Namespace }}"
  opApis:
    # API 的名称
    - name: AppApi
      # API 的 URL
      uri: /api  
status:
  state: active
```

## 申请 Provider 的访问权限

可在 [OlaresManifest.yaml](../package/manifest.md#sysdata) 中配置：

```Yaml
sysData:
- appName: providerapp  # API provider 的 app name。ProviderRegistry v2 版本必填 
  port: 8888  # provider service 的端口号

  # provider 的默认域名格式为 <appName>-svc.<appName>-<username>:<port>。如果 service name 和 app namespace 不是默认格式，可以在以下字段中指定 
  svc: app-svc  # service 名称。ProviderRegistry v2 版本可选
  namespace: ns # app 的 namespace。ProviderRegistry v2版本可选

  version: v2   # ProviderRegistry 的版本
  dataType: legacy_{{ .Release.Name }}  # ProviderRegistry 中定义的 dataType
  group: api.{{ .Release.Name }}   # ProviderRegistry 中定义的组名
  ops:
  - AppApi   # ProviderRegistry 中定义的 opApis 名称
```

配置完成后，你可以将访问密钥(`access key`)和访问密钥(`access secret`)添加到应用 chart 的模板中。它们将在安装过程中被注入以供授权使用。

```yaml
env:
  - name: OS_SYSTEM_SERVER
    value: system-server.user-system-{{ .Values.bfl.username }}
  - name: OS_APP_SECRET

    # 应用名称在应用 chart 中定义
    value: "{{ .Values.os.<appnane>.appSecret }}"
  - name: OS_APP_KEY
    value: "{{ .Values.os.<appname>.appKey }}"
```

你可以在代码中使用这三个环境变量来调用 Provider。以 curl 为例：

1. 获取 access token，有效时间 5 分钟。token 加密算法：bcrypt(`app key` `timestamp` `app secret`) 。默认成本值为 10。

    ```sh
    now=$(date +%s)
    token=$(htpasswd -nbBC 10 USER "${OS_APP_KEY}${now}${OS_APP_SECRET}"|awk -F":" '{print $2}')
    
    curl -X POST http://${OS_SYSTEM_SERVER}/permission/v1alpha1/access -H "content-type: application/json" \
      -d "{ \
      \"app_key\": \"${OS_APP_KEY}\",         \
      \"timestamp\": ${now},                  \
      \"token\": \"${token}\",                \
      \"perm\": {                             \
          \"group\": \"service.bfl\",         \
          \"dataType\": \"app\",              \
          \"version\": \"v1\",                \
          \"ops\": [                          \
          \"InstallDevApp\"                   \
          ]                                   \
      }                                       \
    }'
    ```

2. 系统将返回：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "access_token": "JDJ5JDEwJE5Wbk9vbFpoLjJlSGxhUUpRY1IwRmVZVjFBWmUxUi5LOXNuQWJmVjRnN29xNWVVaFhPWmV5"
      }
    }
    ```

3. 你可以用返回的 token 去调用 provider 的接口：

    ```sh
    # 地址格式 http://${OS_SYSTEM_SERVER}/system-server/v1alpha1/<dataType>/<group>/<version>/<op>
    curl http://${OS_SYSTEM_SERVER}/system-server/v1alpha1/app/service.bfl/v1/InstallDevApp \
      -H "content-type: application/json" \
      -H "X-Access-Token: ${access_token}"  \
      -d '{"data":"post to provider"}'
    ```
