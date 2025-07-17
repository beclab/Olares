---
description: Troubleshooting application issues status by examining the container staus or exporting logs 
---
# 检查容器状态

Pods 页面提供了 Olares 环境中所有 Pod 的全面视图，允许你在 Kubernetes 提供的最小粒度上进行管理。

本文档介绍如何查看容器状态，并导出容器日志。

## 查看容器状态

点击列表中的 Pod 可进入其详情页，你可以：

- 查看容器日志。
- 访问容器环境。
- 查看容器端口和环境变量。
- 在只读模式下查看 Pod 的 YAML 配置。

  :::tip 提示
  无法从该视图直接编辑 YAML 配置。YAML 配置由 Olares 通过应用负载模板和 Webhook 管理。
  :::

![Pod 详情](/images/how-to/olares/controlhub/pods/02.jpg#bordered)

## 导出容器日志以诊断故障

为了有效诊断和解决问题，你可能需要检查容器的详细日志：

![Log 操作](/images/manual/olares/controlhub-log.png)

1. 在浏览列中，导航到你的应用，然后依次进入**部署** > **容器**。
2. 找到状态异常的容器（带有橙色状态下标）。
3. 点击容器旁边的 <i class="material-symbols-outlined">article</i> 按钮。
4. 在弹出的日志窗口中，选择以下操作来管理日志：
   - 点击 <i class="material-symbols-outlined">download_2</i> 按钮下载完整的日志文件。
   - 点击 <i class="material-symbols-outlined">autorenew</i> 按钮刷新并查看最新的日志条目。
   - 点击 <i class="material-symbols-outlined">play_pause</i> 按钮开始或暂停日志的实时更新。






