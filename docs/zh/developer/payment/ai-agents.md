---
outline: [2, 3]
description: 用一条提示词让 AI 智能体构建接入 Olares Payment 的商店，并安装到 Olares OS，得到可公开收款的应用地址。
head:
  - - meta
    - name: keywords
      content: Olares Payment, AI 智能体, olares-cli, Agent Skills, demo 商店, Olares 应用, chart
---

# 用 AI 智能体搭建收款商店

这篇文章带你用 AI 智能体做一个能收稳定币的在线商店，安装到 Olares OS 上，获得一个任何地方都能访问的固定域名。

::: warning 版本要求
本教程需要 `olares-cli` 1.12.6 或更高版本，以及 Olares OS 1.12.6 或更高版本。
:::

全文只有两个值需要确认，先记下来，后面照抄即可：

| 值 | 本文示例 | 说明 |
|---|---|---|
| Olares ID 域名前缀 | `olarespayment` | 你的 Olares ID 中 @ 前的部分 |
| 商店入口名称 | `myshop` | 也是商店的子域名前缀，自取，但全文必须一致 |

按示例，商店的最终地址就是 `https://myshop.olarespayment.olares.com`。

整件事只有三步，写代码、打包、安装交给智能体完成：

| 步骤 | 你做什么 | 智能体做什么 | 做完后的状态 |
|---|---|---|---|
| 1. 准备 | 在 Olares Payment 商户后台创建 API 密钥，登记 webhook，把三个密钥写进 `.env` | — | 密钥全部就绪 |
| 2. 开发部署 | 发一段提示词 | 写代码、打包、安装 Olares 应用 | 商店上线，有了固定域名 |
| 3. 验收 | 打开商店下一单 | — | 订单变为“已支付” |

webhook 为什么能在第 1 步就登记？Olares 应用的域名是确定的，由入口名和 Olares ID 拼成，入口名由你在提示词里指定。所以 `https://myshop.olarespayment.olares.com/webhook` 这个地址在安装之前就是已知的。

## 开始前检查

- **准备好 AI 智能体**：任意 AI 编程智能体，如 Codex、Claude Code、Cursor、DeepSeek Harness。
- **安装 olares-cli 和 Agent Skills**：见[安装 olares-cli](../cli-install) 和[安装与使用 Agent Skills](../cli-agent-skills)。
- **用 olares-cli 登录 Olares OS**：步骤见[登录 Olares](../cli-log-in)。确保当前登录的是文中示例的 Olares ID。

  ```bash
  olares-cli profile list
  ```

  输出示例：

  ```text
      NAME                      OLARES-ID                 STATUS
  *   olarespayment@olares.com  olarespayment@olares.com  logged-in
  ```

  开头的 `*` 标记当前 profile。

- **Docker**：本机已 `docker login` 到 Docker Hub 或其他公共镜像仓库，第 2 步推镜像要用。

## 第 1 步：准备

1. 用 LarePass 扫码登录 Olares Payment 商户后台，首次登录会自动创建商户账号。

2. 进入 **Checkouts** 页面，打开默认商店的 **Advanced settings**，在 **API keys** 一栏创建密钥，拿到 `pk_live_…` 和 `sk_live_…`。（图文步骤见[快速开始](./quickstart)的前两步）

3. 在同一面板的 **Webhooks** 一栏，webhook 地址填 `https://myshop.olarespayment.olares.com/webhook`，保存，拿到签名密钥 `whsec_…`。

4. 建一个仓库，把三个密钥写进 `.env`：

   ```bash
   mkdir myshop && cd myshop
   git init
   echo ".env" >> .gitignore
   ```

   ```plain
   # .env
   PAYMENT_API_KEY=pk_live_…
   PAYMENT_API_SECRET=sk_live_…
   PAYMENT_WEBHOOK_SECRET=whsec_…
   ```

## 第 2 步：开发部署

在仓库目录下启动智能体（`codex`、`claude`，或用 Cursor 打开该文件夹），发送下面这段提示词。整段只有 `myshop` 一处需要改，换成你的入口名，和第 1 步 webhook 地址里的保持一致：

```plain
做一个最小的在线商店（Node.js）：一个商品页面、一个下单接口、一个订单结果页。
用 Olares Payment 收款：
- 下单时调用 createPayment 创建支付，把 checkoutUrl 返回给前端，跳转到托管收银台
- 提供一个 /webhook 接口接收 payment.succeeded，验签通过后把订单标记为已支付
- 密钥从 .env 读取（PAYMENT_API_KEY / PAYMENT_API_SECRET / PAYMENT_WEBHOOK_SECRET），
  不要写进代码

然后把它打包成 Olares 应用，安装到我的 Olares OS 上：
- 构建镜像并推送到公共镜像仓库
- 生成 chart：对外入口名固定为 myshop，入口设为公开访问，
  支付密钥做成可配置的环境变量
- 上传并安装，完成后告诉我商店的访问地址

支付 API 文档：https://www.olares.com/docs/developer/payment/llms-full.txt
```

智能体读完文档后会一口气做完：写代码、本地跑通、打包安装（打包细节在 `olares-chart`、`olares-market` 等 Skills 里，不用你指导）。入口名固定为 `myshop`，装出来的地址就是第 1 步登记 webhook 用的那个。

::: info 地址和预期不一致时
如果安装后智能体报告的地址和预期不一致（入口名被占用时 Olares 会分配别的域名），回商户后台把 webhook 地址改成实际域名即可。
:::

## 第 3 步：验收

打开 `https://myshop.olarespayment.olares.com`，真实下一单：跳转收银台 → 完成支付 → 返回商店，订单显示“已支付”。

订单状态是靠 webhook 通知翻转的，所以看到“已支付”就说明收款、通知、履约整条链路都通了。

执行完毕后，商店页面如下：

![商店页面](/images/payment/demo-store.png#bordered)

::: info 示例来源
本示例由 DeepSeek-V4.1-Flash 生成。
:::

## 下一步

开发更多电商功能，请参考我们的 demo 仓库：[olares-payment-developer-example](https://github.com/beclab/olares-payment-developer-example)。

