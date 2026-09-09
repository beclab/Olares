---
outline: [2, 3]
description: 了解如何在设置的高级页面管理源仓库并查看已下载的镜像。
head:
  - - meta
    - name: keywords
      content: Olares OS, 高级设置, 仓库管理, 镜像管理
---

# 高级设置

**设置**中的**高级**页面专为开发者和高级用户设计，用于管理核心系统资源。导出系统日志排查故障，请参阅[获取技术支持](../../help/request-technical-support.md)。配置系统级环境变量，请参阅[设置系统级环境变量](../../manage-system-env.md)。

## 管理仓库

**仓库管理**页面允许你查看 Olares 用于下载系统镜像和软件包的源仓库。你还可以配置镜像端点以优化下载速度和稳定性。

### 查看仓库

1. 前往**设置** > **高级** > **仓库管理**。
2. 在仓库列表中，你可以查看每个仓库的名称、相关镜像数量和镜像大小。  

   ![仓库管理](/images/zh/manual/olares/repo-management1.png#bordered){width=65%}

### 管理仓库镜像

管理仓库的镜像端点以提高访问速度和稳定性。

1. 在**仓库管理**页面，找到目标仓库，然后在**操作**列点击 <i class="material-symbols-outlined">chevron_forward</i>。
2. 在**镜像站管理**页面，你可以执行以下操作：

    ![镜像站管理](/images/zh/manual/olares/mirror-management1.png#bordered){width=65%}
    
    - 要重新排序镜像端点，点击 <i class="material-symbols-outlined">keyboard_control_key</i> 或 <i class="material-symbols-outlined">keyboard_arrow_down</i>。Olares 会优先使用列表中排名靠前的端点。
    - 要删除不再需要的端点，点击 <i class="material-symbols-outlined">delete</i>。
    - 要添加新的镜像端点，点击**添加镜像站**，输入镜像站URL，然后点击**确认**。

## 管理镜像

**镜像管理**页面提供了 Olares 系统上下载和缓存的所有应用及软件包镜像的全面视图。你可以通过筛选或搜索快速找到特定镜像。

![镜像管理](/images/zh/manual/olares/image-management1.png#bordered){width=65%}

