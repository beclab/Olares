---
outline: [2, 3]
description: 通过 LarePass 在移动设备上打开和管理 Olares 应用，了解安装、更新、克隆、停止、恢复及卸载应用的方法。
---

# 使用 LarePass 管理应用

通过 LarePass 的 **Apps** 标签页，你可以在移动设备上打开、安装和管理 Olares 应用。这些应用安装并运行在 Olares 上。

## 开始前

在移动设备上访问应用前，请确认满足以下条件。

- Olares 已升级至 1.12.8 或更高版本。
- LarePass 移动端已升级至最新版本，且包含 **Apps** 标签页。
- 已在 LarePass 中登录 Olares 账户，且移动设备能够连接到 Olares。远程访问方式参见 [LarePass VPN](private-network.md)。

<!-- 发布前检查：确认支持此功能的 LarePass 移动端最低版本。 -->

## 移动端支持

Olares 的所有系统应用都支持在移动设备上使用。Olares 上已安装的其他应用也都会显示在移动设备的 Apps 标签页中，但有些应用的入口可能不支持移动端。这只影响入口能否在移动端打开。即使你通过手机安装，应用本身也会正常安装并运行在 Olares 上。

不支持的入口会显示为灰色，并带有电脑标记。点击后会弹出 **Entrance unavailable** 面板，提示你在桌面端打开。你仍可以点击 **Manage**（管理），或长按应用图标，执行当前账户有权限进行的管理操作。

## 访问已安装的应用

Apps 标签页显示 Olares 上已安装的应用，默认仅显示运行中的应用。你可以通过筛选菜单查看已停止的应用或全部应用。

1. 打开 LarePass，点击底部的 **Apps** 标签页。

   ![LarePass Apps 标签页](/images/manual/larepass/larepass-apps-tab.png#bordered)

2. 在图标列表中找到应用，或在 **Search apps**（搜索应用）中输入名称。
3. 点击应用图标打开应用。
4. 如需查看已停止的应用或全部应用，点击搜索框旁的 <i class="material-symbols-outlined">more_horiz</i>，选择 **Stopped**（已停止）或 **All**（全部）。

   ![筛选和刷新应用](/images/manual/larepass/larepass-apps-filter-menu.png#bordered)

5. 如需重新加载应用列表，打开同一菜单，选择 **Refresh**（刷新）。

### 返回上一层

打开应用后，可以通过应用工具栏返回上一层。

1. 在已打开的应用中，点击顶部的应用名称（例如 **Market**），显示工具栏。
2. 点击左上角的 **Back** 返回上一层。

## 安装应用

通过 LarePass 中的应用市场查找新应用，并将其安装到 Olares 上。

1. 在 **Apps** 标签页中点击 **Market**（应用市场）。
2. 找到要安装的应用，打开详情页，检查硬件和资源要求。
3. 点击**获取**，再点击**安装**。如有配置提示，按要求完成配置。
4. 安装完成后，返回 **Apps** 查找应用，如有需要，点击搜索框右侧的 <i class="material-symbols-outlined">more_horiz</i>，选择 **Refresh**（刷新）。

## 更新应用

应用市场会列出已安装应用的可用更新。你可以单独更新某个应用，也可以一次更新全部应用。

1. 在 **Apps** 标签页中打开 **Market**（应用市场）。
2. 点击左上角的 <i class="material-symbols-outlined">menu_open</i>，选择**更新**。
3. 单独更新目标应用，或点击**全部更新**。

## 克隆应用

克隆可为已安装的应用创建另一个实例，使用独立配置。只有支持多实例的应用才能克隆。

1. 在 **Apps** 标签页中打开 **Market**（应用市场）。
2. 点击左上角的 <i class="material-symbols-outlined">menu_open</i>，选择**我的 Olares**。
3. 找到目标应用，点击操作按钮右侧的下拉箭头，选择**克隆**。
4. 按提示为新实例命名并完成配置，然后确认。
5. 克隆完成后，返回 **Apps** 查找新实例。

克隆条件和配置说明，参见[克隆应用](../olares/market/clone-apps.md)。

## 停止和恢复应用

暂时不用的应用可以停止，以释放资源，需要时再恢复运行。两种操作都可以从 Apps 或应用市场完成。操作是否可用取决于应用状态和当前账户权限。系统应用不能像普通已安装应用一样管理。

<tabs>
<template #从-Apps-操作>

- **停止应用**：长按应用图标，打开 **Manage [应用名称]** 面板，选择 **Stop app**（停止应用）。
- **恢复应用**：点击搜索框旁的 <i class="material-symbols-outlined">more_horiz</i>，选择 **Stopped**（已停止），然后长按应用图标，选择 **Resume app**（恢复应用）。
</template>
<template #从应用市场操作>

1. 在 **Apps** 标签页中打开 **Market**（应用市场），进入**我的 Olares**。
2. 找到目标应用，点击操作按钮右侧的下拉箭头。
3. 选择**暂停**或**恢复**。
</template>
</tabs>

停止和恢复应用的更多说明，参见[应用市场常见问题](../olares/market/market.md#常见问题)。

## 卸载应用

不再需要的应用可以通过 Apps 或应用市场从 Olares 中卸载。

:::warning 卸载应用
卸载会从 Olares 移除应用，而不只是删除 LarePass 中的图标。如果选择删除应用数据，重新安装也无法恢复这些数据。确认前请备份需要保留的数据。卸载共享服务可能影响其他用户，请检查确认窗口中说明的影响范围。卸载选项的具体含义，参见[卸载应用](../olares/market/market.md#卸载应用)。
:::

只有当前账户有权限移除的应用才能卸载。系统应用无法卸载。

<tabs>
<template #从-Apps-操作>

1. 在 **Apps** 标签页中找到目标应用。如应用已停止，先点击搜索框旁的 <i class="material-symbols-outlined">more_horiz</i>，选择 **All** 或 **Stopped**。
2. 长按应用图标，在 **Manage [应用名称]** 面板中选择 **Uninstall**（卸载）。
3. 检查卸载选项，包括是否删除应用数据、是否移除共享服务，然后确认卸载。
</template>
<template #从应用市场操作>

1. 在 **Apps** 标签页中打开 **Market**（应用市场），进入**我的 Olares**。
2. 找到目标应用，点击操作按钮右侧的下拉箭头，选择**卸载**。
3. 检查卸载选项，包括是否删除应用数据、是否移除共享服务，然后确认卸载。
</template>
</tabs>
