---
outline: [2, 3]
description: 用户环境变量是用户级个性化配置。应用需通过 envs.valueFrom 引用并映射到 .Values.olaresEnv。
---

# 用户环境变量参考

用户环境变量是每个 Olares 用户的个性化配置，由当前用户自行管理。在同一集群中，不同用户的用户环境变量彼此独立。

:::info 信息
用户环境变量是“变量池”。应用需通过 `envs.valueFrom` 映射为自己的 `envName`，再在模板中通过 `.Values.olaresEnv.<envName>` 使用。
:::

## 使用方法

以下示例以 `OLARES_USER_TIMEZONE` 为例演示如何在应用中引用用户环境变量。

1. 在 `OlaresManifest.yaml` 中声明映射关系。在 `envs` 部分声明一个应用变量，并通过 `valueFrom` 引用用户环境变量。

    **示例**：
    ```yaml
    olaresManifest.version: '0.10.0'
    olaresManifest.type: app

    envs:
      - envName: USER_TIMEZONE
        valueFrom:
          envName: OLARES_USER_TIMEZONE
    ```

2. 在 Helm 模板中，在需要使用该变量的位置，通过 `.Values.olaresEnv` 路径进行引用。

    **示例**：
    ```yaml
    value: "{{ .Values.olaresEnv.USER_TIMEZONE }}"
    ```
3. 应用通过 App Service 部署时，系统会自动获取当前用户的环境配置值，并注入到 `values.yaml` 中。

    **示例输出**：

    ```yaml
    olaresEnv:
      USER_TIMEZONE: "Asia/Shanghai"
    ```

## 用户环境变量列表

### 用户信息

#### OLARES_USER_EMAIL

- 类型：`string`
- 是否可编辑：是

#### OLARES_USER_USERNAME

- 类型：`string`
- 是否可编辑：是

#### OLARES_USER_PASSWORD

- 类型：`password`
- 是否可编辑：是

#### OLARES_USER_TIMEZONE

- 类型：`string`
- 是否可编辑：是

### SMTP 设置

#### OLARES_USER_SMTP_ENABLED
是否开启 SMTP
- 类型：`bool`
- 是否可编辑：是

#### OLARES_USER_SMTP_SERVER
SMTP服务商的域名
- 类型：`domain`
- 是否可编辑：是

#### OLARES_USER_SMTP_PORT
SMTP服务端口，通常是"465"或"587"
- 类型：`int`
- 是否可编辑：是

#### OLARES_USER_SMTP_USERNAME
SMTP服务商的用户名
- 类型：`string`
- 是否可编辑：是

#### OLARES_USER_SMTP_PASSWORD
SMTP服务商的密码或授权码
- 类型：`password`
- 是否可编辑：是

#### OLARES_USER_SMTP_FROM_ADDRESS
发送邮件的地址
- 类型：`email`
- 是否可编辑：是

#### OLARES_USER_SMTP_SECURE
是否使用安全协议
- 类型：`bool`
- 默认值：`true`
- 是否可编辑：是

#### OLARES_USER_SMTP_USE_TLS
安全协议类型
- 类型：`bool`
- 是否可编辑：是

#### OLARES_USER_SMTP_USE_SSL
安全协议类型
- 类型：`bool`
- 是否可编辑：是

#### OLARES_USER_SMTP_SECURITY_PROTOCOLS
安全协议，可选值 `tls`, `ssl`, `starttls`, `none`
- 类型：`string`
- 是否可编辑：是

### 镜像 / 代理地址配置

#### OLARES_USER_HUGGINGFACE_SERVICE

- 类型：`url`
- 默认值：`"https://huggingface.co/"`
- 是否可编辑：是

#### OLARES_USER_HUGGINGFACE_TOKEN

- 类型：`string`
- 是否可编辑：是

#### OLARES_USER_PYPI_SERVICE

- 类型：`url`
- 默认值：`"https://pypi.org/simple/"`
- 是否可编辑：是

#### OLARES_USER_GITHUB_SERVICE

- 类型：`url`
- 默认值：`"https://github.com/"`
- 是否可编辑：是

#### OLARES_USER_GITHUB_TOKEN

- 类型：`string`
- 是否可编辑：是

### API Key 设置

#### OLARES_USER_OPENAI_APIKEY

- 类型：`password`
- 是否可编辑：是

#### OLARES_USER_CUSTOM_OPENAI_SERVICE

- 类型：`url`
- 是否可编辑：是

#### OLARES_USER_CUSTOM_OPENAI_APIKEY

- 类型：`password`
- 是否可编辑：是

## 完整变量结构示例

```yaml
userEnvs:
# 用户信息
  - envName: OLARES_USER_EMAIL
    type: string
    editable: true
  - envName: OLARES_USER_USERNAME
    type: string
    editable: true
  - envName: OLARES_USER_PASSWORD
    type: password
    editable: true
  - envName: OLARES_USER_TIMEZONE
    type: string
    editable: true    

# SMTP 设置
    
  - envName: OLARES_USER_SMTP_ENABLED
    type: bool
    editable: true
    
  - envName: OLARES_USER_SMTP_SERVER
    type: domain 
    editable: true
    
  - envName: OLARES_USER_SMTP_PORT
    type: int
    editable: true
    
  - envName: OLARES_USER_SMTP_USERNAME
    type: string
    editable: true
    
  - envName: OLARES_USER_SMTP_PASSWORD
    type: password
    editable: true
    
  - envName: OLARES_USER_SMTP_FROM_ADDRESS
    type: email
    editable: true    
   
  - envName: OLARES_USER_SMTP_SECURE
    type: bool
    editable: true
    default: "true"   
    
  - envName: OLARES_USER_SMTP_USE_TLS
    type: bool
    editable: true
  - envName: OLARES_USER_SMTP_USE_SSL
    type: bool
    editable: true
    
  - envName: OLARES_USER_SMTP_SECURITY_PROTOCOLS
    type: string
    editable: true        
    
# 镜像/代理地址配置
  - envName: OLARES_USER_HUGGINGFACE_SERVICE
    type: url
    editable: true
    default: "https://huggingface.co/"
  - envName: OLARES_USER_HUGGINGFACE_TOKEN
    type: string
    editable: true    
  - envName: OLARES_USER_PYPI_SERVICE
    type: url
    editable: true
    default: "https://pypi.org/simple/"
  - envName: OLARES_USER_GITHUB_SERVICE
    type: url
    editable: true
    default: "https://github.com/"
  - envName: OLARES_USER_GITHUB_TOKEN
    type: string
    editable: true

# API-KEY 设置
  - envName: OLARES_USER_OPENAI_APIKEY
    type: password
    editable: true
  - envName: OLARES_USER_CUSTOM_OPENAI_SERVICE
    type: url
    editable: true    
  - envName: OLARES_USER_CUSTOM_OPENAI_APIKEY
    type: password
    editable: true     
```