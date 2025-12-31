---
outline: [2, 3]
description: 了解如何使用 Studio 设置开发容器，通过 VS Code 访问，并配置端口转发。
---
# 使用开发容器开发应用
Olares Studio 允许你启动预配置的开发容器来编写和调试代码（如 Node.js 脚本或 CUDA 程序），无需管理本地基础设施。它提供了一个与生产运行时一致的、完全隔离的环境。

本文档以 Node.js 项目为例介绍开发和配置流程。

## 前提条件
- Olares 1.12.2 及以上版本。

## 1. 初始化容器
开始编码前，需要配置容器资源并选择运行时环境。
1. 打开 Studio 并选择**创建新应用**。
2. 输入**应用名称**（例如 `My Web`），然后点击**确认**。
3. 选择**在 Olares 上写程序**作为创建方式。
   ![在 Olares 上写程序](/images/zh/manual/olares/studio-coding-on-olares.png#bordered)

4. 配置**开发环境**：

   a. 从下拉列表中选择 `beclab/node20-ts-dev:1.0.0`。

   b. 为容器分配资源，例如：
   - **CPU**：`2 core`
   - **内存**：`4 Gi`
   - **卷大小**：`500 Mi`
5. 在**暴露端口**字段中，输入用于调试的端口，例如 `8080`。
   :::tip 暴露多个端口
   端口 `80` 默认为暴露状态。如需暴露多个额外端口，请用逗号分隔。
   :::
   ![配置开发环境](/images/zh/manual/olares/studio-configure-dev-env.png#bordered)

6. 点击**创建**。等待左下角的状态变为`运行中`。

## 2. 访问工作区
你可以通过浏览器或本地 IDE 访问开发容器。

### 选项 A：基于浏览器的 VS Code
点击 Studio 中的**在线编辑器**，在浏览器中启动一个功能齐全的 VS Code 实例。

![在浏览器中打开 VS Code](/images/manual/olares/studio-open-vs-code-in-browser.png#bordered)
### 选项 B：本地 VS Code 远程隧道
如果更习惯使用本地设置和插件，可以通过隧道连接到容器。
1. 点击 Studio 中的**在线编辑器**以打开基于浏览器的 VS Code。
2. 点击左上角的 <span class="material-symbols-outlined">menu</span>，选择 **Terminal** > **New Terminal** 打开终端。
3. 安装 VS Code Tunnel CLI：
   ```bash
   curl -SsfL https://vscode.download.prss.microsoft.com/dbazure/download/stable/17baf841131aa23349f217ca7c570c76ee87b957/vscode_cli_alpine_x64_cli.tar.gz | tar zxv -C /usr/local/bin
   ```
4. 创建安全隧道：
   ```bash
   code tunnel
   ```
5. 按照终端提示，通过提供的 URL 使用 Microsoft 或 GitHub 帐户进行身份验证。
6. 出现提示时为隧道命名，例如 `myapp-demo`。终端将输出绑定到此远程工作区的 `vscode.dev` URL。
   ![创建安全隧道](/images/manual/olares/studio-create-a-secure-tunnel.png#bordered)

7. 在本地机器上打开 VS Code，点击左下角的 **><** 图标，选择 **Tunnel**。
   ![打开远程窗口](/images/manual/olares/studio-open-remote-window.png#bordered){width=30%}
   ![连接远程隧道](/images/manual/olares/studio-connect-remote-tunnel.png#bordered)

8. 使用上一步中的同一帐户登录。
9. 选择刚才定义的隧道名称 `myapp-demo`。VS Code 可能需要几分钟建立连接。连接成功后，左下角的远程指示器将显示隧道名称。
   ![选择隧道名称](/images/manual/olares/studio-select-tunnel-name.png#bordered)
   ![远程隧道已连接](/images/manual/olares/studio-remote-tunnel-connected.png#bordered){width=30%}

连接成功后，你将拥有对容器文件系统和终端的完全远程访问权限，体验与本地开发一致。
## 3. 编写和运行代码
进入工作区后，无论是通过浏览器还是本地隧道，工作流与标准本地开发无异。 
你可通过以下方式向工作区添加内容：
- 上传文件
- 克隆 Git 仓库
- 手动创建文件

本例演示如何手动创建一个基础网页。

1. 打开 Explorer 侧边栏并导航到 `/root/`。
   :::info
   Studio 将项目文件持久化在 `数据/studio/<app_name>/` 路径下。
   :::

   ![打开根目录](/images/manual/olares/studio-open-root-directory.png#bordered)
2. 在左上角点击 <span class="material-symbols-outlined">menu</span>，选择 **Terminal** > **New Terminal** 打开终端。
3. 运行以下命令初始化项目：
   ```bash
   npm init -y
   ```
4. 安装 Express 框架：
   ```bash
   npm install express --save
   ```
5. 在 `/root/` 中创建文件 `index.js`，内容如下：
   ```js
    // 确保端口与定义的一致
   const express = require('express');
   const app = express();
   app.use(express.static('public/'));
   app.listen(8080, function() {
       console.log('Server is running on port 8080');
   });
   ```
6. 在 `/root/` 中创建 `public` 目录并添加 `index.html` 文件：
   ```html
   <!DOCTYPE html>
    <html>  
        <head>
            <meta charset="UTF-8">
            <title>My Web Page</title>
        </head>
        <body>
            <h1>Hello World</h1>
            <h1>Hello Olares</h1>
        </body>
    </html>
   ```
   
7. 启动服务器：
   ```bash
   node index.js
   ```
8. 打开 VS Code 中的 **Ports** 标签页，点击转发地址查看结果。
   ![查看网页](/images/manual/olares/studio-view-web-page.png#bordered)

## 4. 配置端口转发
如果在创建容器后需要暴露更多端口，例如添加端口 `8081`，需要手动编辑容器配置清单。
:::tip
如果需要更改端口号，可参照相同步骤修改 `OlaresManifest.yaml` 和 `deployment.yaml` 文件。
:::
### 修改配置
1. 在 Studio 中，点击右上角的 **<span class="material-symbols-outlined">box_edit</span>编辑**打开编辑器。
2. 编辑 `OlaresManifest.yaml`。

   a. 将新端口追加到 `entrances` 列表：
   ```yaml
   entrances:
   - authLevel: private
     host: myweb
     icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
     invisible: true
     name: myweb-dev-8080
     openMethod: ""
     port: 8080
     skip: true
     title: myweb-dev-8080
   # 添加以下内容
   - authLevel: private
     host: myweb # 必须匹配 Service metadata name
     icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
     invisible: true
     name: myweb-dev-8081 # 唯一标识符
     openMethod: ""
     port: 8081 # 新端口号
     skip: true
     title: myweb-dev-8081
     ```
   b. 在右上角点击 <span class="material-symbols-outlined">save</span> 保存更改。
3. 编辑 `deployment.yaml`。

   a. 在 `Deployment` > `metadata` 下，将端口映射添加到 `default-thirdlevel-domains`：
   ```yaml
     annotations:
       applications.app.bytetrade.io/default-thirdlevel-domains:
        '[{"appName":"myweb","entranceName":"myweb-dev-8080"},{"appName":"myweb","entranceName":"myweb-dev-8081"}]'
        # # entranceName 必须匹配 OlaresManifest.yaml 中的名称
   ```
   b. 更新 `spec` > `template` > `metadata` 下的 `studio-expose-ports` 注解：
   ```yaml
    template:
      metadata:
        annotations:
          applications.app.bytetrade.io/studio-expose-ports: "8080,8081"
   ```

   c. 在 `Service` > `spec` > `ports` 下添加端口定义：
   ```yaml
   kind: Service
   spec:
     ports:
     - name: "80"
       port: 80
       targetPort: 80
     - name: myweb-dev-8080
       port: 8080
       targetPort: 8080
       # 添加以下内容
     - name: myweb-dev-8081 # 必须匹配 entrance name
       port: 8081
       targetPort: 8081
     selector:
       io.kompose.service: myweb
     ```
   
   d. 在右上角点击 <span class="material-symbols-outlined">save</span> 保存更改。

4. 点击**应用**重新部署容器。

部署成功后，你可以在**服务** > **端口**中看到列出的新端口。
![验证活动端口](/images/zh/manual/olares/studio-verify-active-ports.png#bordered)

### 测试连接
1. 更新 `index.js` 以监听新端口：
   ```js
   const express = require('express');
   const app = express();
   app.use(express.static('public/'));
   app.listen(8080, function() {
       console.log('Server is running on port 8080');
   });
   // 添加以下内容
   const app_new = express();
   app_new.use(express.static('new/'));
   app_new.listen(8081, function() {
       console.log('Server is running on port 8081');
   });
   ```
2. 在 `/root/` 中创建 `new` 目录并添加 `index.html` 文件：
   ```html
   <!DOCTYPE html>
    <html>  
        <head>
            <meta charset="UTF-8">
            <title>My Web Page</title>
        </head>
        <body>
            <h1>This is a new page</h1>
        </body>
    </html>
   ```
3. 重启服务器：
   ```bash
   node index.js
   ```
4. 检查 **Ports** 标签页确认端口 `8081` 处于活动状态且可访问。
   ![查看添加的端口](/images/manual/olares/studio-view-added-port.png#bordered)

5. 点击转发地址查看结果。
   ![查看添加的网页](/images/manual/olares/studio-verify-added-web-page.png#bordered)