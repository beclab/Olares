---
outline: deep
description: 在 Olares 上管理 ComfyUI 的管理员指南，涵盖服务控制、网络配置、模型和插件管理、Python 依赖以及故障排查。
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI Launcher, manage ComfyUI, ComfyUI models, ComfyUI plugins, self-hosted ComfyUI, ComfyUI on Olares
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/comfyui-launcher.md)。
:::

# 在 Olares 上管理 ComfyUI

ComfyUI Launcher 是 ComfyUI 管理员的主要管理工具。使用它来控制整个集群中的 ComfyUI 服务，并管理模型、插件、运行时环境和网络设置。

某些与模型相关的操作，例如从 ComfyUI 客户端的 **Errors** 选项卡下载缺失模型，是在 ComfyUI Launcher 之外完成的。本指南将引导你完成基于 Launcher 的任务和相关的客户端内工作流程。

## 学习目标

在本指南中，你将学习如何：

- 为集群中的所有用户启动和停止 ComfyUI 服务。
- 验证和调整网络设置，确保 GitHub、PyPI 和 Hugging Face 可访问。
- 以不同方式添加模型并了解模型在 Olares Files 中的存储位置。
- 从 ComfyUI Launcher 库管理插件或从 GitHub 安装它们。
- 通过安装、更新、移除包以及分析插件的缺失依赖来管理 Python 依赖。
- 通过检查依赖冲突、重置 ComfyUI、重新安装它以及收集诊断信息以供支持来排查常见问题。

## 启动和停止服务

作为管理员，你必须在其他人可以通过客户端界面访问 ComfyUI 之前启动 ComfyUI 服务。

- **启动 ComfyUI 服务**

    前往 **Home** 并点击右上角的 **START** 按钮。

    ![启动 ComfyUI 服务](/images/manual/use-cases/comfyui-start-service.png#bordered)   
    
    :::tip 首次运行注意事项
    - ComfyUI Launcher 的初始启动通常需要 10-20 秒进行环境初始化。
    - 如果系统提示缺少基本模型，你可以点击 **START ANYWAY**。但是，没有这些基础模型，工作流可能会失败。在启动服务之前下载必需的模型。
    :::

- **停止 ComfyUI 服务**

    当 ComfyUI 不使用时，前往 **Home** 并点击 **STOP** 按钮。这会释放 VRAM 和内存资源以供其他应用使用。

    ![停止 ComfyUI 服务](/images/manual/use-cases/comfyui-stop-service.png#bordered)   

## 配置网络

ComfyUI 需要访问 GitHub 获取插件、PyPI 获取 Python 包以及 Hugging Face 获取模型。在安装任何这些之前，请在 **Network Manager** 中检查连接状态。

1. 前往 **Network**。
2. 检查与 GitHub、PyPI 和 Hugging Face 的连接性。
3. 如果任何服务显示 `Inaccessible`，请修复你的网络或代理设置，然后点击 **SAVE & CHECK** 再次测试。
    
    ![配置网络](/images/manual/use-cases/comfyui-network-config.png#bordered)   

4. 重复操作，直到每个服务的状态变为 `Accessible`。

![网络重新检查](/images/manual/use-cases/comfyui-network-accessible.png#bordered){width=300}   

## 管理文件和目录

ComfyUI Launcher 提供对关键目录的快速访问，并使用共享的 `model` 文件夹，以便模型可以在 ComfyUI 和 SD Web UI 之间复用。

### 访问文件位置

在 **Home** 选项卡上，**File Type** 部分提供对 ComfyUI 使用的主要目录的直接访问：

| 条目 | Files 中的位置 | 描述 |
|--|--|--|
| **Root** | `/Files/Data/comfyuisharev3/comfyui/ComfyUI` | ComfyUI 安装的根目录。 |
| **Plugin** | `/Files/Data/comfyuisharev3/comfyui/ComfyUI/custom_nodes` | 存储已安装的插件（自定义节点和扩展）。 |
| **Model**  | `/Files/Common/comfyui/model` | 由 ComfyUI 和其他 AI 应用共享的 AI 模型的集中存储。 |
| **Output** | `/Files/Common/comfyui/output` | ComfyUI 生成的图像和其他资产的默认目标位置。 |
| **Input**  | `/Files/Common/comfyui/input` | 用于图生图或内绘工作流程的源图像文件夹。 |

点击任何这些条目都会直接在相应的目录中打开 Files 应用。

:::tip 适用于在 Olares 1.12.6 之前安装 ComfyUI 的用户
如果你在 Olares 1.12.5 或更早版本安装了 ComfyUI，文件位置不同：
`/Files/External/<your-hostname>/ai/comfyui/` 用于应用，`/Files/External/<your-hostname>/ai/model/` 用于模型，`/Files/External/<your-hostname>/ai/output/comfyui/` 用于输出。
迁移到新版本后，请使用上方的新路径。
:::

在 ComfyUI v3（通过 Olares 1.12.6 或更高版本安装）中，模型存储在 `/Common/comfyui/model` 中，并在 AI 应用之间共享。从早期版本迁移时，ComfyUI 会自动将模型从旧的 `External/<your-hostname>/ai/model/` 复制到新的 `/Common/comfyui/model/` 位置。

### 理解 `model` 目录结构

Olares 中的 ComfyUI 使用与标准安装不同的文件结构。此更改允许模型在 ComfyUI 和 SD Web UI 之间共享。

手动上传外部模型或在 **Custom Download** 中选择目标时，将文件放入正确的子文件夹。

:::tip
确保模型类型与目标文件夹匹配，以便 ComfyUI 能够正确检测和使用文件。
:::

| 标准 ComfyUI 目录 | Files 中的目录 | 模型类型 |
|--|--|--|
| `checkpoints/` | `checkpoints/` | 基础 checkpoints（SD 1.5, SDXL, Flux） |
| `loras/` | `lora/` | LoRA 权重 |
| `vae/` | `vae/` | VAE 模型 |
| `embeddings/` | `embeddings/` | Textual Inversion 嵌入 |
| `hypernetworks/` | `hypernetworks/` | Hypernetwork 权重 |
| `controlnet/` | `controlnet/` | ControlNet 模型 |
| `clip_vision/` | `clip_vision/` | CLIP Vision 模型（用于 IP-Adapter） |
| `style_models/` | `style_models/` | 风格或效果模型 |
| `upscale_models/` | `upscale_models/` | 放大器（ESRGAN, SwinIR） |
| `ipadapter/` | `ipadapter/` | IP-Adapter 模型 |
| `facerestore_models/` | `facerestore_models/` | 人脸修复模型（GFPGAN, CodeFormer） |
| `inpaint/` | `inpaint/` | 内绘专用模型 |
| `text_encoders/` | `clip/`, `text_encoders/` | 文本编码器模型（例如，CLIP 和 T5 编码器） |
| `diffusion_models/` | `unet/`, `diffusion_models/` | 扩散模型权重，包括基于 UNet 的模型 |

### 关于 `extra_model_paths.yaml`

ComfyUI 使用名为 `extra_model_paths.yaml` 的配置文件来定位模型、自定义节点以及默认目录外的其他资源。在 Olares 中，此文件是自动生成和配置的，指向共享模型目录，因此你通常不需要编辑它。

默认配置包括：

- **`base_path`**：指向 `/Common/comfyui/model` — 共享模型中心，模型存储在此目录并可在 AI 应用之间共享。

#### 何时需要自定义

如果工作流无法识别存储在 `/Common/comfyui/model/` 中的模型，或者你安装的自定义插件需要额外的模型搜索路径，你可能需要编辑此文件。

文件位置：
```
/Data/comfyuisharev3/comfyui/user/extra_model_paths.yaml
```

编辑方法：

1. 打开 Files 并导航到上述文件位置。
2. 在文本编辑器中打开 `extra_model_paths.yaml`。
3. 修改 `base_path` 使其指向你的模型目录。
4. 保存文件并重启 ComfyUI 以使更改生效。

:::warning
对 `extra_model_paths.yaml` 的错误更改可能导致 ComfyUI 无法启动。只有在你熟悉 YAML 语法并了解所添加的路径时，才编辑此文件。
:::

## 管理模型

ComfyUI 支持多种添加模型的方式。选择最适合模型来源和你工作流程的方法。

| 方法 | 最适合 | 说明 |
|--|--|--|
| **使用 ComfyUI Launcher 下载** | 公共模型、资源包或<br/> 不需要登录、访问批准或令牌的直接模型 URL。 | 将标准模型直接下载到 Olares 的最简单方式。 |
| **使用 Server Download** | ComfyUI 客户端中列出的<br/> 具有直接可下载 URL 的缺失模型。 | 直接下载到 Olares 主机。不支持需要登录、访问批准、令牌或其他授权的模型。 |
| **从库中使用** | 你已安装兼容替代方案的<br/> 缺失模型。 | 使用 `/Common/comfyui/model/` 中的模型。无需下载，但你需要自行验证兼容性。 |
| **上传本地模型** | 需要登录、<br/>访问批准、令牌或手动下载的受限模型，或来自不受支持<br/> 来源的模型。 | 先将文件下载到本地设备，然后通过 Files 或 LarePass 上传。 |
| **使用下载器节点** | 提供自己的<br/>内置模型下载器的自定义节点。 | 遵循节点文档。设置、存储位置和要求可能因节点而异。 |

### 使用 ComfyUI Launcher 下载

对不需要登录或令牌的公共 Hugging Face 模型使用此方法。

ComfyUI Launcher 提供三种将模型直接下载到 Olares 的常见方式。

#### 下载资源包

当你想安装 Olares 提供的即用型包（如入门包或工作流专用包）时，使用此方法。

1. 前往 **Home** > **Resource Package**。
2. 找到你想要的包，然后点击 **VIEW**。
3. 在 **Package Details** 页面上，点击 **GET ALL** 下载所有必需的文件。你可以在状态栏中跟踪进度。
    
    ![下载进度](/images/manual/use-cases/comfyui-download-progress1.png#bordered)

#### 从模型库下载

当你想在 ComfyUI Launcher 中直接下载单个公共模型时，使用此方法。

1. 前往 **Models** > **Model library**，然后滚动到 **Available models**。
2. 按名称搜索模型，或按类别浏览。
3. 找到你想要的模型，然后点击 <i class="material-symbols-outlined">download</i> 按钮下载它。
   
   ![库下载](/images/manual/use-cases/comfyui-model-built-in.png#bordered)

#### 通过直接 URL 下载

当你已有 ComfyUI Launcher 可以直接访问的模型文件的直接下载 URL 时，使用此方法。

1. 前往 **Models** > **Custom Download**。
2. 粘贴模型 URL。
3. 根据模型类型选择目标路径。如果你不确定选择哪个文件夹，请参阅[理解 `model` 目录结构](#理解-model-目录结构)。
4. 点击 **DOWNLOAD MODEL**。

   ![自定义下载](/images/manual/use-cases/comfyui-model-link.png#bordered){width=90%}

### 使用 Server Download

使用此方法从 ComfyUI 客户端将缺失模型直接下载到 Olares 主机。

:::info
Server Download 是 ComfyUI 1.0.32 及更高版本中默认启用的 ComfyUI 插件。

它仅支持无需登录、访问批准或其他授权即可直接访问的模型 URL。对于受限模型，请[手动上传](#上传本地模型)。

要关闭此功能，请参阅[禁用 Server Download](#禁用-server-download)。
:::

1. 在 ComfyUI 客户端中，点击画布上的空白区域以打开 **Workflow Overview**。然后，选择 **Errors** 选项卡并找到 **Missing Models** 部分。

    ![Errors 选项卡](/images/manual/use-cases/comfyui-errors-tab.png#bordered)

2. 在列表中找到必需的模型，点击 **Copy URL**，然后点击 **Server Download**。

    Server Download 从你的剪贴板读取 URL。在这两个操作之间不要复制其他文本。

   ![复制模型 URL](/images/manual/use-cases/comfyui-server-download-copy-url.png#bordered)

3. 如果这是你第一次使用 Server Download，浏览器会请求读取剪贴板的权限。点击 **Allow**。

4. 在下载对话框中，验证自动填充的信息：

   - **URL**: 如果 URL 不正确，请清除字段并手动粘贴。
   - **File Name**: 点击 **Auto** 检测并填充文件名。
   - **Model Type**: 确保它与 **Missing Models** 中显示的类别匹配，例如 `checkpoints`、`loras` 或 `diffusion_models`。

   :::warning 验证模型类型
   自动检测的 **Model Type** 可能不正确。如果工作流包含画布上的注释说明模型存储位置，请将其作为额外参考。保存到错误文件夹的模型仍会被报告为缺失。
   :::

   ![Server Download 对话框](/images/manual/use-cases/comfyui-download-model-to-server.png#bordered)

5. 开始下载并等待完成。

   Server Download 目前不支持暂停、恢复或重试下载。如果下载失败，请开始新的下载任务。

6. 下载完成后，刷新 ComfyUI 页面。你现在可以在工作流中使用该模型。

要验证或管理下载的文件，请打开 Files 并前往 `/Common/comfyui/model/` 下的相应子文件夹。

### 从库中使用现有模型

当工作流报告模型缺失，但你在模型库中已有兼容的替代方案时，使用此方法。

**从库中使用** 允许你从 `/Common/comfyui/model/` 中选择模型。仅当相应的模型类型文件夹存在且包含模型文件时，该按钮才可用。

1. 在 ComfyUI 客户端中，点击画布中的空白区域以打开 **Workflow Overview**，然后前往 **Errors** 选项卡并找到 **Missing Models** 部分。
2. 在缺失模型下，点击 **Use from Library** 下拉菜单。
3. 从本地存储中选择一个现有模型来替换缺失的模型。

    ![从库中使用](/images/manual/use-cases/comfyui-use-from-lib.png#bordered)

:::warning 评估兼容性
你必须自行评估模型兼容性。选择不兼容的模型将导致工作流失败或产生意想不到的结果。
:::

### 使用下载器节点

某些自定义节点可以自动下载模型。

遵循节点作者提供的文档进行设置、存储位置和模型要求。

如果下载后工作流仍报告模型缺失，请检查文件是否保存到预期位置以及它是否与工作流期望的模型类型匹配。

### 上传本地模型

当模型需要登录、令牌、批准或手动下载，或模型来源不受 ComfyUI Launcher 或 Server Download 支持时，使用此方法。

1. 将模型文件下载到你的本地设备。如果需要，请参阅[模型无法直接下载到 Olares](./comfyui-common-issues#模型无法直接下载到-olares)。
2. 从 Launchpad 打开 Files。
3. 导航到 `/Common/comfyui/model/`。
4. 打开与模型类型匹配的文件夹。如果你不确定使用哪个文件夹，请参阅[理解 `model` 目录结构](#理解-model-目录结构)。
5. 将模型文件上传到目标文件夹。

### 删除模型

要删除模型：

1. 前往 **Models** > **Model library**。
2. 在 **Installed models** 部分下，找到你要删除的模型，然后点击删除图标。

    ![删除模型](/images/manual/use-cases/comfyui-delete-model.png#bordered)

## 管理插件

ComfyUI Launcher 在 **Plugins** 中提供灵活的方式来管理插件。

![插件状态](/images/manual/use-cases/comfyui-plugin-status.png#bordered){width=90%}

### 管理可用插件

要管理在 ComfyUI Launcher 中注册的可用插件：

1. 前往 **Plugins** > **Plugin library**。
2. 在 **Available Plugins** 下，找到你想要的插件。

对于任何插件，你可以：
- 点击 <i class="material-symbols-outlined">visibility</i> 按钮查看插件详情。
- 如果可用，访问 GitHub 仓库。

根据插件状态，有不同的操作可用：

- **Not installed**: 点击 **Install** 获取最新版本，或 **Switch version** 选择特定版本。
- **Installed**: **Disable**、**Uninstall** 或 **Switch version**。
- **Disabled**: 点击 **Enable** 重新激活它，或 **Uninstall** 移除它。

在部分顶部，你还可以：
   - 点击 **UPDATE ALL PLUGINS** 更新所有已安装的插件。
   - 点击 **REFRESH** 刷新插件列表。

### 从 GitHub 安装插件

要直接从 GitHub 仓库安装插件：

1. 前往 **Plugins** > **Custom Install**。
2. 输入插件的 GitHub 仓库 URL。
3. （可选）指定分支。如果不确定，保持默认。
4. 点击 **INSTALL PLUGIN**。
   
   ![下载插件](/images/manual/use-cases/comfyui-plugin-install.png#bordered)

### 禁用 Server Download

ComfyUI 客户端中的 **Server Download** 按钮由 `comfyui-server-downloader` 插件提供，该插件在 ComfyUI 1.0.32 及更高版本中默认启用。

要禁用它：

1. 在 ComfyUI Launcher 中，前往 **Plugins** > **Plugin library**。
2. 在 **Available Plugins** 下，搜索 `server-download`。
3. 在 `comfyui-server-downloader` 行中，点击 **Actions** 列中的下拉箭头并选择 **Disable**。

   ![禁用 Server Download](/images/manual/use-cases/comfyui-disable-server-download.png#bordered)

4. 前往 **Home** 选项卡并点击 **RESTART** 应用更改。

服务重启后，**Server Download** 按钮将不再出现在 ComfyUI 客户端中。

## 管理环境

ComfyUI 运行在一组 Python 库上。在 **Environment** 页面上管理它们。

### 安装 Python 包

1. 前往 **Environment** > **INSTALL NEW PACKAGE**。
2. 在弹出窗口中，输入包名和版本号（可选），然后点击 **INSTALL**。
    
    ![安装新包](/images/manual/use-cases/comfyui-python-install.png#bordered){width=90%}

### 管理已安装的 Python 包

1. 前往 **Environment**。
2. 在 **Installed Python packages** 选项卡下，找到你要管理的 Python 库。
3. 点击右侧的 <i class="material-symbols-outlined">arrow_upward</i> 按钮更新库，或点击删除按钮移除它。
    
    ![管理已安装的包](/images/manual/use-cases/comfyui-python-manage.png#bordered){width=90%}

### 分析依赖安装状态

1. 前往 **Environment** > **Dependency analysis**。
2. 点击 **ANALYZE NOW** 开始分析。
3. 从左侧的插件列表中，找到以红色高亮显示的问题插件，然后点击它。
4. 从 **Dependency list** 中，找到插件的缺失库，然后点击右侧的 **Install** 按钮。你也可以点击 **FIX ALL** 自动安装所有缺失的库。
    
    ![分析依赖](/images/manual/use-cases/comfyui-dependency-analy.png#bordered)

<!--## 配置 ComfyUI 启动选项

**Lab** 页面允许你自定义 ComfyUI 的启动方式。你可以查看当前启动命令并使用两种方法修改启动参数：手动编辑参数或从预设列表中选择。
    ![Lab](/images/manual/use-cases/comfyui-lab.png#bordered){width=90%}

### 查看启动命令

**Current full launch command** 部分显示用于启动 ComfyUI 的完整命令，包括所有启用的选项。点击 <i class="material-symbols-outlined">content_copy</i> 复制它以供分享或记录。

### 编辑启动选项

你可以使用以下任一方法修改启动参数。

#### 方法 1：手动编辑额外参数

当你需要添加预设列表中不可用的自定义参数时，使用此方法。

1. 在 **Manually edit extra arguments** 中，在文本字段中输入你的额外参数（以空格分隔）。
2. 点击 **SAVE MANUAL ARGS**。
3. 前往 **Home** 并点击 **RESTART** 应用更改。

#### 方法 2：使用启动选项列表

使用此方法从预定义列表中选择参数并设置其值。

:::info
某些选项（如 `--port` 和 `--front-end-version`）由系统固定，无法修改。
:::

要使用列表配置：

1. 在 **Launch options list** 中，找到你要配置的选项。
2. 选中 **Enable** 复选框以将该选项添加到启动命令中。
3. 如果选项需要值，在 **Value** 字段中输入它。
4. 点击 **SAVE LIST ARGS** 保存你的更改。
5. 前往 **Home** 并点击 **RESTART** 应用更改。

### 恢复默认值

如果错误的启动参数导致问题，请将所有选项恢复为默认值。

1. 点击 **RESTORE DEFAULT**。
2. 前往 **Home** 并点击 **RESTART** 应用更改。
-->

## 故障排查流程

当你遇到问题并需要一般恢复路径时，使用以下流程。

:::tip
针对特定症状的解决方案，如 ComfyUI 无法启动、模型无法下载或高 CPU 温度，请先参阅[常见问题](./comfyui-common-issues.md)。
:::

### 检查依赖冲突

如果问题在安装新插件后开始，可能是由依赖冲突引起的。

运行依赖分析以识别和修复问题。有关详细步骤，请参阅[分析依赖安装状态](#分析依赖安装状态)。

### 重置 ComfyUI 配置

如果上述检查后问题仍未解决，请将 ComfyUI 重置为其初始状态。

:::warning 谨慎操作
重置 ComfyUI 是不可逆的。所有插件、自定义配置和 Python 依赖都将被移除。存储在共享 `model` 文件夹中的模型不受影响。
:::
:::tip 获取诊断详情
如果你计划联系支持，请在重置前导出你的 ComfyUI 日志。请参阅[收集支持信息](#收集支持信息)。
:::

要重置 ComfyUI：

1. 在 ComfyUI Launcher 中，前往 **Home** 并点击右上角的 <i class="material-symbols-outlined">more_vert</i>，然后点击 **Wipe and restore**。
2. 在提示窗口中，点击 **WIPE AND RESTORE**。
    
    ![清除并恢复](/images/manual/use-cases/comfyui-wipe-and-restore.png#bordered){width=50%}

3. 输入 `CONFIRM`，然后点击 **CONFIRM**。
    
    ![二次确认](/images/manual/use-cases/comfyui-second-confirm.png#bordered){width=50%}

重置完成后，重启 ComfyUI 以使更改生效。

### 完全重新安装 ComfyUI

如果清除并恢复后问题仍然存在，请完全卸载并重新安装 ComfyUI。

1. 前往 **Market** > **My Olares**。
2. 点击 ComfyUI 操作按钮旁边的下拉箭头并选择 **Uninstall**。
3. 在 **Uninstall** 窗口中，选择 **Also remove all local data**，然后点击 **Confirm**。
4. 从 Launchpad 打开 Files 并前往 `/Data/comfyuisharev3/comfyui/`。
5. 删除 `comfyui` 文件夹。
6. 从 Market 重新安装 ComfyUI。
7. 安装完成后，打开 ComfyUI Launcher 并启动服务。

### 收集支持信息

如果你无法解决问题并需要将其上报给支持团队，请准备以下诊断信息。

#### 导出 ComfyUI 日志

日志包含后端运行状态和错误跟踪。

1. 在 ComfyUI Launcher 中，前往 **Home** 并点击右上角的 <i class="material-symbols-outlined">more_vert</i>，然后点击 **View logs**。
   
   ![查看日志](/images/manual/use-cases/comfyui-view-logs1.png#bordered){width=90%}

2. 点击 <i class="material-symbols-outlined">refresh</i> 按钮以确保你有最新的输出。
3. 点击 <i class="material-symbols-outlined">download</i> 按钮保存日志文件。
   
   ![导出日志](/images/manual/use-cases/comfyui-export-logs.png#bordered){width=90%}

#### 可选：获取工作流错误报告

如果特定工作流失败，请包含工作流错误报告的截图，以帮助支持团队识别问题。

1. 在 ComfyUI 客户端中，点击 **Active** 打开 **Job Queue**。
2. 从列表中选择失败的任务。
3. 点击 **Report error**，然后点击 **Show Report** 展开详情。

    ![工作流错误报告](/images/manual/use-cases/comfyui-workflow-error.png#bordered){width=80%}
