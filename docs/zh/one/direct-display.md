---
outline: [2, 3]
description: 将 Olares One 连接 HDMI 显示器、键盘和鼠标后作为本地电脑使用，同时保留浏览器访问应用和数据的能力。
head:
  - - meta
    - name: keywords
      content: Olares One, HDMI 显示器, 直连显示器, Olares Desktop, Node Display
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../one/direct-display.md)为准。
:::

# 使用直连显示器操作 Olares One <Badge type="warning" text="Alpha"/>

将 HDMI 显示器、键盘和鼠标连接到 Olares One 后，无需使用另一台电脑即可直接访问应用和数据。连接显示器后，仍可同时在浏览器中使用 Olares Desktop，两个访问会话互不影响。

:::warning Alpha 功能
此功能目前处于 **Alpha** 阶段，不建议用于生产环境。默认情况下，此功能处于关闭状态，需要安装预览版 `olares-desktop` 软件包。使用过程中可能会遇到性能问题，也可能需要额外的手动配置。如果遇到问题，请向 [Olares GitHub 仓库](https://github.com/beclab/Olares/issues)提交反馈。
:::

## 前提条件

**硬件**
- Olares One 已完成设置并已开机。
- 一台连接到 Olares One 的 HDMI 显示器、键盘和鼠标。
- 如果需要听游戏声音，还需要连接到 Olares One 的音频输出设备。
- 如果 Olares OS 已激活，需要一台安装了 LarePass 应用的移动设备，以便从 Vault 获取登录密码。

**系统**
- Olares One 运行 Olares OS v1.12.7。

**软件**
- 如果计划使用 Steam，请先将 Steam Headless 更新到最新版本。

## 步骤 1：直接访问 Olares One

1. 准备 Olares One 主机登录密码。

   - 如果 Olares OS 尚未激活，使用默认密码 `olares`。
   - 如果 Olares OS 已激活，在移动设备上打开 LarePass 应用。在 Vault 中找到带有 <span class="material-symbols-outlined">terminal</span> 图标的条目，然后点击查看密码。

   :::info 这不是 Olares Desktop 密码
   这个密码用于登录 Olares One 主机系统，不同于你在浏览器中登录 Olares Desktop 时使用的密码。
   :::

2. 在连接的显示器上显示的文本登录提示中输入用户名 `olares`，然后按 **Enter**。

   ```text
   olares login:
   ```

3. 输入上面准备好的主机登录密码，然后按 **Enter**。出于安全考虑，输入密码时屏幕上不会显示字符。

:::tip
如果希望远程准备 Olares One，可以在连接显示器之前通过 SSH 或 Control Hub 下载并安装软件包。具体可用方式请参阅[访问 Olares One 终端](./access-overview.md)。
:::

## 步骤 2：下载并安装 Olares Desktop 预览版软件包

1. 下载软件包：

   ```bash
   cd /tmp && wget https://cdn.olares.com/olares-one/desktop/olares-desktop_0.1.0_20260928_amd64.deb
   ```

2. 下载完成后，安装软件包：

   ```bash
   sudo bash -c 'apt update && DEBIAN_FRONTEND=noninteractive apt-get install -y -f ./olares-desktop_0.1.0_20260928_amd64.deb'
   ```

3. 如果出现关于配置文件 `/etc/default/apport` 的提示，按 **Enter** 接受默认选项 `N`，保留现有配置。

   ```text
   *** apport (Y/I/N/O/D/Z) [default=N] ?
   ```

4. 等待安装完成，直到终端重新显示命令提示符。

   ```text
   olares@olares:/tmp$
   ```

## 步骤 3：启动本地桌面

1. 运行以下命令：

   ```bash
   sudo start-desktop
   ```

2. 等待 Olares 登录界面出现，然后输入 Olares Desktop 密码登录。

## 步骤 4：为本地显示配置 Steam

如果计划使用 Steam，请在本地桌面启动后完成以下步骤。

:::warning
Steam 兼容模式是临时设置。每次执行 `sudo start-desktop` 后，都要完成[步骤 4](#步骤-4-为本地显示配置-steam)。每次执行 `sudo stop-desktop` 后，都要完成[步骤 7](#步骤-7-将-steam-恢复为默认模式)。
:::

1. 按住 **Ctrl**，然后反复按 **Tab**，直到本地终端（`olares@olares:~`）高亮显示。松开 **Ctrl** 打开终端。
2. 创建 Olares Desktop 配置。此配置用于启用 Steam 兼容模式。

   ```bash
   kubectl apply -f - <<EOF
   kind: ConfigMap
   apiVersion: v1
   metadata:
     name: olares-desktop-config
     namespace: os-framework
   data:
     enabled: 'true'
   EOF
   ```

   命令成功完成后，终端会显示：

   ```text
   configmap/olares-desktop-config created
   ```

3. 更新 Steam Headless 环境。将 `<username>` 替换为 Olares 用户名（例如 `laresprime`）。此命令会在本地桌面运行期间将 Steam Headless 设置为兼容模式。

   ```bash
   kubectl patch appenv steamheadless-<username> -n steamheadless-<username> --type='json' -p='[{"op": "add", "path": "/envs/2/value", "value": "secondary"}]'
   ```

   命令成功完成后，终端会显示：

   ```text
   appenv.sys.bytetrade.io/steamheadless-<username> patched
   ```

4. 按住 **Ctrl**，然后按 **Tab** 切回 Olares Desktop。
5. 打开 Olares Settings，进入 **Applications** > **Steam Headless**，停止应用，然后重新启动应用。
6. 从 Launchpad 打开 Steam Headless，然后登录。

## 步骤 5：在顶层窗口之间切换

按住 **Ctrl** 不松开，反复按 **Tab**，在可用的顶层窗口之间切换。松开 **Ctrl**，打开当前高亮的窗口。

默认窗口包括：

- **Olares Desktop**：本地 Olares Desktop 界面，用法与浏览器中的 Olares Desktop 相同。
- **Node Display**：用于管理 Olares One 设备的本地界面。
- **`olares@olares:~`**：本地终端。

![使用 Ctrl+Tab 在 Olares Desktop、Node Display 和本地终端之间切换](/images/one/direct-display-window-switcher.png#bordered)

## 步骤 6：停止直连桌面

1. 保存所有打开的应用中的工作。
2. 按 **Ctrl+Tab** 切换到本地终端（`olares@olares:~`），然后运行：

   ```bash
   sudo stop-desktop
   ```

## 步骤 7：将 Steam 恢复为默认模式

如果在[步骤 4](#步骤-4-为本地显示配置-steam)中启用了 Steam 兼容模式，请使用以下命令恢复 Steam 设置。只有在[步骤 6](#步骤-6-停止直连桌面)停止直连桌面后，才能运行这些命令。

1. 删除 Olares Desktop 配置。此操作会在停止直连桌面后禁用兼容配置。

   ```bash
   kubectl delete cm -n os-framework olares-desktop-config
   ```

   命令成功完成后，终端会显示：

   ```text
   configmap "olares-desktop-config" deleted
   ```

2. 恢复 Steam Headless 环境。将 `<username>` 替换为 Olares 用户名（例如 `laresprime`）。此操作会移除 Steam Headless 的临时兼容设置。

   ```bash
   kubectl patch appenv steamheadless-<username> -n steamheadless-<username> --type='json' -p='[{"op": "remove", "path": "/envs/2/value"}]'
   ```

   命令成功完成后，终端会显示：

   ```text
   appenv.sys.bytetrade.io/steamheadless-<username> patched
   ```
