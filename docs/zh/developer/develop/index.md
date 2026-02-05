# 开发 Olares 应用

Olares 应用开发依托于通用的 Web 技术与容器化方案。如果你熟悉 Web 应用构建或 Docker 容器技术，即可直接上手开发 Olares 应用，无需额外的技能储备。

本指南涵盖了 Olares 应用开发的全生命周期，从在 Studio 中编写第一行代码开始，一直到在 Market 上完成发布。

## 准备工作
正式开始前，建议先了解以下核心概念：
- [Application](../concepts/application.md)
- [Network](../concepts/network.md)

## 第一步：使用 Studio 开发
Olares Studio 是专为加速应用开发周期而设计的开发平台。它提供了预配置的工作区，支持直接在平台上构建、调试和测试应用。

* **[部署应用](./tutorial/deploy.md)**： 学习如何基于现有 Docker 镜像快速部署应用，并在 Studio 中完成配置与测试。
* **[使用开发容器](./tutorial/develop.md)**： 创建远程开发环境 (Dev Container) 并连接至 VS Code，实现流畅的编码体验。
* **[打包与上传](./tutorial/package-upload.md)**： 将运行中的应用转换为 Olares 兼容的安装包，并上传进行测试。
* **[添加应用素材](./tutorial/assets.md)**： 配置图标、截图和应用描述，完善应用上架准备工作。

## 第二步：打包应用
发布应用到 Olares Market 之前，需要按照 Olares Application Chart (OAC) 规范来组织应用结构。该格式在 Helm Charts 的基础上进行了扩展，增加了权限管理和沙箱机制等 Olares 专属特性。

* **[了解 Olares Application Chart](./package/chart.md)**： 了解应用包的文件结构与具体规范。
* **[了解 `OlaresManifest.yaml`](./package/manifest.md)**： 全面介绍 `OlaresManifest.yaml` 文件，该文件用于定义应用的元数据、权限及与 Olares 系统对接的各项配置。
* **[了解 Helm 扩展](./package/extension.md)**： 了解 Olares 在标准 Helm 部署之上增加的自定义字段与功能。

## 第三步：提交应用
应用构建并打包完毕后，即可将其分享给 Olares 社区。

* **[提交至 Market](/zh/developer/develop/distribute-index.md)**： 了解如何将应用提交至 Olares Market 以进行审核与分发。