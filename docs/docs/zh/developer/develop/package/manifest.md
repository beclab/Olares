---
outline: [2, 3]
---

# OlaresManifest 规范

每一个 Olares 应用的 Chart 根目录下都必须有一个名为 `OlaresManifest.yaml` 的文件。`OlaresManifest.yaml` 描述了一个 Olares 应用的所有基本信息。Olares 应用市场协议和 Olares 系统依赖这些关键信息来正确分发和安装应用。

:::info 提示
最新的 Olares 系统使用的 Manifest 版本为: `0.9.0`
- 在 `options` 中增加 `conflict` 字段, 用于声明不兼容的应用
- 移除 `options` 中 `analytics` 配置项
- 修改 `tailscale` 字段的配置格式
- 增加 `allowedOutboundPorts` 配置，允许通过指定端口进行非 HTTP 协议的对外访问
- 修改 `ports` 部分的配置

:::
:::details Changelog
`0.8.3`
- 在 `dependencies` 配置项里增加 `mandatory` 字段以表示该依赖应用必须安装。
- 增加 `tailscaleAcls` 配置项，允许 Tailscale 为应用开放指定端口

`0.8.2`
- 添加 `runAsUser` 选项，用于限制应用程序在非root权限的用户下运行

`0.8.1`
- 添加 `ports` 选项以指定 UDP 或 TCP 的暴露端口

`0.7.1`
- 添加新的 `authLevel` 值 `internal`
- 将 `spec`>`language` 改为 `spec`>`locale` 并支持 i18n
  :::

一个 `OlaresManifest.yaml` 文件的示例如下：

::: details `OlaresManifest.yaml` 示例

```Yaml
olaresManifest.version: '0.8.0'
olaresManifest.type: app
metadata:
  name: helloworld
  title: Hello World
  description: app helloworld
  icon: https://file.bttcdn.com/appstore/default/defaulticon.webp
  version: 0.0.1
  categories:
  - Utilities
entrances:
- name: helloworld
  port: 8080
  title: Hello World
  host: helloworld
  icon: https://file.bttcdn.com/appstore/default/defaulticon.webp
  authLevel: private
permission:
  appCache: true
  appData: true
  userData:
  - Home/Documents/
  - Home/Pictures/
  - Home/Downloads/BTDownloads/
spec:
  versionName: '0.0.1'
  featuredImage: https://link.to/featured_image.webp
  promoteImage:
  - https://link.to/promote_image1.webp
  - https://link.to/promote_image2.webp
  fullDescription: |
    A full description of your app.
  upgradeDescription: |
    Describe what is new in this upgraded version.
  developer: Developer's Name
  website: https://link.to.your.website
  sourceCode: https://link.to.sourceCode
  submitter: Submitter's Name
  language:
  - en
  doc: https://link.to.documents
  supportArch:
  - amd64
  limitedCpu: 1000m
  limitedMemory: 1000Mi
  requiredCpu: 50m
  requiredDisk: 50Mi
  requiredMemory: 12Mi

options:
  dependencies:
  - name: olares
    type: system
    version: '>=0.1.0'
```
:::

## olaresManifest.type

- 类型：`string`
- 有效值： `app`、`recommend`、`middleware`

Olares 市场目前支持 3 种类型的应用，各自对应不同场景。本文档以 “app” 为例来解释各个字段。其他类型请参考相应的配置指南。
- [推荐算法配置指南](recommend.md)

:::info 示例
```Yaml
olaresManifest.type: app
```
:::

## olaresManifest.version

- 类型：`string`

随着 Olares 更新，`OlaresManifest.yaml` 的配置规范可能会发生变化。你可以通过检查 `olaresManifest.version` 来确定这些更改是否会影响你的应用程序。 `olaresManifest.version` 由三个用英文句点分隔的整数组成。

- 第 1 位数字增加意味着引入了不兼容的配置项，未升级对应 `OlaresManifest.yaml` 的应用将无法分发或安装。
- 第 2 位数字增加意味着分发和安装必须字段存在变化，但 Olares 系统仍兼容之前所有版本配置的应用分发与安装。我们建议开发者尽快更新升级应用的 `OlaresManifest.yaml` 文件。
- 第 3 位数字的改变，不影响应用分发和安装。

开发者可以使用 1-3 位的版本号来标识该应用遵循的配置版本。以下是有效版本的一些示例：
```Yaml
OlaresManifest.yaml.version: 1
OlaresManifest.yaml.version: 1.1.0
OlaresManifest.yaml.version: '2.2'
OlaresManifest.yaml.version: "3.0.122"
```

## Metadata

应用的基本信息，用于在 Olares 系统和应用市场中展示应用。

:::info 示例
```Yaml
metadata:
  name: nextcloud
  title: Nextcloud
  description: The productivity platform that keeps you in control
  icon: https://file.bttcdn.com/appstore/nextcloud/icon.png
  version: 0.0.2
  categories:
  - Utilities
  - Productivity
```
:::

### name

- 类型：`string`
- Accepted Value: `[a-z][a-z0-9]?`

Olares 中的应用的命名空间，仅限小写字母数字字符。最多 30 个字符，需要与 `Chart.yaml` 中的 `FolderName` 和 `name` 字段保持一致。

### title

- 类型：`string`

在应用市场中显示的应用标题。长度不超过 `30` 个字符。

### description

- 类型：`string`

Olares 应用市场中的应用名称下方显示的简短说明。

### icon

- 类型：`url`

应用图标。

图标必须是 `PNG` 或 `WEBP` 格式文件，最大为 `512 KB`，尺寸为 `256x256 px`。

### version

- 类型：`string`

应用的 Chart Version，每次改变 Chart 目录里的内容时应递增。需遵循[语义化版本规范](https://semver.org/)，需要与 `Chart.yaml` 中的 `version` 字段一致。

### categories

- 类型： `list<string>`
- 有效值： `Blockchain`、`Utilities`、`Social Network`、`Entertainment`、`Productivity`

在应用市场的哪个类别下展示应用。

## Entrances

指定此应用访问入口的数量。每个应用允许最少 1 个，最多 10 个入口 。

:::info 示例
```Yaml
entrances:
- name: a
  host: firefox
  port: 3000
  title: Firefox
  authLevel: public
  invisible: false
- name: b
  host: firefox
  port: 3001
  title: admin
```
:::

### name

- 类型：`string`
- Accepted Value: `[a-z]([-a-z0-9]*[a-z0-9])?`

  入口的名称，长度不超过 `63` 个字符。一个应用内不能重复。

### port

- 类型： `int`
- 有效值： `0-65535`

### host

- 类型：`string`
- 有效值： `[a-z]([-a-z0-9]*[a-z0-9])?`

  当前入口的 Ingress 名称，只包含小写字母和数字和中划线`-`，长度不超过 63 个字符。

### title

- 类型：`string`

安装后 Olares 桌面的显示名称。长度不超过 `30` 个字符。

### icon

- 类型： `url`
- 可选

应用安装后 Olares 桌面上的图标。图片文件必须是 `PNG` 或 `WEBP` 格式，不超过 `512 KB`，尺寸为 `256x256 px`。

### authLevel

- 类型：`string`
- 有效值： `public`, `private`, `internal`
- 默认值： `private`
- 可选

指定入口的认证级别。
- **Public**：互联网上的任何人都可以不受限制地访问。
- **Private**：需要从内部和外部网络访问的授权。
- **Internal**：需要授权才能从外部网络访问。从内部网络（通过 LAN/VPN）访问时不需要身份验证。

### invisible

- 类型： `boolean`
- 默认值：`false`
- 可选

当 `invisible` 为` true` 时，该入口不会显示在 Olares 桌面上。

### openMethod

- 类型：`string`
- 有效值： `default`, `iframe`, `window`
- 默认值： `default`
- 可选

指定该入口在桌面的打开方式。

`iframe` 代表在桌面的窗口内通过 iframe 新建一个窗口，`window` 代表在浏览器新的 Tab 页打开。`default` 代表跟随系统的默认选择，系统默认的选择是`iframe`。

### windowPushState
- 类型： `boolean`
- 默认值：`false`
- 可选

将应用嵌入到桌面上的 iframe 中时，应用的 URL 可能会动态更改。由于浏览器的同源策略，桌面（父窗口）无法直接检测到 iframe URL 中的这些变化。因此，如果你重新打开应用程序选项卡，它将显示初始 URL，而不是更新后的 URL。

为了确保无缝的用户体验，你可以通过将其设置为 true 来启用此选项。此操作会提示网关自动将以下代码注入到 iframe 中。每当 iframe 的 URL 发生更改时，此代码都会向父窗口（桌面）发送一个事件。因此，桌面可以跟踪 URL 更改并打开正确的页面。

::: details 代码
```Javascript
<script>
  (function () {
    if (window.top == window) {
        return;
    }
    const originalPushState = history.pushState;
    const pushStateEvent = new Event("pushstate");
    history.pushState = function (...args) {
      originalPushState.apply(this, args);
      window.dispatchEvent(pushStateEvent);
    };
    window.addEventListener("pushstate", () => {
      window.parent.postMessage(
        {type: "locationHref", message: location.href},
        "*"
      );
    });
  })();
</script>
```
:::

## Ports

定义暴露的端口

:::info 示例
```Yaml
ports:
- name: rdp-tcp             # 提供服务的入口名称
  host: windows-svc         # 提供服务的 Ingress 主机名称
  port: 3389                # 提供服务的端口号
  exposePort: 46879         # 暴露的接口，在集群内一次只能分配给一个应用程序。
  addToTailscaleAcl: true   # 自动添加到 Tailscle 的 ACL 列表中
```
:::

Olares 会为你的应用暴露指定的端口，这些端口可通过应用域名在本地网络下访问，如`84864c1f.your_olares_id.olares.com:46879`。对于每个公开的端口，Olares 会自动配置相同端口号的 TCP 和 UDP。

当设置 `addToTailscaleAcl` 字段为 `true`时，该端口会自动增加tailscale acl中，无需再去tailscale部分配置


:::info 提示
暴露的端口只能通过本地网络或 Olares 专用网络访问。
:::


## Permission

:::info 示例
```Yaml
permission:
  appCache: true
  appData: true
  userData:
    - /Home/
  sysData:
    - dataType: legacy_prowlarr
      appName: prowlarr
      port: 9696
      group: api.prowlarr
      version: v2
      ops:
        - All
```
:::

### appCache

- 类型： `boolean`
- 可选

是否需要在 `Cache` 目录创建应用的目录。如需要在部署 yaml 文件中使用`.Values.userspace.appCache`,  `appCache` 必须设为 `true`。

### appData

- 类型： `boolean`
- 可选

是否需要在 `Data` 目录创建应用的目录。如需要在部署 yaml 中使用`.Values.userspace.appData`,  `appData` 必须设为 `true`。

### userData

- 类型：`string`
- 可选

应用是否需要对用户的 `Home` 文件夹进行读写权限。列出应用需要访问的用户 `Home` 下的所有目录。部署 YAML 中配置的所有 `userData` 目录都必须包含在此处。

### sysData

- 类型：`list<map>`
- 可选

声明该应用程序需要访问的 API 列表。

:::info 示例
```Yaml
  sysData:
  - group: service.bfl
    dataType: app
    version: v1
    ops:
    - InstallDevApp
  - dataType: legacy_prowlarr
    appName: prowlarr
    port: 9696
    group: api.prowlarr
    version: v2
    ops:
    - All
```
:::

所有系统 API [providers](../advanced/provider.md) 如下：
| Group | version | dataType | ops |
| ----------- | ----------- | ----------- | ----------- |
| service.appstore | v1 | app | InstallDevApp, UninstallDevApp
| message-disptahcer.system-server | v1 | event | Create, List
| service.desktop | v1 | ai_message | AIMessage
| service.did | v1 | did | ResolveByDID, ResolveByName, Verify
| api.intent | v1 | legacy_api | POST
| service.intent | v1 | intent | RegisterIntentFilter, UnregisterIntentFilter, SendIntent, QueryIntent, ListDefaultChoice, CreateDefaultChoice, RemoveDefaultChoice, ReplaceDefaultChoice
| service.message | v1 | message | GetContactLogs, GetMessages, Message
| service.notification | v1 | message | Create
| service.notification | v1 | token | Create
| service.search | v1 | search | Input, Delete, InputRSS, DeleteRSS, QueryRSS, QuestionAI
| secret.infisical | v1 | secret | CreateSecret, RetrieveSecret
| secret.vault | v1 | key | List, Info, Sign

## Tailscale
- 类型：`map`
- 可选

允许应用在 Tailscale 的ACL(Access Control Lists)中开放指定端口。

:::info 示例
```Yaml
tailscale:
  acls:
  - proto: tcp
    dst:
    - "*:46879"
  - proto: "" # 可选, 如果未指定，则允许使用所有支持的协议
    dst:
    -  "*:4557"  
```
:::

## Spec
记录额外的应用信息，主要用于应用商店的展示。

:::info 示例
```Yaml
spec:
  namespace: os-system 
  # 可选。将应用安装到指定的命名空间，如 os-system、user-space 和 user-system

  versionName: '10.8.11' 
  ## 此 Chart 包含的应用程序的版本。建议将版本号括在引号中。该值对应于 Chart.yaml 文件中的 appVersion 字段。请注意，它与 version 字段无关。

  featuredImage: https://file.bttcdn.com/appstore/jellyfin/promote_image_1.jpg
  # 当应用在应用市场上推荐时，会显示特色图像。

  promoteImage:
  - https://file.bttcdn.com/appstore/jellyfin/promote_image_1.jpg
  - https://file.bttcdn.com/appstore/jellyfin/promote_image_2.jpg
  - https://file.bttcdn.com/appstore/jellyfin/promote_image_3.jpg
  fullDescription: |
    Jellyfin is the volunteer-built media solution that puts you in control of your media. Stream to any device from your own server, with no strings attached. Your media, your server, your way.
  upgradeDescription: |
    upgrade descriptions
  developer: Jellyfin
  website: https://jellyfin.org/
  doc: https://jellyfin.org/docs/
  sourceCode: https://github.com/jellyfin/jellyfin
  submitter: Olares
  locale:
  - en-US
  - zh-CN
  # 列出该应用支持的语言和地区

  requiredMemory: 256Mi
  requiredDisk: 128Mi
  requiredCpu: 0.5
  # 指定安装和运行应用所需的最少资源。安装应用后，系统将保留这些资源以确保最佳性能。

  limitedDisk: 256Mi
  limitedCpu: 1
  limitedMemory: 512Mi
  # 指定应用的最大资源限制。如果应用超出这些限制，它将暂时暂停，以防止系统过载并确保稳定性。

  legal:
  - text: Community Standards
    url: https://jellyfin.org/docs/general/community-standards/
  - text: Security policy
    url: https://github.com/jellyfin/jellyfin/security/policy
  license:
  - text: GPL-2.0
    url: https://github.com/jellyfin/jellyfin/blob/master/LICENSE
  supportClient:
  - android: https://play.google.com/store/apps/details?id=org.jellyfin.mobile
  - ios: https://apps.apple.com/us/app/jellyfin-mobile/id1480192618
```
:::

### i18n

要在 Olares 应用市场中为应用添加多语言支持：

1. 在 Olares Application Chart 根目录中创建一个 `i18n` 文件夹。
2. 在 `i18n` 文件夹中，为每个支持的语言环境创建单独的子目录。
3. 在每个语言环境子目录中，放置 `OlaresManifest.yaml` 文件的本地化版本。

Olares 应用市场将根据用户的区域设置自动显示相应的 `OlaresManifest.yaml` 文件的内容。
:::info 示例
```
.
├── Chart.yaml
├── README.md
├── OlaresManifest.yaml
├── i18n
│   ├── en-US
│   │   └── OlaresManifest.yaml
│   └── zh-CN
│       └── OlaresManifest.yaml
├── owners
├── templates
│   └── deployment.yaml
└── values.yaml
```
:::
目前，你可以为以下字段添加 i18n 内容：
```Yaml
metadata:
  description:
  title:
spec:
  fullDescription:
  upgradeDescription:
```

### supportArch
- 类型: `list<string>`
- 有效值: `amd64`, `arm64`
- 可选

该字段用于声明应用程序支持的 CPU 架构。目前仅支持 `amd64` 和 `arm64` 两种类型。

:::info 示例
```yaml
spec:
  supportArch:
  - amd64
  - arm64
```
:::

:::info 提示
Olares 目前不支持混合架构的集群。
:::

### onlyAdmin
- 类型: `boolean`
- 默认值: `false`
- 可选

设置为 `true` 时，只有管理员可以安装此应用程序。

### runAsUser
- 类型: `boolean`
- 可选

当设置为 `true` 时，Olares 会强制以用户 ID “1000”（非 root 用户）运行应用程序。

## Middleware
- 类型：`map`
- 可选

系统提供了高可用的中间件服务，开发者无需重复安装中间件，只需在此填写对应的中间件信息即可，然后可以直接使用应用程序的 deployment YAML 文件中相应的中间件信息。

使用 `scripts` 字段指定创建数据库后应执行的脚本。此外，使用 `extension` 字段在数据库中添加相应的扩展名。

:::info 示例
```Yaml
middleware:
  postgres:
    username: immich
    databases:
    - name: immich
      extensions:
      - vectors
      - earthdistance
      scripts:
      - BEGIN;                                           
      - ALTER DATABASE $databasename SET search_path TO "$user", public, vectors;
      - ALTER SCHEMA vectors OWNER TO $dbusername;
      - COMMIT;
      # 操作系统提供了两个变量 $databasename 和 $dbusername，命令执行时会被 Olares 应用运行时替换。
  redis:
    password: password
    namespace: db0
  mongodb:
    username: chromium
    databases:
    - name: chromium
      script:
      - 'db.getSiblingDB("$databasename").myCollection.insertOne({ x: 111 });'
      # 请确保每一行都是完整的查询。
```
:::

使用 deployment YAML 中的中间件信息：

```yaml
- name: DB_POSTGRESDB_DATABASE # 你在 OlaresManifest 中配置的数据库名称，在 middleware.postgres.databases[i].name 中指定
  value: {{ .Values.postgres.databases.<dbname> }}
- name: DB_POSTGRESDB_HOST
  value: {{ .Values.postgres.host }}
- name: DB_POSTGRESDB_PORT
  value: "{{ .Values.postgres.port }}"
- name: DB_POSTGRESDB_USER
  value: {{ .Values.postgres.username }}
- name: DB_POSTGRESDB_PASSWORD
  value: {{ .Values.postgres.password }}


# 对于mongodb来说，对应的值如下
host --> {{ .Values.mongodb.host }}
port --> "{{ .Values.mongodb.port }}"  # yaml 文件中的端口和密码需要用双引号括起来。
username --> {{ .Values.mongodb.username }}
password --> "{{ .Values.mongodb.password }}" # yaml 文件中的端口和密码需要用双引号括起来。
databases --> "{{ .Values.mongodb.databases }}" # 数据库的值类型是 map。你可以使用 {{ .Values.mongodb.databases.<dbname> }} 获取数据库。 <dbname> 是你在 OlaresManifest 中配置的名称，在 middleware.mongodb.databases[i].name 中指定


# 对于Redis来说，对应的值如下
host --> {{ .Values.redis.host }}For Redis, the corresponding value is as follow
port --> "{{ .Values.redis.port }}"
password --> "{{ .Values.redis.password }}"

```

## Options

在此部分配置系统相关的选项。

### policies
- 类型：`map`
- 可选

定义应用子域的详细访问控制。

:::info 示例
```yaml
options:
  policies:
    - uriRegex: /$
      level: two_factor
      oneTime: false
      validDuration: 3600s
      entranceName: gitlab
```
:::

### clusterScoped
- 类型：`map`
- 可选

是否为 Olares 集群中的所有用户安装此应用程序。

:::info 服务端示例
```yaml
metadata:
  name: gitlab
options:
  appScope:
    clusterScoped: true
    appRef:
      - gitlabclienta # 客户端的应用名称
      - gitlabclientb
```
:::

:::info 客户端示例
```yaml
metadata:
  name: gitlabclienta
options:
  dependencies:
    - name: olares
      version: ">=0.3.6-0"
      type: system
    - name: gitlab # 服务器端的应用名称
      version: ">=0.0.1"
      type: application
      mandatory: true
```
:::

### dependencies
- 类型：`list<map>`

如果此应用依赖于其他应用或需要特定操作系统版本，请在此处声明。

如果此应用程序需要依赖其他应用程序才能正确安装，则应将 `mandatory` 字段设置为 `true`。

:::info 示例
```yaml
options:
  dependencies:
    - name: olares
      version: ">=1.0.0-0"
      type: system
    - name: mongodb
      version: ">=6.0.0-0"
      type: middleware
      mandatory: true # 如果必须先安装此依赖，请将此字段设为 true。
```
:::

### websocket
- 类型：`map`
- 可选

为应用启用 websocket。请参阅 [websocket](../advanced/websocket.md) 了解更多信息。

:::info 示例
```yaml
options:
  websocket:
    url: /ws/message
    port: 8888
```
:::

### resetCookie
- 类型：`map`
- 可选

如果应用需要 cookie，请启用此功能。更多信息请参考 [cookie](../advanced/cookie.md)。

:::info 示例
```yaml
options:
  resetCookie:
    enabled: true
```
:::

### upload
- 类型： `map`
- 可选

Olares 应用运行时包含一个内置文件上传组件，旨在简化应用程序中的文件上传过程。请参阅 [上传](../advanced/file-upload.md) 了解更多信息。

:::info Example
```yaml
upload:
  # 允许上传的文件类型，*为任意类型， 上传时会指定 file_type，必须在允许的文件类型中
  fileType:
    - pdf
  # dest 的路径必须为某一个 mountPath
  dest: /appdata
  # 文件上传的最大大小，单位为字节
  limitedSize: 3729747942
```
:::

### mobileSupported
- 类型： `boolean`
- 默认值： `false`
- 可选

确定应用是否与移动网络浏览器兼容并且可以在移动版本的 Olares 桌面上显示。如果应用程序针对移动网络浏览器进行了优化，请启用此选项。这将使该应用程序在移动版 Olares 桌面上可见并可访问。

:::info 示例
```yaml
mobileSupported: true
```
:::

### oidc
- 类型：`map`
- 可选

Olares 包含内置的 OpenID Connect 身份验证组件，以简化用户的身份验证。启用此选项可在你的应用中使用 OpenID。
```yaml
# yaml 中 OpenID 相关变量
{{ .Values.oidc.client.id }}
{{ .Values.oidc.client.secret }}
{{ .Values.oidc.issuer }}
```

:::info 示例
```yaml
oidc:
  enabled: true
  redirectUri: /path/to/uri
  entranceName: navidrome
```
:::

### apiTimeout
- 类型：`int`
- 可选

指定 API 提供程序的超时限制（以秒为单位）。默认值为 `15`。使用 `0` 允许无限制的 API 连接。

:::info 示例
```yaml
apiTimeout: 0
```
:::

### allowedOutboundPorts
- 类型： `map`
- 可选

要求开通以下端口进行非 HTTP 协议的对外访问，例如 SMTP 服务等。

:::info 示例
```yaml
allowedOutboundPorts:
  - 465
  - 587
```
:::    