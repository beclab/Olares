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
| ComfyUI 核心数据（插件、工作流等） | `External/<your_hostname>/ai/comfyui/` | `/Data/comfyuisharev3/comfyui/` |
| 模型数据 | `External/<your_hostname>/ai/model/` | `/Common/comfyui/model/` |
| 输出文件 | `External/<your_hostname>/ai/output/comfyui/` | `/Common/comfyui/output/` |
| 输入文件 | `External/<your_hostname>/ai/comfyui/ComfyUI/input/` | `/Common/comfyui/input/` |

:::warning
迁移完成后，需将新的模型和输入文件上传到 `Common/comfyui/` 下的新路径。新版 ComfyUI 不再将 `External/<your_hostname>/ai/` 作为当前使用的文件位置。
:::

数据迁移会在每次 ComfyUI 重启时运行。如果你之后又将文件添加到旧路径，ComfyUI 会在下次重启时将这些文件移动到新路径，并删除 `External/<your_hostname>/ai/` 下的原文件。为避免混淆，需直接使用新路径上传文件。

### 迁移后设置 `extra_model_paths.yaml`

迁移后，ComfyUI 会自动生成 `extra_model_paths.yaml` 配置文件，该文件告诉 ComfyUI 在共享模型中心在哪里查找模型。

文件预配置如下：
- **`base_path`**：指向 `/Common/comfyui/model`（共享模型目录）。

你通常不需要编辑此文件。但是，如果你有需要额外路径的自定义插件或外部资源，可以在此位置手动编辑：
```
/Data/comfyuisharev3/comfyui/user/extra_model_paths.yaml
```

有关编辑此文件的详细信息，请参阅[管理和使用文件目录](/zh/use-cases/comfyui-launcher#关于-extra_model_pathsyaml)。

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

## 工作流无法找到存储在 `/Common/comfyui/model/` 中的模型

迁移到 ComfyUI v3（Olares 1.12.6+）后，工作流可能报告模型缺失，即使模型文件确实存在于 `/Common/comfyui/model/` 中。主要有以下两种原因：

- **原因 1**：`extra_model_paths.yaml` 中未注册模型的子目录，导致 ComfyUI 未扫描该目录。
- **原因 2**：模型已在 ComfyUI 模型目录中，但自定义节点定义的搜索路径不匹配。

### 步骤 1：检查模型的子目录是否被识别

1. 在 ComfyUI 中，打开 **Model Library** 侧边栏，搜索模型文件名。
2. 如果模型未出现在列表中，说明其子目录未在 `extra_model_paths.yaml` 中注册。

   以下面的例子，`ultralytics/bbox/face_yolov8m.pt` 被检测到，但 `detection/mediapipe_face_fp32.safetensors` 未被识别：

   ![模型检测到 vs 未检测到](../public/images/manual/use-cases/comfyui-common-model-detected.png#bordered){width=49%}
   ![模型未检测到](../public/images/manual/use-cases/comfyui-common-model-missing.png#bordered){width=49%}

### 步骤 2：将缺失的子目录添加到 `extra_model_paths.yaml`

1. 打开 Files 并导航到 `/Data/comfyuisharev3/comfyui/user/`。
2. 打开 `extra_model_paths.yaml`。顶部的 `base_path` 指向 `/mnt/olares-shared-model`，对应容器内的 `/Common/comfyui/model/`。

   ![extra_model_paths.yaml 文件](../public/images/manual/use-cases/comfyui-extra-model-paths-file.png#bordered)

   例如，如果你将模型放在 `/Common/comfyui/model/detection/mediapipe_face_fp32.safetensors`，但该模型未出现在 Model Library 中，需添加 `detection` 子目录的映射：

   ```yaml
   base_path: /mnt/olares-shared-model
   detection: detection
   ```

   键（`detection`）为 `/Common/comfyui/model/` 下的子目录名称，值（`detection`）为 ComfyUI 模型搜索路径中的显示名称。

   ![添加 detection 映射](../public/images/manual/use-cases/comfyui-extra-model-paths-add-detection.png#bordered)

3. 保存文件，并在 ComfyUI Launcher 中点击 **Restart** 重启 ComfyUI。
4. 在启动日志中查找类似 `Adding extra search path detection /mnt/olares-shared-model/detection` 的行，确认路径已注册。

   ![启动日志确认 detection 路径已添加](../public/images/manual/use-cases/comfyui-detection-path-added-log.png#bordered)

5. 刷新 ComfyUI 页面，再次检查 Model Library。

   ![重启后模型已被识别](../public/images/manual/use-cases/comfyui-model-recognized-after-restart.png#bordered)

### 步骤 3：排查自定义节点路径不匹配

如果模型已出现在 Model Library，但特定工作流节点仍无法找到模型，说明该自定义节点使用的搜索路径与模型实际位置不同。

1. 查看该自定义节点的文档或源代码，确认其使用的具体搜索路径。

   例如，`ImpactPack/UltralyticsDetectorProvider` 节点监听的路径为 `ultralytics_bbox` 和 `ultralytics_segm`，而非标准的 `ultralytics/` 文件夹。

2. 将所需路径映射添加到 `extra_model_paths.yaml`。例如，要让 bbox YOLO 模型可被该节点使用：

   ```yaml
   base_path: /mnt/olares-shared-model
   ultralytics_bbox:
     models: /Common/comfyui/model/ultralytics/bbox
   ```

   ![添加 ultralytics_bbox 映射](../public/images/manual/use-cases/comfyui-ultralytics-bbox-mapping.png#bordered)

3. 重启 ComfyUI 并重新加载工作流。

   ![face_yolov8m.pt 现已被识别](../public/images/manual/use-cases/comfyui-face-yolov8m-recognized.png#bordered)

### 验证模型文件位置

确保将模型文件放置在 `/Common/comfyui/model/` 下的正确子文件夹中。ComfyUI 通过文件夹名称区分模型类型。例如：

| 模型类型 | 预期文件夹 |
|--|--|
| Checkpoint 模型（SD 1.5、SDXL、Flux） | `/Common/comfyui/model/checkpoints/` |
| LoRA 权重 | `/Common/comfyui/model/lora/` |
| VAE 模型 | `/Common/comfyui/model/vae/` |
| ControlNet 模型 | `/Common/comfyui/model/controlnet/` |

有关模型类型及其预期文件夹的完整列表，请参阅[理解 `model` 目录结构](/zh/use-cases/comfyui-launcher#理解-model-目录结构)。

### 刷新 ComfyUI

验证配置和文件位置后，刷新 ComfyUI 页面（F5 或 Ctrl+R）并再次尝试工作流。

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
