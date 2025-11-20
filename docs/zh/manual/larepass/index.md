---
outline: [2, 3]
description: LarePass 用户文档。了解 LarePass 的核心功能与使用方法，包括账户管理、文件同步、设备与网络管理、系统升级、密码管理，内容收藏等，并提供下载与安装指南。
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
- [升级 Olares](manage-olares.md#升级-olares)
- [双因素登录 Olares](activate-olares.md#使用-larepass-进行双因素验证)
- [管理 Olares](manage-olares.md)
- [切换有线/无线网络](manage-olares.md#有线切换至无线)

### 安全文件访问与同步
- [使用 LarePass 管理文件](manage-files.md)

### 密码与密钥管理
使用 Vault 自动填充凭证、存储密码并生成 2FA 代码。
- [自动填充密码](autofill.md)
- [生成 2FA 代码](two-factor-verification.md)

### 知识收藏
通过 LarePass 收集网页内容并订阅 RSS。
- [通过 LarePass 扩展收集内容](manage-knowledge.md#通过-larepass-扩展收集内容)
- [订阅 RSS 源](manage-knowledge.md#订阅-rss-源)

## 功能对比

<table style="border-collapse: collapse; width: 100%; font-family: Arial, sans-serif;">
  <thead>
    <tr>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">类别</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">功能</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">移动端</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">桌面端</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">Chrome 扩展</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td rowspan="4" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">账户管理</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">创建 Olares ID</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">导入 Olares ID</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">多账户管理</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">SSO 登录</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td rowspan="4" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">设备与网络管理</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">激活 Olares</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">查看资源消耗</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">远程设备控制</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">管理 VPN 连接</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td rowspan="7" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">知识与文件管理</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">跨设备同步文件</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">管理 Olares 上的文件</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">收集网页/视频/播客/PDF/电子书至 Wise</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">下载视频/播客/PDF/电子书至文件</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">添加 RSS 订阅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">沉浸式翻译</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">备份手机上的照片和文件</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td rowspan="5" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">密钥管理</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">生成、共享和自动填充强密码及通行密钥</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">一次性身份验证管理</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Cookies 同步</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">第三方 SaaS 账户集成</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">可验证凭证 (VC) 卡片管理</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
  </tbody>
</table>

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

1. 访问 [LarePass 网站](https://olares.cn/olares) 下载扩展 ZIP 包。
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
