---
description: 在 Linux 系统（如 Ubuntu 或 Debian）上安装 Olares，以在生产环境中获得最佳的性能与稳定性。支持通过 ISO 镜像或一行命令完成部署
---
# 安装 Olares

本文介绍如何在 Linux 系统上安装 Olares。我们推荐在 **Linux 系统**（如 Ubuntu 或 Debian）上部署 Olares，以在生产环境中获得最佳的性能与稳定性。

开始安装前，请先[创建 Olares ID](create-olares-id.md)，并确认操作系统或硬件已满足所选方式的最低要求。

## 安装方式

根据你的硬件和使用场景，Olares 提供以下两种安装方式：

| 安装方式                   | 说明                                                                      | 适用场景                              |
|------------------------|-------------------------------------------------------------------------|-----------------------------------|
| [通过 ISO 镜像安装](.iso.md) | 从官方 ISO 启动镜像安装 Olares, <br/>自动配置宿主机环境（Linux），容器<br/>运行环境、必要驱动及核心依赖。 | 适合在物理机或虚拟化环境（如 PVE） 中部署使用 Olares。 |
| [通过一行命令安装](.script.md) | 在现有的 Linux 系统中执行一条命令<br/>安装 Olares。                                          | 适合在已有 Linux 环境上手动部署。              |


:::info 安装遇到问题？
如果安装过程中遇到问题，[可以提交 GitHub Issue](https://github.com/beclab/Olares/issues/new)。提交时请提供以下信息：
- 使用的平台或环境（如 Ubuntu、Docker、WSL 等）。
- 安装方式（脚本安装或 Docker 镜像）。
- 详细的错误信息（包括日志、错误提示或截图）。
  :::

:::tip 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台上安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::