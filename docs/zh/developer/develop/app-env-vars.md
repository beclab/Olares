---
outline: [2, 3]
description: 通过 OlaresManifest.yaml 的 envs 声明与校验应用配置，并在模板中通过 .Values.olaresEnv 引用。
---

# 自定义环境变量

开发者可以在 `OlaresManifest.yaml` 的 `envs` 中声明应用所需的配置参数。部署或升级时，App Service 会将这些变量的最终值注入到 `values.yaml` 的 `.Values.olaresEnv` 下。

:::tip 注入提示
- 所有 `envs` 中声明的变量，模板引用路径统一为：<code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>
- 若 `values.yaml` 中存在同名字段，会被系统注入值覆盖。
:::

## 使用方法

### 在 Manifest 中声明

**示例**：

```yaml
olaresManifest.version: '0.10.0'
olaresManifest.type: app

envs:
  - envName: ADMIN_PASSWORD
    type: password
    required: true
    editable: false
    description: "Password must be at least 6 characters long"
    regex: '^[\w\-!@#$%^&*()+={}\[\]:,.?~]{6,}$'
```
### 在模板中引用

**示例**：
```yaml
env:
  - name: ADMIN_PASSWORD
    value: "{{ .Values.olaresEnv.ADMIN_PASSWORD }}"
```

## 字段说明

以下字段用于描述变量的取值来源、校验规则与可编辑性。

### applyOnChange

布尔值类型。设置为 `true` 表示该环境变量修改后会自动重启所有使用该变量的应用/组件，使其生效。

:::info 生效机制说明
当 `applyOnChange` 为 `false` 时，该变量变化后，即使手动停止/启动应用，变化也不会生效。仅在应用升级或重装后才会生效。
:::

### default

环境变量的默认值，开发者可在编写时提供，用户不可修改。当该变量未通过用户输入/引用（`valueFrom`）获得值时，系统使用 `default`。

### description

用于描述变量用途、有效值含义等。

### editable

布尔值类型。设置为 `true` 表示该环境变量允许在安装后修改。

### envName

注入到 `values.yaml` 的键名。模板引用方式：<code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>。

### options

允许值列表。变量的值只能从列表中选择，系统通常会以选择框形式让用户输入。

**示例**：
```yaml
envs:
  - envName: VERSION
    options:
      - title: "Windows 11 Pro"
        value: "iso/Win11_24H2_English_x64.iso"
      - title: "Windows 7 Ultimate"
        value: "iso/win7_sp1_x64_1.iso"
```

### regex

正则校验。仅符合 `regex` 的值才允许作为变量值。校验失败会导致设置失败或安装/升级失败。

### remoteOptions

通过 URL 提供 `options` 列表。响应 body 需为 JSON 编码的 `options` 数组。

**示例**：
```yaml
envs:
  - envName: VERSION
    remoteOptions: https://app.cdn.olares.com/appstore/windows/version_options.json
```

### required

布尔值类型。为 `true` 表示安装必需：
- 若未设置 `default`，安装应用前会提示用户输入
- 安装后该环境变量值不能改为空

### type

环境变量值的类型，用于校验值的类型。校验失败时设置失败。目前支持：

- `int`
- `bool`
- `url`
- `ip`
- `domain`
- `email`
- `string`
- `password`

### value

环境变量的值，不支持直接写死任意常量。值可以来自 `default`、用户输入，或通过` valueFrom` 引用其他变量。

### valueFrom

引用 Olares 系统环境变量或用户环境变量的值。

**示例**：
```yaml
envs:
  - envName: APP_CDN_ENDPOINT
    required: true
    applyOnChange: true
    valueFrom:
      envName: OLARES_SYSTEM_CDN_SERVICE
```

:::info 引用继承规则
使用 `valueFrom` 时，当前变量会继承被引用变量的全部属性字段（如 `type`、`editable`、`regex` 等），当前变量中同名属性将失效。并且此时 `default` 与 `options` 配置无效。
:::