---
search: false
---
<!-- 可复用的 .local 域名内容。按行号范围引用。
     顺序：描述、URL+HTTP、Windows、常见问题。
     范围：描述 7-8，URL+HTTP 10-23，Windows 25-40（仅正文 26-40），常见问题 42-75 -->

当设备与 Olares 在同一局域网时，可使用 `.local` 域名访问服务，使流量保持在本地网络。

使用与标准 URL 结构一致的多级 `.local` 主机名。此格式适用于 Olares 系统应用和社区应用。

:::tip
使用 `.local` URL 时使用 `http://`，不要使用 `https://`。
:::

**标准 URL**
```text
https://<entrance_id>.<username>.olares.com
```
**本地 URL**
```text
http://<entrance_id>.<username>.olares.local
```

### Windows

在 Windows 上，`.local` 主机名默认不会解析。使用 LarePass 桌面端在 hosts 文件中添加所需条目，使多级 `.local` URL 解析到你的 Olares 设备。

1. 打开 LarePass 应用，点击头像，选择**设置**。
2. 找到**启用本地服务域名**并点击**添加**。LarePass 会自动更新 hosts 文件。

   ![启用本地服务域名](/images/one/larepass-win-update-hosts.png#bordered)
3. 更新完成后会显示成功提示。如果弹出命令行窗口，直接关闭。
4. （可选）要验证 hosts 变更：

   a. 进入 `C:\Windows\System32\drivers\etc\`。
   b. 用记事本打开 `hosts` 文件，可以看到 LarePass 添加的 `.local` 条目。

   ![LarePass 更新后的 hosts 文件](/images/one/larepass-updated-hosts.png#bordered)

### 为什么在 macOS 的 Chrome 中 .local 域名无法使用？

如果 macOS 没有授予 Chrome 本地网络访问权限，Chrome 可能会拦截本地 URL。

1. 打开 Apple 菜单，进入**系统设置**。
2. 进入**隐私与安全性** > **本地网络**。
3. 找到 **Google Chrome** 和 **Google Chrome Helper**，打开其开关。
4. 重启 Chrome 后再次尝试 `.local` URL。

![启用本地网络](/images/manual/larepass/mac-chrome-local-access.png#bordered){width=400}

### 为什么应用显示“连接不安全”或在 Chrome 中无法加载？

Chrome 有时会强制对 `.local` 主机名使用 HTTPS，而本地并不支持。

在 URL 开头显式使用 `http://`（例如 `http://desktop.<username>.olares.local`）。在家庭网络中，这类未加密本地连接是预期行为，能保证 `.local` 域名正常使用。

![错误的本地地址](/images/manual/get-started/incorrect-local-address.png#bordered)

### 为什么在 Safari 中打开 .local URL 时 iframe 会闪烁？

Safari 对 iframe 中的 `.local` 及其他非 HTTPS 内容处理更严格，可能导致 iframe 闪烁或重载。在**隐私**设置中启用下面两个选项即可解决。

解决方法：

1. 打开 **Safari**，进入**设置**。
2. 打开**隐私**标签页。
3. 启用下面两个选项：
   - 防止跨站跟踪
   - 对跟踪器隐藏 IP 地址

   ![Safari 隐私设置（.local）](/images/manual/get-started/safari-privacy-settings.png#bordered){width=70%}
4. 重新加载 `.local` 页面。
