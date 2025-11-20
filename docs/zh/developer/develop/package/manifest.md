---
outline: [2, 3]
---

# OlaresManifest 规范

每一个 Olares 应用的 Chart 根目录下都必须有一个名为 `OlaresManifest.yaml` 的文件。`OlaresManifest.yaml` 描述了一个 Olares 应用的所有基本信息。Olares 应用市场协议和 Olares 系统依赖这些关键信息来正确分发和安装应用。

:::info 提示
最新的 Olares 系统使用的 Manifest 版本为: `0.10.0`
- 修改 `categories` 分类
- 增加 Permission 部分中 `provider` 权限的申请
- 增加 Provider 部分，用于让应用对集群内暴露指定服务接口
- 移除 Spec 部分已不支持的一些配置项
- 移除 Option 部分已不支持的一些配置项
- 增加 `allowMultipleInstall` 配置，允许应用克隆出多个独立的实例
- 增加 Envs 部分，支持应用声明需要的环境变量
:::
:::details Changelog
`0.9.0`
- 在 `options` 中增加 `conflict` 字段, 用于声明不兼容的应用
- 移除 `options` 中 `analytics` 配置项
- 修改 `tailscale` 字段的配置格式
- 增加 `allowedOutboundPorts` 配置，允许通过指定端口进行非 HTTP 协议的对外访问
- 修改 `ports` 部分的配置

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
olaresManifest.version: '0.10.0'
olaresManifest.type: app
metadata:
  name: helloworld
  title: Hello World
  description: app helloworld
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
  version: 0.0.1
  categories:
  - Utilities
entrances:
- name: helloworld
  port: 8080
  title: Hello World
  host: helloworld
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
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

:::info 提示
Olares Market 目前不展示 `recommend` 类型的应用，但你可以上传自定义 Chart 来给 Wise 安装推荐算法
:::

## olaresManifest.version

- 类型：`string`

随着 Olares 更新，`OlaresManifest.yaml` 的配置规范可能会发生变化。你可以通过检查 `olaresManifest.version` 来确定这些更改是否会影响你的应用程序。 `olaresManifest.version` 由三个用英文句点分隔的整数组成。

- 第 1 位数字增加意味着引入了不兼容的配置项，未升级对应 `OlaresManifest.yaml` 的应用将无法分发或安装。
- 第 2 位数字增加意味着分发和安装必须字段存在变化，但 Olares 系统仍兼容之前所有版本配置的应用分发与安装。我们建议开发者尽快更新升级应用的 `OlaresManifest.yaml` 文件。
- 第 3 位数字的改变，不影响应用分发和安装。

开发者可以使用 1-3 位的版本号来标识该应用遵循的配置版本。以下是有效版本的一些示例：
```Yaml
olaresManifest.version: 1
olaresManifest.version: 1.1.0
olaresManifest.version: '2.2'
olaresManifest.version: "3.0.122"
```

## Metadata

应用的基本信息，用于在 Olares 系统和应用市场中展示应用。

:::info 示例
```Yaml
metadata:
  name: nextcloud
  title: Nextcloud
  description: The productivity platform that keeps you in control
  icon: https://app.cdn.olares.com/appstore/nextcloud/icon.png
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

描述在应用市场的哪个类别下展示应用。

OS 1.11 有效值：
- `Blockchain`, `Utilities`, `Social Network`, `Entertainment`, `Productivity`

OS 1.12 有效值：
- `Creativity`：设计创作
- `Productivity_v112`：工作效率
- `Developer Tools`：开发工具
- `Fun`：休闲娱乐
- `Lifestyle`：生活方式
- `Utilities_v112`：实用工具
- `AI`：AI



:::info 提示
Olares OS 1.12.0 版本对应用商店的应用分类进行了调整，因此如果应用需要同时兼容 1.11 和 1.12 版本，请同时填写两个版本所需的分类。
:::

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

当将 `addToTailscaleAcl` 字段设置为 `true` 时，系统会为该端口分配一个随机端口，并自动将其加入到 Tailscale 的 ACL 中。

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

:::info 提示
从 1.12.0 版本开始，该权限配置已经被废弃。
:::

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

### provider

- 类型：`list<map>`
- 可选

用于声明本应用需访问的其他应用接口。被访问的应用需在其 `provider` 部分声明对外开放的 `providerName`，详见下方 Provider 章节。

此处 `appName` 应填写目标应用的 `name`，`providerName` 填写目标应用 `provider` 配置中的 `name` 字段。`podSelectors` 字段用于指定本应用中哪些 pod 需要访问目标应用。如果未声明此字段，则默认为本应用的所有 pod 注入 `outbound envoy sidecar`。

:::info 调用应用示例
```Yaml
# 需要调用其他应用的应用，如 sonarr
permission:  
  provider:
  - appName: bazarr
    providerName: bazarr-svc
    podSelectors:
      - matchLabels:
          io.kompose.service: api
```
:::
:::info 被调用应用示例
```Yaml
# 被调用方应用，如 bazarr
provider:
- name: bazarr-svc
  entrance: bazarr-svc
  paths: ["/*"]
  verbs: ["*"]
```
:::


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
  versionName: '10.8.11' 
  # 此 Chart 包含的应用程序的版本。建议将版本号括在引号中。该值对应于 Chart.yaml 文件中的 appVersion 字段。请注意，它与 version 字段无关。

  featuredImage: https://app.cdn.olares.com/appstore/jellyfin/promote_image_1.jpg
  # 当应用在应用市场上推荐时，会显示特色图像。

  promoteImage:
  - https://app.cdn.olares.com/appstore/jellyfin/promote_image_1.jpg
  - https://app.cdn.olares.com/appstore/jellyfin/promote_image_2.jpg
  - https://app.cdn.olares.com/appstore/jellyfin/promote_image_3.jpg
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

:::info 提示
MongoDB，MySQL，MariaDB，MinIO，RabbitMQ需要管理员从 Market 安装后才能被其他应用使用
:::

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
```Yaml
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
```Yaml
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
```Yaml
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

### mobileSupported
- 类型： `boolean`
- 默认值： `false`
- 可选

确定应用是否与移动网络浏览器兼容并且可以在移动版本的 Olares 桌面上显示。如果应用程序针对移动网络浏览器进行了优化，请启用此选项。这将使该应用程序在移动版 Olares 桌面上可见并可访问。

:::info 示例
```Yaml
mobileSupported: true
```
:::

### oidc
- 类型：`map`
- 可选

Olares 包含内置的 OpenID Connect 身份验证组件，以简化用户的身份验证。启用此选项可在你的应用中使用 OpenID。
```Yaml
# yaml 中 OpenID 相关变量
{{ .Values.oidc.client.id }}
{{ .Values.oidc.client.secret }}
{{ .Values.oidc.issuer }}
```

:::info 示例
```Yaml
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
```Yaml
apiTimeout: 0
```
:::

### allowedOutboundPorts
- 类型： `map`
- 可选

要求开通以下端口进行非 HTTP 协议的对外访问，例如 SMTP 服务等。

:::info 示例
```Yaml
allowedOutboundPorts:
  - 465
  - 587
```
:::    

### allowMultipleInstall
- 类型： `boolean`
- 默认值： `false`
- 可选

该应用支持在同一 Olares 集群中部署多个独立实例。此设置对付费应用和共享应用客户端无效。

## Envs

在此声明应用运行所需的环境变量，既支持用户手动输入，也可以直接引用已有的系统环境变量值。

:::info 提示
该配置需要 Olares OS 版本在 1.12.2 以上才生效
:::    

:::info 示例
```Yaml
envs:
  - envName: ENV_NAME
    # 在部署应用时，注入的value的key，最终注入为.Values.olaresEnv.ENV_NAME

    required: true
    # 是否安装必须有值，若必填，则：若开发者未设置default，用户安装应用时必须提供，且修改value时不允许清空

    default: "DEFAULT"
    # 环境变量的默认值，开发者可在编写时提供，用户不可修改。

    type: string
    # 环境变量的类型，目前有int/bool/url/ip/domain/email/string/password。如果声明，会对value进行类型校验

    editable: true    
    # 是否可在应用部署后编辑

    options:
    - title: Windows11
      value: "11"
    - title: Windows10
      value: "10"
    # 允许值列表，此环境变量的值只允许从该列表中选择
    # title表示展示给用户的描述，value是实际提供给系统的值

    remoteOptions: https://xxx.xxx/xx
    # 提供允许值列表的一个url，response的body格式为JSON编码的options列表

    regex: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    # 该环境变量的值必须符合此正则表达式

    valueFrom:
      envName: OLARES_SYSTEM_CLUSTER_DNS_SERVICE
    # 引用系统环境变量的值。如果采用该方式，将不允许用户手动指定/修改其value
    # 引用包含了带引用的环境变量的所有可声明字段(type,editable)，应用环境变量里的同名属性会失效，default/value字段也会失效
    
    applyOnChange: true
    # 是否在此环境变量的值变化时自动重新部署应用，使变化生效
    # 若该字段为false，在环境变量变化时，即使停止/启动应用，也不会生效，只有升级/重装会生效
    
    description: "DESCRIPTION"
    # 对环境变量的描述
```
:::

如需在部署 YAML 文件中使用环境变量的值，只需在相应位置使用 `.Values.olaresEnv.ENV_NAME` 即可。系统会在应用部署时自动将对应的 olaresEnv 变量注入到 values 中。例如

:::info 示例
```Yaml
BACKEND_MAIL_HOST: "{{ .Values.olaresEnv.MAIL_HOST }}"
BACKEND_MAIL_PORT: "{{ .Values.olaresEnv.MAIL_PORT }}"
BACKEND_MAIL_AUTH_USER: "{{ .Values.olaresEnv.MAIL_AUTH_USER }}"
BACKEND_MAIL_AUTH_PASS: "{{ .Values.olaresEnv.MAIL_AUTH_PASS }}"
BACKEND_MAIL_SECURE: "{{ .Values.olaresEnv.MAIL_SECURE }}"
BACKEND_MAIL_SENDER: "{{ .Values.olaresEnv.MAIL_SENDER }}"
```
:::

## Provider

在此声明本应用向其他应用开放的接口。系统会自动为这些接口生成 Service，让集群内其他应用能够通过内部网络访问。如果其他应用要调用这些接口，需要在 permission 部分申请访问该 provider 的权限。

:::info 示例
```Yaml
provider:
- name: bazarr
  entrance: bazarr-svc   # 该服务的入口名称
  paths: ["/api*"]       # 开放的接口路径，不能只包含通配符 *
  verbs: ["*"]           # 支持post,get,put,delete,patch，"*"允许所有方法

```
:::
