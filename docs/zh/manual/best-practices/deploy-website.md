---
outline: [2, 3]
description: 通过 OpenCode 等 AI 编程助手把已开发好的网站部署到 Olares，本教程以 Portfolio Landing Page 为例。
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, Claude Code, 部署网站, 自定义域名, 镜像仓库, 预览, portfolio
---

# 通过 AI Agent 部署网站到 Olares

Olares 支持直接通过 AI 编程助手把已开发好的网站部署到设备上。你把项目告诉 Agent，它会帮你打包成 Olares 应用、推送镜像、安装到 Olares 设备上，并绑定自定义域名，其他人就能通过 HTTPS 访问这个网站。

本教程以 **OpenCode** 和 **Portfolio Landing Page** 项目作为运行示例。你可以用同样的步骤在 Claude Code 中操作，或替换为自己的网站项目。

本教程只讲**部署和分享**，不教如何开发网站。开始前，你需要已经有一个能在本地跑起来的网站项目。

## 学习目标

完成本教程后，你将学会：

- 准备项目源码；
- 部署前通过临时预览链接查看网站效果；
- 把网站构建并部署到 Olares；
- 配置自定义域名并上传 RSA 证书。

## 前提条件

开始前，请确认已有：

- 一台资源足够运行该网站的 Olares 设备；
- 已安装 OpenCode 应用；
- 已安装并登录 Olares CLI 和 Agent Skills；
- 一个已经能运行的网站项目源码；
- 以下镜像仓库之一：
  - Docker Hub 账号；
  - 本地 Olares 镜像仓库访问权限；
- 如果要公开分享，还需要一个自定义域名和 RSA SSL 证书。

:::warning 本教程假设网站已可运行
下文从一个已经能在本地构建运行的项目开始。如果你还没有网站项目，请先完成开发工作。
:::

## 部署流程概览

具体打包工作大部分由 Agent 完成，但整体流程分为以下阶段：

1. **准备源码** —— Agent 读取本地、Olares Files 或 GitHub 上的项目。
2. **预览验证** —— 在 OpenCode 内启动网站，通过预览 URL 确认效果。
3. **构建并部署** —— Agent 生成生产构建、`Dockerfile`、容器镜像和 Olares chart，然后安装应用。
4. **分享** —— 绑定自定义域名、上传 RSA 证书，并把入口设为 public。

## 步骤 1：准备项目源码

告诉 Agent 项目在哪。不同来源的处理方式如下。

:::info 示例项目
本教程以 Portfolio Landing Page 仓库 `https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend` 为例。实际操作时，把 URL 和目录名替换为你自己的项目。
:::

### 源码在本地电脑

如果源码在自己电脑上，需要让 Olares 上的 OpenCode 能访问到它。你可以自己完成，也可以让 Agent 帮你做：

- **直接使用**：如果你用的是能读取本地文件的本地 OpenCode CLI；
- **推送到 GitHub**：自己推送到 GitHub，或让 Agent 创建仓库并推送。然后把仓库地址告诉 Agent；
- **上传到 Olares Files**：自己上传到 `Home/Code` 目录，或让 Agent 复制过去。然后在 OpenCode 里以该目录创建 workspace。

示例：

> “把当前项目上传到 Olares Files 的 Home/Code 目录，并在 OpenCode 里作为 workspace 打开。”

### 源码已在 Olares Files

OpenCode 可以直接把项目目录作为 workspace 打开。Agent 会在这个 workspace 里执行构建、打包和部署命令。

### 源码在 GitHub

把仓库地址告诉 Agent。公开仓库直接贴链接让它克隆；私有仓库需要先给 Agent 提供 GitHub personal access token 或 SSH key。

示例：

> “克隆 Portfolio Landing Page 仓库 `https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend`，并作为 workspace 打开。”

Agent 会克隆项目并在该目录下继续工作。

## 步骤 2：在 OpenCode 中预览网站

项目准备好后，向 Agent 请求预览。

如果 Agent 主动询问是否要安装依赖并启动开发服务器，直接同意即可。如果没有主动询问，直接告诉它：

> “安装依赖、启动开发服务器，并把预览链接给我。”

Agent 会执行对应命令，然后返回一个临时预览 URL，例如：

```text
https://<host>/__preview/<listen-port>/
```

或

```text
<appid>-<listen-port>.<userzone>
```

例如：

```text
https://1f47cd9b0.laresprime.olares.com/__preview/5173/#home
```

在浏览器中打开预览 URL，检查网站是否正常。

:::tip 预览仅用于开发阶段
预览 URL 是临时的，只能验证网站能否运行。要让网站长期可用并对外分享，需要继续执行后面的部署步骤。
:::

## 步骤 3：构建并部署

确认预览效果后，让 Agent 直接部署：

> “预览没问题，部署到 Olares。”

Agent 会完成剩余工作：生产构建、`Dockerfile`、镜像构建与推送、Olares chart、安装。安装完成后，应用会出现在 Launchpad 中。

如果安装卡住，请参考下方的[常见问题](#常见问题)。

### 选择镜像仓库

Agent 需要一个镜像仓库来存放镜像。如果你不明确告诉它用哪个仓库，它可能会直接选一个默认值而不询问。建议在构建前就声明偏好。

- **Docker Hub**：告诉 Agent 你的 Docker Hub 用户名，例如：

  > “把镜像推送到 Docker Hub 我的用户名 `myusername` 下面。”

  如果系统里 `~/.docker/config.json` 已经配置了 Docker 凭证，Agent 可能会默认使用 Docker Hub 和该用户名，而不会再次询问。

- **Olares 本地镜像仓库**：如果在 Olares 系统内操作，可以让 Agent 使用本地仓库：

  > “使用 Olares 本地镜像仓库，不要推送到 Docker Hub。”

  Agent 通常可以直接推送到 `mirrors.olares.com`，无需额外凭证。

推送完成后，确认镜像最终存到了哪里：

> “镜像推送到哪里了？把完整的镜像名和仓库告诉我。”

答案会显示在 Agent 的输出里，也会在 chart 的 image 字段中体现。

### 避免架构不匹配

:::warning 注意 CPU 架构
如果在 Apple Silicon Mac（arm64）上构建镜像，而 Olares 设备是 AMD64，容器启动会报 `exec format error`。
:::

确保镜像架构与 Olares 节点一致。例如：

> “我的 Olares 节点是 AMD64，镜像请按 `linux/amd64` 构建。”

Agent 会执行类似下面的命令：

```bash
docker buildx build --platform linux/amd64 -t <image>:<tag> .
```

如果希望镜像同时支持 ARM 和 AMD64 设备，可以让 Agent 构建多架构镜像：

> “构建一个同时支持 ARM64 和 AMD64 的多架构镜像。”

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t <image>:<tag> --push .
```

:::tip 每次重新构建都要换一个镜像标签
修复问题后重新构建镜像时，请使用新 tag，例如 `0.1.0`、`0.1.1`、`0.1.2`。如果复用 tag，Olares 节点可能缓存旧层，继续运行有问题的镜像。
:::

:::tip 每次上传都要更新 chart 版本号
修改 chart 后重新上传时，请提升 `Chart.yaml` 和 `OlaresManifest.yaml` 中的版本号。Olares 可能对重复版本号的 chart 拒绝上传或表现异常。
:::

## 步骤 4：配置自定义域名

告诉 Agent 你想使用的真实子域名：

> “把这个应用的入口域名设为 `website.bellame.online`。”

:::warning 使用你真实拥有的域名
请把 `yourdomain.com` 替换为你实际拥有的域名。
:::

Agent 会先把入口访问级别设为 public，然后向你索要 TLS 证书和私钥。如果你还没有证书，它会给出 certbot 命令并说明必须使用 RSA：

```text
First, I need to set the entrance auth level to public (required for custom domains):
Auth level set to public.

To register website.bellame.online, I need the TLS certificate and private key for this domain. Do you have these files available? I need:

Certificate file (full chain PEM, e.g., cert.pem)
Private key file (RSA PEM, e.g., key.pem)
If you don't have them yet, you can generate one with certbot:

certbot certonly -d website.bellame.online --key-type rsa
Note: RSA key type is required (certbot defaults to ECDSA, which won't work).
```

按 Agent 的提示继续。接下来需要：生成证书、添加 CNAME 记录、把证书复制到 Agent 或 `olares-cli` 能读取的位置，然后上传。

### 申请 RSA 证书

Olares 要求使用 RSA 证书。certbot 默认可能会申请 ECDSA 证书，配置后会导致 BFL 服务异常。

你可以在本地终端自己运行 certbot：

```bash
sudo certbot certonly --manual --preferred-challenges dns --key-type rsa -d <your-domain>
```

按提示操作，certbot 要求添加 DNS TXT 记录时，在 DNS 服务商处添加对应记录。

### 添加 CNAME 记录

在 DNS 服务商处添加一条 CNAME 记录，把域名指向 Olares：

| 类型 | 名称 | 值 |
| :--- | :--- | :--- |
| CNAME | `<your-subdomain>` | `laresprime.olares.com` |

### 上传证书

在本地电脑上，certbot 把证书保存在 `/etc/letsencrypt/live/<your-domain>/`。该目录需要 root 权限，`olares-cli` 无法直接读取。

1. 先把证书复制到当前用户能读取的位置：

   ```bash
   sudo cp /etc/letsencrypt/live/<your-domain>/fullchain.pem ~/cert.pem
   sudo cp /etc/letsencrypt/live/<your-domain>/privkey.pem ~/key.pem
   sudo chown $(whoami) ~/cert.pem ~/key.pem
   ```

2. 确保你已经在本地安装并登录了 `olares-cli`（v1.12.6+）。

3. 在本地运行域名绑定命令：

   ```bash
   olares-cli settings apps domain set <app-name> <entrance-name> \
     --third-party <your-domain> \
     --cert-file ~/cert.pem \
     --key-file ~/key.pem
   ```

   例如：

   ```bash
   olares-cli settings apps domain set portfolio7 portfolio7 \
     --third-party website.bellame.online \
     --cert-file ~/cert.pem \
     --key-file ~/key.pem
   ```

### 验证入口

让 Agent 验证域名是否可以通过 HTTPS 访问：

> “自定义域名已经绑定，请验证 `website.bellame.online` 是否能通过 HTTPS 访问。”

你也可以让 Agent 列出所有入口确认 URL：

> “列出这个应用的所有入口，告诉我当前的 URL。”

Agent 设置好访问级别后会自动验证域名，你看到的输出类似：

```text
set auth level for portfolio7/portfolio7 to "public"
Auth level is set to public.
https://website.bellame.online is now returning HTTP 200 and serving your portfolio site directly.
The portfolio site is live and accessible at https://website.bellame.online.
```

:::warning public 入口仍需登录 Olares
设为 public 后，任何能登录你的 Olares 设备的用户都能访问该网站。未登录的访问者会先被重定向到 Olares 的认证页面。
:::

## 步骤 5：分享网站

在浏览器中打开自定义域名，确认网站能正常加载，然后把 URL 分享给需要访问的人。

例如：

```text
https://website.bellame.online
```

## 常见问题

### 安装一直卡在 `initializing`

常见原因：

- 镜像架构与 Olares 节点不一致；
- 镜像 tag 被复用，节点缓存了旧层；
- 重新上传 chart 时未更新版本号。

可以尝试卸载旧版本、提升版本号后重新上传安装：

```bash
olares-cli market uninstall <app-name>
olares-cli market delete --version <old-version>
olares-cli market upload <new-chart>
olares-cli market install <app-name> -s upload --watch
```

### nginx 权限错误

如果容器以非 root 用户运行，nginx 可能无法写入 PID 文件。让 Dockerfile 修改 PID 路径，并把相关目录权限交给 nginx 用户：

```dockerfile
RUN sed -i 's|/run/nginx.pid|/tmp/nginx.pid|' /etc/nginx/nginx.conf \
    && chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx /etc/nginx/conf.d
USER nginx
```

### 上传证书后 BFL 服务无法访问

通常是证书类型为 ECDSA 而不是 RSA。用 `--key-type rsa` 重新申请证书并上传。

### DNS TXT 记录验证失败

在 DNS 服务商添加 TXT 记录时，记录名称必须包含完整子域。例如域名是 `n1.monster`，证书申请的是 `portfolio.n1.monster`，则 TXT 名称应填 `_acme-challenge.portfolio`，而不是只填 `_acme-challenge`。

### 恢复应用后 auth level 变回 private

修改访问级别后如果恢复应用，chart 中定义的 `authLevel` 可能会覆盖手动设置。需要修改 chart 源文件中的 `OlaresManifest.yaml`，把 `authLevel` 设为 `public`，提升版本号后重新部署。

### Agent 无法读取 certbot 证书

certbot 把证书存在 `/etc/letsencrypt`，只有 root 能访问。先把证书复制到项目目录并修改所有者，再让 Agent 读取。
