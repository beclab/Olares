---
outline: [2, 3]
description: 通过 LarePass 的 Apps 标签页打开 Olares 应用，在应用市场中安装和更新应用，并停止、恢复或卸载已安装的应用。
---

# 使用 LarePass 管理应用

通过 LarePass 手机端的 **Apps** 标签页，你可以访问和管理 Olares 上的应用。应用运行在 Olares 上。通过 LarePass 安装应用时，应用会安装到 Olares，而不是手机上。

## 开始前

- Olares 已升级至 1.12.8 或更高版本。
- LarePass 手机端已包含 **Apps** 标签页。
- 已在 LarePass 中登录 Olares 账户，且手机能够连接到 Olares。远程访问方式参见 [LarePass VPN](private-network.md)。

<!-- 发布前检查：确认支持此功能的 LarePass 手机端最低版本。 -->

本文使用 Apps 页的英文界面名称，与截图保持一致。

## 移动端支持

所有系统应用都支持在手机上使用。其他已安装应用的移动端支持正在逐步开放。应用安装成功不代表能在手机上打开。

不支持移动端的入口会显示为灰色，并带有电脑标记。点击后会提示在桌面端打开。你仍可以点击 **Manage**（管理），或长按应用图标，执行当前账户有权限进行的管理操作。

![不支持移动端的应用入口](/images/zh/manual/larepass/apps-mobile-unsupported.png#bordered){width=40%}

## 查找和打开应用

1. 打开 LarePass，点击底部的 **Apps** 标签页。列表默认显示运行中的应用。

   ![LarePass Apps 标签页](/images/manual/larepass/apps-home.png#bordered){width=40%}

2. 在图标列表中找到应用，或在 **Search apps**（搜索应用）中输入名称。
3. 点击应用图标打开应用。
4. 如需按状态筛选或重新加载列表，点击搜索框旁的 <i class="material-symbols-outlined">more_horiz</i>，选择 **All**（全部）、**Running**（运行中）、**Stopped**（已停止）或 **Refresh**（刷新）。

   ![筛选和刷新应用](/images/manual/larepass/apps-filter.png#bordered){width=40%}

### 在应用内返回

1. 在已打开的应用中，点击顶部的应用名称（例如 **Market**），显示工具栏。
2. 点击左上角的 **Back** 返回上一层。

## 安装应用

1. 在 **Apps** 标签页中点击 **Market**（应用市场）。
2. 浏览**发现**页面，或打开左上角菜单，点击**搜索**查找应用。
3. 打开应用详情，检查硬件和资源要求。
4. 点击**获取**，再点击**安装**。如有配置提示，按要求完成配置。
5. 等待安装完成。返回 **Apps** 查找应用，如有需要，从 more_horiz 菜单选择 **Refresh**（刷新）。

## 停止、恢复或卸载应用

1. 在 **Apps** 标签页中长按目标应用图标。
2. 在底部弹出的 **Manage [应用名称]** 面板中选择操作：

   | 操作 | 作用 |
   | --- | --- |
   | **Stop app** | 停止 Olares 上的应用。 |
   | **Resume app** | 恢复已停止的应用。 |
   | **Uninstall** | 从 Olares 卸载应用。 |

   :::info
   操作是否可用取决于应用状态和当前账户权限。系统应用不能像普通已安装应用一样管理。
   :::

3. 如出现确认窗口，阅读操作说明后再继续。

   :::warning 卸载应用
   卸载会从 Olares 移除应用，而不只是删除 LarePass 中的图标。如果选择删除应用数据，重新安装也无法恢复这些数据。确认前请备份需要保留的数据。共享服务的操作可能影响其他用户。请检查确认窗口中说明的影响范围。
   :::

## 更新应用

1. 在 **Apps** 标签页中打开 **Market**（应用市场）。
2. 打开左上角菜单，点击**更新**。
3. 单独更新目标应用，或点击**全部更新**。
