---
outline: [2, 3]
description: 了解如何在 Olares 控制面板中管理容器和排查应用问题。本文档涵盖如何查看容器组详情、检查容器状态以及导出容器日志以进行故障诊断。
---
# 管理容器

当应用出现 CrashLoopBack、启动缓慢或资源飙升时，容器级别的排查最为直接。控制面板的容器组详情页面聚合特定容器组下所有容器，配合事件分析、可视化图表以及日志，帮你快速定位、解决问题。

## 查看容器组详情

1. 左侧导航点击**浏览**，并在第一列**命名空间**树中根据应用或服务选择目标命名空间。
2. 在第二栏**部署 / 有状态副本集 / 守护进程集**下逐级点击目标**工作负载** > **容器组**。
3. 在第三栏详情页面查看容器组详情信息：

| 区块   | 内容                                    |
|--------|---------------------------------------|
| **信息** | Pod 状态、重启次数、IP、节点、QoS、镜像等基础数据         |
| **容器** | 各容器的实时 CPU / 内存用量曲线与镜像版本；提供日志导出和命令行操作 |
| **卷**   | 挂载的持久卷、路径、访问模式                        |
| **环境变量** | 注入到容器的全部变量，支持展开查看                     |
| **事件** | 最近一小时容器组调度 / 探针 / 网络等事件日志             |

![containers](/images/how-to/olares/controlhub/browse/04.jpg#bordered)

:::tip 提示
无法从该视图直接编辑容器组的 YAML 配置。YAML 配置由 Olares 通过应用负载模板和 Webhook 管理。
:::

## 查看容器状态

对于容器列表里的单个容器，你可以：

- 点击条目以查看该容器详细信息，包含状态、镜像信息、镜像拉取策略、端口、容器等。
- 查看容器端口和环境变量。
- 访问容器命令行环境。
- 实时查看或导出容器日志。

![容器详情](/images/how-to/olares/controlhub/pods/02.jpg#bordered)

## 导出问题容器日志

使用导出的容器日志结合事件与容器组资源图表，可帮你快速定位 CrashLoopBack、探针失败或 OOM 等错误。

![Log operation](/images/manual/olares/controlhub-export-log.png)

1. 在浏览列中，点击问题应用对应的命名空间。
2. 在第二栏里，展开部署列表，并依次点击**目标部署** > **容器组**。
3. 在容器组详情页面，找到状态异常的容器（带有橙色状态下标），并点击容器旁边的 <i class="material-symbols-outlined">article</i> 按钮。
4. 在弹出的日志窗口中，选择以下操作来管理日志：
   ![Log 操作](/images/manual/olares/controlhub-log.png)
   - 点击 <i class="material-symbols-outlined">download_2</i> 按钮下载完整的日志文件。
   - 点击 <i class="material-symbols-outlined">autorenew</i> 按钮刷新并查看最新的日志条目。
   - 点击 <i class="material-symbols-outlined">play_pause</i> 按钮开始或暂停日志的实时更新。






