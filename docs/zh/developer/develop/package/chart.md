# Olares 应用 Chart 包的结构

Olares 应用 Chart 基于 Helm Chart 的基础结构，扩展 Olares 特有信息，主要为：
```
|-- Chart.yaml                   # chart 的 metadata
|-- OlaresManifest.yaml          # Olares 应用的配置
|-- templates                    # chart 安装部署模版文件
|   |-- deployment.yaml          # 应用部署脚本
|-- values.yaml                  # chart 安装部署参数
```
:::info 注意
为了使 `templates` 目录更易于理解，你可以将部署拆分为多个文件。
:::

- 应用 Chart 包示例：
```
AppName
|-- Chart.yaml                # 必选: 包含了 chart 信息的 YAML文件
|-- OlaresManifest.yaml       # 必选: 应用的配置文档
|-- values.yaml               # 必选: chart 默认的配置值
|-- templates                 # 必选: 模板目录， 当和 values 结合时，可生成有效的 Kubernetes manifest 文件
|   |-- NOTES.txt             # 可选: 包含简要使用说明的纯文本文件
|   |-- deployment.yaml       # 定义应用安装的 Deployment
|   |-- service.yaml          # 定义应用提供 Entrance 的 Service
|   |-- provider.yaml         # 可选：如果需要暴露 Provider 接口
|-- LICENSE                   # 可选: 包含 chart 许可证的纯文本文件
|-- README.md                 # 可选: 可读的 README 文件
```

- 推算算法 Chart 包示例：

```
RecommendName
|-- Chart.yaml                # 必选: 包含了 chart 信息的 YAML 文件
|-- OlaresManifest.yaml     # 必选: 推荐算法的配置文档
|-- values.yaml               # 必选: chart 默认的配置值
|-- templates                 # 必选: 模板目录， 当和 values 结合时，可生成有效的 Kubernetes manifest 文件
|   |-- NOTES.txt             # 可选: 包含简要使用说明的纯文本文件
|   |-- train.yaml            # 定义推荐算法 workflows 中的 train 流程
|   |-- prerank.yaml          # 定义推荐算法 workflows 中的 prerank 流程
|   |-- rank.yaml             # 定义推荐算法 workflows 中的 rank 流程
|   |-- embedding.yaml        # 定义推荐算法 workflows 中的 embedding 流程
|-- LICENSE                   # 可选: 包含 chart 许可证的纯文本文件
|-- README.md                 # 可选: 可读的 README 文件

```