---
outline: [2, 3]
description: 了解 Olares 应用在部署期间的变量注入机制：声明式环境变量（.Values.olaresEnv）与系统自动注入的运行时 Helm Values（.Values.*）。
---

# 环境变量概览

Olares 应用通过 App Service 将运行时信息与配置项注入到应用的 `values.yaml` 中（Helm Values）。应用在 Helm 模板中通过 `.Values.*` 引用这些值。

:::info 信息
本文提到的“变量”主要指 Helm values。它们不会自动进入容器环境变量。如需在容器内使用，请在模板中显式映射到 `env:`。
:::

## 变量注入通道

Olares 的变量注入分为两条通道：

1. **声明式环境变量**

    - 开发者在 `OlaresManifest.yaml` 的 `envs` 中声明变量。  
    - App Service 会在部署时把值注入到 `values.yaml` 的 `.Values.olaresEnv` 下。  
    - 模板引用：<code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>

2. **系统自动注入的运行时变量**  
   - 由系统在部署时直接注入到 `values.yaml` 的 `.Values` 根层级或指定子树下（例如 `.Values.postgres.*`）。  
   - 无需在 Manifest 中为“变量本身”声明（但部分变量仅在你声明依赖后才会注入，例如中间件）。  
   - 模板引用：<code v-pre>{{ .Values.* }}</code>（如 <code v-pre>{{ .Values.postgres.host }}</code>）

## 分类对比

| 类别 | 来源 | 声明方式 | 
| -- | --  | -- |
| **自定义环境变量** | 开发者业务配置 | `envs` 直接定义 | 
| **系统/用户环境变量映射值** | 系统变量池/用户变量池 | `envs.valueFrom` 引用后映射 |
| **预定义运行时变量** | 用户/系统/硬件/存储/依赖/中间件 | 系统自动注入（部分需声明依赖） | 

:::info 信息
系统环境变量与用户环境变量是“变量池”。应用必须通过 `envs.valueFrom` 映射到自己的 `envName`，才能在 `.Values.olaresEnv` 下使用。
:::

## 下一步

- 配置与管理：[自定义环境变量](app-env-vars.md)
- 引用变量池：
    - [系统环境变量](app-sys-env-vars.md)
    - [用户环境变量](app-user-env-vars.md)
- 直接使用注入值：[预定义运行时变量](runtime-values.md)