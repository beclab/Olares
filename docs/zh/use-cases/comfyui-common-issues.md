---
outline: [2, 3]
description: Olares 上 ComfyUI 的常见问题及解决方案，包括启动问题、启动器日志消息、模型下载限制、工作流错误以及 Olares One 上的高 CPU 温度。
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, troubleshooting, common issues, self-hosted
---
# ComfyUI 常见问题

使用本页面识别和解决 Olares 上 ComfyUI 的常见问题。

:::tip 需要更多帮助？
如果你遇到的问题未在此处列出，请参阅[故障排查流程](./comfyui-launcher#故障排查流程)。
:::

## 升级到 Olares 1.12.6 后如何迁移到新版 ComfyUI

如果你已升级到 Olares 1.12.6，且之前安装过 ComfyUI Shared，可参考本节完成迁移。如果你是在 Olares 1.12.6 或更高版本上首次安装 ComfyUI，直接从应用市场安装 ComfyUI 即可。

Olares 1.12.6 更新了共享应用架构。旧版 ComfyUI Shared 在升级后仍可继续运行，但无法接收后续更新。要继续获取更新，需在保留本地数据的情况下卸载旧版应用，然后从应用市场安装新版 ComfyUI。

:::warning
卸载旧版应用时，不要勾选**同时删除所有本地数据**。如果勾选此项，你的模型、插件、工作流以及输入/输出文件会被删除。
:::


### 迁移步骤

1. 打开应用市场，进入**我的 Olares**。
2. 找到 ComfyUI Shared，点击操作按钮旁的下拉箭头，然后选择**卸载**。
3. 在卸载窗口中，确保未勾选**同时删除所有本地数据**，然后点击**确认**。
4. 返回应用市场，搜索 “ComfyUI”，然后点击**安装**。
5. 在应用详情页，查看**信息**下的**兼容性**。如果显示 `Olares >=1.12.6-0`，说明这是新版 ComfyUI。
6. 安装完成后，打开 ComfyUI，确认模型、插件、工作流以及输入/输出文件是否可正常使用。

### 迁移内容说明

新版 ComfyUI 安装后，将自动完成数据的迁移工作，具体规则如下：

| 数据类型 | 原路径 | 新路径 |
|:---|:---|:---|
| ComfyUI 核心数据（插件、工作流等） | `External/<your_hostname>/ai/comfyui/` | `Data/comfyuisharev3/comfyui/` |
| 模型数据 | `External/<your_hostname>/`<br /> `ai/model/` | `Common/comfyui/model/` |
| 输出文件 | `External/<your_hostname>/`<br /> `ai/output/comfyui/` | `Common/comfyui/output/` |
| 输入文件 | `External/<your_hostname>/`<br /> `ai/comfyui/ComfyUI/input/` | `Common/comfyui/input/` |

:::warning
迁移完成后，需将新的模型和输入文件上传到 `Common/comfyui/` 下的新路径。新版 ComfyUI 不再将 `External/<your_hostname>/ai/` 作为当前使用的文件位置。
:::

数据迁移会在每次 ComfyUI 重启时运行。如果你之后又将文件添加到旧路径，ComfyUI 会在下次重启时将这些文件移动到新路径，并删除 `External/<your_hostname>/ai/` 下的原文件。为避免混淆，需直接使用新路径上传文件。

:::info 迁移后的模型路径
新版 ComfyUI 会自动生成 `extra_model_paths.yaml`，并将标准模型类别映射到共享模型目录。迁移后通常无需手动配置此文件。

自动生成的映射可能不包含自定义类别或特定节点使用的模型类别。如果迁移后的模型已位于 `/Files/Common/comfyui/model/`，但 ComfyUI 或自定义节点仍无法找到该模型，可参阅[配置其他模型路径](./comfyui-launcher#配置其他模型路径)。
:::

## ComfyUI 无法启动

ComfyUI 无法启动、意外停止或行为异常。

这通常由资源不足或 GPU 分配不正确引起。要解决此问题：

1. 检查你的系统资源。如果你的 CPU 或内存使用率已满，请停止其他资源密集型应用。
2. 如果系统资源看起来正常，前往**设置** > **AI 算力**检查你的 GPU 模式：
   - 如果你使用的是**容量分片**，需确保 ComfyUI 已绑定到 GPU 并有足够的显存分配。
   - 如果你使用的是**独占分配**，需确保独占应用设置为 ComfyUI。
3. 等待片刻，然后再次尝试启动 ComfyUI。

## 启动器日志显示错误

启动器日志中的 `Error` 消息不一定表示系统故障。在启动和插件扫描期间，ComfyUI 通常会记录关于缺失的可选依赖或环境检查的非致命错误，即使在正常运行时也是如此。

如果 ComfyUI 成功启动，大多数这些消息不需要采取行动。仅在 ComfyUI 无法启动、工作流无法运行或插件停止工作时才调查日志。

## ComfyUI 无法找到共享模型目录中的模型

迁移到 ComfyUI v3，或在 Olares 1.12.6 及更高版本中直接安装新版 ComfyUI 后，工作流可能会报告模型缺失，即使模型文件已存储在 `/Files/Common/comfyui/model/` 下。

此问题通常由以下两种原因之一导致：

- `extra_model_paths.yaml` 中未映射对应的模型类别，因此 ComfyUI 不会从该文件夹加载模型。
- ComfyUI 已检测到模型，但自定义节点查找的是另一个模型类别。

以下示例使用共享模型目录中已有的两个模型：

```text
/Files/Common/comfyui/model/
├── detection/
│   └── mediapipe_face_fp32.safetensors
└── ultralytics/
    └── bbox/
        └── face_yolov8m.pt

```

这两个模型分别用于说明以下问题：

- `mediapipe_face_fp32.safetensors` 未显示在 **Model Library** 中，因为未映射 `detection` 类别。
- `face_yolov8m.pt` 显示在 **Model Library** 中，但 `UltralyticsDetectorProvider` 无法找到它，因为该节点查找的是 `ultralytics_bbox` 类别。

### 模型未显示在 Model Library 中

在此示例中，虽然 `mediapipe_face_fp32.safetensors` 已存储在对应文件夹中，但 **Model Library** 的 `detection` 类别下没有显示任何模型。

![detection 类别中未检测到模型](/images/manual/use-cases/comfyui-common-shared-model-not-detected.png#bordered)

要解决此问题：

1. 在 `extra_model_paths.yaml` 的 `olares_shared_models` 下添加以下映射：

   ```yaml
   detection: detection
   ```

   ![添加 detection 映射](/images/manual/use-cases/comfyui-extra-model-paths-add-detection.png#bordered)

   有关编辑配置文件的说明，请参阅[配置其他模型路径](./comfyui-launcher#配置其他模型路径)。

2. 保存配置并重启 ComfyUI。
3. 在 ComfyUI 启动日志中查找以下记录，确认映射已加载：

   ```text
   Adding extra search path detection /mnt/olares-shared-model/detection
   ```

   ![启动日志确认已添加 detection 路径](/images/manual/use-cases/comfyui-detection-path-added-log.png#bordered)

4. 刷新 ComfyUI 页面并再次搜索该模型。模型应显示在 `detection` 类别下，并可供工作流节点使用。

   ![重启后成功识别 detection 模型](/images/manual/use-cases/comfyui-model-recognized-after-restart.png#bordered)

:::info 如果模型仍未显示
- 如果日志中没有上述记录，检查类别名称、相对文件夹路径、YAML 缩进以及 `extra_model_paths.yaml` 的位置。
- 如果日志中有上述记录，检查模型文件的位置、文件名和文件格式。
:::

### 模型已显示，但自定义节点无法找到

在此示例中，ComfyUI 已在 `ultralytics/bbox` 类别下检测到 `face_yolov8m.pt`，但 `ImpactPack/UltralyticsDetectorProvider` 节点查找的是 `ultralytics_bbox` 类别。

![模型已被检测到，但自定义节点无法使用](/images/manual/use-cases/comfyui-common-model-detected.png#bordered)

要解决此问题：

1. 查看自定义节点的文档或错误消息，确定节点所需的模型类别。
2. 将该类别与模型当前在 **Model Library** 中所属的类别进行比较。
3. 如果类别不同，请将节点所需的类别映射到模型所在的现有文件夹。

   在此示例中，在 `olares_shared_models` 下添加以下映射：

   ```yaml
   ultralytics_bbox: ultralytics/bbox
   ```

   此映射允许节点从现有文件夹加载模型，无需移动或复制模型文件。

   有关添加映射的说明，请参阅[配置其他模型路径](./comfyui-launcher#配置其他模型路径)。

4. 保存配置并重启 ComfyUI。
5. 重新打开工作流，确认 `face_yolov8m.pt` 已显示在 `UltralyticsDetectorProvider` 的模型选择器中。

   ![自定义节点可以使用 Ultralytics 模型](/images/manual/use-cases/comfyui-common-model-recognized.png#bordered)

## 升级到 v1.0.37 或更高版本后 ComfyUI 无法启动

升级到 ComfyUI v1.0.37 或更高版本后可能会出现此问题。

升级后，ComfyUI 应用可能无法启动并显示如下错误：

```
main.py: error: unrecognized arguments: --normalvram
```

这意味着来自先前版本的自定义启动参数仍在使用，但新版本不再支持它。

要解决此问题：

1. 打开 **ComfyUI Launcher** 并从左侧边栏前往 **Lab**。
2. 在 **Manually edit extra arguments** 字段中，手动删除 `--normalvram` 并点击 **SAVE MANUAL ARGS**。或者，点击 **RESTORE DEFAULT** 重置为默认启动参数。
3. 验证顶部的 **Current full launch command** 不再包含 `--normalvram`。
4. 返回 ComfyUI Launcher 中的 **Home**，然后点击 **Start** 启动 ComfyUI。

## 模型无法直接下载到 Olares

某些模型需要登录、访问批准、令牌或手动确认才能下载。这些模型无法通过 ComfyUI Launcher 或 Server Download 直接下载到 Olares。

要解决此问题，请使用以下方法之一找到下载链接。然后手动下载模型并[上传](/use-cases/comfyui-launcher.md#上传本地模型)到 Olares Files 中的正确文件夹。

### 方法 1：检查模板备注或 Model Links 部分

某些官方模板包含备注或 **Model Links** 部分，列出：

- 所需的模型文件
- 下载 URL
- 预期的存储位置

如果可用，复制下载 URL 或直接打开模型页面。

![模型链接](/images/manual/use-cases/comfyui-model-links.png#bordered){width=90%}

### 方法 2：使用浏览器辅助扩展

如果模板显示缺失模型对话框但未暴露完整 URL，请使用浏览器辅助扩展，如 [WAN Download URL Helper](https://github.com/carlric/wan-download-url-helper)：

1. 在 ComfyUI 中打开缺失模型对话框。
2. 将鼠标悬停在下载图标上。
3. 右键点击图标并选择 **Show download URL**。
4. 复制 URL，然后在下载器中使用它或保存以供手动下载。

![ComfyUI 下载 URL 助手](/images/manual/use-cases/comfyui-download-url-helper.png#bordered){width=80%}

### 方法 3：在浏览器开发者工具中检查页面

如果 URL 未在模板备注或对话框中显示，请在浏览器开发者工具中检查页面，并查找由模板或缺失模型对话框触发的网络请求。

![检查 URL](/images/manual/use-cases/comfyui-inspect-url.png#bordered){width=80%}

## Olares One 上的 CPU 温度异常升高

当工作流需要的 VRAM 超过显卡所拥有的容量时，系统会将重负载放在单个 CPU 核心上进行数据交换，导致温度升高。

长期解决方案是减少工作流的 VRAM 占用，例如降低分辨率、使用更小的模型或启用模型卸载。作为临时解决方法，可以在工作负载运行期间限制最大 CPU 频率。

### Olares OS 1.12.6 或更高版本

Olares One 配备的 CPU 默认最大频率为 5.4 GHz。使用**限制 CPU 频率**开关，可在工作负载期间将其降至 5.0 GHz。工作负载完成后，再关闭该开关。

1. 打开**设置**。
2. 点击左上角头像，打开**我的 Olares**。
3. 在**硬件**下，开启**限制 CPU 频率**。
4. 在 ComfyUI 中运行任务。
5. 工作负载完成后，关闭**限制 CPU 频率**。

更多信息请参阅[限制 CPU 频率](/zh/manual/olares/settings/my-olares#limit-cpu-frequency)。

### Olares OS 1.12.5 或更早版本

如果设备运行 Olares OS 1.12.5 或更早版本，请使用终端命令在工作负载期间降低最大 CPU 频率，并在完成后恢复。

1. 打开控制面板。
2. 在左侧边栏的**终端**下，点击 **Olares**。

   ![打开终端](/images/zh/manual/use-cases/comfyui-ts-terminal.png#bordered){width=90%}

3. 运行以下命令将最大 CPU 频率降低到 5.0 GHz：
    ```bash
    echo 5000000 | sudo tee /sys/devices/system/cpu/cpufreq/policy*/scaling_max_freq
    ```
   在其他设备上，根据 CPU 最大频率调整目标值。先运行 `cat /sys/devices/system/cpu/cpufreq/policy0/cpuinfo_max_freq` 来检查。
4. 在 ComfyUI 中运行任务。
5. 工作负载完成后，运行以下命令恢复默认的 5.4 GHz 最大 CPU 频率：
    ```bash
    echo 5400000 | sudo tee /sys/devices/system/cpu/cpufreq/policy*/scaling_max_freq
    ```
