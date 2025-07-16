---
description: 获取关于 Olares 的常见问题解答。
---
# 常见问题

## Olares 采用了什么协议？

Olares 由一系列项目组成，采用了分层授权的方式，基本原则是：

- 区块链上运行的项目，采用 Apache 2.0，例如 [Snowinning Protocol](https://github.com/beclab/olaresdid-contract-system)。
- 协议相关的项目，采用 Apache 2.0，例如 [r4](https://github.com/beclab/r4)。
- Olares 和 LarePass 相关的项目，采用 Olares License。
- 在 Olares 上运行的第三方应用，由开发者自己决定开源与否与协议选择。

- 每个项目具体的情况，可以在 [GitHub](https://github.com/beclab) 上查询。

## Olares 许可证是开源许可证吗？

Olares 主要项目的许可证选择受到了 [fair code](https://faircode.io/) 的启发。[Olares 许可证](https://github.com/beclab/Olares/blob/main/LICENSE.md)同样遵循以下原则:

> - 可供任何人免费使用和分发
> - 源代码公开可见
> - 任何人都可以在公共和私有社区中进行扩展
> - 作者对商业使用有所限制

## 如果助记词丢失了为什么不能恢复账号？

从 1Password 的主密钥到加密钱包的助记词，十多年来助记词存储这个问题一直没有很好的解决方案。

Olares 的助记词会被加密存储在所有安装了 LarePass 的设备上。通常情况下，只有当所有安装了 LarePass 的设备同时丢失时，才会失去助记词。

安全性是我们系统设计中最重要的原则。我们会在未来继续改进，为大家提供一个在便利性和安全性之间更好平衡的解决方案。

## Olares 和目前运行在 NAS 上的操作系统有什么区别？

在 Olares（前身是 Terminus）诞生之初，市面上已经有很优秀的 NAS 操作系统，如 [Synology](https://www.synology.com/en-global/dsm/packages)、[CasaOS](https://github.com/IceWhaleTech/CasaOS) 和 [Umbrel](https://github.com/getumbrel/umbrel)。它们确实给了我们很多启发。

但我们认为运行在边缘端的操作系统应该能够：

- 为多个硬件编排资源
- 在沙箱中管理应用

这些功能很难通过上述基于 Docker Compose 构建的 NAS 操作系统实现。

同时，Olares 致力于提供一站式自托管解决方案，这已经超出了普通 NAS 操作系统的范畴。

## 使用 Olares 需要付费吗？

在自托管场景下，基本上可以免费使用 Olares。

但对于以下两个功能，由于成本因素，我们可能会引入合理的收费（目前都是免费提供）：

- **备份**

  我们为每个 Olares ID 在 Olares Space 上提供 10G 的免费备份空间。当存档大小超过这个限制时，我们会收取一定费用来支付云服务商的费用。

- **快速反向代理（FRP）**

   通过本地访问 Olares 或使用 Olares VPN 基本是免费的。然而，如果你通过 Olares 提供外部服务（例如博客），流量会先转发到 FRP 服务器再到达 Olares。此种情况下：

   - 如果你使用自己的 FRP 服务，Olares 不会收取任何费用。
   - 如果你选择使用 Olares 默认的 FRP 服务，我们每月提供 2GB 的免费流量额度。对于不通过 Olares 提供外部服务的用户，这些流量通常是足够的。如果使用超出此限额，可能会产生额外费用。

## 什么时候支持其他语言？

目前我们只支持英语和简体中文。

实际上，我们已经在所有前端项目中完成了 i18n 的替换工作。但我们缺乏通过开源社区维护快速迭代项目的翻译资源的经验，这方面我们还在学习中。

## 各种“密码”之间有什么区别？

为了保证安全性，Olares 确实有多种密码，包括：

- 私钥
- LarePass 的密码：
  - 手机上可以使用生物识别登录
  - 电脑和浏览器插件需要手动输入
- Olares 首次激活使用的一次性密码
- Olares 的登录密码
- Olares 登录时的二次验证码

不用慌！日常使用时，需要输入的只是登录 Olares 时的二次验证码。

## 如何部署多用户应用？

这取决于是要对外提供服务，还是仅供内部 Olares 用户使用。

- 如果要对外提供服务，可以选择“**公开**”作为应用的访问入口。这样应用就可以自行管理用户注册和认证。
- 如果只提供内部访问，可以在 Olares 上部署这类产品的集群版本。

对于 Gitlab，我们提供了两个移植版本：[Gitlab Pure](https://github.com/beclab/apps/tree/main/gitlabpure) 和 [Gitlab Fusion](https://github.com/RLovelett/gitlab-fusion)。

## 如何使用相同的 Olares ID 重新激活 Olares?

如果你在同一台设备上重新安装了 Olares，你之前激活的 Olares 实例将无法再访问。你可以使用同一个 Olares ID 重新激活 Olares：

:::tip 使用相同的 Olares ID 安装
请确保在安装过程中输入了与之前完全相同的域名和 Olares ID。
:::

![重新激活](/images/zh/manual/help/reactivate.png)

1. 在手机上打开 LarePass 并进入之前的账户。你应该会在顶部看到一个红色状态提示：“未发现运行中的 Olares”。
2. 点击**了解更多**>**重新激活**，进入二维码扫描界面。
3. 点击**扫描二维码**来扫描向导页面上的二维码并激活 Olares。