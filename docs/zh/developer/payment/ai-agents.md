---
outline: [2, 3]
description: 用一条提示词让 AI 智能体构建接入 Olares Payment 的商店，并安装到 Olares OS，得到可公开收款的应用地址。
head:
  - - meta
    - name: keywords
      content: Olares Payment, AI 智能体, olares-cli, Agent Skills, demo 商店, Olares 应用, chart
---

# 用 AI 智能体搭建收款商店

本文说明如何借助 AI 智能体构建接入 Olares Payment 的最小商店，并将其打包安装到 Olares OS。安装完成后，商店将获得可公开访问的域名，能够接收任意地方的下单。

下文以 `olarespayment@olares.com` 为例演示。请将其替换为你自己的 Olares ID。

:::warning
本教程需要 Olares OS 1.12.5 或更高版本。
:::

## 准备工作

- **准备好 AI 智能体**：任意 AI 编程智能体，如 Codex、Claude Code、Cursor、DeepSeek Harness。
- **安装 olares-cli 和 Agent Skills**：见[安装 olares-cli](../cli-install) 和[安装与使用 Agent Skills](../cli-agent-skills)。
- **用 olares-cli 登录 Olares OS**：步骤见[登录 Olares](../cli-log-in)。确保当前登录的是文中示例的 Olares ID，如图。

  ![olares-cli profile list 登录状态](/images/payment/cli-profile-list.png#bordered)

- **能把镜像推送到远程仓库**：本机已登录 Docker Hub 或其他镜像仓库，可将镜像推送到仓库，供 Olares 节点拉取。
- **在 Olares Payment 后台创建 API key**：拿到 `pk_live_…`（API key）和 `sk_live_…`（API secret）。步骤见[快速开始](./quickstart)。Webhook 地址填 `https://demo.olarespayment.olares.com/webhook`，保存后会生成 `whsec_…`。

  ![Olares Payment 后台中的 API key 与 webhook](/images/payment/dashboard-api-webhook.png#bordered)

  地址里的 `olarespayment.olares.com` 由 Olares ID `olarespayment@olares.com` 派生，把 `@` 换成 `.`。替换成你自己的 Olares ID。

- **准备工作目录**：建一个空目录。把上面的 API key、API secret 和 webhook secret 写进 `.env`：

```plain
# .env
PAYMENT_API_KEY=pk_live_…
PAYMENT_API_SECRET=sk_live_…
PAYMENT_WEBHOOK_SECRET=whsec_…
```

## 构建并安装到 Olares OS

把下面的支付 API 文档和提示词发给智能体。跑完后，商店会装到你的 Olares OS 上，打开就能下单收款。

<PaymentLlmsLink />

```plain
做一个能收款的最小商店（Node.js），商品 0.01 美元，
密钥从 .env 读取。按我给的 Olares Payment API 文档接入。

打包安装到我已用 olares-cli 登录的 Olares OS。
商店地址用 https://demo.olarespayment.olares.com，
打开就能下一单并完成收款。
```

:::tip
提示词里的商店地址是示例。把 `olarespayment.olares.com` 换成你的 Olares ID 派生地址。
:::

执行完毕后，商店页面如下：

![demo 商店](/images/payment/ds-v4-flash-store.png#bordered)

:::tip
本 demo 由 DeepSeek-V4.1-Flash 生成。
:::

## 下一步

最小商店已经能收款。接下来可以继续扩展，比如让智能体做一个订单管理后台；等打磨好了，再把它[发布到 Olares 应用市场](../develop/submit-apps)，让更多人用上。

