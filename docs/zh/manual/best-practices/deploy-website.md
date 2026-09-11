---
outline: [2, 3]
description: 通过由 olares-cli agent skills 驱动的 AI 智能体，把已开发好的网站项目发布到自定义域名。
head:
  - - meta
    - name: keywords
      content: Olares, Lares, 发布网站, 自定义域名, olares-cli, olares-cli agent skills
---

# 发布网站到自定义域名

在 Olares 上，你可以让 AI 智能体帮你部署已经开发好的网站。在 Olares CLI 的支持下，智能体会完成打包、推送镜像、安装到设备、绑定自定义域名等一系列操作，让网站以一个固定、公开的 HTTPS 地址对外可用。

本教程以 Lares 和一个托管在 GitHub 上的网站项目为例，演示完整的部署流程。通过其他配备了 Olares CLI agent skills 的 AI 智能体部署你自己的网站项目时，步骤同样适用。

## 前提条件

开始前，请确认已有：

**Olares 环境**

- 一台运行 v1.12.7 或更高版本的 Olares 设备；
- 已安装 Lares 和 Router，并连接了本地模型；
- Olares CLI（v1.12.7 或更高版本）和 Agent Skills。Olares CLI 已随 Olares 内置，运行在 Olares 设备上的智能体都自带；如果你在电脑上使用 AI 智能体，先安装 [Olares CLI 和 Agent Skills](/zh/developer/cli-overview.md)，用你的 Olares ID 登录后，本地的智能体都可以使用它们。

**你的项目**

- 一个已开发完成、源码可获取的网站项目。

**账号与权限**

- 以下镜像仓库之一：
  - Docker Hub 账号；
  - 拥有创建 `write:packages` 权限 personal access token 的 GitHub 账号；
- 一个你拥有的自定义域名，以及它的 DNS 管理控制台访问权限。

## 部署流程概览

大部分打包工作由智能体完成，整体流程分为以下阶段：

1. **源码**：智能体从 GitHub、Olares Files 或你的电脑读取项目。
2. **预览**：在智能体中运行网站，打开预览 URL 确认效果。
3. **构建并部署**：智能体生成生产构建、Dockerfile、容器镜像和 Olares chart，然后安装应用。
4. **绑定并分享**：智能体申请 RSA 证书并绑定自定义域名。你只需添加两条 DNS 记录，网站即可通过 HTTPS 访问。

## 步骤 1：准备项目源码

告诉智能体项目源码的位置。不同来源的处理方式不同。

### 源码在 GitHub

把仓库地址告诉智能体。

- 公开仓库：直接提供链接，让它克隆；
- 私有仓库：按提示先给智能体提供 GitHub personal access token 或 SSH key。

**示例：**

```text
Clone the repository from https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend
```

**结果：**

Lares 会把项目克隆到它在 Olares Files 中的默认 workspace `Data/lares/data/workspace/`，并在该目录下工作。之后它会简单介绍项目，并询问如何继续。例如：

```text
Cloned successfully to Portfolio-Landing-Page-7-React-Frontend in the workspace. Quick overview: ...

Want me to install dependencies and run it locally, or explore the source code?
```

### 源码在 Olares Files

在 Lares 中选择 Olares Files 里的项目目录作为 workspace。智能体会在这个 workspace 中执行构建、打包和部署命令。

### 源码在本地电脑

如果源码在你电脑上，需要先让它能被 Olares 上的 Lares 访问到。你可以自己完成，也可以让智能体来做：

- **推送到 GitHub**：把项目推送到仓库，然后把仓库地址告诉智能体。
- **上传到 Olares Files**：上传到任意目录，然后把路径告诉智能体。

## 步骤 2：预览网站

项目就位后，Lares 会询问如何继续。告诉它运行网站。Lares 会在需要时安装项目依赖、启动开发服务器，并返回一个临时预览 URL。在浏览器中打开该 URL，确认网站运行正常。

**示例：**

```text
The app is up and running 🎉

Dev server — Vite v7.3.1:
Local: http://localhost:5173/
Verified: HTTP 200, page serves correctly
```

## 步骤 3：构建、部署并发布

确认预览效果后，告诉智能体把网站发布到你的自定义域名。

剩下的工作由智能体完成。它会检查你的环境（Olares 版本、节点架构、Docker 配置），构建生产版本，并直接在 Olares 设备上打包容器镜像，镜像架构与节点一致。

**示例：**

```text
The preview looks good. Publish it to `website.bellame.online`.
```

### 提供镜像仓库

智能体需要一个镜像仓库来存放镜像，它会询问你使用哪一个。

- **Docker Hub（推荐）**：把你的 Docker Hub 用户名告诉智能体。当它要求凭证时，按它的说明创建 access token 并粘贴。智能体登录后会推送镜像，并验证镜像可以匿名拉取——Olares 节点正是通过这种方式下载镜像。推送完成后，记得在 Docker Hub 中删除该 token，它只用于这一次推送。

- **GitHub Container Registry**：智能体可以把镜像推送到你 GitHub 账号下的 `ghcr.io`。当它要求凭证时，创建一个带 `write:packages` 权限的 personal access token 并粘贴。镜像仓库必须设为 public，Olares 节点才能匿名拉取。

镜像推送完成后，智能体会创建一个入口访问级别为 `public` 的 Olares chart（自定义域名要求如此），上传 chart 并安装应用。完成后，应用会出现在 Launchpad 中。

如果安装卡住，请参考[常见问题](#常见问题)。

## 步骤 4：绑定自定义域名

应用运行起来后，智能体开始绑定你的自定义域名。它会为域名申请 RSA 证书，并挂载到应用入口。你只需要在 DNS 服务商的控制台添加两条记录，具体的值由智能体提供。绑定过程中应用会重启一次，这是正常现象。

Olares 只接受 RSA 证书。某些工具默认生成的 ECDSA 证书会导致 BFL 服务异常，所以智能体一律申请 RSA 证书。

### 1. 添加 TXT 记录

为了证明你拥有该域名，智能体会让你在 DNS 控制台添加一条 TXT 记录。

**示例：**

- Type: `TXT`
- Name: `_acme-challenge.website`（必须包含完整的子域）
- Value: 智能体提供的具体值

智能体会持续查询公共 DNS，一旦查到这条 TXT 记录，就会自动签发 RSA 证书。这个过程可能需要几分钟。TXT 记录只是临时验证用的，证书签发完成后就可以删除。

### 2. 添加 CNAME 记录

添加一条 CNAME 记录，把你的域名指向 Olares。这样，访问你域名的人就会被导到你的 Olares 设备。

**示例：**

- Type: `CNAME`
- Name: `website`
- Value: `laresprime.olares.com`

### 3. 验证

智能体进行最后的检查：证书、域名绑定、CNAME 记录，以及端到端的 HTTPS 测试。全部通过后，它会告诉你网站已上线。

**示例：**

```text
✓ RSA certificate issued and valid
✓ Domain website.bellame.online bound to the app entrance
✓ CNAME record detected and active
✓ HTTPS check passed (HTTP 200)
The site is live at https://website.bellame.online 🎉
```

在浏览器中打开线上地址确认，然后把 URL 分享给需要访问的人。

:::tip 证书续期
Let's Encrypt 证书的有效期为 90 天。续期时，让智能体用新证书重新绑定域名即可。重新绑定过程中应用会短暂重启。
:::

## 步骤 5：更新网站（可选）

网站上线后，你可以持续改进它。

告诉智能体你想改什么。智能体会用新的镜像标签重新构建网站、使用新的 chart 版本，并重新上传和安装应用。你的域名、证书和 DNS 记录都保持不变。

## 常见问题

### 安装一直卡在 `Initializing`

常见原因：

- 镜像架构与 Olares 节点不一致；
- 镜像标签被复用，节点缓存了旧层；
- 重新上传 chart 时未更新版本号。

可以尝试卸载旧版本、删除旧版本号、使用新的镜像标签和 chart 版本后重新安装：

```bash
olares-cli market uninstall <app-name>
olares-cli market delete --version <old-version>
olares-cli market upload <new-chart>
olares-cli market install <app-name> -s upload --watch
```

### 镜像架构与节点不一致

智能体自动识别节点架构，并直接在 Olares 设备上构建镜像，所以这种情况很少发生。如果错误架构的镜像被安装（例如在 AMD64 节点上装了 `linux/arm64` 的镜像），容器会报 `exec format error` 崩溃。让智能体按节点架构重新构建镜像并重新安装应用即可。

### nginx 权限错误

如果容器以非 root 用户运行，nginx 可能无法写入 PID 文件。让 Dockerfile 修改 PID 路径，并把相关目录的所有权交给 nginx 用户：

```dockerfile
RUN sed -i 's|/run/nginx.pid|/tmp/nginx.pid|' /etc/nginx/nginx.conf \
    && chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx /etc/nginx/conf.d
USER nginx
```

### 上传证书后 BFL 服务无法访问

通常是证书类型为 ECDSA 而不是 RSA。用 `--key-type rsa` 重新申请证书并上传。

### DNS TXT 记录验证失败

在 DNS 服务商添加 TXT 记录时，记录名称必须包含完整的子域。例如域名是 `n1.monster`，证书申请的是 `portfolio.n1.monster`，则 TXT 名称应填 `_acme-challenge.portfolio`，而不是只填 `_acme-challenge`。

如果记录正确但验证仍失败，可能是公共解析器缓存了上一次尝试的旧 TXT 值。等待缓存过期，或用你域名的权威名称服务器核对记录值。
