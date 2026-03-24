---
outline: [2,4]
description: 通过检查 nginx 重启、网络连接、GPU 模式、工作模式、游戏兼容性以及 Steam Headless 资源上限，排查并解决 Olares 上的 Steam 串流卡顿或延迟问题。
---
# Steam 串流卡顿或延迟

从 Olares 串流 Steam 游戏时，如果感觉卡顿或有延迟，可参考本指南进行排查。

## 适用情况

串流游玩 Steam 游戏时出现卡顿或延迟，具体表现可能为高延迟、画面卡顿、突然断连或输入响应迟缓。

## 原因

Steam 串流卡顿或延迟，可能由以下一个或多个原因导致：

- **nginx 重启**：nginx 重启会打断串流。
- **网络连接**：Olares 设备使用的是 Wi-Fi 而不是有线网络，导致带宽较低或延迟较高。
- **GPU 分配**：GPU 资源被其他应用占用。
- **GPU 功率**：GPU 未以满功率运行。
- **兼容性**：当前 Proton 版本下，游戏兼容性或运行表现不佳。
- **资源限制**：Steam Headless 已接近其配置的 CPU 或内存上限。

## 解决方案

请按优先级依次检查以下项目。

### 避免 nginx 重启导致的中断

- 如果你当前使用的是 Olares 1.12.4，请升级到 1.12.5。
- 串流游戏时，避免安装、卸载或更新应用。

### 使用有线网络连接

检查 Olares 设备当前是否通过**有线网络**连接，而不是 Wi-Fi。

如果设备当前使用的是 Wi-Fi，请切换为有线网络再测试游戏串流效果。

### 将 GPU 设置为应用独占模式

将 GPU 设置为**应用独占**模式，并确保 Steam Headless 为独占应用。

1. 前往**设置** > **GPU**，在 **GPU 模式**下拉菜单中选择**应用独占**。
2. 在**选取独占应用**区域，检查当前独占应用是否为 Steam Headless。如果是，则继续下一项检查。

   ![检查 GPU 模式](/images/zh/manual/help/ts-steam-stream-gpu-mode.png#bordered)

如果当前选中的是其他应用：

1. 前往**设置** > **应用**，选择该应用，然后点击**暂停**。
2. 返回**设置** > **GPU**，然后点击 <i class="material-symbols-outlined">link_off</i> 解除绑定。
3. 再次前往**设置** > **应用**，选择 Steam Headless，然后点击**恢复**。
4. 返回**设置** > **GPU**，点击 <i class="material-symbols-outlined">sync</i> 刷新应用列表。如果系统仍未自动选中 Steam Headless，请点击**绑定应用**手动设置。

   ![设置应用独占模式](/images/zh/manual/help/ts-steam-stream-exclu.png#bordered)

### 切换至性能模式（仅适用于 Olares One）

如果你使用的是 Olares One，可以尝试切换到**性能模式**。

1. 前往**设置** > **我的 Olares** > **我的硬件**。
2. 点击**工作模式**旁边的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，然后选择**性能模式**。

   ![切换至性能模式](/images/zh/manual/help/ts-steam-stream-lag-power-mode.png#bordered)

### 尝试使用其他 Proton 版本

如果问题只出现在某一款游戏中，可以尝试在 Steam 中更换该游戏使用的 Proton 版本。

:::info 什么是 Proton？
Proton 是 Steam 的兼容层，用于在 Linux 系统上运行 Windows 游戏。不同的 Proton 版本可能会影响游戏兼容性和串流表现。
:::
1. 在 Steam 的**库**里打开目标游戏页面，然后点击 <i class="material-symbols-outlined">settings</i> > **属性...**。
2. 前往**兼容性**选项卡。
3. 勾选**强制使用特定 Steam Play 兼容性工具**。
4. 从下拉列表中选择一个 Proton 版本。
   :::tip
   可以访问 [ProtonDB](https://www.protondb.com/)，查看特定游戏的兼容性报告和推荐的 Proton 版本。
   :::
   ![Proton 版本](/images/zh/manual/help/ts-steam-stream-lag-proton.png#bordered)
   
5. 重新启动游戏，检查串流性能是否有所改善。

### 检查 CPU 和内存使用情况

#### 检查运行时资源占用

启动游戏，然后检查 Steam Headless 实际占用了多少 CPU 和内存。

1. 从启动台打开控制面板。
2. 在左侧边栏，点击**浏览**。
3. 在资源树中展开你的项目，然后展开**部署**。
4. 选择 Steam Headless 对应的部署。
5. 在详情面板的右上角，点击 <i class="material-symbols-outlined">more_vert</i>。
6. 点击**信息**，记录游戏运行过程中 CPU 和内存的最大使用值。
   - CPU 使用量单位为 `m`，其中 1000 m = 1 个 CPU 核心。
   - 内存使用量单位为 `Gi`。
   
   ![检查 Steam 资源使用](/images/zh/manual/help/ts-steam-stream-details.png#bordered)

#### 对比使用量与配置上限

检查运行时的资源占用是否接近配置的资源上限。

1. 关闭**信息**页，返回 Steam Headless 部署页面。
2. 在右侧详情面板中，点击 <i class="material-symbols-outlined">edit_square</i> 编辑 YAML 文件。
3. 找到 `limits` 下的 `cpu` 和 `memory`。
   
   示例：
   ```yaml
   limits:
      cpu: '18'
      memory: 64Gi
   ```
   ![检查 CPU 和内存上限](/images/zh/manual/help/ts-steam-stream-limit.png#bordered)
4. 对比配置上限与刚才记录的实际使用量。
5. 如果实际使用量持续接近当前配置上限，请根据设备容量，适当调高 `cpu` 或 `memory` 的值。
6. 点击 **Confirm** 保存更改，然后重新测试游戏串流效果。

## 如果问题仍然存在

如果完成以上排查步骤后，问题仍然存在，在 [Olares GitHub 仓库](https://github.com/beclab/Olares/issues) 提交 Issue，并提供以下信息：

- Steam Headless 当前配置的 `cpu` 和 `memory` 上限。
- 游戏运行时，Steam Headless 的 CPU 和内存使用截图。
- 游戏名称及使用的 Proton 版本。
- 问题的简短描述。
- 串流目标设备类型，例如笔记本电脑、掌机、手机或电视。

这些信息将帮助团队更快排查并定位问题。