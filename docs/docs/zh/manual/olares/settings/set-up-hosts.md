---
description: 在 Olares 中修改 Hosts 配置，解决网络访问限制问题。配置 IP 映射和验证解析的方法，确保应用正常访问外部资源。
---
# 修改 Hosts 配置

运行应用时，可能需要访问特定网站（例如 GitHub）来下载资源或获取数据。然而，由于网络环境的限制，这些网站可能无法正常访问。为了解决这一问题，Olares 支持通过系统设置更新系统 Hosts 文件（例如添加 GitHub 的 IP 映射），以支持 Olares 在运行时访问特定网站。

## 如何修改 Hosts 配置

1. 打开**设置**，进入**系统** > **Hosts**。
2. 右上角点击**添加 Hosts**，输入域名和对应的 IP 地址：
    - **Host 名称**：输入目标网站的域名，例如 `github.com`。
    - **IP 地址**：输入对应的 IP 地址，例如 `20.205.243.166`。

      ![添加 Host](/images/zh/manual/tasks/add-host.png#bordered)

3. 点击**确认**保存变更。

:::info DNS 缓存延迟
由于可能存在 DNS 缓存，Hosts 地址配置后可能需要等待一段时间才能生效。
:::
## Hosts 配置示例
以下是 GitHub 网站的常用 Hosts 配置：

| Host 名称               | IP 地址          |
|-------------------------|------------------|
| `github.com`            | `20.205.243.166` |
| `raw.githubusercontent.com` | `185.199.111.133` |

:::tip 查找可用 GitHub Hosts
示例的 IP 地址可能会随时间变化，请以 [https://hosts.gitcdn.top/hosts.txt](https://hosts.gitcdn.top/hosts.txt) 网站提供的最新地址为准。
:::

## 验证 Hosts 配置

在运行应用之前，可以使用 `nslookup` 命令来验证域名是否已正确解析为配置的 IP 地址。

打开命令提示符或终端，输入以下命令，替换 `[domain-name]` 为要验证的域名：
   ```shell
   nslookup [domain-name]
   ```
查看输出结果。例如：
   ```shell
   Name:    github.com
   Address: 20.205.243.166
   ```

确认 `Address` 字段是否与配置的 IP 地址一致。如果一致，则说明修改已生效。