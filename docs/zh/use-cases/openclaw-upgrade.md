---
outline: [2, 3]
description: 了解升级 OpenClaw 时各版本的变更内容和故障排除步骤。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw 升级, 升级故障排除
app_version: "1.0.36"
doc_version: "1.4"
doc_updated: "2026-09-04"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/openclaw-upgrade.md)为准。
:::

# 升级 OpenClaw

升级 OpenClaw 前，建议先查看本页面的版本变更内容和故障排除步骤，确保升级顺利。

## 升级到 v2026.9.1

OpenClaw 2026.09.01 更新可能会导致部分插件不兼容，使网关无法启动（应用显示启动失败）。该更新还引入了更严格的代理归因检查，可能会在升级后阻止 Control UI 访问。

### 升级后网关无法启动

请按以下步骤解决。

1. 从启动台打开控制面板，找到 **clawdbot** 容器，然后点击其名称旁的终端图标。

    ![打开 OpenClaw 容器终端](/images/manual/use-cases/openclaw-container-terminal.png#bordered)

2. 在容器终端中运行修复命令，并按屏幕提示操作：

    ```bash
    openclaw update repair
    ```

3. 如果修复后网关仍无法启动，请逐个将不兼容的插件更新到最新版本：

    ```bash
    openclaw plugins update <插件名称>
    ```

    所有不兼容的插件更新完成后，网关应能正常启动。

### Control UI 显示代理归因错误

升级后，Control UI 可能会显示以下错误：

```text
Proxy client attribution is required. Configure gateway.trustedProxies narrowly and make the proxy overwrite or safely rebuild forwarded client headers.
```

请按以下步骤解决。

1. 打开 Files 应用，进入 **Data** > **clawdbot** > **config**，然后打开 `openclaw.json` 文件。
2. 点击右上角的编辑图标进入编辑模式。
3. 找到 `gateway` 部分，然后添加 `trustedProxies` 数组：

    ```json
    "trustedProxies": [
    "10.233.0.0/16",
    "10.244.0.0/16",
    "10.42.0.0/16",
    "127.0.0.1",
    "::1"
    ],
    ```

    ![添加 OpenClaw 配置](/images/manual/use-cases/openclaw-config-file.png#bordered)

4. 保存更改。
5. 返回控制面板，点击**部署**下的 **clawdbot**，然后点击右上角的**重启**。

     ![重启 OpenClaw](/images/zh/manual/use-cases/restart-openclaw.png#bordered)

6. 重新打开 Control UI，验证其是否正常加载。

更多信息，请参阅 [OpenClaw 发布说明](https://github.com/openclaw/openclaw/releases/tag/v2026.9.1)。

## 升级到 v2026.6.5

OpenClaw 2026.06.05 版本将认证配置文件、认证状态和定时任务从旧版 JSON 文件迁移到了内部 SQLite 数据库。网关现在从 SQLite 中读取这些配置，而非原始 JSON。

:::warning 强制数据迁移
升级完成后，你必须运行自动修复工具来迁移数据。如果你使用云端托管的模型提供商，在完成迁移之前，你的智能体可能无法与外部 API 通信，并报告以下错误：

```text
Missing API key for the selected provider on the gateway. Configure provider
auth, then try again.
```
:::

升级完成后，执行以下步骤迁移数据并恢复网关完整功能：
1. 从启动台打开 OpenClaw CLI。
2. 运行自动修复工具：

    ```bash
    openclaw doctor --fix
    ```

3. 查看输出日志。成功迁移后将显示与以下结构相符的确认信息：

    ```text
    |  Migrated auth profile JSON for ~/.openclaw/agents/main/agent/auth-profiles.json into  |
    |  SQLite (backups:                                                                      |
    |  ~/.openclaw/agents/main/agent/auth-profiles.json.sqlite-import.1781088154476.bak,     |
    |  ~/.openclaw/agents/main/agent/auth-state.json.sqlite-import.1781088154484.bak).       |
    ```

更多信息，请参阅 [OpenClaw 发布说明](https://github.com/openclaw/openclaw/releases/tag/v2026.6.5)。

## 升级到 v2026.5.26

OpenClaw 2026.05.26 版本引入了重大的架构变更。升级到该版本后，你的智能体可能会暂时失去部分功能，直到你将已安装的插件和技能更新到最新的兼容版本。

要恢复智能体的功能，打开 OpenClaw CLI，使用以下方法之一：

- **运行自动诊断工具（推荐）**：运行以下命令，让系统自动检测并修复兼容性问题。

    ```bash
    openclaw doctor --fix
    ```
- **更新所有插件**：运行以下命令，一次性批量更新所有已安装的插件。

    ```bash
    openclaw plugins update --all
    ```

    或者，你也可以根据需要逐个更新插件。

- **手动更新自定义插件**：如果你是通过手动方式安装插件的（例如，使用 `npx` 或直接上传文件），自动化 CLI 命令无法更新它们。你必须参考原始插件开发者的官方文档获取具体的升级说明。

更多信息，请参阅 [OpenClaw 发布说明](https://github.com/openclaw/openclaw/releases/tag/v2026.5.26)。

## 升级到 v2026.3.22

:::tip 前提条件
在将 OpenClaw 升级到 2026.03.22 之前，你必须先将 Olares OS 升级到 V1.12.5。
:::

OpenClaw 2026.03.22 版本引入了多项限制插件权限的变更。由于这一安全增强，旧版插件可能不再兼容。更多信息，请参阅 [OpenClaw 发布说明](https://github.com/openclaw/openclaw/releases/tag/v2026.3.22)。

如果你发现某个之前正常工作的插件在升级到该版本后不可用，可尝试以下解决方案：
- **更新插件**：检查是否有符合更新后权限限制的新版本可用。
- **验证配置方式**：咨询插件提供方，了解 OpenClaw 2026.03.22 及更高版本是否需要新的配置。

## 升级到 v2026.2.25

OpenClaw 2026.02.25 版本引入了一项安全增强，要求现有用户显式声明允许的 Control UI 访问地址。因此，如果升级后 Control UI 无法启动，可按照以下步骤解决。

1. 从启动台上打开控制面板，查看 **clawdbot** 的容器日志。

    ![查看容器日志](/images/zh/manual/use-cases/check-container-logs.png#bordered)

2. 查找以下错误信息。如果出现，继续下一步。

    ```text
    Gateway failed to start: Error: non-loopback Control UI requires gateway.controlUi.allowedOrigins (set explicit origins), or set gateway.controlUi.dangerouslyAllowHostHeaderOriginFallback=true to use Host-header origin fallback mode
    ```

    ![错误日志](/images/manual/use-cases/container-logs.png#bordered)

3. 打开设置，进入**应用** > **OpenClaw** > **Control UI** > **端点配置**，复制**端点**地址。

    ![OpenClaw 端点地址](/images/zh/manual/use-cases/openclaw-control-ui-endpoint.png#bordered){width=70%}

4. 打开文件管理器，进入**应用** > **数据** > **clawdbot** > **config**，右键点击 `openclaw.json` 文件，然后选择**下载**。

    ![OpenClaw 配置文件](/images/zh/manual/use-cases/openclaw-config-json.png#bordered)

5. 在文本编辑器中打开下载的文件，找到 `gateway` 部分，然后添加一个包含你端点地址的 `controlUi` 块。

    ```json
    "controlUi": {
      "allowedOrigins": ["Endpoint-Address"]
    },
    ```
    ![更新配置文件](/images/manual/use-cases/add-control-ui-endpoint.png#bordered)

    :::info
    如果你使用多个地址访问 Control UI（例如本地 URL 或自定义域名），可将它们以逗号分隔添加到 `allowedOrigins` 数组中。例如，`["https://url-one.com", "https://url-two.com"]`。
    :::

6. 返回文件管理器，将原始的 `openclaw.json` 文件重命名以保留备份，然后上传修改后的 `openclaw.json` 文件。
7. 返回控制面板，点击**部署**下的 **clawdbot**，然后点击右上角的**重启**。

     ![重启 OpenClaw](/images/zh/manual/use-cases/restart-openclaw.png#bordered)

8. 在**重启 clawdbot** 窗口中，准确输入 `clawdbot`，然后点击**确认**。等待程序状态显示为**Running**（以绿色圆点表示）。
9. 再次查看容器日志，验证网关是否已成功启动。

      ![验证容器日志](/images/manual/use-cases/verify-container-logs.png#bordered)

10. 打开 Control UI。如果仍然显示错误，请刷新浏览器页面。
