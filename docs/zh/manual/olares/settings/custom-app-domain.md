---
outline: [2, 3]
description: 自定义 Olares 应用的访问地址，通过设置路由 ID 和域名，实现更简洁专业的应用访问方式。
---

# 自定义应用 URL
无论是在本地还是远程，你都可以随时随地访问 Olares 应用。本文档将介绍如何为应用添加自定义域名。

## 开始之前
在开始之前，建议你先熟悉一些与 Olares 应用相关的概念：

- [端点 (Endpoints)](../../../developer/concepts/network.md#端点)
- [路由 ID (Route ID)](../../../developer/concepts/network.md#路由-id)

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
仅认证级别为**内部**或**公开**的应用支持设置自定义三方域名。若想要无需登录即可通过自定义域名公开访问应用，请将认证级别设置为**公开**。
:::

:::info
在中国大陆境内使用，请确保你的自定义域名已完成备案，否则可能会影响正常访问。
- 如果反向代理为 Olares Tunnel，请在腾讯云完成备案。
- 如果反向代理为为自建 FRP，请在工信部完成备份。
:::

要为应用配置自定义域名：

1. 打开**设置**，在左侧边栏中选择 **应用**。
2. 右侧应用列表里点击您想要配置的应用程序，进入其详情页面。
3. 前往**入口** > **设置端点**，然后点击**设置自定义域名**旁边的 <i class="material-symbols-outlined">add</i> 按钮。
4. 在**第三方域名**对话框里，输入你的自定义域名，上传域名的有效 HTTPS 证书及私钥，之后点击**确认**。
   :::tip 注意
   如果反向代理是 Cloudflare Tunnel，仅填写域名即可。
   :::

   ![输入三方域名及证书](/images/zh/manual/olares/enter-custom-domain.jpeg#bordered)

5. 点击**激活**，进入激活引导窗口。
   ![激活第三方域名](/images/zh/manual/olares/activate-custom-domain.jpeg#bordered)

6. 按照弹出窗口中的说明，在你的域名托管服务商处创建一条 CNAME 记录。
   ![添加 CNAME](/images/zh/manual/olares/add-cname.png#bordered)

   :::tip 为 Cloudflare 关闭代理状态
   如果你的反向代理使用的是 Cloudflare Tunnel，关闭 DNS 记录旁边的代理状态选项以便 Olares 实时接收您域名解析状态的更新。
   :::

7. 在激活引导窗口中点击**确认**，完成激活。

此时自定义域名状态为“等待 CName 激活”。DNS 解析生效时间可能从几分钟到 48 小时不等。
   
Olares 会自动验证 DNS 记录是否生效。验证通过后，自定义域名状态会变为“已激活”。此时你就可以通过自定义 URL 访问应用了。