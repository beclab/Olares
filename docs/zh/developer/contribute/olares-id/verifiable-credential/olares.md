# Olares 的案例

虽然 DID 解决了去中心化网络中的身份问题，但它们通常难以记忆，不便在日常情况下使用。Olares ID 提供了一种类似于电子邮件地址的、易于记忆的熟悉格式，同时仍然利用了 DID 的强大功能和安全性。每个 Olares ID 都与一个 DID 绑定。

## 潜在的公平性问题

Olares ID 注册中的**先到先得**系统可能会带来一些公平性问题，例如：

1.  **欺诈：** 例如，Olares ID `elonmusk@myterminus.com` 可能会被并非埃隆·马斯克本人注册。
2.  **投机：** 抢注热门名称可能会导致投机行为，这虽然可能促进早期网络活跃度，但会牺牲公平性。

## Olares ID 的 VC 流程

为了解决潜在的公平性问题，我们采纳了 Web5 团队提出的**自主身份 (Self-Sovereign Identity, SSI)** 服务原则，并结合了 Olares ID 的 [VC 流程](/zh/manual/concepts/vc.md#六步验证流程)。这引导我们设计了一个**发行方 (Issuer) 和验证方 (Verifier)** 流程，以协助用户申请 **Olares ID**。

![alt text](/images/developer/contribute/vc-process.jpeg)

### Gmail 发行方服务

我们利用谷歌的 OAuth 流程来促进**可验证凭证 (Verifiable Credentials, VCs)** 的发行。简化流程如下：

1.  Alice 在钱包客户端 LarePass 中通过 OAuth 登录她的 Gmail 账户。
2.  谷歌将 OAuth 凭证返回给 LarePass。
3.  LarePass 将 OAuth 凭证提交给发行方。
4.  发行方与谷歌服务器确认凭证的有效性，并检索基本信息（例如，电子邮件名称）。
5.  发行方向 Alice 发行一个与其 Gmail 地址本地部分相匹配的 VC。

现在，Alice 就可以将已发行的 VC 存储在 LarePass 中。

:::tip 注意
- 在整个过程中，Alice 仅向 LarePass 和发行方服务透露凭证授权范围内的基本账户数据，密码和隐私由谷歌的 OAuth 协议确保。
- 所有用于设置 Gmail 发行方服务或其他 Web2 服务发行方服务的代码都在 GitHub 上开源。
  :::

### Olares ID 验证方服务

以下是 **验证方服务 (Verifier Service)** 在 Olares 端的运作方式：

1.  Alice 将她的 DID、Olares ID 和 Gmail VC 打包成一个**可验证表述 (Verifiable Presentation, VP)**，并将其与签名一同提交给验证方服务。
2.  验证方服务会检查：
    -   签名的有效性。
    -   VP 中 VC 的有效性。
    -   该 Olares ID 是否可以在区块链上注册（如果使用多个渠道，如 Gmail 和 Twitter，获取 VC 信息，可能会产生冲突）。
3.  所有检查通过后，验证方服务会将 Alice 的信息提交到区块链，并支付 Gas 费用。

至此，Alice 成功获得了她的 **Olares ID**。例如，如果你使用 Gmail 地址 “hello@gmail.com” 进行申请，所有检查完成后，你将获得 Olares ID “hello@olares.com”。