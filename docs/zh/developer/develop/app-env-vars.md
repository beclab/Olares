---
outline: [2, 4]
description: 通过 `OlaresManifest.yaml` 中的 `envs` 声明并校验应用配置，并在模板中通过 `.Values.olaresEnv` 引用变量值。
---

# 声明式环境变量

在 `OlaresManifest.yaml` 中使用 `envs` 字段来声明配置参数，例如密码、API 端点或功能开关。在部署过程中，app-service 会解析这些变量，并将其注入到 `values.yaml` 的 `.Values.olaresEnv` 中。你可以在 Helm 模板中通过 <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code> 引用。

## 变量来源

声明式变量可以从应用外部管理的配置中获取值：

- **系统变量**：由管理员管理的集群级基础设施配置，例如 CDN 端点或根路径。
- **用户变量**：由用户个人管理的配置，例如时区、SMTP 设置或 API Key。

应用本身无法直接修改这些变量。如需使用，需通过 `valueFrom` 字段映射。

## 映射环境变量

系统环境变量和用户环境变量都通过 `valueFrom` 使用相同的映射机制。

以下示例演示了如何将系统变量 `OLARES_SYSTEM_CDN_SERVICE` 映射为应用变量 `APP_CDN_ENDPOINT`：

1. 在 `OlaresManifest.yaml` 中，在 `envs` 下声明一个应用变量，并将 `valueFrom.envName` 设置为系统变量名。

    ```yaml
    # 将系统变量 OLARES_SYSTEM_CDN_SERVICE 映射为应用变量 APP_CDN_ENDPOINT
    olaresManifest.version: '0.10.0'
    olaresManifest.type: app

    envs:
      - envName: APP_CDN_ENDPOINT
        required: true
        applyOnChange: true
        valueFrom:
          envName: OLARES_SYSTEM_CDN_SERVICE
    ```

2. 在 Helm 模板中，通过 `.Values.olaresEnv.<envName>` 引用该应用变量。

    ```yaml
    # 在容器环境变量中使用 APP_CDN_ENDPOINT
    env:
      - name: CDN_ENDPOINT
        value: "{{ .Values.olaresEnv.APP_CDN_ENDPOINT }}"
    ```

部署时，app-service 会解析引用的变量，并将值注入到 `values.yaml` 中：

```yaml
# 由 app-service 在部署时注入
olaresEnv:
  APP_CDN_ENDPOINT: "https://cdn.olares.com"
```

可用环境变量的完整列表，请参阅[变量参考](#变量参考)。

## 声明字段

`envs` 列表中的每个条目支持以下字段。

### envName

注入到 `values.yaml` 中的变量名。在模板中通过 <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code> 引用。

### value

变量解析后的最终值。不能直接将其设置为固定常量。该值来源于 `default`、用户输入或通过 `valueFrom` 引用的变量。

### default

变量的默认值。由开发者在编写应用时提供。用户不可修改。当用户未提供值或未通过 `valueFrom` 引用时使用。

### valueFrom

将当前变量映射到系统或用户环境变量。设置后，当前变量将继承所引用变量的所有字段（如 `type`、`editable`、`regex` 等）。当前变量本地定义的同名字段将被忽略。使用 `valueFrom` `时，default` 和 `options` 不生效。

**示例**： 将应用变量 `APP_CDN_ENDPOINT` 映射到系统变量 `OLARES_SYSTEM_CDN_SERVICE`。

```yaml
envs:
  - envName: APP_CDN_ENDPOINT
    required: true
    applyOnChange: true
    valueFrom:
      envName: OLARES_SYSTEM_CDN_SERVICE
```

### required

布尔值。为 `true` 时，安装必须提供该变量值。如果未设置 `default`，系统会提示用户输入。安装后该值不能设为空。

### editable

布尔值。为 `true` 时，该变量在安装后允许修改。

### applyOnChange

布尔值。为 `true` 时，修改该变量会自动重启使用该变量的应用或组件。为 `false` 时，修改仅在应用升级或重装后生效。手动停止和启动应用不会使其生效。

### type

值的预期类型。用于在接受输入前进行校验。支持的类型包括：`int`、`bool`、`url`、`ip`、`domain`、`email`、`string`、`password`。

### regex

值必须匹配的正则表达式。如果校验失败，则无法设置该值，且安装或升级可能失败。

### options

将变量限制为固定的可选值列表。系统会在界面中提供选择界面。

**示例**：显示标题与实际存储值不同的下拉菜单。

```yaml
envs:
  - envName: VERSION
    options:
      - title: "Windows 11 Pro"
        value: "iso/Win11_24H2_English_x64.iso"
      - title: "Windows 7 Ultimate"
        value: "iso/win7_sp1_x64_1.iso"
```

### remoteOptions

从 URL 加载选项列表，而不是在行内定义。响应体必须是与 `options` 格式相同的 JSON 编码数组。

**示例**：在安装时从远程端点获取选项列表。

```yaml
envs:
  - envName: VERSION
    remoteOptions: https://app.cdn.olares.com/appstore/windows/version_options.json
```

### description

变量用途及合法取值范围的说明。显示在 Olares 界面中。

## 变量参考

### 系统环境变量

下表列出了可通过 `valueFrom` 引用的系统级环境变量。

| 变量 | 类型 | 默认值 | 可编辑 | 必填 | 描述 |
| --- | --- | --- | --- | --- | --- |
| `OLARES_SYSTEM_REMOTE_SERVICE` | `url` | `https://api.olares.com` | `true` | `true` | Olares 远程服务端点，例如应用商店与 Olares Space。 |
| `OLARES_SYSTEM_CDN_SERVICE` | `url` | `https://cdn.olares.com` | `true` | `true` | 系统资源 CDN 端点。 |
| `OLARES_SYSTEM_DOCKERHUB_SERVICE` | `url` | 无 | `true` | `false` | Docker Hub 镜像或加速端点。 |
| `OLARES_SYSTEM_ROOT_PATH` | `string` | `/olares` | `false` | `true` | Olares 根目录路径。 |
| `OLARES_SYSTEM_ROOTFS_TYPE` | `string` | `fs` | `false` | `true` | Olares 文件系统类型。 |
| `OLARES_SYSTEM_CUDA_VERSION` | `string` | 无 | `false` | `false` | 主机 CUDA 版本。 |

### 用户环境变量

所有用户环境变量均可由用户编辑。

#### 用户信息

| 变量 | 类型 | 默认值 | 描述 |
| --- | --- | --- | --- |
| `OLARES_USER_EMAIL` | `string` | 无 | 用户邮箱地址。 |
| `OLARES_USER_USERNAME` | `string` | 无 | 用户名。 |
| `OLARES_USER_PASSWORD` | `password` | 无 | 用户密码。 |
| `OLARES_USER_TIMEZONE` | `string` | 无 | 用户时区，例如 `Asia/Shanghai`。 |

#### SMTP 设置

| 变量 | 类型 | 默认值 | 描述 |
| --- | --- | --- | --- |
| `OLARES_USER_SMTP_ENABLED` | `bool` | 无 | 是否启用 SMTP。 |
| `OLARES_USER_SMTP_SERVER` | `domain` | 无 | SMTP 服务器域名。 |
| `OLARES_USER_SMTP_PORT` | `int` | 无 | SMTP 服务端口，通常为 `465` 或 `587`。 |
| `OLARES_USER_SMTP_USERNAME` | `string` | 无 | SMTP 用户名。 |
| `OLARES_USER_SMTP_PASSWORD` | `password` | 无 | SMTP 密码或授权码。 |
| `OLARES_USER_SMTP_FROM_ADDRESS` | `email` | 无 | 发件人邮箱地址。 |
| `OLARES_USER_SMTP_SECURE` | `bool` | `"true"` | 是否使用安全协议。 |
| `OLARES_USER_SMTP_USE_TLS` | `bool` | 无 | 使用 TLS。 |
| `OLARES_USER_SMTP_USE_SSL` | `bool` | 无 | 使用 SSL。 |
| `OLARES_USER_SMTP_SECURITY_PROTOCOLS` | `string` | 无 | 安全协议类型，可选值包括：`tls`、`ssl`、`starttls`、`none`。 |

#### 镜像与代理端点

| 变量 | 类型 | 默认值 | 描述 |
| --- | --- | --- | --- |
| `OLARES_USER_HUGGINGFACE_SERVICE` | `url` | `https://huggingface.co/` | Hugging Face 服务地址。 |
| `OLARES_USER_HUGGINGFACE_TOKEN` | `string` | 无 | Hugging Face 访问令牌。 |
| `OLARES_USER_PYPI_SERVICE` | `url` | `https://pypi.org/simple/` | PyPI 镜像地址。 |
| `OLARES_USER_GITHUB_SERVICE` | `url` | `https://github.com/` | GitHub 镜像地址。 |
| `OLARES_USER_GITHUB_TOKEN` | `string` | 无 | GitHub 个人访问令牌。 |

#### API keys

| 变量 | 类型 | 默认值 | 描述 |
| --- | --- | --- | --- |
| `OLARES_USER_OPENAI_APIKEY` | `password` | 无 | OpenAI API key。 |
| `OLARES_USER_CUSTOM_OPENAI_SERVICE` | `url` | 无 | 自定义 OpenAI 兼容服务地址。 |
| `OLARES_USER_CUSTOM_OPENAI_APIKEY` | `password` | 无 | 自定义 OpenAI 兼容服务的 API key。 |