---
outline: [2, 3]
description: NemoClaw 在 Olares 上的常见问题和解决方法。
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, 常见问题, 故障排除
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-05-11"
---

# 常见问题

本文列出 NemoClaw 在 Olares 上的常见问题及其解决方法。

## Discord 频道卡在 `startup-not-ready` 状态

在 NemoClaw CLI 沙盒中配置 Discord 后，频道可能会在 Web UI 中显示 `startup-not-ready`。

![OpenClaw Web UI 中的 startup-not-ready 错误](/images/manual/use-cases/nemoclaw-startup-not-ready.png#bordered){width=60%}

要恢复，请从 NemoClaw CLI 重启 gateway：

1. 从启动台打开 NemoClaw CLI 应用。

2. 在 shell 提示符下停止 gateway：

   ```bash
   docker exec openshell-cluster-nemoclaw kubectl -n openshell exec my-assistant -c agent -- \
     sh -lc 'openclaw gateway stop 2>/dev/null || pkill -9 -f "openclaw.*gateway|openclaw-gateway|gateway run" 2>/dev/null || true'
   ```

3. 启动 gateway：

   ```bash
   sh /opt/nemoclaw/sandbox-ensure-gateway.sh
   ```

等待约 10 到 15 秒，然后刷新页面。频道应会恢复就绪状态。

## 缺少默认 workspace 文件

NemoClaw 在安装过程中可能无法创建默认 workspace 文件。作为临时解决方法，请参考 [OpenClaw 默认 Agent 文档](https://docs.openclaw.ai/reference/AGENTS.default)和[官方模板](https://github.com/openclaw/openclaw/tree/main/docs/reference/templates)，手动创建所需文件。

## Olares CLI 登录和技能在重启后不会保留

NemoClaw 不会在重启后保留你的 Olares CLI 登录状态或已安装的 ClawHub 技能。重启 NemoClaw 后，请重新登录 Olares CLI，并重新安装 Olares 技能。详情参见[使用 Olares CLI 管理 Olares](nemoclaw-olares-cli.md)。

## OpenClaw Web UI 显示 `unauthorized: gateway token missing`

NemoClaw 重启后，OpenClaw Web UI 可能会显示以下错误：

```text
unauthorized: gateway token missing (open the dashboard URL and paste the token in Control UI settings)
```

要想恢复，需从 NemoClaw CLI 获取 gateway token，并将其粘贴到 Control UI 设置中：

1. 从启动台打开 NemoClaw CLI 应用。

2. 在 shell 提示符下运行以下命令，打印 gateway token：

   ```bash
   nemoclaw my-assistant gateway-token --quiet
   ```

3. 复制终端中显示的 token。

4. 返回 OpenClaw Web UI，将 token 粘贴到 **Gateway Token** 字段，然后点击 **Connect**。

   ![在 OpenClaw Web UI 中粘贴 gateway token](/images/manual/use-cases/nemoclaw-gateway-token.png#bordered){width=60%}
