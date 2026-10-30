---
outline: [2, 3]
description: 通过自定义路由 ID，让 Olares 应用的访问地址更简洁、更易记。
head:
  - - meta
    - name: keywords
      content: Olares, 设置, 自定义路由 ID, 应用 URL
---

# 自定义应用 URL

为 Olares 应用设置自定义路由 ID，把访问地址中自动生成的路由 ID 换成更易记的名称。

## 设置自定义路由 ID

[路由 ID](../../../developer/concepts/network.md#路由-id) 是 Olares 应用 URL 中用于标识应用的部分。例如，在 `https://7e89d2a1.laresprime.olares.com` 中，`7e89d2a1` 就是路由 ID。Olares 默认为社区应用分配一串随机的数字和字母作为路由 ID，不容易记住。

你可以为应用设置自定义路由 ID，获得更简洁的 URL。以 Jellyfin 为例：

:::warning 使用唯一的路由 ID
自定义路由 ID 不能与同一 Olares 账号下其他应用入口的默认路由或自定义路由 ID 重复。如果修改路由 ID 后应用 URL 返回 `404`，请参考[自定义路由 ID 导致应用无法访问](../../help/ts-custom-route-domain.md)。
:::

1. 在 Olares 上，打开**设置**，前往 **应用** > **Jellyfin**。
2. 在**入口**下，点击 **Jellyfin**。
3. 在**端点配置**下，点击**设置自定义路由 ID** 旁的 <i class="material-symbols-outlined">add</i>。
4. 输入一个更易记的路由 ID，例如 `jellyfin`。
5. 点击**确认**。

设置完成后，你可以通过两个 URL 访问 Jellyfin：原始的 `https://7e89d2a1.laresprime.olares.com` 和新的 `https://jellyfin.laresprime.olares.com`。
