# Other

## Change Provider Registry

If you need to connect the `provider` of the original system application to your development environment, you can manually change the address of the `ProviderRegistry`. In the **'CRDs'** page of the control hub, locate `sys.bytetrade.io` and then find the `ProviderRegistry`.

![image](/images/developer/develop/contribute/system-app/other/provider_registry.jpg)

In the list, navigate to the app you wish to replace (in this case, **desktop-notification**). Click on the **'...'** button on the right and select **'Edit YAML'**. Modify the `endpoint` in the **YAML** file and direct it towards the service address of your **developing app**. Click **'OK'** to save and apply the changes.

![image](/images/developer/develop/contribute/system-app/other/edit_yaml.jpg)


## Vite configuration

If your frontend project uses **Vite**, you need to add an **HMR** configuration. In development mode, **Vite** initiates a **WebSocket** to receive code reload notifications from the server. The default **WebSocket** port matches the server's startup port. However, if the development app uses an **Nginx proxy** it will operate on the default port 443. Therefore, some modifications are required.

If you are using **Quasar** + **Vite**, add the following in the `quasar.config.js`:

```js
extendViteConf(viteConf) {
    viteConf.server.hmr = {clientPort: 443};
},
```
If it is a standalone **Vite** project, modify `vite.config.js` as:
```js
export default defineConfig({
  server: {
    hmr: {
      clientPort: 443,
    },
  },
});
```

## Use System Database

You can add **system databases** by adding `MiddlewareRequest` in the `deployment.yaml`.
Using the **Postgres** in Dify as an example:
```Yaml
apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: dify-pg
  namespace: os-system
spec:
  app: dify
  appNamespace: os-system
  middleware: postgres
  postgreSQL:
    user: dify_os_system
    password:
      valueFrom:
        secretKeyRef:
          key: pg_password
          name: dify-secrets
    databases:
    - name: dify
```

You need to set a `secret` of password for adding a `MiddlewareRequest`, 

```Yaml
apiVersion: v1
kind: Secret
metadata:
  name: dify-secrets
  namespace: {{ .Release.Namespace }}
type: Opaque
data:
  pg_password: {{ $pg_password }}   # Password can be randomly generated, then Base64 encoded
```

Configure the following in the pod where you want to use the database.

```Yaml
env:
  - name: DB_USERNAME
    value: dify_user_space_{{ .Values.bfl.name }}   # Please note, you need to add a namespace suffix when use the username configured in the MiddlewareRequest above. For username in Postgres, you also need to replace - with _
  - name: DB_PASSWORD
    value: {{ $pg_password | b64dec }}  #  The decoded password configured above
  - name: DB_HOST
    value: citus-master-svc.user-system-{{ .Values.bfl.username }}   # HOST address,
    # For Redis: redis-cluster-proxy.user-system-{{ .Values.bfl.username }}
    # For Mongo: mongo-cluster-mongos.user-system-{{ .Values.bfl.username }}
  - name: DB_PORT
    value: '5432'   
    # For Redis: 6379
    # For Mongo: 27017
  - name: DB_DATABASE
    value: os_system_dify  # Please note, you need to add a namespace suffix when use the database name configured in the MiddlewareRequest above. For username in Postgres, you also need to replace - with _
```

You can also concatenate the `dsn` link:

```Yaml
postgres://dify_{{ .Values.bfl.username }}:{{ $pg_password_data }}@citus-master-svc.user-system-{{ .Values.bfl.username }}/user_space_{{ .Values.bfl.username }}_dify?sslmode=disable

mongodb://dify-{{ .Values.bfl.username }}:{{ $mongo_password_data }}@mongo-cluster-mongos.user-system-{{ .Values.bfl.username }}:27017/{{ .Release.Namespace }}_dify

redis://:{{ $redis_password | b64dec }}@redis-cluster-proxy.user-system-{{ .Values.bfl.username }}:6379/0  # Please note, Since the system uses Redis Cluster, the database name must set to '0'.
```

To register a provider, you need to add a `ProviderRegistry` in the `deployment.yaml` file.

```Yaml
apiVersion: sys.bytetrade.io/v1alpha1
kind: ProviderRegistry
metadata:
  name: desktop-provider-dev  # The name cannot be duplicated with the existing one.
  namespace: user-system-{{ .Values.bfl.username }}
spec:

  dataType: config-dev   # The dataTypes cannot be duplicated with an existing one. If you want to replace an existing dataType, you need to delete it first.
  deployment: desktop-dev
  description: Set Desktop Config
  endpoint: desktop-svc-dev.{{ .Release.Namespace }}  # The address of the provider, pointing to your developing app
  group: service.desktop
  kind: provider
  namespace: {{ .Release.Namespace }}
  opApis:
  - name: Update
    uri: /server/updateDesktopConfig
  version: v1
status:
  state: active
```
