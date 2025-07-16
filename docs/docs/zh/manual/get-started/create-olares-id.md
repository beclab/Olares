---
description: 使用 LarePass 移动端应用创建 Olares ID。
---
# 创建 Olares ID

Olares ID 是 Olares 生态系统中的唯一标识符，作为你的数字身份，可用于访问各种服务和功能。

Olares ID 由本地名称和域名两部分组成。以 `alice123@olares.cn` 为例：
- `alice123`：本地名称
- `olares.cn`：域名

:::tip
要了解为什么需要 Olares ID，请参阅 [Olares ID](../concepts/olares-id.md)。
:::

## 下载并安装 LarePass 应用

在手机应用商店中搜索“LarePass”并下载。

## 创建 Olares ID

::: tip
本节主要介绍如何创建个人 Olares ID。如需创建用于组织用途的 Olares ID，请参阅[创建组织 Olares ID](../../space/host-domain.md#创建组织-olares-id)。
:::

:::warning `.com`域名 与 `.cn`域名
为了保证良好的激活和使用体验，Olares 为中国大陆境内用户设置了专属反向代理节点和对应的 `.cn` 的域名。首次创建 Olares ID 时，LarePass 会根据手机系统语言，默认分配 Olares ID 的域名。如果你的手机语言为英文，则会创建 `.com` 域名的 Olares ID, 可能会遇到 DNS 解析问题，进而影响后续的激活和使用。此时，需要你从 LarePass 的账号创建页面右上角进入高级账号创建模式，切换域名默认值为 `.cn`后再返回创建。
:::

1. 打开 LarePass 应用，点击**创建账号**。
2. 输入想要使用的 Olares ID。需要满足以下要求：
   * 之前从未被注册过
   * 长度不少于 8 个字符
   * 仅可使用小写字母和数字
3. 点击**继续**完成创建。

![快速创建](/images/manual/get-started/create-olares-id.png)

## 后续步骤

请妥善保管新创建的 Olares ID 和 LarePass 应用，后续步骤中会用到。

如果要以管理员身份安装和激活 Olares：
- [安装 Olares](./install-olares)

如果你的团队已经部署了 Olares，需要以成员身份加入：
- [激活 Olares](./activate-olares)

如果要登录 Olares Space 或创建基于云的 Olares：
- [管理 Olares Space 账号](../../space/manage-accounts)