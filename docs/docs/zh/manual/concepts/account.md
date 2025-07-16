---
outline: [2, 3]
description: Olares 账户系统的同步机制、账户阶段和统一认证原理。包括多因素认证机制、密码系统及多设备账户同步原理。
---

# 账户

本文介绍 Olares 账户系统的相关概念和设计。

## 账户同步

LarePass、Olares 和 Olares Space 之间的账户同步机制如下：

- 创建 Olares 时需要提供 Olares ID，并使用该 ID 登录 LarePass 进行激活。
- 登录 Olares Space 时，需要使用 LarePass 扫描二维码。

## 账户的状态

每个账户都有三个状态：

### 未绑定 Olares ID（DID 阶段）
已在本地创建 DID、助记词和私钥，但尚未关联 Olares ID。

在这个阶段，你可以导出助记词，也可以访问 Olares Space 配置自定义域名或申请组织域名。

但此时无法将账户导入到其他 LarePass 客户端。
:::tip 提示
在 LarePass 上，当你点击**创建账户**时，就已经处于未绑定 Olares ID 状态。
:::
### 已绑定 Olares ID
当账户绑定了 Olares ID 后，系统会在区块链上记录你的 Olares ID 与 DID 之间的关联。

在这个阶段，你可以通过命令行在本机上安装 Olares，或在 Olares Space 申请并激活 Olares。

你也可以使用导出的助记词将账户导入其他设备，实现应用间的统一认证。

### 已绑定 Olares
最后一个阶段是账户与 Olares 设备建立关联，此时你可以完整访问 Olares 上的所有功能，例如在 LarePass 上查看该 Olares ID 下的机器系统资源。

## 统一账户系统

Olares 支持多用户系统的统一认证。

1. 用户在登录页面完成登录后，后续所有请求都会自动包含认证信息。
2. 每个用户请求都会先经过 Authelia 服务进行认证。
3. 如果认证失败，应用会将用户重定向到登录页面重新认证。
4. 如果认证成功，[Backend for Launcher (BFL)](https://github.com/beclab/bfl) 会附加用户的基本信息并将请求转发给应用服务。这样应用本身就不需要处理认证逻辑。
5. 对于[共享应用（Shared application）](./application.md#共享应用)，开发者需要构建额外的 `Auth Server` 来连接应用账号与 BFL 账号。

## 多因素认证（MFA）

Olares 集成了多种不同安全等级的认证因素，以确保系统中用户身份认证的安全性。

### 密码

首次激活或创建子用户时，Olares 会生成一个随机密码用于初始设置。完成身份验证后，系统会提示用户通过 LarePass 将这个初始密码更换为更强的自定义密码。

### 一次性密码

当用户执行登录等敏感操作时，Olares 要求用户输入 LarePass 生成的一次性双因素认证码。

## 了解更多

### 用户

- [创建 Olares ID](../get-started/create-olares-id)
- [用户角色](../olares/settings/roles-permissions.md)

### 开发者

- [账户系统回调](../../developer/develop/advanced/account.md)
