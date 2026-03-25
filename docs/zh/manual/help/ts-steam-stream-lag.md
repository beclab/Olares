---
outline: deep
description: 排查并解决 Steam Headless 游戏串流卡顿或延迟的问题。
---
# Steam 串流卡顿或延迟

从 Olares 串流 Steam 游戏时出现卡顿或延迟，可参考本指南进行排查。

## 适用情况

串流游玩 Steam 游戏时出现卡顿或延迟，具体表现可能为高延迟、画面卡顿、突然断连或输入延迟。

## 原因

Steam 串流卡顿或延迟，可能由以下一个或多个原因导致：

- **nginx 重启**：nginx 重启会打断串流。
- **网络连接**：Olares 设备使用的是 Wi-Fi 而不是有线网络。
- **GPU 分配**：GPU 资源被其他应用占用。
- **GPU 功率**：GPU 性能因电源管理设置被降频。
- **兼容性**：当前 Proton 版本下，游戏运行表现不佳。
- **资源限制**：Steam Headless 已达到其配置的 CPU 或内存上限。

## 解决方案

按照优先级依次检查以下项目。

### 避免 nginx 重启导致的中断

- 如果你当前使用的是 Olares 1.12.4，请升级到 1.12.5 或更高版本。
- 串流游戏时，避免安装、卸载或更新应用。

### 使用有线网络连接

如果 Olares 设备当前通过 Wi-Fi 连接，请改用有线网络。

### 将 GPU 设置为应用独占模式

为 Steam Headless 分配完整的 GPU 访问权限，以获得最佳性能。

1. 前往**设置** > **GPU**，在 **GPU 模式**下拉菜单中选择**应用独占**。
2. 选择 Steam Headless 作为独占应用。如果当前已绑定其他应用，请先前往**设置** > **应用**停止该应用，然后返回绑定 Steam Headless。

   ![设置应用独占模式](/images/zh/manual/help/ts-steam-stream-exclu.png#bordered)

### 检查 GPU 功率状态

你设备上的 GPU 可能因电源管理设置而被降频。请确保 GPU 以全功率运行。

对于 Olares One 用户：

1. 前往**设置** > **我的 Olares** > **我的硬件**。
2. 点击**工作模式**旁边的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，然后选择**性能模式**。

   ![切换至性能模式](/images/zh/manual/help/ts-steam-stream-lag-power-mode.png#bordered)

### 尝试使用其他 Proton 版本

如果问题只出现在某一款游戏中，可以尝试在 Steam 中更换该游戏使用的 Proton 版本。Proton 是 Steam 用于在基于 Linux 的系统上运行 Windows 游戏的兼容层，不同版本可能会影响游戏性能。

1. 在 Steam 的**库**里打开目标游戏页面，然后点击 <i class="material-symbols-outlined">settings</i> > **属性…**。
2. 前往**兼容性**选项卡。
3. 启用**强制使用特定 Steam Play 兼容性工具**。
4. 从下拉列表中选择一个 Proton 版本。

   :::tip
   可以访问 [ProtonDB](https://www.protondb.com/)，查看特定游戏的兼容性报告和推荐的 Proton 版本。
   :::
   ![Proton 版本](/images/zh/manual/help/ts-steam-stream-lag-proton.png#bordered)
   
5. 重新启动游戏，检查串流性能是否有所改善。

### 检查并调整 CPU 和内存限制

Steam Headless 可能已达到配置的 CPU 或内存上限。先启动游戏，然后检查实际使用量是否接近上限。

1. 从启动台打开控制面板。在左侧边栏点击**浏览**。展开你的项目，然后展开**部署**，选择 Steam Headless 对应的部署。
2. 在详情面板右上角，点击 <i class="material-symbols-outlined">more_vert</i> > **信息**，记录游戏运行过程中 CPU 和内存的最大使用值。
   - CPU 使用量的单位为 `m`（毫核），其中 1000 m = 1 个 CPU 核心。
   - 内存使用量单位为 `Gi`。
 
   ![检查 Steam 资源使用](/images/zh/manual/help/ts-steam-stream-details.png#bordered)

3. 返回 Steam Headless 部署页面，点击 <i class="material-symbols-outlined">edit_square</i> 编辑 YAML 文件。找到 `limits` 下的 `cpu` 和 `memory`。
   
   示例：
   ```yaml
   limits:
      cpu: '18'
      memory: 64Gi
   ```
   ![检查 CPU 和内存上限](/images/zh/manual/help/ts-steam-stream-limit.png#bordered)

4. 如果实际使用量持续接近配置的上限，请根据设备容量适当调高 `cpu` 或 `memory` 的值。
5. 点击 **Confirm** 保存更改，然后重新测试游戏串流效果。

## 如果问题仍然存在

如果完成以上排查步骤后，问题仍然存在，在 [Olares GitHub 仓库](https://github.com/beclab/Olares/issues) 提交 Issue，并提供以下信息：

- Steam Headless 当前配置的 `cpu` 和 `memory` 上限。
- 游戏运行时，Steam Headless 的 CPU 和内存使用截图。
- 游戏名称及使用的 Proton 版本。
- 问题的简短描述。
- 串流目标设备类型，例如笔记本电脑、掌机、手机或电视。

这些信息将帮助团队更快定位问题原因。