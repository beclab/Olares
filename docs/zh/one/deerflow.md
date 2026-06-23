---
outline: [2, 3]
description: 了解如何在 Olares 设备上设置 DeerFlow，包括 Ollama 集成和用于网络研究的 Tavily。
head:
  - - meta
    - name: keywords
      content: Deerflow, Ollama, Deep research
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/deerflow.md)为准。
:::

# 使用 DeerFlow 进行深度研究 <Badge type="tip" text="20 min" />
DeerFlow 是一个开源框架，能将简单的研究主题转化为全面、详细的报告。

本指南将引导你完成在 Olares 设备上设置 DeerFlow 的过程，将其与本地 Ollama 模型和 Tavily 搜索引擎集成，以实现支持网络搜索的研究。

## 学习目标
- 配置 DeerFlow 通过 Ollama 与本地 LLM 通信。
- 配置 Tavily 搜索 API 以实现网络访问。
- 执行深度研究任务并管理报告。

## 开始之前
DeerFlow 使用 `.com` 域名连接 Ollama。
* 本地访问：如果 Ollama API 的 **Authentication model** 为 **None**，则不需要 LarePass VPN。
* 远程访问：你必须启用 LarePass VPN。

## 前提条件
**硬件** <br>
- Olares One 连接到稳定的网络。 
- 有足够的磁盘空间来下载模型。

**用户权限**
- 拥有 Admin 权限，以便从 Market 安装 Ollama。

**外部服务** <br>
- 需要 [Tavily](https://www.tavily.com/) 账户来生成网络搜索的 API key。免费计划即可满足需求。

**LarePass**（远程访问需要）
- 设备上已安装 LarePass 应用。本指南使用桌面应用来演示从桌面端的配置和使用。

## 步骤 1：安装 Ollama 和 DeerFlow
1. 打开 Market，搜索 "Ollama"。
   ![安装 Ollama](/images/manual/use-cases/ollama.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

3. 重复相同步骤安装 "DeerFlow"。
   ![安装 DeerFlow](/images/manual/use-cases/deerflow.png#bordered)

## 步骤 2：安装模型并配置 Ollama

### 安装语言模型
1. 从 Launchpad 打开 Ollama。
2. 使用 `pull` 命令下载模型：
   ```bash
   ollama pull [model]
   ```
   例如：
   ```bash
   ollama pull cogito:14b
   ```
   :::warning 不支持推理模型
   DeerFlow 目前仅支持非推理模型。此外，由于缺乏工具使用能力，Gemma-3 模型也不受支持。
   :::
   :::tip 查看 Ollama 库
   如果你不确定下载哪个模型，请查看 [Ollama Library](https://ollama.com/library) 以探索可用模型。
   :::

3. 验证下载：
   ```bash
   ollama ls
   ```
如果模型出现在列表中，则表示已准备好使用。

### 配置 API 访问
要将 Ollama 用作 DeerFlow 的后端，你必须配置 API 以允许其他应用访问。

### 验证认证级别
默认情况下，Ollama API 的认证级别设置为 **Internal**，允许同一本地网络上的应用访问它。

作为 Super Admin，你可以验证或修改认证级别：
1. 打开 Settings，然后导航到 **Applications** > **Ollama** > **Ollama API**。
2. 确认 **Authentication level** 设置为 **Internal**。
3. （可选）如果你不想启用 LarePass VPN，将 **Authentication model** 设置为 **None**。
4. 如果做了更改，点击 **Submit**。
   ![验证认证级别](/images/manual/use-cases/ollama-authentication-level.png#bordered)

### 获取端点

1. 在同一设置页面，点击 **Set up endpoint**。
2. 复制地址以备后用。

   ![获取 Ollama 端点](/images/manual/use-cases/ollama-endpoint.png#bordered)

### 可选：启用 LarePass VPN

如果你从远程网络访问 Olares，或者选择了 **None** 以外的认证模型，则必须在客户端设备上启用 LarePass VPN 以建立连接。

1. 打开 LarePass 应用，点击左上角的头像打开用户菜单。
2. 打开 **VPN connection** 的开关。

   ![在桌面端启用 LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered)


## 步骤 3：配置 DeerFlow
DeerFlow 需要 LLM 的连接详情。你将通过图形界面或命令行编辑 `conf.yaml` 文件来配置。

### 配置 DeerFlow 使用 Ollama

<tabs>
<template #Use-graphical-interface>

1. 打开 Files 应用，导航到 `/Applications/Data/Deerflow/app/`。
2. 找到 `conf.yaml` 文件并双击打开。
   ![在 Files 中查找 `conf.yaml`](/images/manual/use-cases/deerflow-conf-yaml-in-files.png#bordered)

3. 点击右上角的 <span class="material-symbols-outlined">box_edit</span> 打开文本编辑器。
4. 修改默认模型设置：
   ```yaml
    BASIC_MODEL:
      base_url:  # 你的 Ollama API 端点（确保包含 /v1 后缀）
      model: # 模型名称
      api_key: # 任意非空字符串
   ```
   例如：
   ```yaml
    BASIC_MODEL:
      base_url: https://a5be22681.laresprime.olares.com
      model: "cogito:14b"
      api_key: ollama
   ```
5. 点击 <span class="material-symbols-outlined">box_edit</span> 保存文件。

</template>

<template #Use-command-line>

你可以直接在主机上通过终端编辑配置文件。
1. 打开 Control Hub，从侧边栏选择 DeerFlow 项目。
2. 导航到 **Deployments** > **deerflow** 并点击运行中的 pod。
3. 展开 **deerflow** 容器详情以查看 **Volumes** 部分。
   ![定位 DeerFlow 的容器](/images/manual/use-cases/deerflow-locate-containers.png#bordered)

   ![查找 app 文件夹](/images/manual/use-cases/deerflow-app-volume.png#bordered)

4. 复制此路径。
5. 从 Control Hub 打开 Olares 终端，使用 `cd` 命令切换到复制的路径：
   ```bash
   # 替换为你的实际路径
   cd /olares/rootfs/userspace/pvc-userspace-laresprime-raizlofhiszoin5c/Data/deerflow/app
   ```
6. 使用命令行文本编辑器（如 `nano` 或 `vi`）编辑 `conf.yaml` 文件。例如：
   ```Bash
   nano conf.yaml
   ```
7. 修改默认模型设置：
   ```yaml
    BASIC_MODEL:
      base_url:  # 你的 Ollama API 端点（确保包含 /v1 后缀）
      model: # 模型名称
      api_key: # 任意非空字符串
   ```
   例如：
   ```yaml
    BASIC_MODEL:
      base_url: https://a5be22681.laresprime.olares.com/v1
      model: "cogito:14b"
      api_key: ollama
   ```
8. 保存更改并退出编辑器。
</template>
</tabs>


### 配置 DeerFlow 使用 Tavily
要启用网络搜索，请将你的 Tavily API key 添加到应用配置中。
1. 在 Control Hub 中，选择 DeerFlow 项目。
2. 在资源列表中点击 **Configmaps**，然后选择 **deerflow-config**。
   ![浏览到 DeerFlow 的 configmaps](/images/manual/use-cases/deerflow-configmap.png#bordered)

3. 点击右上角的 <span class="material-symbols-outlined">edit_square</span> 打开编辑器。
4. 在 `data` 部分下添加以下键值对：
   ```yaml
   SEARCH_API: tavily
   TAVILY_API_KEY: tvly-xxx # 你的 Tavily API Key
   ```
   ![配置 Tavily](/images/manual/use-cases/deerflow-configure-tavily.png#bordered)

5. 点击 **Confirm** 保存更改。

### 重启 DeerFlow
重启服务以应用新的模型和搜索配置。

1. 在 Control Hub 中，选择 DeerFlow 项目。
2. 在 **Deployments** 下，找到 **deerflow** 并点击 **Restart**。
   ![重启 DeerFlow](/images/manual/use-cases/deerflow-restart.png#bordered)

3. 在确认对话框中，输入 `deerflow` 并点击 **Confirm**。 
4. 等待状态图标变为绿色，表示服务已成功重启。

## 步骤 4：运行 DeerFlow
### 运行深度研究任务
1. 从 Olares Launchpad 打开 DeerFlow。
2. 点击 **Get Started** 并在提示框中输入你的研究主题。
   ![输入研究提示](/images/manual/use-cases/deerflow-enter-prompt.png#bordered)

3. （可选）点击魔杖图标，让 DeerFlow 优化你的提示以获得更好的结果。
4. 启用 **Investigation**。
5. （可选）选择你喜欢的写作风格（例如 **Popular Science**）。
6. 点击 <span class="material-symbols-outlined">arrow_upward</span> 发送请求。

DeerFlow 将生成一份初步研究计划。如有必要，请审阅并编辑此计划，或允许它继续执行。
![生成研究计划](/images/manual/use-cases/deerflow-generate-research-plan.png#bordered)

流程完成后，将显示一份详细的分析报告。
![查看研究报告](/images/manual/use-cases/deerflow-generate-research-report.png#bordered)

要审计所使用的来源和步骤，请点击 **Activities** 标签页。
![审阅研究活动](/images/manual/use-cases/deerflow-review-research-activities.png#bordered)

### 编辑并保存报告
:::info 验证引用
AI 模型偶尔可能会生成不准确的引用或"幻觉"链接。请务必在引用部分手动验证重要来源。
:::

1. 点击右上角的 <span class="material-symbols-outlined">edit</span> 进入编辑模式。
2. 你可以使用 Markdown 调整格式，或选择某个部分让 AI 改进或扩展它。
   ![让 AI 编辑报告](/images/manual/use-cases/deerflow-ask-ai-to-edit.png#bordered)
3. 点击右上角的 <span class="material-symbols-outlined">undo</span> 退出编辑模式。
4. 点击 <span class="material-symbols-outlined">download</span> 将报告作为 Markdown 文件保存到本地。

## 高级用法
### 添加 MCP 服务器
模型上下文协议（MCP）通过集成外部工具来扩展 DeerFlow 的能力。例如，添加 Fetch 服务器可以让代理抓取网页内容并将其转换为 Markdown 进行分析。

1. 打开 DeerFlow 应用，点击 <span class="material-symbols-outlined">settings</span> 打开 **Settings** 对话框。
2. 选择 **MCP** 标签页并点击 **Add Servers**。
3. 粘贴服务器的 JSON 配置。以下示例添加 fetch 服务器：
   ```json
    {
      "mcpServers": {
        "fetch": {
          "command": "uvx",
          "args": ["mcp-server-fetch"]
        }
      }
    }
   ```
4. 点击 **Add**。服务器将自动启用，并可供研究代理使用。
   ![添加 MCP 服务器](/images/manual/use-cases/deerflow-add-mcp-server.png#bordered)

### 将研究报告转为播客（TTS）
DeerFlow 可以使用文本转语音（TTS）服务（如火山引擎 TTS）将报告转换为 MP3 音频。这需要将 API 凭据添加到应用环境中。

1. 从 [火山引擎](https://console.volcengine.com) 控制台获取你的 **Access Token** 和 **App ID**。
2. 在 Control Hub 中，选择 DeerFlow 项目并前往 **Configmaps** > **deerflow-config**。
3. 点击右上角的 **Edit** 图标。
4. 在 `data` 部分下添加以下键：
   ```yaml
   VOLCENGINE_TTS_ACCESS_TOKEN: # 你的 Access Token
   VOLCENGINE_TTS_APPID: # 你的 App ID
   ```
5. 点击 **Confirm** 保存更改。
6. 导航到 **Deployments** > **deerflow** 并点击 **Restart**。

重启后，DeerFlow 应能检测到这些键，播客/TTS 功能将可用。

## 故障排除
### DeerFlow 没有生成响应
如果代理无法启动或挂起：
* **重启 Ollama**：在 **Control Hub** 中重启 Ollama 服务。
- **检查模型兼容性**：DeerFlow 不支持推理模型（例如 DeepSeek R1）。切换到标准聊天模型后重试。
- **检查端点配置**：确保 `conf.yaml` 中的 Ollama API 端点包含 `/v1` 后缀。

### 研究过程中没有网络搜索结果
如果报告内容泛化且缺乏外部数据：
- **检查模型能力**：所选 LLM 可能缺乏强大的工具调用能力。切换到以有效工具使用著称的模型，如 Qwen 2.5 或 Llama 3.1。
- **验证 API Key**：确保 ConfigMap 中的 `TAVILY_API_KEY` 正确，且账户仍有剩余额度。

## 资源
- [DeerFlow GitHub](https://github.com/bytedance/deer-flow)
- [通过 Ollama 下载和运行本地 AI 模型](../use-cases/ollama.md)
