# 其它

## ProviderRegistry

如果需要将原系统应用的 provider 接入你的开发环境可以手工修改 ProviderRegistry 的地址。在控制面板中的 **CRDs**，找到 `sys.bytetrade.io`，找到 `ProviderRegistry`。

![image](/images/developer/develop/contribute/system-app/other/provider_registry.jpg)

进入列表后找到你要替换的应用（比如 desktop-notification），点击右侧的功能按钮，选择**编辑 YAML**. 将 yaml 中的` endpoint`，指向你的开发应用的 service 地址. 保存之后即可生效。

![image](/images/developer/develop/contribute/system-app/other/edit_yaml.jpg)


## vite 配置

如果前端项目采用了 vite，需要增加 hmr 配置。原因是 vite 在 dev 状态，会启动 websocket 监听服务器端发送的代码更新 reload 通知。默认 ws 端口为 server 启动的端口。而 dev app 启动了 nginx 代理，采用了标准的 443 端口。所以需要做相应修改。

如果是 quasar + vite，在 `quasar.config.js` 里增加一段

```js
extendViteConf(viteConf) {
    viteConf.server.hmr = {clientPort: 443};
},
```
如果是独立 vite 项目，需要修改 `vite.config.js` 文件。
```js
export default defineConfig({
  server: {
    hmr: {
      clientPort: 443,
    },
  },
});
```

## 使用系统数据库

在 `deployment.yaml` 中，添加 `MiddlewareRequest`。
以 Dify 中申请 `postgres` 为例:
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

添加 Middleware request，需要设置一个密码的 secret。

```Yaml
apiVersion: v1
kind: Secret
metadata:
  name: dify-secrets
  namespace: {{ .Release.Namespace }}
type: Opaque
data:
  pg_password: {{ $pg_password }}   # password 可以随机生成，然后做 Base64 encode
```

在使用的地方配置如下字段：

```Yaml
env:
  - name: DB_USERNAME
    value: dify_user_space_{{ .Values.bfl.name }}   # 注意，在上面配置的Middleware request配置的用户名，在实际使用时，需要加上namespace 后缀。Postgres的用户名，还需要把 - 换成 _
  - name: DB_PASSWORD
    value: {{ $pg_password | b64dec }}  #  上面配置的密码原文
  - name: DB_HOST
    value: citus-master-svc.user-system-{{ .Values.bfl.username }}   # 数据库地址，  redis的地址为 redis-cluster-proxy.user-system-{{ .Values.bfl.username }}， mongo 地址为 mongo-cluster-mongos.user-system-{{ .Values.bfl.username }}
    # Redis：redis-cluster-proxy.user-system-{{ .Values.bfl.username }}
    # Mongo：mongo-cluster-mongos.user-system-{{ .Values.bfl.username }}
  - name: DB_PORT
    value: '5432'   # redis 端口 6379，mongo 端口 27017
    # Redis：6379
    # Mongo：27017
  - name: DB_DATABASE
    value: os_system_dify  # 注意，在上面配置的 Middleware request 配置的数据库名，在实际使用时，需要加上 namespace 前缀。Postgres 的数据库名，还需要把 - 换成 _
```

也可以拼接 `dsn` 链接：

```Yaml
postgres://dify_{{ .Values.bfl.username }}:{{ $pg_password_data }}@citus-master-svc.user-system-{{ .Values.bfl.username }}/user_space_{{ .Values.bfl.username }}_dify?sslmode=disable

mongodb://dify-{{ .Values.bfl.username }}:{{ $mongo_password_data }}@mongo-cluster-mongos.user-system-{{ .Values.bfl.username }}:27017/{{ .Release.Namespace }}_dify

redis://:{{ $redis_password | b64dec }}@redis-cluster-proxy.user-system-{{ .Values.bfl.username }}:6379/0  # 注意，由于系统采用的是 redis cluster，db 只能写 0
```

如果提供 provider 需要手工在 `deployment.yaml` 文件里配 `ProviderRegistry`。

```Yaml
apiVersion: sys.bytetrade.io/v1alpha1
kind: ProviderRegistry
metadata:
  name: desktop-provider-dev  # 名字不能和现有的重名
  namespace: user-system-{{ .Values.bfl.username }}
spec:

  dataType: config-dev   # 类型不要现在的重名，如果想替换现在的，需要把已有的删除
  deployment: desktop-dev
  description: Set Desktop Config
  endpoint: desktop-svc-dev.{{ .Release.Namespace }}  # provider 的访问地址，指向开发的应用
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
