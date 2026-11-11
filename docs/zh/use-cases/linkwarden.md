---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/linkwarden
outline: [2, 3]
description: 在 Olares 上用 Linkwarden 保存网页、整理收藏夹并阅读归档
head:
  - - meta
    - name: keywords
      content: "Olares, Linkwarden, 自托管, 书签, 网页归档, 收藏夹"
app_version: "0.0.5"
doc_version: "2.0"
doc_updated: "2026-11-11"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/linkwarden.md)为准。
:::

# 使用 Linkwarden 建立阅读资料库

Linkwarden 是一款书签管理工具，可以将链接整理到收藏夹，并保存网页副本供日后阅读。在 Olares 上，资料库和归档内容保存在自己的设备上。

本示例将 Jellyfin、Navidrome 和 Immich 的 Olares 指南保存到同一个收藏夹，方便比较这三款家庭媒体应用。

## 前提条件

- Olares 1.12.7 或更高版本。

## 安装 Linkwarden

1. 打开应用市场，搜索“Linkwarden”。
2. 点击 **Get**，再点击 **Install**，等待安装完成。

## 打开 Linkwarden

1. 从启动台打开 Linkwarden。
2. 首次使用时注册账号。如果已有 Linkwarden 账号，直接登录。

Olares 账号用于保护应用入口。Linkwarden 账号与 Olares 账号是分开的，用于访问书签资料库。

## 创建收藏夹并保存链接

先为三篇指南创建收藏夹，再将它们保存为带有相同标签的链接。

1. 在侧边栏点击 **Collections**，然后点击 **Collections you own** 旁边的 **+**。
2. 在 **Name** 中输入收藏夹名称，点击 **Create Collection**。本指南以 `Home media apps` 为例。
3. 点击 **New Link**，在 **Link** 中输入 `https://docs.olares.com/use-cases/jellyfin`。
4. 在 **Collection** 中选择 `Home media apps`，然后点击 **More Options**。
5. 在 **Name** 中输入 `Jellyfin on Olares`。在 **Tags** 中输入 `home-media` 并从菜单中选择该标签。如果标签尚不存在，选择 **Create "home-media"**。点击 **Create Link**。
6. 按照第 3 至第 5 步再保存两条链接，使用同一个收藏夹和标签：

   - Navidrome：链接为 `https://docs.olares.com/use-cases/navidrome`，名称为 `Navidrome on Olares`。
   - Immich：链接为 `https://docs.olares.com/use-cases/immich`，名称为 `Immich on Olares`。

:::info
收藏夹默认不公开。只有希望其他人阅读其中的链接时才分享收藏夹。
:::

## 阅读网页归档

打开已保存的 Jellyfin 指南，以可读格式查看归档正文。

1. 从侧边栏打开 `Home media apps`，找到 `Jellyfin on Olares` 卡片。
2. 卡片上的四个图标从左到右分别对应 **Webpage**、**Screenshot**、**PDF** 和 **Readable** 归档格式。点击 **Readable** 打开阅读视图。

   ![从链接卡片打开可读格式的网页归档](/images/manual/use-cases/linkwarden-formats.png#bordered)

3. 打开归档页面后，点击顶部的 **Readable** 菜单，可以切换到 **Webpage**、**Screenshot** 或 **PDF**。

   ![在归档页面切换格式](/images/manual/use-cases/linkwarden-format-menu.png#bordered)

:::tip 关于网页归档
Linkwarden 会在后台抓取网页。如果卡片上尚未显示归档格式图标，稍等片刻后刷新收藏夹。需要登录或阻止自动访问的网站可能无法生成完整副本。
:::

## 查找已保存的链接

需要再次查找链接时，可以搜索资料库。这里以 Jellyfin 页面为例。

1. 点击侧边栏中的 **Search for Links**。
2. 输入 `Jellyfin` 并按回车键。
3. 在搜索结果中找到 `Jellyfin on Olares`。

   ![搜索已保存的 Jellyfin 页面](/images/manual/use-cases/linkwarden-search-results.png#bordered)

## 可选：通过助手使用资料库

本示例使用 Lares 和社区 MCP 适配器搜索 Linkwarden 资料库、读取归档页面。其他支持本地 stdio MCP 服务的助手也可以使用同一个适配器，但需要按照各自的方式配置。

### 前提条件

- 已通过 Router 为 Lares 接入可用的模型。参见[使用 Lares](lares.md)。
- 运行 Lares 本地 MCP 命令的环境中有 Node.js 22 或更高版本以及 npm。

### 获取 API 令牌和服务地址

MCP 适配器需要服务地址来连接 Linkwarden API，还需要访问令牌来验证请求。

**API 令牌**

1. 在 Linkwarden 左下角打开账号菜单，选择 **Settings** > **Access Tokens**。
2. 点击 **New Access Token**，为令牌命名、设置有效期并妥善保存。令牌具有对应 Linkwarden 账号的访问权限。

**服务地址**

1. 在 Olares **Settings** 中打开 **Applications** > **Linkwarden**。
2. 在 **Entrances** 下选择 **Linkwarden Agent API**。
3. 在 **Endpoint settings** 下复制 **Endpoint** 地址，稍后用作 `LINKWARDEN_URL`。

   ![在 Olares 设置中复制 Linkwarden Agent API 的 Endpoint 地址](/images/manual/use-cases/linkwarden-agent-api-endpoint.png#bordered)

### 配置 MCP 连接

1. 在 Lares 中打开 **Settings** > **MCP**，点击 **Add server**。
2. 在 **Server name** 中输入 `linkwarden`，在 **Transport** 中选择 **Local command (stdio)**。
3. 在 **Command** 中输入 `npx`。在 **Arguments** 中输入以下 JSON 数组：

   ```json
   ["-y", "linkwarden-mcp@0.4.0"]
   ```

4. 在 **Environment variables** 中输入以下 JSON 对象。将 `LINKWARDEN_URL` 设为刚才从 Olares Settings 复制的 Endpoint 地址，将 `LINKWARDEN_TOKEN` 设为在 Linkwarden 中创建的访问令牌：

   ```json
   {
     "LINKWARDEN_URL": "https://api.example.olares.com",
     "LINKWARDEN_TOKEN": "<your-token>",
     "LINKWARDEN_READ_ONLY": "true",
     "LINKWARDEN_ALLOW_TOOLS": "search_links,get_link_content,list_collections"
   }
   ```

   将令牌保存在 MCP 配置中，不要粘贴到聊天中。只读模式限制适配器提供的工具，不会缩减令牌本身的权限。

5. 点击 **Save**。在 **Settings** > **MCP** 中等待 `linkwarden` 显示为已连接。
6. 新建 Lares 会话并提问：“List my Linkwarden collections.” 确认 Lares 返回 Linkwarden 收藏夹。

   ![Lares 返回包含三条链接的 Home media apps 收藏夹](/images/manual/use-cases/linkwarden-lares-collections.png#bordered)

### 比较已保存的资料

1. 连接成功后，向 Lares 提问：

   > 在我的 Home media apps 收藏夹中找到 Jellyfin、Navidrome 和 Immich 的页面。读取归档正文，比较每款应用帮助我整理什么内容，并为每个结论附上来源链接。如果某个页面没有可读归档，请说明，并标明哪些信息来自实时网页。不要修改书签或收藏夹。

2. 检查 Lares 是否找到全部三条链接、说明使用的是归档正文还是实时网页，并为结论提供来源链接。

   ![Lares 使用可读归档比较三条已保存的链接并附上来源](/images/manual/use-cases/linkwarden-lares-compare-sources.png#bordered)

## 连接浏览器扩展

连接浏览器扩展后，可以在浏览网页时将页面保存到同一个 Linkwarden 资料库。

1. 在 Olares **Settings** 中打开 **Applications** > **Linkwarden**。在 **Entrances** 下选择 **Linkwarden**，复制其 **Endpoint** 地址。
2. 安装并打开 [Linkwarden 浏览器扩展](https://docs.linkwarden.app/getting-started/browser-extension)。在 **Initial Setup** 页面点击 **Configure**。按下表填写，然后点击 **Save**。

   | 字段 | 填写内容 |
   | --- | --- |
   | **URL** | 第 1 步复制的 Endpoint 地址 |
   | **Method** | **Username and Password** |
   | **Username or Email** | 在 Olares 上的 Linkwarden 中创建的账号所用的用户名或邮箱 |
   | **Password** | 该 Linkwarden 账号的密码 |

## 了解更多

- [Linkwarden 文档](https://docs.linkwarden.app/)：了解收藏夹、网页归档和集成功能。
