# 开发 Olares 应用

Olares 应用开发依托于通用的 Web 技术与容器化方案。如果你熟悉 Web 应用构建或 Docker 容器技术，即可直接上手开发 Olares 应用，无需额外的技能储备。

本指南涵盖了 Olares 应用开发的全生命周期，从规划应用包结构开始，一直到在 Market 上完成发布。

## 准备工作
正式开始前，建议先了解以下核心概念：
- [Application](../concepts/application.md)
- [Network](../concepts/network.md)

## 第一步：打包应用
发布应用到 Olares Market 之前，需要按照 Olares Application Chart (OAC) 规范来组织应用结构。该格式在 Helm Charts 的基础上进行了扩展，增加了权限管理和沙箱机制等 Olares 专属特性。

* **[了解 Olares Application Chart](./package/chart.md)**： 了解应用包的文件结构与具体规范。
* **[了解 `OlaresManifest.yaml`](./package/manifest.md)**： 全面介绍 `OlaresManifest.yaml` 文件，该文件用于定义应用的元数据、权限及与 Olares 系统对接的各项配置。
* **[了解 Helm 扩展](./package/extension.md)**： 了解 Olares 在标准 Helm 部署之上增加的自定义字段与功能。

## 第二步：提交应用
应用构建并打包完毕后，即可将其分享给 Olares 社区。

* **[提交至 Market](/zh/developer/develop/distribute-index.md)**： 了解如何将应用提交至 Olares Market 以进行审核与分发。
