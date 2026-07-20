---
outline: deep
description: 了解如何启用和配置 OpenClaw 沙盒，以实现安全的代码执行。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, 沙盒, 安全, 代码执行
app_version: "1.0.2"
doc_version: "1.1"
doc_updated: "2026-06-09"
---

# 可选：启用沙盒

默认情况下，OpenClaw 在其主容器内直接执行命令和代码。虽然这对于日常任务通常是安全的，但赋予智能体运行任意代码或安装外部依赖的能力仍存在固有风险。

为了最大程度地提升安全性并隔离潜在的危险操作，你可以启用 OpenClaw 沙盒。沙盒提供了一个专门用于代码执行的隔离的、临时的环境，确保你的核心系统始终受到保护。

:::tip 版本要求
要使用此功能，你的系统必须满足以下要求：
- **Olares OS**：升级至 V1.12.5 或更高版本。
- **OpenClaw**：升级至 V0.1.31 或更高版本。
:::

## 了解沙盒模式

配置沙盒时，`mode` 设置指定了沙盒的触发时机：
- **off**：沙盒已禁用。所有命令都在主容器中运行。
- **non-main**：沙盒会隔离通过 Discord 等外部渠道执行的命令。在 Control UI 的 **Chat** 页面中直接执行的命令将绕过沙盒，在主容器中运行。
- **all**：无论使用哪个界面或渠道，所有命令都在沙盒内运行。

## 启用沙盒

OpenClaw 沙盒默认处于禁用状态。你可以通过修改配置文件或使用 Control UI 来启用它。

<Tabs>
<template #通过配置文件启用>

1. 打开文件管理器，然后进入**数据** > **clawdbot** > **config**。
2. 双击打开 `openclaw.json` 文件。
3. 点击右上角的 <i class="material-symbols-outlined">edit_square</i> 进入编辑模式。
4. 找到 `agents` > `defaults` 部分，然后将以下 `sandbox` 配置块添加进去。

    :::info 新安装用户与升级用户的区别
    <ul><li>如果你是全新安装，沙盒配置块已存在，只需将 <code>mode</code> 的值从 <code>off</code> 改为 <code>non-main</code>。</li><li>如果你是从旧版本升级，则需要粘贴整个配置块。</li></ul>
    :::

      ```json
          "sandbox": {
            "mode": "non-main",
            "backend": "docker",
            "scope": "agent",
            "workspaceAccess": "rw",
            "docker": {
              "image": "beclab/harveyff-openclaw-sandbox-common:2026.4.7",
              "network": "bridge",
              "user": "1000:1000"
            },
            "prune": {
              "idleHours": 24,
              "maxAgeDays": 7
            }
          }
    ```

    ![通过配置文件启用沙盒](/images/zh/manual/use-cases/openclaw-edit-config-file.png#bordered)

5. 点击右上角的 <i class="material-symbols-outlined">save</i> 保存。
6. 重启 OpenClaw 使更改生效。
</template>
<template #通过-Control-UI-启用>

1. 打开 Control UI，从左侧边栏选择 **Settings**。
2. 点击 **AI & Agents**。
3. 向下滚动找到 **Sandbox** 部分，然后将其展开。
4. 按如下方式配置设置：
    - **Backend**：输入 `docker`。
    - 展开 **Docker**：
      - **Image**：输入 `beclab/harveyff-openclaw-sandbox-common:2026.4.7`。
      - **Network**：输入 `bridge`。
      - **User**：输入 `1000:1000`。
    - **Mode**：选择 **non-main**。
    - 展开 **Prune**：
      - **Idle Hours**：输入 `24`。
      - **Max Age Days**：输入 `7`。
    - **Scope**：选择 **agent**。
    - **Workspace Access**：选择 **rw**。

    ![在 Control UI 中启用沙盒](/images/manual/use-cases/openclaw-sandbox-enable-ui1.png#bordered)

5. 点击右上角的 **Save**。
6. 重启 OpenClaw 使更改生效。

</template>
</Tabs>

## 使用沙盒

启用后，每当需要执行命令时，OpenClaw 会自动创建并使用隔离的沙盒环境。

要验证沙盒是否正常工作，请使用 Discord 等外部渠道进行测试：

1. 确保你已[集成 Discord](/zh/use-cases/openclaw-integration.md)。
2. 在 Discord 中，向智能体发送以下私信：

    ```text
    Clone the repo https://github.com/beclab/core, read the package.json, then summarize what version it is and list its dependencies
    ```
3. 智能体会启动隔离的沙盒，安全地克隆仓库、读取文件并返回摘要。
4. 在智能体运行期间，打开 OpenClaw CLI 并运行以下命令以验证活动的沙盒：

    ```bash
    openclaw sandbox list
    ```

    终端将显示当前运行的沙盒容器，确认隔离已生效。

    ![验证沙盒](/images/manual/use-cases/openclaw-sandbox-verify.png#bordered)

## 授予额外的目录访问权限

默认情况下，`workspaceAccess: "rw"` 设置仅允许沙盒访问 OpenClaw 自身的工作区，以便智能体更新其记忆文件。

如果你希望沙盒与 Olares 文件交互，则必须使用自定义绑定挂载显式授予访问权限。这会将特定目录直接挂载到临时沙盒容器中。

### 授予访问权限

例如，要授予沙盒对 **Home** 目录的只读（`ro`）访问权限：
1. 确保 OpenClaw 可以通过启用 `ALLOW_HOME_DIR_ACCESS` 环境变量来访问 **Home** 目录中的本地文件：

    a. 打开设置，然后进入**应用** > **OpenClaw** > **管理环境变量**。

    b. 点击变量 `ALLOW_HOME_DIR_ACCESS` 右侧的 <i class="material-symbols-outlined">edit_square</i>，将变量值修改为 `true`，然后点击**确认**。

    c. 点击**应用**。

2. 打开文件管理器，然后进入**数据** > **clawdbot** > **config**。
3. 双击打开 `openclaw.json` 文件。
4. 点击右上角的 <i class="material-symbols-outlined">edit_square</i> 进入编辑模式。
5. 找到 `agents` > `defaults` > `sandbox` > `docker` 部分。
6. 按如下方式添加 `binds` 和 `dangerouslyAllowExternalBindSources` 行。确保在前面的 `"user": "1000:1000"` 行后面添加了逗号，以保持 JSON 语法有效。

    ```json
    "binds": ["/home/userdata/home:/home/userdata/home:ro"],
    "dangerouslyAllowExternalBindSources": true //允许沙盒访问默认工作区之外的目录
    ```

    ![授予对 Home 目录的只读访问权限](/images/manual/use-cases/openclaw-sandbox-readonly.png#bordered)

7. 点击右上角的 <i class="material-symbols-outlined">save</i> 保存更改。
8. 重启 OpenClaw 使更改生效。

### 测试访问权限

在上一步中，沙盒模式为 `non-main`，绑定挂载设置为 `ro`。要了解这些设置如何协同工作，你可以通过两个不同的界面进行测试。

#### 测试主会话

打开 Control UI 中的 **Chat** 页面，然后发送以下消息：

```text
Write a self-instruction file in txt format, and save it to the Documents folder in my Olares
```

**结果**：文件成功创建在指定目录中。

![在指定目录中成功创建文件](/images/manual/use-cases/openclaw-sandbox-file-created.png#bordered)

**原因**：通过 Control UI 的 **Chat** 页面发送的命令属于"主"会话。由于你将沙盒模式设置为 `non-main`，该会话完全绕过沙盒。智能体使用 OpenClaw 的默认系统权限写入文件。

#### 测试非主会话

打开 Discord，然后发送类似的消息：

```text
Write a sci-fi story outline in txt format, and save it to the Documents folder in my Olares Files
```

**结果**：文件创建失败。

![在指定目录中创建文件失败](/images/manual/use-cases/openclaw-sandbox-file-failure.png#bordered)

**原因**：通过 Discord 等外部渠道发送的命令会触发沙盒。由于你为 **Home** 目录配置了只读（`ro`）绑定挂载，智能体被阻止写入或修改任何文件。

## 了解更多

- [OpenClaw 沙盒隔离文档](https://docs.openclaw.ai/zh-CN/gateway/sandboxing)
- [自定义绑定挂载](https://docs.openclaw.ai/zh-CN/gateway/sandboxing#%E8%87%AA%E5%AE%9A%E4%B9%89%E7%BB%91%E5%AE%9A%E6%8C%82%E8%BD%BD)
