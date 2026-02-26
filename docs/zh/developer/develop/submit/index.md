---
outline: [2, 3]
description: 了解 Olares 中的应用分发机制。
---
# 分发 Olares 应用

Olares 的应用分发基于开放标准和自动化校验机制。如果你的应用已经打包为Olares 应用图表（OAC），即可发布到 Olares 应用市场，让用户能够轻松获取和安装。

本指南将带你了解 Olares 应用分发的生命周期，理解应用市场的索引机制，如何发布、维护与推广应用。

## 开始之前

在分发应用之前，先来了解以下核心概念：

- **[Olares 应用图表（OAC）](/zh/developer/develop/package/chart.md)**  
  用于描述 Olares 应用的打包格式，包含元数据、所有权、版本信息与安装配置等。

- **Olares 应用市场**  
  用于索引、展示、分发与安装 Olares 应用的市场平台。

- **应用索引**  
  向 Olares 应用市场提供应用元数据的服务。Olares 提供默认的公共索引服务，你也可以部署自己的索引服务。

- **GitBot**  
  自动验证系统，负责校验提交的应用并执行分发规则。

## 应用分发流程

要将应用提交到 Olares 默认索引，请按以下步骤操作：

1. 在 Olares 上开发并测试你的应用，并根据[指南](/zh/developer/develop/package/chart.md)创建 OAC。
2. Fork [官方 Olares 应用市场仓库 `beclab/apps`](https://github.com/beclab/apps)，并将你的应用 OAC 添加到你的 Fork 仓库中。
3. 创建一个指向 `beclab/apps:main` 的拉取请求（Pull Request，PR），并等待 GitBot 校验你的 OAC 是否符合要求。
4. 如有需要，根据反馈更新 PR，直到所有检查通过并自动合并。
5. 稍等片刻，你就可以在 Olares 应用市场中找到你提交的应用。