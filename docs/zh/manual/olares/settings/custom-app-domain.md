---
outline: [2, 3]
description: 通过自定义路由 ID 或使用自己的域名，让 Olares 应用的访问地址更简洁、更易记。
---

# 自定义应用 URL

Olares 提供两种方式来自定义应用的访问地址：
- 自定义路由 ID
- 自定义域名

## 开始之前

建议先了解以下概念：

- [端点 (Endpoints)](../../../developer/concepts/network.md#端点)
- [路由 ID (Route ID)](../../../developer/concepts/network.md#路由-id)

## 自定义路由 ID

路由 ID 是 Olares 应用访问 URL 的一部分：

`https://{routeID}.{OlaresDomainName}`

Olares 为预装的系统应用分配了易记的路由 ID。对于社区应用，你可以自定义路由 ID，获得更简洁的 URL。以 Jellyfin 为例：

1. 在 Olares 上，打开**设置**，前往 **应用** > **Jellyfin**。
2. 在**入口**下，点击 **Jellyfin**。
3. 在**端点配置**下，点击**设置自定义路由 ID** 旁的 <i class="material-symbols-outlined">add</i>。
4. 输入一个更易记的路由 ID，例如 `jellyfin`。
5. 点击**确认**。

   ![自定义路由 ID](/images/zh/manual/olares/custom-route-id1.png#bordered){width=90%}

完成后，你就可以通过新的 URL 访问 Jellyfin：`https://jellyfin.alexmiles.olares.com`。

## 自定义域名

除了使用默认的 Olares 域名，你也可以用自己的域名访问应用。

:::info
仅认证级别为**内部**或**公开**的应用支持设置自定义三方域名。若希望无需登录即可通过自定义域名公开访问，请将认证级别设置为**公开**。
:::

:::info
如需在中国大陆境内使用，请确保自定义域名已完成备案，否则可能影响正常访问。
- 如果使用 Olares Tunnel 作为反向代理，请在腾讯云完成备案。
- 如果使用自建 FRP 作为反向代理，请在工信部完成备案。
:::

要为应用配置自定义域名：

1. 在 Olares 上，打开**设置**，前往 **应用** > 目标应用。
2. 在**入口**下，点击目标入口。
3. 在**端点配置**下，点击**设置自定义域名**旁的 <i class="material-symbols-outlined">add</i>。
4. 在**第三方域名**对话框中，输入你的自定义域名，上传该域名的有效 HTTPS 证书和私钥，然后点击**确认**。

   :::tip 注意
   如果反向代理使用 Cloudflare Tunnel，只需填写域名，无需上传证书。
   :::

   ![输入三方域名及证书](/images/zh/manual/olares/enter-custom-domain.jpeg#bordered)

5. 点击**激活**，打开激活引导弹窗。

   ![激活第三方域名](/images/zh/manual/olares/activate-custom-domain.jpeg#bordered)

6. 按照弹窗中的说明，在你的域名托管服务商处创建一条 CNAME 记录。

   ![添加 CNAME](/images/zh/manual/olares/add-cname.png#bordered)

   :::tip 为 Cloudflare 关闭代理状态
   如果反向代理使用 Cloudflare Tunnel，请关闭 DNS 记录旁的代理状态选项，以便 Olares 及时获取域名解析状态。
   :::

7. 在激活引导弹窗中点击**确认**，完成激活。

此时自定义域名状态显示为“等待 CName 激活”。DNS 解析生效时间通常从几分钟到 48 小时不等。

Olares 会自动验证 DNS 记录。验证通过后，自定义域名状态将更新为“已激活”，你就可以通过自定义 URL 访问应用了。
