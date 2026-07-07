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

## 迁移到 1.12.6 的新版 ComfyUI

ComfyUI 1.12.6 使用了新的[共享应用机制](/zh/manual/olares/market/shared-apps)（参见 [PR #3485](https://github.com/beclab/Olares/pull/3485)）。要更新到新版本，你必须迁移现有的 ComfyUI。迁移过程是自动的——新版 ComfyUI 会在迁移过程中自动转移你的模型、插件、工作流以及输入输出文件。

已安装的 ComfyUI 在升级到 1.12.6 后仍可继续使用，但无法接收后续更新。我们建议你尽快迁移到新版 ComfyUI。

### 迁移步骤

1. 卸载当前的 ComfyUI。在弹窗中，**不要**勾选"Also remove all local data"。
2. 打开应用市场，搜索 **ComfyUI**，点击**安装**。

   在应用详情页查看 **Information** 下的 **Compatibility** 项，若显示为 **"Olares >=1.12.6-0"**，说明安装的是新版本。

:::info
卸载后，你的模型、插件、工作流和输入输出文件仍保留在设备上。新版 ComfyUI 会自动找到并迁移它们。
:::

### 迁移内容说明

新版 ComfyUI 安装后，将自动完成数据的迁移工作，具体规则如下：

| 数据类型 | 原路径 | 新路径 |
|:---|:---|:---|
| ComfyUI 核心数据（插件、工作流等） | `External/<your_hostname>/ai/comfyui/` | `/Data/comfyuisharev3/comfyui/` |
| 模型数据 | `External/<your_hostname>/ai/model/` | `Common/comfyui/model/` |
| 输出文件 | `External/<your_hostname>/ai/output/comfyui/` | `Common/comfyui/output/` |
| 输入文件 | `External/<your_hostname>/ai/comfyui/ComfyUI/input/` | `Common/comfyui/input/` |

:::warning
迁移完成后，请勿再将模型或输入文件上传至 `External/<your_hostname>/ai/` 目录下。新版 ComfyUI 不再挂载这两个目录，因此无法识别这些文件。
:::

数据迁移会在每次 ComfyUI 应用重启时进行。如果原路径下有新增数据，将在重启后按上述规则迁移到新目录位置，并删除 `External/<your_hostname>/ai/` 目录下对应的文件。为避免文件冲突和不必要的数据迁移等待，请使用新目录上传文件。

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
