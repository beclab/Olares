---
outline: [2, 3]
description: Learn how to use Studio to set up a dev container, access it via VS Code, and configure port forwarding.
---
# Develop in a dev container
Olares Studio allows you to spin up a pre-configured dev container to write and debug code (such as Node.js scripts or CUDA programs) without managing local infrastructure. This provides an isolated environment identical to the production runtime.

The following guide shows the development and setup workflow using a Node.js project as an example.

## Prerequisite
- Olares version 1.12.2 or later.

## 1. Initialize the container
To start coding, you must provision the container resources and select your runtime environment.
1. Open Studio and select **Create a new application**.
2. Enter an **App name**, for example: `My Web`, and click **Confirm**.
3. Select **Coding on Olares** as the creation method.
   ![Coding on Olares](/images/manual/olares/studio-coding-on-olares.png#bordered)

4. Configure the **Dev Env**:

   a. From the drop-down list, select `beclab/node20-ts-dev:1.0.0`. 

   b. Allocate the resources for the container. For example:
     - **CPU**: `2 core`
     - **Memory**: `4 Gi`
     - **Volume Size**: `500 Mi`
5. In the **Expose Ports** field, enter the ports you intend to use for debugging (e.g., `8080`).
   :::tip Expose multiple ports
   Port `80` is exposed by default. Separate multiple additional ports with commas.
   :::
   ![Configure Dev Env](/images/manual/olares/studio-configure-dev-env.png#bordered)

6. Click **Create**. Wait for the status in the bottom-left corner to change to `Running`.

## 2. Access the workspace
You can access your dev container via the browser or your local IDE.

### Option A: Browser-based VS Code
Click **code-server** in Studio. This launches a fully functional VS Code instance inside your browser.

![Open VS Code in browser](/images/manual/olares/studio-open-vs-code-in-browser.png#bordered)
### Option B: Local VS Code (Remote Tunnel)
If you prefer your local settings and extensions, you can tunnel into the container.
1. Click **code-server** in Studio to open the browser-based VS code.
2. Click <span class="material-symbols-outlined">menu</span> in the top left, and select **Terminal** > **New Terminal** to open the terminal.
3. Install the VS Code Tunnel CLI:
   ```bash
   curl -SsfL https://vscode.download.prss.microsoft.com/dbazure/download/stable/17baf841131aa23349f217ca7c570c76ee87b957/vscode_cli_alpine_x64_cli.tar.gz | tar zxv -C /usr/local/bin
   ```
4. Create a secure tunnel:
   ```bash
   code tunnel
   ```
5. Follow the terminal prompts to authenticate using a Microsoft or GitHub account via the provided URL.
6. Assign a name to the tunnel when prompted (e.g., `myapp-demo`). This will output a `vscode.dev` URL tied to this remote workspace.
   ![Create a secure tunnel](/images/manual/olares/studio-create-a-secure-tunnel.png#bordered)

7. Open VS Code on your local machine, click the **><** icon in the bottom-left, and select **Tunnel**.
   ![Open remote window](/images/manual/olares/studio-open-remote-window.png#bordered){width=30%}
   ![Connect remote tunnel](/images/manual/olares/studio-connect-remote-tunnel.png#bordered)

8. Log in with the same account used in the previous step.
9. Select the tunnel name you defined (e.g., `myapp-demo`). It may take a few minutes for VS Code to establish the connection. Once successful, the remote indicator in the bottom-left will display your tunnel name.
   ![Select tunnel name](/images/manual/olares/studio-select-tunnel-name.png#bordered)
   ![Remote tunnel connected](/images/manual/olares/studio-remote-tunnel-connected.png#bordered){width=30%}

Once connected, you have full remote access to the container's file system and terminal, mirroring a local development experience.
## 3. Write and run code
Once inside the workspace, either via browser or local tunnel, the workflow mirrors standard local development.
You can populate your workspace by:
- Uploading files
- Cloning a Git repository, or
- Creating files manually

This example demonstrates creating a basic web page manually.

1. Open the **Explorer** sidebar and navigate to `/root/`.
   :::info
   Studio persists project data at `Data/studio/<app_name>/`.
   :::

   ![Open root directory](/images/manual/olares/studio-open-root-directory.png#bordered)
2. Click <span class="material-symbols-outlined">menu</span> in the top left, and select **Terminal** > **New Terminal** to open the terminal.
3. Run the following command to initialize the project:
   ```bash
   npm init -y
   ```
4. Install the Express framework:
   ```bash
   npm install express --save
   ```
5. Create a file named `index.js` in `/root/` with the following content:
   ```js
   // Ensure the port matches what you defined
   const express = require('express');
   const app = express();
   app.use(express.static('public/'));
   app.listen(8080, function() {
       console.log('Server is running on port 8080');
   });
   ```
6. Create a `public` directory in `/root/` and add an `index.html` file:
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
   
7. Start the server:
   ```bash
   node index.js
   ```
8. Open the **Ports** tab in VS Code and click the forwarded address to view the result.
   ![View web page](/images/manual/olares/studio-view-web-page.png#bordered)

## 4. Configure port forwarding
If you need to expose additional ports after the container is created (e.g., adding port `8081`), you must manually edit the container configuration manifests.
:::tip
You can follow the same steps to modify `OlaresManifest.yaml` and `deployment.yaml` to change the port number.
:::
### Modify configuration manifests
1. In Studio, click **<span class="material-symbols-outlined">box_edit</span>Edit** in the top-right to open the editor.
2. Edit `OlaresManifest.yaml`.

   a. Append the new port to the `entrances` list:
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
   # Add the following
   - authLevel: private
     host: myweb # Must match Service metadata name
     icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
     invisible: true
     name: myweb-dev-8081 # Unique identifier
     openMethod: ""
     port: 8081 # The new port number
     skip: true
     title: myweb-dev-8081
     ```
   b. Click <span class="material-symbols-outlined">save</span> in the top-right to save changes.
3. Edit `deployment.yaml`.
   
   a. Add the port mapping to `default-thirdlevel-domains` under `Deployment` > `metadata`:
   ```yaml
     annotations:
       applications.app.bytetrade.io/default-thirdlevel-domains:
        '[{"appName":"myweb","entranceName":"myweb-dev-8080"},{"appName":"myweb","entranceName":"myweb-dev-8081"}]'
        # entranceName must match the name used in OlaresManifest.yaml
   ```
   b. Update the `studio-expose-ports` annotation under `spec` > `template` > `metadata`:
   ```yaml
    template:
      metadata:
        annotations:
          applications.app.bytetrade.io/studio-expose-ports: "8080,8081"
   ```

   c. Add the port definition under `Service` > `spec` > `ports`:
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
       # Add the following
     - name: myweb-dev-8081 # Must match entrance name
       port: 8081
       targetPort: 8081
     selector:
       io.kompose.service: myweb
     ```
   
   d. Click <span class="material-symbols-outlined">save</span> in the top-right to save changes.

4. Click **Apply** to redeploy the container.

Once deployed, go to **Services** > **Ports**. You can see your new port listed here.
![Verify active ports](/images/manual/olares/studio-verify-active-ports.png#bordered)

### Test the connection
1. Update `index.js` to listen on the new port:
   ```js
   const express = require('express');
   const app = express();
   app.use(express.static('public/'));
   app.listen(8080, function() {
       console.log('Server is running on port 8080');
   });
   // Add the following
   const app_new = express();
   app_new.use(express.static('new/'));
   app_new.listen(8081, function() {
       console.log('Server is running on port 8081');
   });
   ```
2. Create a `new` directory in `/root/` and add an `index.html` file:
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
3. Restart the server:
   ```bash
   node index.js
   ```
4. Check the **Ports** tab to confirm port `8081` is active and accessible.
   ![View added port](/images/manual/olares/studio-view-added-port.png#bordered)

5. Click the forwarded address to view the result.
   ![Verify added web page](/images/manual/olares/studio-verify-added-web-page.png#bordered)