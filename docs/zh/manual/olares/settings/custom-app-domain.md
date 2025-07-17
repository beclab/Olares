---
outline: [2, 3]
description: 自定义 Olares 应用的访问地址，通过设置路由 ID 和域名，实现更简洁专业的应用访问方式。
---

# 自定义应用 URL
无论是在本地还是远程，你都可以随时随地访问 Olares 应用。本文档将介绍如何为应用添加自定义域名。

## 开始之前
在开始之前，建议你先熟悉一些与 Olares 应用相关的概念：

- [端点 (Endpoints)](../../concepts/network.md#端点)
- [路由 ID (Route ID)](../../concepts/network.md#路由-id)

## 为应用自定义域名

Olares 提供两种方法来优化应用的访问地址：
* 自定义路由 ID
* 自定义域名

### 自定义路由 ID
路由 ID 是访问 Olares 应用的重要组成部分，和 Olares 域名一起构成了你通过 Web 浏览器访问应用的 URL：

`https://{routeID}.{OlaresDomainName}`

为方便起见，Olares 为预安装的系统应用使用了易于记忆的路由 ID。对于社区应用，你可以通过更改路由 ID 快速获得一个简单易记的 URL。

以 Jellyfin 为例：

1. 打开**设置**，在左侧边栏中选择**应用**。
2. 点击右侧的 **Jellyfin** 查看应用详情。
3. 前往**入口** > **设置端点**。可以看到 Jellyfin 的默认路由 ID，是一个数字和字母的组合 `7e89d2a1`。
4. 在**设置自定义路由 ID** 旁，点击 **<i class="material-symbols-outlined">add</i>** 打开设置对话框。
5. 输入一个更易记的路由 ID，例如 `jellyfin`。

   ![自定义路由 ID](/images/zh/manual/tasks/custom-route-id.png#bordered)
6. 点击**确认**。

现在，你可以通过新的 URL 访问 Jellyfin：`https://jellyfin.bob.olares.cn`。

### 自定义域名

除了使用默认的 Olares 域名，你还可以使用自己的域名访问应用，使其看起来更专业、更易记。
:::info
仅认证级别为**内部**或**公开**的应用支持设置自定义三方域名。
:::


:::info
- 中国大陆网络环境下，如果当前反向代理为 Olares Tunnel 或自建 FRP，配置三方域名时需要手动上传 HTTPS 证书及私钥。
- 请确保你的自定义域名已完成备案，否则可能会影响正常访问。
:::

以 Affine 为例：

1. 打开**设置**，在左侧边栏中选择 **应用**。
2. 点击右侧的 **Affine** 查看应用详情。
3. 前往**入口** > **设置端点**。
4. 在**设置自定义域名**旁，点击 <i class="material-symbols-outlined">add</i> 打开设置对话框。

   ![设置三方域名](/images/zh/manual/tasks/set-custom-domain.png#bordered)
5. 输入你的自定义域名，例如 `hello.coffee`，并粘贴 HTTPS 证书及私钥，然后点击**确认**。

   ![输入三方域名及证书](/images/zh/manual/tasks/enter-custom-domain.png#bordered){width=70%}
6. 点击**激活**，按提示在你的域名托管网站上添加 CNAME 记录，点击**确认**。

   ![激活第三方域名](/images/zh/manual/tasks/activate-custom-domain.png#bordered)
   
   此时自定义域名状态为“等待 CName 激活”。DNS 解析生效时间可能从几分钟到 48 小时不等。
   
Olares 会自动验证 DNS 记录是否生效。验证通过后，自定义域名状态会变为“已激活”。此时你就可以通过新的 URL `hello.coffee` 访问 Affine 了。

::: tip 提示
若要允许无需登录即可通过自定义域名公开访问应用，请按以下步骤更新访问策略：
1. 前往**设置** > **应用**，点击目标应用。
2. 点击**入口**，在**创建访问策略**下，将**认证级别**设置为**公开**。
3. 点击**提交**使变更生效。

   ![设置认证级别为公开](/images/zh/manual/tasks/set-auth-level-to-public.png){width=50%}
:::