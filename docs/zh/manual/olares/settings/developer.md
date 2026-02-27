---
outline: [2, 3]
description: 了解如何管理仓库、查看已下载的镜像、导出系统日志以及配置系统级环境变量
head:
  - - meta
    - name: keywords
      content: Olares OS, 高级设置, 仓库管理, 镜像管理, 系统日志, 系统级环境变量管理
---

# 高级设置

**设置**中的**高级**页面专为开发者和高级用户设计，用于管理核心系统资源和诊断问题。

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

## 导出系统日志

系统日志记录了各个系统组件的运行状态，为故障排除提供关键的诊断信息。

导出系统日志的步骤如下：
1. 前往**设置** > **高级** > **日志**。
2. 点击**收集**生成日志文件。日志会自动保存到 `/Home/pod_logs` 目录。
3. 点击**打开**在新窗口中打开日志目录。
4. 右键单击生成的日志文件，选择**下载**将其保存到本地。

   ![生成日志](/images/zh/manual/olares/export-log1.png#bordered){width=65%}

    :::tip
    你可以将此日志文件附加到 GitHub Issue 中，以帮助 Olares 团队更快地定位和解决你的问题。
    :::

## 设置系统级环境变量

从 Olares 1.12.2 版本开始，你可以为应用配置系统级环境变量。这样就能全局应用通用设置，而无需为每个应用单独配置。

变量主要分为两大类：

| 类别 | 描述 | 权限 |
|:----|:----|:----|
| 系统配置 | 核心系统运行所需的预定义变量。<br><br>例如：`OLARES_SYSTEM_CDN_SERVICE` | 不能删除。为保障稳定性，系统会锁定部分变量并将其置灰，但你可以编辑其他变量的值。 |
| 用户信息 | 自定义或预定义的用户级变量，例如第三方 API 密钥。<br><br>例如：`OLARES_USER_CUSTOM_OPENAI_APIKEY` | 可以添加、编辑和删除。系统会自动为你创建的任何自定义键添加 `OLARES_USER_` 前缀。 |

管理环境变量的步骤如下：

1. 前往**设置** > **高级** > **系统环境变量**。

    ![系统级环境变量](/images/zh/manual/olares/sys-env-var1.png#bordered){width=65%}

2. 添加新的自定义变量：

    a. 点击**添加环境变量**。

    b. 在**键**字段中，输入你的自定义键名，该名称将附加到 `OLARES_USER_` 前缀之后。

    c. 在**值**字段中，填入数值。
    
    d. 从**类型**列表中，选择数据类型。
    
    e. 可选：提供描述。
    
    f. 点击**保存**。

    g. 点击**系统环境变量**底部右侧的**应用**。

3. 修改变量：

    a. 在列表中找到目标变量，然后点击 <i class="material-symbols-outlined">edit_square</i>。没有此图标的变量已被锁定，无法更改。

    b. 在**编辑环境变量**窗口中，更新变量的值。

    c. 点击**确认**保存更改。

    d. 点击**系统环境变量**底部右侧的**应用**。

4. 删除变量：

    a. 找到目标用户级变量，点击 <i class="material-symbols-outlined">delete</i>。

    b. 点击**确认**。

    c. 点击**系统环境变量**底部右侧的**应用**。
