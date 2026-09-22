---
search: false
head:
  - - meta
    - name: keywords
      content: Olares, hosts 映射, .local 域名, 局域网访问, hosts 文件
---
<!-- 可复用的本地域名内容。请通过命名 region 引用。 -->

<!-- #region local-domain-overview -->
当设备与 Olares 位于同一局域网时，可以通过局域网直连。你既可以继续使用标准 `olares.com` 地址，也可以使用 `.local` 地址。

<!-- #region local-domain-url-format -->
多级 `.local` 主机名与标准 Olares 地址的结构相同，适用于系统应用和社区应用。

:::tip
标准地址使用 `https://`，`.local` 地址使用 `http://`。
:::

**标准地址**
```text
https://<entrance_id>.<username>.olares.com
```
**`.local` 地址**
```text
http://<entrance_id>.<username>.olares.local
```
<!-- #endregion local-domain-url-format -->
<!-- #endregion local-domain-overview -->

<!-- #region larepass-local-domains-summary -->
在 Windows 和 macOS 上，LarePass 桌面端可以为当前电脑配置局域网直连。它会将 `olares.com` 和 `olares.local` 主机名写入 hosts 文件，并将它们映射到 Olares 的局域网 IP。

1. 确保电脑与 Olares 位于同一局域网。
2. 在 LarePass 桌面端关闭**专用网络连接**。
3. 通过以下任一入口开始更新：
   - 点击左下角的**配置 hosts**。
   - 点击头像，前往**设置** > **hosts 映射**，然后点击**启用**。
4. 在**更新 hosts 映射**弹窗中检查条目，并按需编辑。请勿修改以 `#` 开头的行，LarePass 使用这些标记管理条目。完成后点击**更新**。
5. 输入电脑的管理员密码并确认更改。
6. 等待**成功**提示出现。

之后，可以使用标准 `https://<entrance_id>.<username>.olares.com` 地址，或对应的 `http://<entrance_id>.<username>.olares.local` 地址。

:::info 停用 hosts 映射
1. 在 LarePass 桌面端点击头像，然后进入**设置**。
2. 在**hosts 映射**下点击**关闭**。
3. 在**停用 hosts 映射**确认窗口中再次点击**关闭**。

LarePass 会删除此前由它添加到这台电脑的全部 hosts 条目。
:::
<!-- #endregion larepass-local-domains-summary -->

<!-- #region windows-local-domain -->
在 Windows 上，使用 LarePass 桌面端将所需条目添加到 hosts 文件，使多级 `.local` 主机名解析到 Olares 的局域网 IP。

1. 确保电脑与 Olares 位于同一局域网。
2. 在 LarePass 桌面端关闭**专用网络连接**。
3. 通过以下任一入口开始更新：
   - 点击左下角的**配置 hosts**。
   - 点击头像，前往**设置** > **hosts 映射**，然后点击**启用**。
4. 在**更新 hosts 映射**弹窗中检查条目，并按需编辑。请勿修改以 `#` 开头的行，LarePass 使用这些标记管理条目。完成后点击**更新**。
5. 输入电脑的管理员密码并确认更改。
<!-- #endregion windows-local-domain -->

<!-- #region larepass-local-domains -->
使用 LarePass 桌面端的**hosts 映射**管理 Olares 的 hosts 条目。LarePass 会将 `olares.com` 和 `olares.local` 主机名映射到 Olares 的局域网 IP，使这台电脑发出的请求留在局域网内。

此模式仅对当前电脑生效，并且只会在电脑与 Olares 位于同一局域网时显示。

:::warning 请先关闭专用网络
LarePass 专用网络与 hosts 映射不能同时启用。添加或更新 hosts 条目前，请先关闭 **专用网络连接**。
:::

1. 确保电脑与 Olares 位于同一局域网。
2. 在 LarePass 桌面端关闭 **专用网络连接**。
3. 通过以下任一入口开始更新：
   - 点击左下角的**配置 hosts**。
   - 点击头像，前往**设置** > **hosts 映射**，然后点击**启用**。
4. 在**更新 hosts 映射**弹窗中检查条目，并按需编辑。请勿修改以 `#` 开头的行，LarePass 使用这些标记管理条目。完成后点击**更新**。
5. 出现密码提示后，输入电脑的管理员密码并确认更改。
6. 等待**成功**提示出现。

现在可以打开上面列出的任一地址。

当 Olares 的局域网 IP 发生变化，或安装、卸载应用导致所需主机名变化时，LarePass 会提示更新 hosts 条目。请先完成更新，再使用受影响的地址。

:::tip
无需手动编辑 hosts 文件。LarePass 会在 Windows 和 macOS 上管理这些条目。
:::

:::info 停用 hosts 映射
1. 点击头像并进入**设置**。
2. 找到**hosts 映射**，然后点击**关闭**。
3. 在**停用 hosts 映射**确认窗口中再次点击**关闭**。

LarePass 会删除此前由它添加到这台电脑的全部 hosts 条目。标准 `olares.com` 地址随后恢复使用常规网络路径；在不支持原生解析的系统上，多级 `.local` 地址可能不再能够解析。
:::
<!-- #endregion larepass-local-domains -->

<!-- #region larepass-local-domain-faq -->
### hosts 映射

#### 为什么在 LarePass 中找不到 hosts 映射？

此选项只会在电脑与 Olares 位于同一局域网时显示。请检查两台设备的网络连接，然后重新打开 LarePass。

#### 为什么无法启用 LarePass 专用网络？

hosts 映射与 LarePass 专用网络不能同时启用。前往 **设置** > **hosts 映射**，先停用 hosts 映射。

#### 为什么 LarePass 再次提示更新 hosts 文件？

当 Olares 的局域网 IP 发生变化，或安装、卸载应用时，所需条目可能随之变化。检查条目并点击 **更新**，确保所有由 LarePass 管理的主机名都指向当前局域网 IP。

#### 停用 hosts 映射后会发生什么？

LarePass 会删除由它管理的 hosts 条目。标准 `olares.com` 地址随后恢复使用常规网络路径；在不支持原生解析的系统上，多级 `.local` 地址可能不再能够解析。

#### 如何查看 hosts 文件或恢复备份？

在**更新 hosts 映射**弹窗中点击**在文件夹中显示**，可打开 hosts 文件所在文件夹。如果 hosts 文件自上次更新后发生变化，LarePass 会在更新前创建备份。你可以手动恢复备份。
<!-- #endregion larepass-local-domain-faq -->

<!-- #region local-domain-faq -->
### `.local` 浏览器问题

#### 为什么在 macOS 的 Chrome 中无法使用 .local 域名？

如果 macOS 没有授予 Chrome 本地网络访问权限，Chrome 可能会拦截本地地址。

1. 打开 Apple 菜单，进入**系统设置**。
2. 进入**隐私与安全性** > **本地网络**。
3. 找到 **Google Chrome** 和 **Google Chrome Helper**，打开相应开关。
4. 重启 Chrome，然后再次尝试 `.local` 地址。

![启用本地网络](/images/manual/larepass/mac-chrome-local-access.png#bordered){width=400}

#### 为什么应用显示“连接不安全”或在 Chrome 中无法加载？

Chrome 有时会强制对 `.local` 主机名使用 HTTPS，但 `.local` 地址不支持 HTTPS。

请在地址开头明确输入 `http://`，例如 `http://desktop.<username>.olares.local`。

![错误的本地地址](/images/manual/get-started/incorrect-local-address.png#bordered)

#### 为什么在 Safari 中打开 .local 地址时 iframe 会闪烁？

Safari 对 iframe 中的 `.local` 及其他非 HTTPS 内容处理更严格，可能导致 iframe 闪烁或重新加载。

1. 打开 **Safari**，进入**设置**。
2. 打开**隐私**标签页。
3. 启用**防止跨站跟踪**和**对跟踪器隐藏 IP 地址**。

   ![Safari 的 .local 隐私设置](/images/manual/get-started/safari-privacy-settings.png#bordered){width=70%}
4. 重新加载 `.local` 页面。
<!-- #endregion local-domain-faq -->
