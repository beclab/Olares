---
outline: [2, 3]
description: 系统环境变量是集群级全局配置，由管理员维护。应用需通过 envs.valueFrom 引用并映射到 .Values.olaresEnv。
---

# 系统环境变量参考

系统环境变量是每个 Olares 集群实例的全局配置。它们在系统安装时设定，或由管理员在系统中维护。集群内所有用户共享同一套系统环境变量。

:::info 信息
系统环境变量是“变量池”，应用不能直接修改它们，由管理员维护。应用需通过 `envs.valueFrom` 将其映射为自己的 `envName`，再在模板中通过 `.Values.olaresEnv.<envName>` 使用。
:::

## 使用方法

以下示例以 `APP_CDN_ENDPOINT` 为例，演示如何在应用中引用系统环境变量。

1. 在 `OlaresManifest.yaml` 中声明映射关系。在 `envs` 部分声明一个应用变量，并通过 `valueFrom` 引用系统环境变量。 

    **示例**：
    ```yaml
    olaresManifest.version: '0.10.0'
    olaresManifest.type: app

    envs:
      - envName: APP_CDN_ENDPOINT
        required: true
        applyOnChange: true
        valueFrom:
          envName: OLARES_SYSTEM_CDN_SERVICE
    ```

2. 在 Helm 模板中，在需要使用该变量的位置，通过 `.Values.olaresEnv` 路径进行引用。

    **示例**：
    ```yaml
    value: "{{ .Values.olaresEnv.APP_CDN_ENDPOINT }}"
    ```

3. 应用通过 App Service 部署时，系统会自动获取当前用户的环境配置值，并注入到 `values.yaml` 中。

    **示例输出**：

    ```yaml
    olaresEnv:
      APP_CDN_ENDPOINT: "https://cdn.olares.com"
    ```

## 系统环境变量列表

### OLARES_SYSTEM_REMOTE_SERVICE
Olares 系统远程服务地址（如应用商店、Olares Space 等）

- 类型：`url`
- 默认值：`https://api.olares.com`
- 是否可编辑：是
- 是否必填：是

### OLARES_SYSTEM_CDN_SERVICE
系统资源 CDN 地址
- 类型：`url`
- 默认值：`https://cdn.olares.com`
- 是否可编辑：是
- 是否必填：是

### OLARES_SYSTEM_DOCKERHUB_SERVICE
Docker Hub 镜像加速地址
- 类型：`url`
- 是否可编辑：是
- 是否必填：否

### OLARES_SYSTEM_ROOT_PATH
Olares 根目录地址
- 类型：`string`
- 默认值：`/olares`
- 是否可编辑：否
- 是否必填：是

### OLARES_SYSTEM_ROOTFS_TYPE
Olares 文件系统类型
- 类型：`string`
- 默认值：`fs`
- 是否可编辑：否
- 是否必填：是

### OLARES_SYSTEM_CUDA_VERSION
宿主机 CUDA 版本
- 类型：`string`
- 是否可编辑：否
- 是否必填：否

## 完整变量结构示例

```yaml
systemEnvs:
  - envName: OLARES_SYSTEM_REMOTE_SERVICE
    default: "https://api.olares.com"
    type: url
    editable: true
    required: true
    
  - envName: OLARES_SYSTEM_CDN_SERVICE
    default: "https://cdn.olares.com"
    type: url
    editable: true
    required: true

  - envName: OLARES_SYSTEM_DOCKERHUB_SERVICE
    type: url
    editable: true
    required: false

  - envName: OLARES_SYSTEM_ROOT_PATH
    default: /olares
    editable: false
    required: true

  - envName: OLARES_SYSTEM_ROOTFS_TYPE
    default: fs
    editable: false
    required: true

  - envName: OLARES_SYSTEM_CUDA_VERSION
    editable: false
    required: false
```