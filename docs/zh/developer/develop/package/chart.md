---
description: Olares 应用 Chart 基于标准 Helm Chart 结构，并扩展了 Olares 独有的信息。
---
# Olares 应用 Chart 结构

Olares 应用 Chart 基于标准 **Helm Chart** 结构，并扩展了 Olares 独有的信息。通常，`App` 和 `Middleware` 的标准应用 Chart 目录包含以下文件：

```
AppName
|-- Chart.yaml                   # chart 的元数据
|-- OlaresManifest.yaml          # Olares 应用专属配置
|-- templates/                   # 部署资源的模板
|   |-- deployment.yaml          # Deployment 资源定义
|-- owners                       # 提交到 Market 时必需；列出允许维护与更新此应用的 GitHub 账号
|-- crds/                        # 可选：Custom Resource Definitions
|-- values.yaml                  # 可选：此 chart 的默认部署参数
|-- values.schema.json           # 可选：用于约束 values.yaml 文件结构的 JSON Schema
|-- README.md                    # 可选：关于该应用的可读文档
```
:::info 注意
为了使 `templates` 目录更易于理解，你可以将部署拆分为多个文件。
:::
