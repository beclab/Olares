# Service Provider

App developers can define the `ProviderRegistry` in the application chart or request permission to call other providers.

## Define Provider

```yaml
apiVersion: sys.bytetrade.io/v1alpha1
kind: ProviderRegistry
metadata:
  # Provider name. A namespace is required to prevent duplication.
  name: provider-{{ .Release.Namespace }}

  # provider registry needs to be installed under user-system
  namespace: user-system-{{ .Values.bfl.username}}
spec:
  version: v2   #The latest version is v2, but the system remains compatible with v1.

  # dataType of provider, it is recommended to add app name to prevent duplication.
  dataType: legacy_{{ .Release.Name }}
  deployment: {{ .Release.Name }}
  description: {{ .Release.Name }} legacy api v2

  # accessible service from the provider. Usually it is <appServiceName>.<appNameSpace>:<servicePort>
  endpoint: {{ .Release.Name }}-svc.{{ .Release.Namespace }}:1234

  # group of the provider; it is recommended to add the app name to prevent duplication.
  group: api.{{ .Release.Name }}
  kind: provider
  namespace: "{{ .Release.Namespace }}"
  opApis:
    # name of the provided API
    - name: AppApi
      # URL of the API
      uri: /api  
status:
  state: active
```

## Request Permission to Call Provider

You can configure it in the [OlaresManifest.yaml](../package/manifest.md#sysdata) as follows:

```Yaml
sysData:
- appName: providerapp  # The appname of the api provider. Required for ProviderRegistry v2. 
  port: 8888  # The port of the provider service

  # The default domain of provider is <appName>-svc.<appName>-<username>:<port>, if the service name and app namespace is not in default format, you can specify it in following field  
  svc: app-svc  # Name of the service. Optional for ProviderRegistry v2.
  namespace: ns # Namespace of the app. Optional for ProviderRegistry v2.

  version: v2   # version of the ProviderRegistry
  dataType: legacy_{{ .Release.Name }}  # dataType defined in ProviderRegistry
  group: api.{{ .Release.Name }}   # group defined in ProviderRegistry
  ops:
  - AppApi   # name of opApis defined in ProviderRegistry
```

Once configured, you can add the `access key` and `access secret` to the templates in the application chart. They will be injected during installation for authorized usage.

```yaml
env:
  - name: OS_SYSTEM_SERVER
    value: system-server.user-system-{{ .Values.bfl.username }}
  - name: OS_APP_SECRET

    # The appname is defined in the application chart
    value: "{{ .Values.os.<appnane>.appSecret }}"
  - name: OS_APP_KEY
    value: "{{ .Values.os.<appname>.appKey }}"
```

You can use these three environment variables in the code to call the **Provider**. Take `curl` as an example:

1. Get the `access token`, which has a valid duration of 5 minutes. Token encryption algorithm: bcrypt(`app key` `timestamp` `app secret`), default cost 10.

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

2. You will receive a response like:
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "access_token": "JDJ5JDEwJE5Wbk9vbFpoLjJlSGxhUUpRY1IwRmVZVjFBWmUxUi5LOXNuQWJmVjRnN29xNWVVaFhPWmV5"
      }
    }
    ```

3. You can then use the token to call the provider's API

    ```sh
    # API URL format http://${OS_SYSTEM_SERVER}/system-server/v1alpha1/<dataType>/<group>/<version>/<op>
    curl http://${OS_SYSTEM_SERVER}/system-server/v1alpha1/app/service.bfl/v1/InstallDevApp \
      -H "content-type: application/json" \
      -H "X-Access-Token: ${access_token}"  \
      -d '{"data":"post to provider"}'
    ```
