---
connectionVersion: 1.12.7
connectionLatestPath: /zh/use-cases/linkwarden
outline:
- 2
- 3
description: 在 Olares 上保存和整理实用网页
head:
- - meta
  - name: keywords
    content: Olares, Linkwarden, 自托管, 书签, 网页存档, 收藏夹
app_version: 0.0.3
doc_version: '1.0'
doc_updated: '2026-10-29'
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/linkwarden.md)为准。
:::

# 使用 Linkwarden 建立阅读资料库

Linkwarden 将链接归入收藏夹，并保存网页副本，便于日后阅读。在 Olares 上运行后，书签库和网页归档都保存在自己的设备中。本示例为一个家庭项目收集几篇参考资料。

## 学习目标

- 创建 Linkwarden 账号。
- 使用收藏夹和标签整理链接。
- 打开网页归档，并再次找到已保存的链接。

## 前提条件

- Olares 1.12.7 或更高版本。
- 已连接 [Olares 专用网络](../manual/larepass/private-network.md)的浏览器。

## 安装并登录

1. 打开应用市场，搜索并安装 **Linkwarden**。
2. 从启动台打开 Linkwarden。
3. 点击 **Sign up**，为当前实例创建账号，然后登录。

Olares 登录保护应用入口。在 Linkwarden 内创建的账号属于这个书签库。

## 创建收藏夹并保存链接

1. 创建名为 **Home projects** 的收藏夹。
2. 添加实用网页的链接，例如 `https://docs.olares.com/`。
3. 选择 **Home projects**，填写清晰的标题，并添加 `Olares` 等标签。
4. 保存链接，然后添加其他相关网页。

![Linkwarden 中的私有阅读收藏夹](/images/manual/use-cases/linkwarden-library.png#bordered)

## 阅读网页归档

通过已保存链接的归档控件打开 **Readable**。顶部栏可切换其他可用格式。如果浏览器无法预览 PDF，请下载后在本地打开。后台任务需要一定时间才能完成抓取。如果网站禁止自动访问，或要求单独登录，可能无法生成完整副本。

使用收藏夹、标签筛选，或搜索标题中的部分文字，再次找到需要的网页。

## 保留资料库

书签和账号记录保存在 Olares PostgreSQL 中。网页归档和搜索数据保存在应用的私有数据目录中。备份时应同时包含数据库和应用数据。

共享收藏夹会改变可以查看内容的用户范围。如果不打算分享，请保持收藏夹私有。浏览器扩展和 Olares 网络以外的访问需要另行配置。

## 了解更多

- [Linkwarden 文档](https://docs.linkwarden.app/)
