---
outline: [2,3]
description: 本页面记录了在 Olares 上使用 Windows 时可能遇到的已知问题和意外行为，以及相应的解决方案或变通方法。
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/windows-issues.md)为准。
:::

# Windows VM 已知问题

使用本页面识别和排查 Olares 上 Windows 的当前已知问题。

## 升级到 Olares 1.12.5 后远程桌面连接失败

将 Olares 从版本 1.12.4 升级到 1.12.5 后，设备重启后您可能无法远程访问 Windows。

这是因为过时的 Tailscale 访问控制列表（ACL）设置导致的，该设置控制哪些设备可以访问您的 Windows VM。由于此设置在升级期间不会自动更新，它会阻止远程桌面连接。

### 变通方法

要恢复远程访问，请更新访问控制设置并重启所需的服务。

1. 打开 Control Hub。
2. 在左侧边栏中，在 **Terminal** 下点击 **Olares**。
   ![Olares terminal](/images/developer/develop/controlhub-terminal.png#bordered){width=90%}

3. 逐个运行以下命令。将 `<username>` 替换为您的 Olares ID 中 `@` 之前的部分。例如，如果您的 Olares ID 是 `alice123@olares.com`，则 `<username>` 为 `alice123`。

   ```bash
   kubectl annotate configmap tailscale-acl -n user-space-<username> tailscale-acl-md5-

   kubectl patch deployment headscale -n user-space-<username> --type=json -p='[{"op":"remove","path":"/spec/template/metadata/annotations/tailscale-acl-md5"}]'

   kubectl rollout restart sts app-service -n os-framework
   ```

4. 重启完成后，等待几分钟，然后再次尝试远程连接 Windows。
