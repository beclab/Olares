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
app_version: "0.0.5"
doc_version: "1.2"
doc_updated: "2026-11-01"
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

- 让助手根据归档页面整理带来源的简报。

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

## 通过助手使用阅读资料库

Lares 或 Hermes Agent 可以通过社区 MCP 适配器搜索 Linkwarden，并读取已保存页面的正文。你可以先收集某个项目的资料，再让助手比较收藏内容，整理带来源链接的简报。

### 获取 API 令牌和服务地址

在 Linkwarden 中打开 **Settings** 下的 **Access Tokens** 页面，为助手创建令牌并私下保存。该令牌以你的 Linkwarden 账号权限访问资料。

在已配置当前设备连接的 Olares CLI 电脑终端中运行：

```bash
olares-cli settings apps get linkwarden -o json
```

在 `entrances` 中找到名为 `api` 的条目。将 `https://` 与该条目的 `url` 拼接，作为服务地址。API 请求使用该地址下的 `/api/v1/`，并设置 `Authorization: Bearer <linkwarden-token>`。从电脑连接前，先开启 [LarePass VPN](../manual/larepass/private-network.md)。

### 配置只读 MCP 连接

应用本身提供 REST 接口。可选的社区适配器为 `linkwarden-mcp@0.4.0`。在助手运行的环境中配置：

| 配置项 | 内容 |
| --- | --- |
| 连接方式 | stdio |
| 命令 | `npx` |
| 参数 | `-y`、`linkwarden-mcp@0.4.0` |
| `LINKWARDEN_URL` | internal API 服务地址，不包含 `/api/v1/` |
| `LINKWARDEN_TOKEN` | 你的 Linkwarden 令牌 |
| `LINKWARDEN_READ_ONLY` | `true` |
| `LINKWARDEN_ALLOW_TOOLS` | `search_links,get_link_content,list_collections` |

该环境需要 Node.js 和 npm。将令牌保存在助手的私有 MCP 环境配置中。只读模式限制适配器提供的工具，不会缩减令牌本身的权限。

### 整理带来源链接的简报

连接后，可以这样提问：

> 在我的 Linkwarden 资料库中搜索与家庭项目相关的页面。读取相关页面的归档正文，合并重复网址，整理简短对比，为每个结论附上来源链接。如果页面没有可读归档，请说明。不要修改书签或收藏夹。

助手可以使用 `list_collections`、`search_links` 和 `get_link_content`。请对照来源检查结论。仅保存书签但未完成正文归档的页面，可能没有足够文字供助手比较。

## 连接浏览器扩展或其他客户端

Linkwarden 客户端连接的是你部署在 Olares 上的同一个服务和账号。选择支持自定义自托管服务地址的客户端。所需地址和登录方式取决于客户端是打开网页还是直接调用 API。上面的 internal API 地址只提供 `/api/v1/`，不提供网页登录页面。需要浏览器登录的客户端应使用应用的主网页入口。

## 了解更多

- [Linkwarden 文档](https://docs.linkwarden.app/)
