---
outline: [2, 3]
---

# OlaresManifest 规范

每一个 Olares 应用的 Chart 根目录下都必须有一个名为 `OlaresManifest.yaml` 的文件。`OlaresManifest.yaml` 描述了一个 Olares 应用的所有基本信息。Olares 应用市场协议和 Olares 系统依赖这些关键信息来正确分发和安装应用。

:::info 提示
最新的 Olares 系统使用的 Manifest 版本为： `0.12.0`
- 要求 Olares 系统版本在 1.12.6 以上
- **不再支持模版渲染**：`OlaresManifest.yaml` 中不能使用 <code v-pre>{{ ... }}</code> 等模板渲染函数
- 修改 `apiVersion` 字段有效值，增加`v3`，原 `v2` 格式将在 1.12.6 不再支持
- 增加 `options.shared` 字段，用于标识共享应用
- 增加 `spec.accelerator` 字段，用于 GPU 资源声明
- 增加 `workloadReplicas` 字段，声明所有 workload 的副本数
- 增加 `overlayGateway` 字段，支持 L2 overlay 局域网发现
- 增加 `LLMGatewaySupported` 选项，支持 LLM Gateway
- 增加 `appCommon` 和 `externalData` 权限（均默认为 `false`）
- 增加 `templateOnly` 字段，用于标记模版类应用
- 移除 [已废弃字段](#废弃字段-0-12-0)
:::

:::tip 升级至 Manifest 0.12.0
如果你正在维护已有的应用 Chart，在升级至新版本前，务必重点关注 `apiVersion`、`workloadReplicas` 以及 `permission.externalData` 这几个字段。它们涉及显著变更或新增的强制要求，可能需要你根据实际情况对现有配置进行适配和调整。
:::

:::details Changelog
`0.11.0`
- 移除已不支持的 `sysData` 配置项
- 修改共享应用的案例
- 增加 `apiVersion` 字段说明
- 增加共享入口的配置说明

`0.10.0`
- 修改 `categories` 分类
- 增加 Permission 部分中 `provider` 权限的申请
- 增加 Provider 部分，用于让应用对集群内暴露指定服务接口
- 移除 Spec 部分已不支持的一些配置项
- 移除 Option 部分已不支持的一些配置项
- 增加 `allowMultipleInstall` 配置，允许应用克隆出多个独立的实例
- 增加 Envs 部分，支持应用声明需要的环境变量

`0.9.0`
- 在 `options` 中增加 `conflict` 字段，用于声明不兼容的应用
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

```yaml
olaresManifest.version: '0.12.0'
olaresManifest.type: app
apiVersion: 'v3'

workloadReplicas:
  helloworld: 1

metadata:
  name: helloworld
  title: Hello World
  description: app helloworld
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
  version: 0.0.1
  categories:
  - AI

entrances:
- name: helloworld
  port: 8080
  title: Hello World
  host: helloworld
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
  authLevel: private
  openMethod: default

sharedEntrances:
- name: helloworld
  host: sharedentrances-api
  port: 0
  title: Hello World API
  invisible: true
  authLevel: internal
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp

permission:
  appCache: true
  appData: true
  appCommon: true
  userData:
  - Home/Documents/
  externalData: true
  
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
  locale:
  - en-US
  - zh-CN
  doc: https://link.to.documents
  supportArch:
  - amd64
  - arm64
  onlyAdmin: true

  accelerator:
  - mode: nvidia
    requiredCpu: "1"
    limitedCpu: "7"
    requiredMemory: 13Gi
    limitedMemory: 34Gi
    requiredDisk: 36Gi
    limitedDisk: 100Gi
    requiredGPUMemory: 12Gi
    limitedGPUMemory: 24Gi
  - mode: cpu  
    requiredCpu: 50m
    limitedCpu: 1000m
    requiredMemory: 12Mi
    limitedMemory: 1000Mi
    requiredDisk: 50Mi
    limitedDisk: 100Gi

options:
  dependencies:
  - name: olares
    type: system
    version: '>=1.12.6-0'
  shared: true
  apiTimeout: 0
  conflicts:
    - name: conflictapp
      type: application

envs:
  - envName: USERNAME
    required: true
    type: string
    editable: true
    applyOnChange: true
    description: 'default username'
    regex: '^[\w\-!@#$%^&*()+={}\[\]:,.?~]{6,}$'
```
:::

## olaresManifest.version

- 类型：`string`

随着 Olares 更新，`OlaresManifest.yaml` 的配置规范可能会发生变化。你可以通过检查 `olaresManifest.version` 来确定这些更改是否会影响你的应用程序。 `olaresManifest.version` 由三个用英文句点分隔的整数组成。

- 第 1 位数字增加意味着引入了不兼容的配置项，未升级对应 `OlaresManifest.yaml` 的应用将无法分发或安装。
- 第 2 位数字增加意味着分发和安装必须的字段存在变化，但 Olares 系统仍兼容之前版本配置的应用分发与安装。我们建议开发者尽快更新升级应用的 `OlaresManifest.yaml` 文件。
- 第 3 位数字的改变，不影响应用分发和安装。

请使用 3 位的版本号来标识该应用遵循的配置版本。以下是有效版本的一些示例：
```yaml
olaresManifest.version: 0.1.0
olaresManifest.version: '0.1.2'
olaresManifest.version: "0.12.0"
```

## olaresManifest.type

- 类型：`string`
- 有效值： `app`、`middleware`

Olares 市场目前支持 2 种类型的应用，在字段上有一定的区别。本文档以 `app` 类型为例来解释各个字段。

:::info
`recommend` 类型已不再包含在最新的 OlaresManifest 规范中。
:::

## apiVersion
- 类型：`string`
- 有效值：`v1`、`v3`
- 默认值：`v1`

从 olaresManifest.version: '0.12.0' 版本开始，apiVersion升级至 `v3` 版本，采用新的资源声明和共享应用配置格式，移除了OlaresManifest的模版渲染。如果没有填写该字段，将默认按照 `v1` 格式进行解析。

Olare OS 1.12.6 同时支持 `v1` 和 `v3` 格式应用的安装，但不再兼容 `v2` 格式的共享应用安装。

:::warning `apiVersion: 'v2'`将废弃
- `apiVersion: 'v2'` 用于 Olares OS 1.12.5 及以前版本的共享应用，将在 1.12.6 发布后逐步停止支持。
- 升级到 Olares OS 1.12.6 版本后：
  - 已安装的`apiVersion: 'v2'`不受影响，仍可继续使用，但无法继续升级。建议尽快迁移到 `apiVersion: 'v3'` 版本的共享应用。
  - 商店列表中 `apiVersion: 'v2'` 的应用不再显示，也无法安装
:::

## metadata

应用的基本信息，用于在 Olares 系统和应用市场中展示应用。

:::info 示例
```yaml
metadata:
  name: nextcloud
  title: Nextcloud
  description: The productivity platform that keeps you in control
  icon: https://app.cdn.olares.com/appstore/nextcloud/icon.png
  version: 0.0.2
  categories:
  - Utilities_v112
  - Productivity_v112
```
:::

### name

- 类型：`string`
- 有效值：`^[a-z][a-z0-9]{0,29}$`

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

OS 1.12 有效值：
- `Creativity`：设计创作
- `Productivity_v112`：工作效率
- `Developer Tools`：开发工具
- `Fun`：休闲娱乐
- `Lifestyle`：生活方式
- `Utilities_v112`：实用工具
- `AI`：AI

## entrances

指定此应用访问入口的数量。每个应用允许最少 1 个，最多 10 个入口 。

:::info 示例
```yaml
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
- **Internal**：需要授权才能从外部网络访问。从内部网络(通过 LAN/专用网络)访问时不需要身份验证。

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

将应用嵌入到桌面上的 iframe 中时，应用的 URL 可能会动态更改。由于浏览器的同源策略，桌面(父窗口)无法直接检测到 iframe URL 中的这些变化。因此，如果你重新打开应用程序选项卡，它将显示初始 URL，而不是更新后的 URL。

为了确保无缝的用户体验，你可以通过将其设置为 true 来启用此选项。此操作会提示网关自动将以下代码注入到 iframe 中。每当 iframe 的 URL 发生更改时，此代码都会向父窗口(桌面)发送一个事件。因此，桌面可以跟踪 URL 更改并打开正确的页面。

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

## sharedEntrances

共享入口是共享应用为集群内其他应用调用提供的接口地址。共享入口的字段配置和常规入口基本一致，一个典型的共享入口配置如下

:::info 示例
```yaml
sharedEntrances:
  - name: ollamav2
    host: sharedentrances-ollama
    port: 0
    title: Ollama API
    icon: https://app.cdn.olares.com/appstore/ollama/icon.png
    invisible: true
    authLevel: internal
```
:::

## ports

定义暴露的端口

:::info 示例
```yaml
ports:
- name: rdp-tcp             # 提供服务的入口名称
  host: windows-svc         # 提供服务的 Ingress 主机名称
  port: 3389                # 提供服务的端口号
  protocol: udp             # 暴露端口使用的协议
  exposePort: 46879         # 暴露的端口，在集群内一次只能分配给一个应用程序
  addToTailscaleAcl: true   # 自动添加到 Tailscale 的 ACL 列表中
```
:::

### exposePort
- 类型： `int`
- 可选
- 有效值： `0-65535`,保留端口 `22`, `80`, `81`, `443`, `444`, `2379`, `18088` 除外
Olares 会为你的应用暴露指定的端口，这些端口可通过应用域名在本地网络下访问，如`84864c1f.your_olares_id.olares.com:46879`。对于每个公开的端口，Olares 会自动配置相同端口号的 TCP 和 UDP。

:::info 提示
暴露的端口只能通过本地网络或 Olares 专用网络访问。
:::

### protocol
- 类型： `string`
- 可选
- 有效值： `udp`、`tcp`

暴露端口使用的协议 ，如果不填默认同时开通udp和tcp。

### addToTailscaleAcl
- 类型： `boolean`
- 可选
- 默认值：`false`

当将 addToTailscaleAcl 字段设置为 true 时，系统会为该端口分配一个随机端口，并自动将其加入到 Tailscale 的 ACL 中。

## tailscale
- 类型：`map`
- 可选

允许应用在 Tailscale 的 ACL(Access Control Lists) 中开放指定端口。

:::info 示例
```yaml
tailscale:
  acls:
  - proto: tcp
    dst:
    - "*:46879"
  - proto: "" # 可选，如果未指定，则允许使用所有支持的协议
    dst:
    -  "*:4557"
```
:::

## permission

### appCache

- 类型： `boolean`
- 默认值：`false`
- 可选

应用是否需要在 `Cache` 目录创建应用的目录。设置为 `true` 时，应用在部署 YAML 中可通过 `.Values.userspace.appCache` 获取 `Cache` 目录的路径。

### appData

- 类型： `boolean`
- 默认值：`false`
- 可选

应用是否需要 `Data` 目录创建应用的目录。设置为 `true` 时，应用在部署 YAML 中可通过 `.Values.userspace.appData` 获取 `Data` 目录的路径。

### appCommon

- 类型： `boolean`
- 默认值：`false`
- 可选

应用是否需要 `App Common` 目录的读写权限。设置为 `true` 时，应用在部署 YAML 中可通过 `.Values.userspace.appCommon` 获取 App Common 目录的路径。从而在多个应用或节点间共享的文件（如 AI 模型等）。

### externalData

- 类型： `boolean`
- 默认值：`false`
- 可选

应用是否需要 `External` 目录（通常用于访问挂载的 NAS 或其他外部磁盘数据）的读写权限。设置为 `true` 时，应用在部署 YAML 中可通过 `.Values.sharedlib` 获取 `External` 目录的路径。

:::warning
从 Manifest 0.12.0 版本开始，`externalData` 权限默认不会开启。如果您的应用需要访问挂载的 NAS 或外部磁盘，必须显式声明 `externalData: true`。若未声明，将无法获得该权限。
:::

### userData

- 类型： `list<string>`
- 可选

应用是否需要用户 `Home` 文件夹中的特定目录的读写权限。请在此字段列出应用需访问的所有 `Home` 子目录。部署 YAML 文件中，可通过 `.Values.userspace.userData` 获取 Home 目录的根路径。所有需要挂载的 Home 子目录，必须在本字段中明确声明后方可访问。


## spec
记录额外的应用信息

:::info 示例
```yaml
spec:
  versionName: '10.8.11'
  # 此 Chart 包含的应用程序的版本。建议将版本号括在引号中。该值对应于 Chart.yaml 文件中的 appVersion 字段。请注意，它与 version 字段无关。

  featuredImage: https://app.cdn.olares.com/appstore/jellyfin/promote_image_1.jpg
  # 当应用安装后，会在 My Olares 页展示此图片。

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
  # 列出该应用商店页支持的语言和地区
  
  accelerator:
  - mode: nvidia
    limitedCpu: 7000m
    requiredCpu: 150m
    requiredDisk: 50Mi
    limitedDisk: 500Gi
    limitedMemory: 40Gi
    requiredMemory: 5Gi
    requiredGPUMemory: 1Gi
    limitedGPUMemory: 24Gi
  - mode: cpu
    limitedCpu: 7000m
    requiredCpu: 150m
    requiredDisk: 50Mi
    limitedDisk: 500Gi
    limitedMemory: 40Gi
    requiredMemory: 5Gi
  # 如果应用需要使用 GPU 或其他硬件加速设备，请在此处详细列出所有支持的加速模式及相关资源需求。仅使用 CPU 的场景，请填 mode: cpu, 下方配置与之前版本一致。
  
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
```yaml
metadata:
  description:
  title:
spec:
  fullDescription:
  upgradeDescription:
```

### supportArch
- 类型： `list<string>`
- 有效值： `amd64`, `arm64`
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
- 类型： `boolean`
- 默认值： `false`
- 可选

设置为 `true` 时，只有管理员可以安装此应用程序。

### runAsUser
- 类型： `boolean`
- 可选

当设置为 `true` 时，Olares 会强制以用户 ID "1000"（非 root 用户）运行应用程序。

### accelerator
- 类型： `map`
- 可选

声明应用所需的加速计算资源（如 GPU，核显等）。对于需要加速能力的应用（如大语言模型、图像生成、视频处理或人工智能模型服务等），应该使用 `spec.accelerator` 字段声明资源，不能包含原来的 `spec.requiredMemory` 等字段。

:::info 提示
当使用 accelerator 字段时，所有相关的资源需求（CPU、内存、磁盘、GPU 显存等）都必须在该字段内部进行声明，而不能再依赖于 spec.requiredMemory 或 spec.requiredCpu 这些根级别字段。
:::

支持的模式

- `nvidia`、`nvidia-gb10`、`apple-m`、`strix-halo`、`intel`、`amd`、`intel-gpu`、`amd-gpu`：适用于支持特定的 GPU/NPU 硬件加速的应用。
- `cpu`：用于常规应用，或在支持加速的应用中强制只使用 CPU 计算。

:::info 示例
```yaml
spec:
  accelerator:
  - mode: nvidia  # 支持的 mode：nvidia、nvidia-gb10、apple-m、strix-halo、mthreads-m1000、intel、amd、intel-gpu、amd-gpu、cpu
    limitedCpu: 7000m
    requiredCpu: 150m
    requiredDisk: 50Mi
    limitedDisk: 500Gi
    limitedMemory: 40Gi
    requiredMemory: 5Gi
    requiredGPUMemory: 1Gi
    limitedGPUMemory: 24Gi
```
:::

不需要 GPU 的应用，建议使用 cpu mode 来声明：
:::info 示例
```yaml
spec:
  accelerator:
  - mode: cpu
    requiredMemory: 2Gi
    requiredDisk: 50Mi
    requiredCpu: 0.25
    limitedMemory: 10240Mi
    limitedCpu: '4'
```
:::

## workloadReplicas
- 类型： `map`

声明 chart 中所有 workload 的副本数。从 0.12.0 开始，应用需要在`workloadReplicas` 中声明 chart 中所有 workload 的副本数量。

开发者还需要在 `values.yaml` 文件下对应的变量，并保证在 `workloadReplicas` 中指定的 workload 名称必须与 `values.yaml` 文件下 `workloads` 中的 workload 名称完全一致。从而保证兼容性。
:::info OlaresManifest.yaml
```yaml 
workloadReplicas:
  affine: 1
```
:::
:::info values.yaml
```yaml
workloads:
  affine:
    replicaCount: 1
```
:::

部署时，请确保引用该值来设置每个 workload 的副本数量。
:::info deployment.yaml
```yaml 
spec:
  replicas: {{ .Values.workloads.affine.replicaCount }}
```
:::

## middleware
- 类型：`map`
- 可选

系统提供了高可用的中间件服务，开发者无需重复安装中间件，只需在此填写对应的中间件信息即可，然后可以直接使用应用程序的 deployment YAML 文件中相应的中间件信息。

使用 `scripts` 字段指定创建数据库后应执行的脚本。此外，使用 `extension` 字段在数据库中添加相应的扩展名。

:::info 提示
MongoDB、MySQL、MariaDB、MinIO、RabbitMQ 需要管理员从 Market 安装后才能被其他应用使用
:::

### PostgreSQL
:::info 示例
```yaml
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
```
:::
使用 deployment YAML 中的中间件信息：
```yaml
# 对于 PostgreSQL，对应值如下
- env:
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
```

### Redis
:::info 示例
```yaml
middleware:
  redis:
    password: password
    namespace: db0
```
:::
使用 deployment YAML 中的中间件信息：
```yaml
# 对于 Redis，对应的值如下
host --> {{ .Values.redis.host }}For Redis, the corresponding value is as follow
port --> "{{ .Values.redis.port }}"
password --> "{{ .Values.redis.password }}"
```
### MongoDB
:::info 示例
```yaml
middleware:
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
# 对于 MongoDB，对应的值如下
host --> {{ .Values.mongodb.host }}
port --> "{{ .Values.mongodb.port }}"  # yaml 文件中的端口和密码需要用双引号括起来。
username --> {{ .Values.mongodb.username }}
password --> "{{ .Values.mongodb.password }}" # yaml 文件中的端口和密码需要用双引号括起来。
databases --> "{{ .Values.mongodb.databases }}" # 数据库的值类型是 map。你可以使用 {{ .Values.mongodb.databases.<dbname> }} 获取数据库。 <dbname> 是你在 OlaresManifest 中配置的名称，在 middleware.mongodb.databases[i].name 中指定
```
### MinIO
:::info 示例
```yaml
middleware:
  minio:
    username: miniouser
    buckets:
    - name: mybucket
```
:::
使用 deployment YAML 中的中间件信息：
```yaml
# 对于 MinIO，对应的值如下
- env:
  - name: MINIO_ENDPOINT
    value: '{{ .Values.minio.host }}:{{ .Values.minio.port }}'
  - name: MINIO_PORT
    value: "{{ .Values.minio.port }}"
  - name: MINIO_ACCESS_KEY
    value: {{ .Values.minio.username }}
  - name: MINIO_SECRET_KEY
    value: {{ .Values.minio.password }}
  - name: MINIO_BUCKET
    value: {{ .Values.minio.buckets.mybucket }}
```
### RabbitMQ
:::info 示例
```yaml
middleware:
  rabbitmq:
    username: rabbitmquser
    vhosts:
    - name: aaa
```
:::
使用 deployment YAML 中的中间件信息：
```yaml
# 对于 RabbitMQ，对应的值如下
- env:
  - name: RABBITMQ_HOST
    value: '{{ .Values.rabbitmq.host }}'
  - name: RABBITMQ_PORT
    value: "{{ .Values.rabbitmq.port }}"
  - name: RABBITMQ_USER
    value: "{{ .Values.rabbitmq.username }}"
  - name: RABBITMQ_PASSWORD
    value: "{{ .Values.rabbitmq.password }}"
  - name: RABBITMQ_VHOST
    value: "{{ .Values.rabbitmq.vhosts.aaa }}"

user := os.Getenv("RABBITMQ_USER")
password := os.Getenv("RABBITMQ_PASSWORD")
vhost := os.Getenv("RABBITMQ_VHOST")
host := os.Getenv("RABBITMQ_HOST")
portMQ := os.Getenv("RABBITMQ_PORT")
url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, portMQ, vhost)
```
### MariaDB
:::info 示例
```yaml
middleware:
  mariadb:
    username: mariadbclient
    databases:
    - name: aaa
```
:::
使用 deployment YAML 中的中间件信息：
```yaml
# 对于 MariaDB，对应的值如下
- env:
  - name: MDB_HOST
    value: '{{ .Values.mariadb.host }}'
  - name: MDB_PORT
    value: "{{ .Values.mariadb.port }}"
  - name: MDB_USER
    value: "{{ .Values.mariadb.username }}"
  - name: MDB_PASSWORD
    value: "{{ .Values.mariadb.password }}"
  - name: MDB_DB
    value: "{{ .Values.mariadb.databases.aaa }}"
```
### MySQL
:::info 示例
```yaml
middleware:
  mysql:
    username: mysqlclient
    databases:
    - name: aaa
```
:::
使用 deployment YAML 中的中间件信息：

```yaml
# 对于 MySQL，对应的值如下
- env:
  - name: MDB_HOST
    value: '{{ .Values.mysql.host }}'
  - name: MDB_PORT
    value: "{{ .Values.mysql.port }}"
  - name: MDB_USER
    value: "{{ .Values.mysql.username }}"
  - name: MDB_PASSWORD
    value: "{{ .Values.mysql.password }}"
  - name: MDB_DB
    value: "{{ .Values.mysql.databases.aaa }}"
```

## options

此部分用于配置与Olares系统相关的选项。

### policies
- 类型：`list<map>`
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

### conflicts
- 类型：`list<map>`
- 可选

请在此处声明与该应用冲突的其他应用。必须卸载冲突应用后才能安装此应用。

:::info 示例
```yaml
options:
  conflicts:
  - name: comfyui
    type: application
  - name: comfyuiclient
    type: application
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

指定 API 提供程序的超时限制(以秒为单位)。默认值为 `15`。使用 `0` 允许无限制的 API 连接。

:::info 示例
```yaml
apiTimeout: 0
```
:::

### allowedOutboundPorts
- 类型： `list<int>`
- 可选

要求开通以下端口进行非 HTTP 协议的对外访问，例如 SMTP 服务等。

:::info 示例
```yaml
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

### shared
- 类型： `boolean`
- 默认值： `false`
- 可选

设置为 `true` 时，表示该应用为共享应用。应用将被安装在`<appname>-shared`命名空间下。 

当 `options.shared: true` 时：
- `apiVersion` 必须设置为 `'v3'`
- `onlyAdmin` 必须设置为 `true`
- 不允许包含以下 `apiVersion: 'v3'` 字段：
  ```yaml
  spec:
    subCharts:
    - name: ollamaserver
      shared: true
    - name: ollamav2
  options:
    appScope:
      clusterScoped: true
      appRef:
      - ollamav2
  ```

### templateOnly
- 类型： `boolean`
- 默认值： `false`
- 可选

设置为 `true` 时，应用会被标记为模板应用，无法直接安装，必须先创建一个实例后才能使用。由于模板应用主要用于生成多个实例，因此 `allowMultipleInstall` 也必须设为 `true`。

### LLMGatewaySupported
- 类型： `boolean`
- 默认值： `false`
- 可选

设置为 `true` 时，表示应用支持 LLM Gateway 调用。主要用于模型相关的应用服务。

### overlayGateway
- 类型： `map`
- 可选

声明应用支持 L2 overlay 局域网发现。开启后，其他应用可通过 IP 地址在局域网内访问该应用。
:::info 示例
```yaml
overlayGateway:
  enable: true  # 开启 L2 overlay 局域网发现，默认为 false
  entrances:
  - port: 8096  # Overlay 监听端口号
    title: Jellyfin  # Overlay 入口名称
    workload: jellyfin  # Workload 的名字，oac 需要校验
    description: "Access Jellyfin using IP address in LAN"
    protocol: tcp  # 支持的协议：tcp/udp，不填默认 tcp/udp 都支持
```
:::

## envs

在此声明应用运行所需的环境变量，既支持用户手动输入，也可以直接引用已有的系统环境变量值。

:::info 提示
该配置需要 Olares OS 版本在 1.12.2 及以上才生效
:::

:::info 示例
```yaml
envs:
  - envName: ENV_NAME
    # 在部署应用时，该键会被注入为.Values.olaresEnv.ENV_NAME

    required: true
    # 是否为必填项
    # 若为true且未设置default，则用户安装应用时必须填写此值，且修改value时不允许清空

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
    # title为展示给用户的名称，value为实际注入系统的值

    remoteOptions: https://xxx.xxx/xx
    # 提供允许值列表的一个url，response body需为JSON编码的options列表

    multiSelect: true
    splitter: ","
    # 当 options 或 remoteOptions 字段为 true 时可启用，允许用户进行多项选择。所有选择项会通过 splitter 字符拼接为字符串传入应用。

    regex: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2，}$'
    # 该环境变量的值必须匹配此正则表达式

    valueFrom:
      envName: OLARES_SYSTEM_CLUSTER_DNS_SERVICE
    # 引用系统环境变量的值。如果采用该方式，将不允许用户手动指定/修改其value
    # 引用后，此环境变量的可声明字段(type，editable)将被系统环境变量的对应属性覆盖，default/value字段也会失效

    applyOnChange: true
    # 是否在此环境变量的值变化时自动重新部署应用，使变化生效
    # 若该字段为false，在环境变量变化时，即使停止/启动应用，也不会生效，只有升级/重装会生效

    description: "DESCRIPTION"
    # 对环境变量的描述
```
:::

如需在部署 YAML 文件中使用环境变量的值，只需在相应位置使用 `.Values.olaresEnv.ENV_NAME` 即可。系统会在应用部署时自动将对应的 olaresEnv 变量注入到 values 中。例如

:::info 示例
```yaml
BACKEND_MAIL_HOST: "{{ .Values.olaresEnv.MAIL_HOST }}"
BACKEND_MAIL_PORT: "{{ .Values.olaresEnv.MAIL_PORT }}"
BACKEND_MAIL_AUTH_USER: "{{ .Values.olaresEnv.MAIL_AUTH_USER }}"
BACKEND_MAIL_AUTH_PASS: "{{ .Values.olaresEnv.MAIL_AUTH_PASS }}"
BACKEND_MAIL_SECURE: "{{ .Values.olaresEnv.MAIL_SECURE }}"
BACKEND_MAIL_SENDER: "{{ .Values.olaresEnv.MAIL_SENDER }}"
```
:::

## 废弃字段（0.12.0）

以下字段已在 OlaresManifest 0.12.0 版本中废弃，请务必及时移除。若安装包中包含这些字段，应用在安装时可能会被**拒绝通过**。

| 字段 | 修改方式 |
|------|----------|
| `metadata.appid` | 移除，会自动根据 metadata.name 创建 |
| `provider`（顶层） | 移除，改为使用 authLevel: interal 的 entrance 供其他应用调用 |
| `permission.provider` | 移除 |
| `permission.sysData` | 移除 |
| `options.appScope` | 移除，使用 `apiVersion: 'v3'` 格式声明共享应用 |
| `spec.subCharts` | 移除，使用 `apiVersion: 'v3'` 格式声明共享应用 |
| `spec.requiredMemory`等 | 修改，请在 `spec.accelerator` 的 `mode: cpu` 配置下声明 |
| OS 1.11 分类值：`Blockchain`、`Utilities`、`Social Network`、`Entertainment`、`Productivity` | 使用 OS 1.12 分类值 |
