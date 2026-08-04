---
outline: [2, 3]
description: 通过自定义路由 ID 或自定义域名，让 Olares 应用的访问地址更简洁、更易记。
head:
  - - meta
    - name: keywords
      content: Olares, 自定义域名, 自定义路由 ID, 应用 URL, HTTPS 证书, CNAME 记录
---

# 自定义应用 URL

当你希望让 Olares 应用的访问地址更易记时，可以参考本文。Olares 提供两种自定义方式：

- **自定义路由 ID**：把 Olares URL 中自动生成的路由 ID 换成更易记的名称。
- **自定义域名**：使用你自己的域名，替代默认的 Olares 域名。

## 前提条件

要使用自定义域名，你需要：

- 一个你已注册并可控制的域名。
- 可以登录该域名的 DNS 管理后台，用于添加 CNAME 记录。
- 该域名的有效 HTTPS 证书及其 RSA 私钥，格式为 PEM。

:::info 中国大陆用户
如需在中国大陆境内使用，请确保自定义域名已完成备案，否则可能影响正常访问。
- 如果使用 Olares Tunnel 作为反向代理，请在腾讯云完成备案。
- 如果使用自建 FRP 作为反向代理，请在工信部完成备案。
:::

## 设置自定义路由 ID

[路由 ID](../../../developer/concepts/network.md#路由-id) 是 Olares 应用 URL 中用于标识应用的部分。例如，在 `https://7e89d2a1.laresprime.olares.com` 中，`7e89d2a1` 就是路由 ID。Olares 默认为社区应用分配一串随机的数字和字母作为路由 ID，不容易记住。

你可以为应用设置自定义路由 ID，获得更简洁的 URL。以 Jellyfin 为例：

1. 在 Olares 上，打开**设置**，前往 **应用** > **Jellyfin**。
2. 在**入口**下，点击 **Jellyfin**。
3. 在**端点配置**下，点击**设置自定义路由 ID** 旁的 <i class="material-symbols-outlined">add</i>。
4. 输入一个更易记的路由 ID，例如 `jellyfin`。
5. 点击**确认**。

   ![为 Jellyfin 设置自定义路由 ID](/images/manual/olares/custom-route-id1.png#bordered){width=90%}

设置完成后，你可以通过两个 URL 访问 Jellyfin：原始的 `https://7e89d2a1.laresprime.olares.com` 和新的 `https://jellyfin.laresprime.olares.com`。

## 设置自定义域名

:::warning 自定义域名不支持 Olares 身份认证
使用自定义域名的应用不支持 Olares 身份认证。仅**认证级别**为**公开**的应用入口支持自定义域名。
:::

本文介绍如何为单个应用设置自定义域名。如需为整个 Olares 系统设置统一的自定义域名，需要在激活设备前在 Olares Space 中完成设置。详见[为 Olares 设置自定义域名](../../best-practices/set-custom-domain.md)。

例如，把 Jellyfin 的访问地址从 `https://7e89d2a1.laresprime.olares.com` 改为 `https://media.n1.monster`：

1. 在 Olares 上，打开**设置**，前往 **应用** > **Jellyfin**。
2. 在**入口**下，点击 **Jellyfin**。
3. 在**访问策略**下，将**认证级别**设置为**公开**，然后点击**提交**以应用更改。
4. 在**端点配置**下，点击**设置自定义域名**旁的 <i class="material-symbols-outlined">add</i>。
5. 在**设置自定义域名**对话框中，输入你的自定义域名，并粘贴该域名的 HTTPS 证书和私钥。两个文件都应以 `-----BEGIN-----` 开头、以 `-----END-----` 结尾。

   ![输入自定义域名及证书](/images/manual/olares/settings-add-custom-domain.png#bordered)

6. 点击**确认**。
7. 点击**激活**，打开激活引导弹窗。
8. 按照弹窗中的说明，在你的域名注册商处添加一条 CNAME 记录。本例中，添加一条 Name 为 `media`、Value 为 `laresprime.olares.com` 的 CNAME 记录。

   ![添加 CNAME 记录](/images/manual/olares/settings-add-cname.png#bordered)

9. 在激活引导弹窗中点击**确认**。此时自定义域名状态会显示为“等待 CNAME 激活”。

DNS 解析生效通常需要几分钟到几小时，具体时间取决于你的域名注册商。

DNS 记录验证通过后，自定义域名状态会自动更新为“已激活”。

现在，你可以通过两个 URL 访问 Jellyfin：原始的 Olares URL `https://7e89d2a1.laresprime.olares.com` 和自定义域名 `https://media.n1.monster`。

![自定义域名已激活](/images/manual/olares/settings-custom-domain-activated.png#bordered)
