# VC 服务

本文档旨在帮助你了解如何使用以下工具，为你自己的场景配置各种 schema 并创建发行方 (Issuers) 和验证方 (Verifiers)：
-   Olares 的 [verifiable-credential-gate](https://github.com/Above-Os/verifiable-credential-gate) 和 [did-gate](https://github.com/Above-Os/did-gate)
-   tbd 团队的 [SSI Service](https://github.com/TBD54566975/ssi-service)

::: tip
如果你想进行更底层的开发，请结合 [SSI SDK](https://github.com/TBD54566975/ssi-sdk) 中的源代码阅读[协议标准](#reference)。
:::

## 简介

我们已经了解了 [VC](/zh/manual/concepts/vc.md) 的概念以及申请 VC 的基本流程。

在深入实现细节之前，我们可以先熟悉一下在钱包 (Wallet)、验证方 (Verifier) 和发行方 (Issuer) 的实际通信过程中会遇到的术语。
| 术语                      | 定义                                                                                                                                                                                                                                                                                                       |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **持有方 (Holder)** | 持有一个或多个可验证凭证的实体。持有方也是向验证方提交证明以满足其在“呈现定义 (Presentation Definition)”中描述的需求的实体。                                                                                                                                                                                  |
| **发行方 (Issuer)** | 一个实体可以扮演的角色，通过对一个或多个主体断言声明，从这些声明中创建可验证凭证，并将其传输给持有方。                                                                                                                                                                                                     |
| **验证方 (Verifier)** | 定义了需要从持有方那里获取何种证明（通过“呈现定义”）以进行交互的实体。                                                                                                                                                                                                                                     |
| **可验证凭证 (VC)** | 一种防篡改的凭证，其所有权可以被加密验证。可验证凭证可用于构建可验证表述 (Verifiable Presentations)，该表述也可被加密验证。凭证中的声明可以关于不同的主体。PEX 接受三种形式的可验证凭证：1. JSON_LD，在我们的系统中称为 IVerifiableCredential；2. JWT 包装的 VC，在我们的系统中称为 JwtWrappedVerifiableCredential；或 3. 字符串，即一个有效的可验证凭证 jwt。 |
| **可验证表述 (VP)** | 一种防篡改的表述，其编码方式使得数据的所有权在经过加密验证后可以被信任。                                                                                                                                                                                                                                         |
| **清单 (Manifest)** | 凭证清单 (Credential Manifests) 用于描述哪些凭证可供发行。                                                                                                                                                                                                                                               |
| **申请 (Application)** | 持有方提供给发行方的格式。                                                                                                                                                                                                                                                                               |
| **呈现定义 (Presentation Definition)** | 阐明验证方需要何种证明的对象。                                                                                                                                                                                                                                                             |
| **呈现 (Presentation)** | 源自一个或多个发行方所发行的一个或多个可验证凭证的数据。                                                                                                                                                                                                                                                 |
| **提交 (Submission)** | TBC (待补充)                                                                                                                                                                                                                                                                                           |
| **定义 (Definition)** | TBC (待补充)                                                                                                                                                                                                                                                                                           |
| **Schema** | 所有不同的清单、申请、凭证、呈现、定义都需要使用 JSON 来定义 Schema。服务将验证所提交数据的正确性，然后进入业务流程。                                                                                                                                                                                           |

## 参考资料

以下参考资料来自 [SSI SDK](https://github.com/TBD54566975/ssi-sdk) 项目。

### 规范

以下是本库目前支持的一系列规范参考。这是一个动态集合，会随着库的演进而变化。

-   [去中心化标识符 (DIDs) v1.0](https://www.w3.org/TR/did-core/) W3C 建议提案 2021年8月3日
-   [可验证凭证数据模型 v1.1](https://www.w3.org/TR/vc-data-model/) W3C 推荐标准 2021年11月9日
    -   支持 [链接数据证明 (Linked Data Proof)](https://www.w3.org/TR/vc-data-model/#data-integrity-proofs) 格式。
    -   支持 [VC-JWT 和 VP-JWT](https://www.w3.org/TR/vc-data-model/#json-web-token) 格式。
-   [可验证凭证 JSON Schema 规范](https://w3c-ccg.github.io/vc-json-schemas/v2/index.html) 社区组报告草案 2021年9月21日
-   [呈现交换 (Presentation Exchange) 2.0.0](https://identity.foundation/presentation-exchange/) 工作组草案 2022年3月
-   [钱包渲染草案 (Wallet Rendering Strawman)](https://identity.foundation/wallet-rendering/) 2022年6月
-   [凭证清单 (Credential Manifest)](https://identity.foundation/credential-manifest/) 草案 2022年6月
-   [状态列表 (Status List) 2021](https://w3c-ccg.github.io/vc-status-list-2021/) 社区组报告草案 2022年4月4日

### 签名方法

> -   [数据完整性 (Data Integrity) 1.0](https://w3c.github.io/vc-data-integrity/) 社区组报告草案
> -   [链接数据加密套件注册表 (Linked Data Cryptographic Suite Registry)](https://w3c-ccg.github.io/ld-cryptosuite-registry/) 社区组报告草案 2020年12月29日
> -   [JSON Web 签名 (JSON Web Signature) 2020](https://w3c-ccg.github.io/lds-jws2020/) 社区组报告草案 2022年2月9日
>     -   [VC 证明格式测试套件，使用 JSON Web 签名的 VC 数据模型](https://identity.foundation/JWS-Test-Suite/) 非官方草案 2022年3月9日 本实现对 JWS 测试套件的合规性可以在此处找到。
>     -   支持 JWT 和链接数据证明两种格式，并符合 [JOSE 规范](https://jose.readthedocs.io/en/latest/)。