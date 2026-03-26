---
outline: [2, 3]
description: 了解 Olares 应用在部署期间的变量注入机制：声明式环境变量（.Values.olaresEnv）与系统自动注入的运行时 Helm Values（.Values.*）。
---

# 环境变量概览

Olares 应用通过 app-service 将运行时信息与配置项注入到应用的 `values.yaml` 中。应用在 Helm 模板中通过 `.Values.*` 引用这些值。

:::info 变量与 Helm 值
本文提到的“变量”主要指 Helm 值。它们不会自动进入容器环境变量。如需在容器内使用，请在模板中显式映射到 `env:`。
:::

## 变量注入通道

Olares 通过两种通道注入变量：

- **声明式环境变量**：开发者在 `OlaresManifest.yaml` 的 `envs` 下声明变量。在部署时，app-service 会解析这些值并将其注入到 `values.yaml` 的 `.Values.olaresEnv` 路径下。
- **系统注入的运行时变量**：由 Olares 在部署时自动注入，无需声明。不过，某些值（例如中间件）只有在声明相关依赖后才可用。

## 下一步

1. [声明式环境变量](app-env-vars.md)：`envs` 字段说明、变量映射以及变量参考。
2. [系统注入的运行时变量](app-sys-injected-variables.md)：所有系统注入运行时变量的完整参考。