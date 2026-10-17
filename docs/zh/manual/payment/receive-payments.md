---
description: 跟着视频和截图快速上手 Olares Payment 收款——LarePass 扫码登录、收款地址自动就绪、分享收款链接、查看到账记录。
head:
  - - meta
    - name: keywords
      content: Olares Payment 上手, LarePass 登录, 收款地址, 收款链接, 商户后台
---

# 收款快速上手

从开通到收到第一笔钱,全程不需要写代码。开始前,请确认你已有 [Olares ID](/zh/manual/get-started/create-olares-id) 并装好了 LarePass 应用。

<video controls preload="none" poster="/images/payment/nocode-demo-poster.png" style="width:100%;border-radius:8px"><source src="/videos/payment/nocode-demo.mp4" type="video/mp4" /></video>

*两分钟完整流程:扫码登录 → 收款地址就绪 → 生成 invoice 链接 → 买家用 MetaMask 付款 → 后台看到账。*

## 第 1 步:扫码登录

在浏览器打开 [https://www.olares.com/payment/dashboard](https://www.olares.com/payment/dashboard/),用 LarePass 扫二维码确认登录。

![LarePass 扫码登录](/images/payment/dashboard-login.png#bordered)

**第一次登录即完成开通**:系统自动创建你的商户账户,并把你 LarePass 钱包的 EVM 地址配置为全部支持链上的收款地址——不用填表,不用手抄地址。

## 第 2 步:确认收款地址

登录后首页会显示 **Ready to receive** 卡片:这就是你的收款地址,USDC / USDT 通用,七条链(Optimism、Base、BNB Smart Chain、Ethereum、Arbitrum One、Polygon、Avalanche)同一个地址。

![首页:收款已就绪](/images/payment/dashboard-home.png#bordered)

这个地址就是你 LarePass 钱包的 EVM 地址——货款到链上后直接进你自己的钱包,平台不经手。

::: tip 想换收款地址?
默认地址不是固定的:到 **Checkouts** 页删掉默认配置,换成你自己的其他 EVM 地址即可。
:::

## 第 3 步:管理收银台

进入 **Checkouts** 页,可以看到默认收银台下七条链的收款地址与各自接受的币种,需要时也可以更换地址或关掉某条链。

![Checkouts:七条链收款地址](/images/payment/dashboard-checkouts.png#bordered)

页面下方的 **Share & collect** 可以生成**定额账单**——为某一笔具体金额生成一次性收款链接,发给买家即可收款。

## 第 4 步:查看到账

买家付款并在链上确认后,款项直接进入你的收款地址。**Transactions** 页和首页的 Recent transactions 会列出每一笔到账,含买家、金额、币种与链。

::: tip 想在自己的网站或应用里自动收款?
创建 API 密钥、调用接口、接收 webhook,见[开发者文档快速上手](/zh/developer/payment/quickstart)。
:::
