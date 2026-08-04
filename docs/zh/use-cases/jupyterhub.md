---
outline: [2, 3]
description: 在 Olares 上设置 JupyterHub，提供多用户 Jupyter Notebook 环境，用于数据科学、研究和协作编程。
head:
  - - meta
    - name: keywords
      content: Olares, JupyterHub, Jupyter, notebook, data science, multi-user, self-hosted, Python
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-05-08"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/jupyterhub.md)。
:::

# 使用 JupyterHub 设置多用户 Notebook 环境

JupyterHub 是一个开源的多用户 Jupyter Notebook 服务器。它让你可以为多个用户提供计算环境，每个用户都有自己的工作空间，而无需他们在本地安装任何软件。在 Olares 上运行 JupyterHub 可为你提供一个自托管的 Notebook 平台，用于数据科学、研究或团队协作。

## 学习目标

在本指南中，你将学习如何：
- 安装 JupyterHub 并设置管理员账户。
- 启动 Notebook 服务器并在 JupyterLab 中编写代码。
- 添加和管理 JupyterHub 用户。
- 以管理员身份自定义 Notebook 配置文件。

## 安装 JupyterHub

1. 打开 Market 并搜索 "JupyterHub"。

   ![JupyterHub in Market](/images/manual/use-cases/jupyterhub.png#bordered){width=90%}

2. 点击 **获取**，然后点击 **安装**。

3. 出现提示时，设置以下环境变量：

   - **JUPYTERHUB_ADMIN_USERNAME**：输入默认管理员账户的用户名。安装完成后，使用相同的用户名注册并创建管理员密码。

4. 等待安装完成。

## 设置管理员账户

安装后，管理员用户名已被保留，但尚未创建密码。要激活管理员账户，请使用安装时输入的相同用户名注册。

1. 从 Launchpad 打开 JupyterHub。你将看到登录页面。

   ![JupyterHub 登录页面](/images/manual/use-cases/jupyterhub-signin.png#bordered){width=40%}

2. 点击 **注册** 进入账户创建页面。

3. 输入安装时指定的管理员用户名并设置密码，然后点击 **创建用户**。

   例如，如果你将 `JUPYTERHUB_ADMIN_USERNAME` 设置为 `olares`，则在此处输入 `olares` 作为用户名。

   ![JupyterHub 注册页面](/images/manual/use-cases/jupyterhub-signup.png#bordered){width=40%}

4. 返回登录页面，使用管理员用户名和密码登录。

## 启动和使用 Notebook 服务器

登录后，你将看到 JupyterHub 仪表板。从这里，你可以启动自己的 Notebook 服务器并访问管理页面。

![JupyterHub 仪表板](/images/manual/use-cases/jupyterhub-dashboard.png#bordered){width=90%}

### 启动 Notebook 服务器

1. 在仪表板中，点击 **启动我的服务器**。

2. 选择一个 Notebook 配置文件，然后点击 **启动**。

   默认选项是 **基础环境**。它提供一个最小化的纯 Python Jupyter 环境，适合大多数用户。

   ![选择 Notebook 配置文件](/images/manual/use-cases/jupyterhub-select-profile.png#bordered){width=90%}

   :::info
   当你首次使用选定的配置文件启动 Notebook 服务器时，JupyterHub 需要拉取相应的 Notebook 镜像。这可能需要几分钟，具体取决于镜像大小和网络连接。

   如果你想使用不同的 Notebook 镜像，请让管理员 [自定义 Notebook 配置文件](#可选：自定义-notebook-配置文件)。
   :::

3. 等待服务器启动。服务器启动后，Jupyter Notebook 界面将打开。

### 创建 Notebook

从 Jupyter Notebook 界面，你可以创建新 Notebook、打开终端、上传文件和处理现有文件。

要创建 Notebook：

1. 在 Jupyter Notebook 界面中，点击 **新建**，然后选择 **Notebook**。

   ![Jupyter Notebook 界面](/images/manual/use-cases/jupyterhub-notebook-interface.png#bordered){width=90%}

2. 选择一个内核，然后点击 **选择**。

   ![选择内核](/images/manual/use-cases/jupyterhub-new-notebook.png#bordered){width=90%}

一个新的 Notebook 将在当前工作区中打开。

### 打开 JupyterLab 并编写代码

你也可以在 JupyterLab 中工作，它提供了更丰富的编程环境，包括文件浏览器、多标签页、终端、Notebook 和扩展。

1. 在 Jupyter Notebook 界面中，点击顶部导航栏中的 **视图** > **打开 JupyterLab**。

2. 在 Launcher 中，点击 **Notebook** 下的 **Python 3 (ipykernel)**。

   ![JupyterLab launcher](/images/manual/use-cases/jupyterhub-lab.png#bordered){width=90%}

3. Notebook 打开后，在单元格中输入代码，然后点击 <i class="material-symbols-outlined">play_arrow</i> 或按 **Shift + Enter** 运行。

   ![在 JupyterLab 中编写代码](/images/manual/use-cases/jupyterhub-lab-notebook.png#bordered){width=90%}

### 返回 Hub

要从 JupyterLab 返回 JupyterHub 仪表板，点击 **文件** > **Hub 控制面板**。

![Hub 控制面板](/images/manual/use-cases/jupyterhub-control-panel.png#bordered){width=90%}

## 管理用户

作为管理员，你可以控制谁可以使用 JupyterHub。

推荐的工作流程是先在 **管理** 页面上创建用户名。然后用户可以使用该确切用户名注册并设置自己的密码。

### 添加新用户

1. 在 JupyterHub 仪表板中，前往 **管理** > **添加用户**。

2. 输入新用户的用户名并点击 **添加用户**。

   ![添加用户](/images/manual/use-cases/jupyterhub-add-user.png#bordered){width=90%}

3. 让用户打开 JupyterHub，点击 **注册**，并使用你添加的确切用户名创建密码。

   登录后，用户可以启动自己的 Notebook 服务器。

### 添加自注册用户

如果用户在管理员添加其用户名之前注册，他们可能会出现在隐藏的授权页面上。然而，在此页面上显示为已授权并不自动允许他们登录并启动 Notebook 服务器。

要允许用户使用 JupyterHub：

1. 在浏览器地址栏中，将 `/hub/authorize` 附加到你的 JupyterHub URL。

   ![授权页面](/images/manual/use-cases/jupyterhub-authorize.png#bordered){width=90%}

2. 检查授权页面上显示的用户名。

3. 返回 JupyterHub 仪表板并前往 **管理** > **添加用户**。

4. 添加授权页面上显示的相同用户名。

用户在 **管理** 页面上被添加后，就可以使用他们注册的账户登录。

![管理页面 - 创建用户](/images/manual/use-cases/jupyterhub-admin-create-user.png#bordered){width=90%}

## 可选：自定义 Notebook 配置文件

每个 Notebook 配置文件定义了用户启动 Notebook 服务器时使用的镜像和资源限制。大多数用户不需要更改这些设置。

作为管理员，你可以从 JupyterHub ConfigMap 自定义配置文件。

:::warning 不支持 GPU 加速
Olares 上的 JupyterHub 目前不支持 Notebook 服务器的 GPU 加速。仅使用基于 CPU 的 Notebook 镜像。不要使用 CUDA 启用的镜像标签，例如以 `cuda12-` 或 `cuda-` 为前缀的标签。
:::

默认配置文件及其资源限制如下：

| 配置文件 | CPU (保证 / 限制) | 内存 (保证 / 限制) |
|:--------|:------------------------|:---------------------------|
| 基础环境 | 0.1 / 1 | 1 GB / 1 GB |
| 最小环境 | 0.2 / 1 | 1 GB / 1 GB |
| 科学计算 | 0.5 / 1 | 1 GB / 2 GB |
| 数据科学 (Python + R) | 1 / 1 | 2 GB / 2 GB |
| 深度学习 (TensorFlow) | 2 / 4 | 2 GB / 4 GB |
| 大数据 (PySpark) | 2 / 4 | 4 GB / 8 GB |
| 全 Spark (完整) | 2 / 4 | 4 GB / 8 GB |
| R 环境 | 1 / 2 | 2 GB / 4 GB |

要自定义 Notebook 配置文件：

1. 在 Control Hub 中，前往 **浏览**，然后选择 JupyterHub 项目。

2. 在 **Configmaps** 下，选择 `jupyterhub-config`，然后点击详情面板右上角的 <i class="material-symbols-outlined">edit_square</i> 打开 YAML 编辑器。

   ![JupyterHub ConfigMap](/images/manual/use-cases/jupyterhub-configmap.png#bordered){width=90%}

3. 在 YAML 编辑器中，找到 `c.KubeSpawner.profile_list`。

4. 找到你要修改的配置文件，然后更新 `kubespawner_override` 下的值。

   - 要更改 Notebook 镜像，修改 `image`。该值必须是完整的镜像地址。
   - 要更改资源限制，修改 `cpu_guarantee`、`mem_guarantee`、`cpu_limit` 或 `mem_limit`。

   配置文件条目包含以下字段：

   ```python
   'kubespawner_override': {
       'image': 'docker.io/beclab/jupyter-base-notebook:notebook-7.0.6',
       'cpu_guarantee': 0.1,
       'mem_guarantee': '1G',
       'cpu_limit': 1,
       'mem_limit': '1G',
   }
   ```

5. 点击 **确认** 保存更改。

6. 返回 **部署** > **jupyterhub**，然后点击右侧面板中的 **重启**。

等待状态图标变为绿色。更新的配置文件设置将在用户为该配置文件启动新的 Notebook 服务器时生效。

## 常见问题

### 为什么我的 Notebook 服务器卡在 "Pending" 状态？

当集群没有足够的 CPU 或内存资源来启动服务器容器时，Notebook 服务器可能会保持在 **Pending** 状态。

作为管理员，你可以：

- 停止未使用的 Notebook 服务器以释放资源。
- 如果选定的 Notebook 配置文件需要比集群能提供的更多资源，请 [自定义 Notebook 配置文件](#可选：自定义-notebook-配置文件)，然后让用户重新启动服务器。

### 如何使用不同的 Notebook 镜像？

要使用不同的 Notebook 镜像，管理员需要更新 JupyterHub ConfigMap 中的 `image` 值。详细步骤请参阅 [自定义 Notebook 配置文件](#可选：自定义-notebook-配置文件)。

仅使用基于 CPU 的 Notebook 镜像，因为 Notebook 服务器目前不支持 GPU 加速。

### 可以使用 GPU 加速吗？

Olares 上的 JupyterHub 目前不支持 Notebook 服务器的 GPU 加速。默认的 Notebook 配置文件使用基于 CPU 的镜像。

不要使用 CUDA 启用的镜像标签，例如以 `cuda12-` 或 `cuda-` 为前缀的标签。有关更多信息，请参阅 Jupyter Docker Stacks 文档中的 [CUDA 启用变体](https://jupyter-docker-stacks.readthedocs.io/en/latest/using/selecting.html#cuda-enabled-variants)。

## 了解更多

- [JupyterHub 文档](https://jupyterhub.readthedocs.io)：官方 JupyterHub 文档和指南。
- [Zero to JupyterHub with Kubernetes](https://z2jh.jupyter.org)：在 Kubernetes 上运行 JupyterHub 的综合指南。
