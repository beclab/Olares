---
outline: [2, 3]
title: 在 Olares 上设置 FlareSolverr 与 Prowlarr
description: 在 Olares 上设置 FlareSolverr 与 Prowlarr，绕过 Cloudflare 防护，为媒体栈访问被阻止的索引器站点。
head:
  - - meta
    - name: keywords
      content: Olares, FlareSolverr, Prowlarr, Cloudflare, indexer, proxy, self-hosted, 设置 FlareSolverr 与 Prowlarr, FlareSolverr Prowlarr, Prowlarr FlareSolverr, FlareSolverr Kodi, ghcr.io FlareSolverr
app_version: "1.0.4"
doc_version: "1.1"
doc_updated: "2026-07-29"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/flaresolverr.md)。
:::

# 在 Prowlarr 中使用 FlareSolverr 访问 Cloudflare 保护的站点

FlareSolverr 是一个代理服务器，可以代替 Prowlarr 等应用解决 Cloudflare 和 DDoS-GUARD 挑战。当索引器站点显示 CAPTCHA 或 JavaScript 挑战时，FlareSolverr 会使用无头浏览器完成挑战并返回有效响应。

在 Olares 上，你可以将 FlareSolverr 设置为 Prowlarr 的索引器代理。配置完成后，Prowlarr 会将受保护索引器的请求路由到 FlareSolverr，让你的 *arr 应用能够继续搜索和下载内容。

本文档将引导你安装 FlareSolverr、将其连接到 Prowlarr，并添加受 Cloudflare 保护的索引器（例如 1337x）。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上安装 FlareSolverr。
- 将 FlareSolverr 配置为 Prowlarr 中的索引器代理。
- 添加受 Cloudflare 保护的索引器站点。
- 验证 FlareSolverr 是否正确解决挑战。

## 前提条件

- 已设置并运行 Olares 设备。
- Prowlarr 已安装并运行。安装步骤请参考[设置 *Arrs 媒体栈](./arrs.md)。

## 安装 FlareSolverr

1. 打开 Market，搜索 "FlareSolverr"。
   ![安装 FlareSolverr](/images/manual/use-cases/install-flaresolverr.png#bordered){width=90%}

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

FlareSolverr 作为后台服务运行，因此不会出现在 Launchpad 上。

在 **Settings** > **Applications** 的应用列表中找到 FlareSolverr，确认它正在运行。

![Settings 中的 FlareSolverr](/images/manual/use-cases/flaresolverr-installed-in-settings.png#bordered){width=60%}

或者在 **Market** > **My Olares** 的已安装应用列表中查看 FlareSolverr，确认它正在运行。

![My Olares 中的 FlareSolverr](/images/manual/use-cases/flaresolverr-installed-in-market.png#bordered){width=60%}

## 在 Prowlarr 中配置 FlareSolverr

### 获取 FlareSolverr Endpoint

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

对于 FlareSolverr：

1. 前往 Olares **Settings** > **Applications** > **FlareSolverr** > **Entrances**。
2. 选择 **FlareSolverr**，然后复制 **Endpoint** URL。

### 将 FlareSolverr 添加为索引器代理

1. 从 Launchpad 打开 Prowlarr。
   :::tip 首次设置
   首次打开 Prowlarr 时，会出现 **Authentication Required** 对话框。选择一种认证方法，设置用户名和密码，然后点击 **Save** 继续。
   :::
2. 前往 **Settings** > **Indexers**。
3. 在 **Indexer Proxies** 下，点击 <i class="material-symbols-outlined">add_2</i>，然后选择 **FlareSolverr**。

   ![添加 FlareSolverr 代理](/images/manual/use-cases/prowlarr-add-flaresolverr.png#bordered){width=80%}

4. 配置代理设置：
    - **Tags**：输入小写标签名称，例如 `flaresolverr`。Prowlarr 使用此标签来决定哪些索引器应通过 FlareSolverr 路由请求。
    - **Host**：粘贴你之前复制的 FlareSolverr Endpoint。
5. 点击齿轮图标，将 **Request Timeout** 设为 `180` 秒。
6. 点击 **Test** 验证连接，然后点击 **Save**。
   ![FlareSolverr 代理设置](/images/manual/use-cases/prowlarr-flaresolverr-settings.png#bordered){width=80%}

### 添加受 Cloudflare 保护的索引器

本示例使用 1337x，一个受 Cloudflare 保护的流行索引器站点。

1. 在 Prowlarr 中，点击 **Indexers** > **Add Indexer** 并搜索 "1337x"。
2. 选择 **1337x**，然后将 **Base URL** 设为 `1337x.to`。
3. 在底部的 **Tags** 字段中，输入你分配给 FlareSolverr 代理的相同标签（例如 `flaresolverr`）。
   ![1337x 索引器设置](/images/manual/use-cases/prowlarr-1337x-tags.png#bordered){width=80%}

4. 点击 **Test**。

:::info
挑战解决过程可能需要一些时间，并且不一定在第一次尝试时就成功。如果初始测试失败，请多尝试几次。
:::

### 验证 FlareSolverr 是否正常工作

你可以检查 FlareSolverr 的日志，以确认它是否正在接收和解决 Cloudflare 挑战。

1. 打开 Control Hub，从侧边栏选择 **FlareSolverr** 项目。
2. 在 **Deployments** 下，点击 **flaresolverr** 的运行中 pod，然后展开容器以查看其日志。

   ![FlareSolverr 容器日志](/images/manual/use-cases/flaresolverr-logs.png#bordered){width=80%}

3. 点击播放按钮以流式传输实时日志。
4. 返回 Prowlarr，点击 1337x 索引器上的 **Test**。你应该会在 FlareSolverr 的日志中看到传入的请求。
5. 在日志中查找 `Challenge solved`。这确认 FlareSolverr 已绕过 Cloudflare 保护。
   ![FlareSolverr 挑战已解决](/images/manual/use-cases/flaresolverr-challenge-solved.png#bordered){width=80%}

6. 在 Prowlarr 的搜索栏中搜索内容。如果结果出现，FlareSolverr 正在正常工作。
   ![Prowlarr 搜索结果](/images/manual/use-cases/prowlarr-search-results.png#bordered){width=80%}

## 将 FlareSolverr 用于你的媒体栈

FlareSolverr 最常与 Prowlarr 一起使用，但同样的代理设置也让自托管媒体工作流的其他部分受益：

- **Prowlarr** 在索引器标签与代理标签匹配时，会通过 FlareSolverr 发送索引器查询。
- **Sonarr**、**Radarr** 和 **Lidarr** 依赖 Prowlarr 的索引器，因此它们会间接受益于 FlareSolverr，无需额外配置。
- **Kodi**、**Jellyfin** 和 **Plex** 等媒体中心使用 Sonarr 和 Radarr 获取的文件，因此正常运行的 FlareSolverr 设置有助于为整个媒体栈持续提供新内容。

如果在其他 *arr 应用中遇到受 Cloudflare 保护的索引器，请确保该索引器通过 Prowlarr 管理并正确标记。

## 常见问题

### Prowlarr 测试失败，但 FlareSolverr 日志显示 "Challenge solved"

如果 FlareSolverr 成功解决了 Cloudflare 挑战，但 Prowlarr 仍然报告站点被阻止，你可以强制保存索引器。在某些情况下，即使测试失败，索引器也会返回搜索结果。

要强制保存：

1. 在索引器设置中，根据需要配置所有其他参数。
2. 取消选中 **Enable** 复选框，然后点击 **Save**。索引器现在将以禁用状态添加到列表中。
3. 在 **Indexers** 页面上，在列表中找到该索引器，点击最右侧的扳手图标进行编辑。
   ![从 Indexers 列表编辑索引器](/images/manual/use-cases/prowlarr-cloudflare-edit-indexer.png#bordered){width=80%}
   
4. 选中 **Enable** 复选框。
5. 点击 **Save**。
6. 如果 Prowlarr 保持编辑页面打开并显示警告（例如 `Unable to access 1337x.to, blocked by CloudFlare Protection`），**Save** 按钮将变为红色警告图标。再次点击此图标以保存索引器，尽管连接检查失败。

重新启用后，尝试在 Prowlarr 中搜索。如果结果出现，则索引器正在工作，尽管测试失败。

### FlareSolverr 没有收到 Prowlarr 的请求

请检查以下内容：

- Prowlarr 的 FlareSolverr 代理设置中的 **Host** 字段与你从 Olares Settings 复制的 API 端点一致。
- 索引器的 **Tags** 字段包含分配给 FlareSolverr 代理的相同标签。
- FlareSolverr pod 正在 Control Hub 中运行。

### 如何更新 FlareSolverr

当 Olares Market 发布 FlareSolverr 新版本时，打开 Market，前往 **My Olares**，然后点击 FlareSolverr 旁的 **Update**。Olares 会自动拉取最新的 `ghcr.io/flaresolverr/flaresolverr:latest` 镜像。

## 了解更多

- [使用 Jellyfin 流媒体播放](jellyfin.md)：在 Olares 设备上托管和播放电影、剧集和音乐。
