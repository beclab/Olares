---
outline: [2, 3]
description: LarePass 用户文档。了解 LarePass 的核心功能与使用方法，包括账户管理、文件同步、设备与网络管理、密码管理，内容收藏等，并提供下载与安装指南。
---

# LarePass 使用文档

**LarePass** 是 Olares 的官方跨平台客户端，为用户与 Olares 系统之间建立安全桥梁。无论是移动端、桌面端还是浏览器，你都可以随时随地借助 LarePass 实现无缝访问、身份、密码管理、文件同步、内容管理。

![LarePass](/images/manual/larepass/larepass.png)

## 主要功能

### 账户与身份管理
创建和管理 Olares ID，安全备份凭证并连接外部服务。
- [创建 Olares ID](create-account.md)
- [备份助记词](back-up-mnemonics.md)
- [设置或重置本地密码](back-up-mnemonics.md#设置本地密码)
- [管理集成服务](integrations.md)

### 启用专用网络
随时随地通过 LarePass 专用网络访问 Olares。
- [打开专用网络](private-network.md#在-larepass-中启用专用网络)
- [排查连接问题](private-network.md#故障排查)

### 设备管理
激活并管理 Olares 设备，通过 LarePass VPN 安全连接。
- [激活 Olares 设备](activate-olares.md)
- [双因素登录 Olares](activate-olares.md#使用-larepass-进行双因素验证)
- [管理 Olares](manage-olares.md)
- [切换有线/无线网络](manage-olares.md#有线切换至无线)

### 安全文件访问与同步
跨设备访问并同步 Olares 文件。
- [使用 LarePass 管理文件](manage-files.md)
- [同步与共享文件](sync-share.md)

### 密码与密钥管理
使用 Vault 自动填充凭证、存储密码并生成 2FA 代码。
- [自动填充密码](autofill.md)
- [生成 2FA 代码](two-factor-verification.md)

### 知识收藏
通过 LarePass 收集网页内容并订阅 RSS。
- [通过 LarePass 扩展收集内容](manage-knowledge.md#通过-larepass-扩展收集内容)
- [订阅 RSS 源](manage-knowledge.md#订阅-rss-源)

---

## 下载与安装 LarePass

前往 [LarePass 官网](https://www.olares.cn/larepass) 获取适用于你设备的最新版本。

### 安装 LarePass 浏览器扩展

<tabs>
<template #从-Chrome-Web-Store-安装>

1. 在 [Chrome 网上应用店](https://chrome.google.com/webstore) 搜索 **LarePass**。
2. 打开详情页并点击 **添加至 Chrome**。
3. 通过导入 Olares ID 登录扩展：
    - 打开 LarePass 扩展，点击 **导入账户**。
    - 输入 Olares ID 的助记词。
    - 输入 Olares 密码完成登录。

</template>

<template #离线安装>

1. 访问 [olares.com/larepass](https://olares.com/larepass) 下载扩展 ZIP 包。
2. 在浏览器地址栏输入 `chrome://extensions/`。
3. 打开右上角 **开发者模式**。
4. 点击 **加载已解压的扩展程序**，选择解压后的 LarePass 文件夹。
5. 登录流程：
    - 打开 LarePass 扩展，点击 **导入账户**。
    - 输入 Olares ID 的助记词。
    - 输入 Olares 密码完成登录。
</template>
</tabs>

::: tip 快速访问
安装完成后，可在 Chrome 扩展菜单中固定 LarePass，方便一键启动。
:::
