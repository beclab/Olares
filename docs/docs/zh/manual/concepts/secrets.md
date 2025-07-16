---
description: Olares 密钥管理体系说明，包括 Vault 项目、凭据、密钥和集成凭据的分类与安全机制，以及讲解敏感数据的存储策略。
---
# 密钥

Olares 根据使用场景对密钥进行分类并采用不同的管理方式。

|          | 数据类型                                | 存储位置                               | 泄露风险                                                | 使用方式                                             |
|----------|-------------------------------------|------------------------------------|-----------------------------------------------------|--------------------------------------------------|
| Vault 项目 | 包括网站密码、数<br/>据库密码、区块链<br/>私钥等       | Vault                              | Olares 中的加密数据确保第三方即使登录也无法查看                         | 每次使用都需要 LarePass 签名                              |
| 凭证       | 安全认证后获取<br/>的系统访问凭据：<br/>令牌、Cookie 等 | [Infisical](https://infisical.com/) | 第三方在 Olares 中通过特定步骤认证后可查看                           | 应用程序获得 Provider 权限后可通过 API 使用                    |
| 密钥       | Pod 容器中使用的<br/>敏感数据，如数据<br/>库连接和管理账号     | ETCD                               | 可在[控制面板](../olares/controlhub/navigate-control-hub.md#保密字典)直接查看 | 用于 Helm 部署模板，通过 `valueFrom -> secretKeyRef` 注入环境变量 |

## 集成凭据

用户可以通过在设置中登录第三方服务账号，让 Olares 中的应用访问外部服务凭据。例如：

- 登录 Olares Space 后，备份服务可以请求令牌用于自动后台备份
- 登录 Google 后，文件功能可以与 Google Drive 中的数据同步

Olares 中的应用程序可以通过[Service Provider](../../developer/develop/advanced/provider.md) 获取这些第三方服务凭据。

## 应用凭据

- Olares 中的应用可以通过系统提供的接口管理和使用[密钥](../../developer/develop/advanced/secret.md)
- 应用生成的凭据仅限该应用程序使用