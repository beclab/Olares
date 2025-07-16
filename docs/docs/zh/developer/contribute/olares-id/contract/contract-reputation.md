# 声誉 (Reputation)

#### 额外需求：抽象的声誉系统

:::info 所需功能

-   **映射到现实世界对象的标识符**
    所有信用系统都基于一些特定场景，例如对应用的评价或对商家的投诉。这不可避免地需要我们为一些现实世界或抽象的对象建立唯一的链上标识符。

-   **身份验证**
    信用系统最核心的功能是评价，而评价者的身份验证是其先决条件之一。

-   **必要信息的存储**
    用户评价后，我们可能需要记录评论或计算一些加权统计数据。

-   **历史记录与更新**
    在某些情况下，我们可能需要记录评价行为，并允许将来更新现有评价。
:::

本章为每个需求提供了相应的解决方案。

## 待评价的对象

对于待评价的对象，可能会出现以下两种情况：

-   **是一个商家或单一个人**
    在这种情况下，该对象必须拥有一个“个人 (Individual)”类型的 DID，这样我们就可以直接操作该对象，而无需创建其他 DID。
-   **是一个现实世界的对象或抽象概念**
    在这种情况下，我们首先需要为此对象创建一个“实体 (Entity)”类型的 DID，然后对其进行操作。

## 身份验证

身份验证有两种解决方案：

-   使用 DID 的所有者 (owner) 提交交易 (tx)，并让 Tagger 调用 Olares DID 合约的接口进行身份验证。
-   使用 DID 的所有者签署一个自定义的 EIP-712 消息，并使用一个转发器 (forwarder) 将交易发送上链。Tagger 将使用消息的签名者进行身份验证。

:::tip 提示
我们推荐第二种方案，因为交易费用由转发器支付，而不是评价者。
:::

## 必要信息的存储

对于“实体 (Entity)”类型的 DID，我们定义该 DID 自身的标签，并将必要的数据写入其中。对于“个人 (Individual)”类型的 DID，我们将该场景抽象为一个“实体 (Entity)”类型的 DID，并向其标签中写入数据。

## 历史记录与更新

我们可以遵循上一节的建议，将数据存储在实体的标签中。但在某些场景中并不需要链上查询，此时链上存储就是一种浪费。对于这些情况，我们建议使用以太坊事件 (Ethereum events)进行记录，并在 Tagger 中自定义具体实现。

## 示例 - OtmoicReputation

:::info

```mermaid
  flowchart LR

    otmoic{{OtmoicReputation}}
    did{{TerminusDID}}
    complaints[/tag-complaints/]
    otmoicdid((otmoic.reputation))

    otmoic-- Authentication -->did
    did-.->otmoicdid
    otmoicdid-.-complaints
    otmoic-- read/write -->complaints
````

Otmoic Reputation 合约使用 DID 所有者的 EIP-712 签名进行身份验证，并将被投诉的出价 `bidid` 存储在实体 `otmoic.reputation` 的 `complaints` 标签中。
:::

## 示例 - TerminusAppMarketReputation

:::info
```mermaid
  flowchart TD

    app((app.myterminus.com))
    appname((appname.app.myterminus.com))
    version1((version1.appname.app.myterminus.com))
    version2((version2.appname.app.myterminus.com))
    reputation{{TerminusAppMarketReputation}}
    ratings[/tag-ratings/]
    version1ratings[/version1-tag-ratings/]
    event[[event]]

    app-->appname
    appname-->version1
    appname-->version2
    app-.->ratings
    ratings-.-version1ratings
    version1-.-version1ratings
    reputation-.- read/write -.->version1ratings
    reputation-.->event
```
Olares App Market Reputation 合约同样使用 DID 所有者的 EIP-712 签名进行身份验证。合约的评分数据存储在实体 `<version>.<appname>.app.myterminus.com` 的 `ratings` 标签中，而评论数据则以以太坊事件的形式发布，而不是存储在链上。
:::
