---
outline: [2, 4]
---

# TerminusDID 合约系统

## 架构

TerminusDID 合约系统分为两个部分：DID 管理和标签管理。除了标签管理的核心功能外，我们还实现了一个官方 Tagger 和一个信誉系统。

```mermaid
graph TB

    RootTagger{RootTagger}

    RSAPubKey[/RSAPubKey/]
    DNSARecord[/DNSARecord/]
    AuthAddress[/AuthAddress/]
    otherTag[/.../]




    TerminusDID{TerminusDID}
    Reputations[/Reputations\]
    XXXReputation{...}
    AppMarketReputation{AppMarketReputation}
    Tag[[Tag]]
    Tagger([Tagger])


    DID[DID]
    Domain((Domain))
    com((com))
    net((net))
    io((io))

    TerminusDID--->DID
    TerminusDID--->Tag
    Tag--->Tagger

    Tagger-.-RootTagger
    Tagger-.-AppMarketReputation
    Tagger-.-XXXReputation


    Domain-.->Tag


    subgraph ide4 [Tag]

    subgraph ide1 [OfficialTag]
    RootTagger--->RSAPubKey
    RootTagger--->DNSARecord
    RootTagger--->AuthAddress
    RootTagger--->otherTag
    end

    subgraph ide2 [Reputations]
    Reputations-.->AppMarketReputation
    Reputations-.->XXXReputation
    end

    end

    subgraph ide3 [DID]
    DID-.-Domain
    Domain--->com
    Domain--->net
    Domain--->io
    end

````

关于 DID/标签管理的用法，请参考[此处](https://www.google.com/search?q=./contract.md)；关于 TerminusDID 合约的设计细节，请参考[此处](https://www.google.com/search?q=././did/design.md)；关于信誉系统的推荐实现和示例，请参考[此处](https://www.google.com/search?q=./contract-reputation.md)。

## 设计细节

### Multicall

考虑到我们合约使用场景的复杂性和不确定性，我们添加了一个内置的 multicall 函数来简化链上交互。你可以在单笔交易中访问多个接口，而无需外部辅助合约。

### EIP-7201

我们遵循 EIP-7201 进行合约数据存储，这使得合约升级更轻松、更安全，也有利于实现对数据的精细化控制。

### Olare ID 的验证

在注册时，Olare ID 是以 `string` 类型提交的。尽管它通过 `.` 来进行层级分离，但这无法保证其正确性和可读性。我们在合约代码中实现了额外的验证，以确保提交的名称是经过 UTF-8 编码的可读字符串。

### 标签中结构体的字段名

为了提高 Gas 效率，如果标签类型中包含结构体，其字段名将通过以太坊事件 (Ethereum events) 发布在链上，而不占用合约存储。事件将记录定义该标签类型时的区块高度。使用区块高度、合约地址、事件签名和布隆过滤器 (Bloom filters)，可以精确地获取所需事件。

### 内联汇编 (Inline Assembly)

我们使用基于内联汇编的切片类型来解析和遍历 Olares ID 的层级，这可以避免复制子字符串并降低 Gas 消耗。

## 附录 - 需求

### DID

- 链上 [DID](https://www.w3.org/TR/did-core/) 管理（偏好于 EVM 兼容链）
- 为 DID 记录 IPv4/IPv6、头像、RSA 公钥等信息
- 为未来可能出现的新需求扩展存储数据

### Olares ID

- 为 Olares ID 区分两种 DID：组织 (Organization) 或个人 (Individual)
- 为父级组织派生 DID 管理权限

### 信誉 (Reputation)

- 使用另一种称为“实体 (Entity)”的 DID 来代表现实世界中的对象（例如，用于应用市场信誉的应用版本和用于 Otmoic 信誉的投诉）

:::tip 提示
我们希望基于 DID 合约设计一个去中心化的信用体系。起初，我们设想的是一个能适应各种场景的通用系统，但随着逐步实现，我们发现这个庞大而全面的系统会带来许多不必要的资源消耗，并降低在不同场景下的灵活性。因此我们改变了方向：我们提供一个用于抽象和必要组件的信誉 (Reputation) 系统，并提供一些推荐实现。用户可以组合和定制实现，以满足自己的场景需求。
:::

