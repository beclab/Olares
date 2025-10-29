---
description: 了解如何利用开发者页面管理仓库、查看系统镜像、导出系统日志以排查问题。
---

# 开发者资源

**开发者**页面位于 Olares **设置**项目的底部，专为开发者和高级用户设计，提供了用于管理核心系统资源和诊断问题的工具。包含以下功能：
* **仓库管理**
* **镜像管理**
* **日志导出**
* **系统环境变量管理**

## 仓库管理

**仓库管理** 页面维护了 Olares 下载系统镜像和其他软件包的核心来源仓库。通过此功能，你可以查看已有仓库、添加新仓库、管理现有端点，以优化 Olares 的软件包获取性能。

![仓库管理](/images/zh/manual/olares/repo-management.png#bordered)

在仓库列表页面，您可以查看到已添加的仓库名称、相关镜像数量以及镜像大小。

### 添加新仓库
添加新仓库步骤如下：
1. 从桌面进入**设置** > **开发者** > **仓库管理**。 
2. 点击右上角的 **+ 添加仓库**按钮。 
3. 在弹出的对话框中，填写以下信息：
   * **仓库名称**：输入仓库的唯一名称，例如 `docker.io` 或 `quay.io`。
   * **初始 Endpoint**：输入该仓库的初始 URL。 
4. 点击**确认**完成添加。

### 管理仓库端点

您可以通过添加或排序仓库的访问端点来优化特定仓库的访问速度和稳定性。

![端点管理](/images/zh/manual/olares/repo-endpoint-management.png#bordered)

1.  在**仓库管理**页面列表中，点击目标仓库右侧的 <i class="material-symbols-outlined">table_edit</i>按钮。
2. 进入**Endpoint 管理**页面，你可以：
   * **排序**：使用上下箭头对端点进行排序。Olares 将按照列表顺序优先使用排在前面的端点。
   * **删除**：点击 <i class="material-symbols-outlined">delete</i>图标以删除不再需要的端点。

## 镜像管理

镜像管理功能为您提供了一个全面的视图，用于查看 Olares 系统中所有已下载和缓存的应用程序和软件包镜像。

![镜像管理](/images/zh/manual/olares/image-management.png#bordered)

## 导出系统日志

日志记录了各系统组件的运行情况。在排查 Olares 问题时，系统日志可提供关键的诊断信息。要下载系统日志：

1. 在 Olares 桌面启动**设置** > **开发者** > **日志**。  
2. 点击**收集**生成日志文件。日志将自动保存至默认目录 `/Home/pod_logs`。 
3. 点击**打开**，在新窗口打开日志目录。  

   ![生成日志](/images/zh/manual/olares/export-log.png#bordered)

4. 右键选择生成的日志文件，点击**下载**将其下载到本地。  

   ![下载日志](/images/zh/manual/olares/download-log.png#bordered){width=70%}
下载后，可在 GitHub 反馈帖中附加日志文件，与 Olares 团队共享以加速问题定位。

## 设置系统环境变量

系统环境变量以`OLARES_SYSTEM_`开头，用于**集中管理** Olares 的核心功能配置，例如 SMTP、API Key 等通用设置。

你可以在此页面查看和修改这些变量，以优化网络访问和系统性能。

:::tip 说明

- 你无法在此处新增或删除系统环境变量，也无法修改它们的类型或其他属性。

- 如果某个环境变量的属性为`required: true`且尚未设置值，系统会在安装依赖该变量的应用时，提示你先完成配置。
::: 

### 自动配置

激活 Olares 时，系统会根据你的 Olares ID 自动设置部分环境变量，以确保连接到最适合的网络区域，从而提高下载与访问速度。

- 如果 Olares ID 以`.cn`结尾（如`user@olares.cn`），系统会将环境变量值设置为使用`.cn`的服务地址。
- 如果 Olares ID 以`.com`结尾（如`user@olares.com`），系统会将环境变量值设置为使用`.com`的服务地址。

系统会自动设置以下环境变量：

- `OLARES_SYSTEM_REMOTE_SERVICE`：用于调用组件接口。
- `OLARES_SYSTEM_CDN_SERVICE`：用于下载和内容分发。

这些变量会在系统激活时自动生效，确保 Olares 使用最快、最合适的网络服务。

### 修改变量

如果你需要手动调整系统环境变量，请按以下步骤操作：

1.  从 Olares 桌面进入**设置 > 开发者 > 系统环境变量**。

2.  在列表中找到你要修改的变量。

3.  点击<i class="material-symbols-outlined">edit_square</i>图标。置灰的变量不可修改。

4.  输入新值后，点击**确认**保存更改。
   ![设置系统环境变量](/images/zh/manual/olares/sys-env-var-cn.png#bordered)