---
outline: [2, 4]
description: Learn how to locate and modify application environment variables in Control Hub for debugging, updates, or configuration changes.
---

# Configure environment variables

Control Hub allows you to modify application environment variables to support debugging, feature updates, or temporary adjustments.

## Before you begin

- **Global vs. local variables**: Before making changes here, check **Settings** > **Advanced** > **System environment variables**. Global variables like API keys or mail server details configured there persist after application updates. For more information, see [Set system environment variables](../settings/developer.md#set-system-environment-variables).
- **Updates overwrite local changes**: Any changes you make directly in Control Hub are temporary. They will be overwritten and lost when the application is upgraded.

## 1. Locate environment variables

Identify which Kubernetes resource contains the variable you want to change. Variables typically reside in one of following locations:
- **Deployments**: Within the workload YAML for direct, simple variables
- **Secrets**: For sensitive data like database connections and admin accounts
- **Configmaps**: For general configuration data or files

### Locate variables in standard apps

1. Open Control Hub and click **Browse** from the left sidebar.
2. Select the target namespace in the first column.
3. View its associated resources, which are grouped under categories like **Deployments**, **Secrets**, and **Configmaps**.
4. Select the specific resource instance under **Deployments** to open its details page.
5. Click <span class="material-symbols-outlined">edit_square</span> next to the resource name.
6. In the YAML editor, locate the `spec` > `containers` section to see how variables and values are injected:
    - **`env`**: If the variable appears as a name/value pair, you can modify it directly here.
    - **`envFrom`**: If the variable refers to a `configMapRef`, the variable is actually stored in the referenced Configmap. You must locate that specific Configmap resource instance to make modifications instead.
    - **`valueFrom`**: If the variable refers to a `secretKeyRef`, the variable is actually stored in the referenced Secret. You must locate that specific Secret resource instance to make modifications instead.

    Example: In the following YAML, `envFrom` references the Configmap `lobechat-config`, while `env` directly defines `PGID` (group ID), `PUID` (user ID), and `TZ` (time zone).

    ```yaml
        spec:
        containers:
            - name: lobechat
            image: docker.io/beclab/lobehub-lobehub:2.1.18
            ports:
                - name: http
                containerPort: 3210
                protocol: TCP
            envFrom:
                - configMapRef:
                    name: lobechat-config
            env:
                - name: PGID
                value: '1000'
                - name: PUID
                value: '1000'
                - name: TZ
                value: Etc/UTC
    ```    
7. (Optional) If the variable comes from a referenced Configmap or Secret:

    a. Locate that resource in the corresponding resource group, and then edit it following the same steps above.

    b. Restart the deployment to apply the changes. 
    
    :::tip
    If you are not sure which container to restart, you can stop and then resume the app via the Market or Settings to apply the changes.
    :::

### Locate variables in C/S apps

Some applications use a Client/Server (C/S) architecture, such as Ollama. Their environment variables are distributed across two different namespaces:

| Namespace category | Role | Common content | Function |
|:---|:---|:---|:---|
| User | Client/Proxy | Nginx config, Envoy sidebar config | Controls external access, routing, and load balancing |
| System | Server/Application core | Application parameters such as model settings, concurrency, and debug switches | Controls core app behaviors |

The steps to locate these variables are the same, but you must look in the correct namespace depending on what you want to modify:
- To modify external access behavior, go to the User namespace, and then look for Configmaps typically containing `nginx` or `sidecar` in the name.
- To modify core application parameters, go to the System namespace, and then look for Configmaps typically containing `env`, `config`, or similar identifiers.

## 2. Modify environment variables

Depending on whether your target variables are stored in a Deployment, a Configmap, or a Secret, select the corresponding method below to apply your changes.

### Modify variables in a Deployment

Use this method for direct, temporary adjustments to a workload, such as changing log levels or time zones.

**Example scenario**

Change the time zone for Jellyfin so the media library displays local timestamps instead of UTC.

**Procedure**

1. Go to **Browse** > **demo0002** > **jellyfin-demo0002** > **Deployments** > **jellyfin**, and then click <span class="material-symbols-outlined">edit_square</span>.

    ![Browse to Jellyfin's deployment](/images/manual/olares/jellyfin-env-var.png#bordered)

2. In the YAML editor, find the `containers` section, locate the `env` field for jellyfin, and then change the value of `TZ`:

    ```yaml
    env:
    - name: PGID
        value: '1000'
    - name: PUID
        value: '1000'
    - name: UMASK
        value: '002'
    - name: TZ
        value: Asia/Shanghai   # changed from Etc/UTC
    ```
3. Click **Confirm**. The pod restarts automatically to apply the changes.

### Modify variables in a Configmap

Configmaps store environment variables, startup parameters, and configuration files. Use this method to add third-party API keys or modify service parameters.

#### Update a standard app

**Example scenario**

Add a Tavily API key to DeerFlow's configuration to enable web search.

**Procedure**

1. In Control Hub, go to **Browse** > **laresprime** > **deerflow-laresprime**.
2. From the resource list, expand **Configmaps**, and then click `deerflow-config`.
    ![Browse to DeerFlow's configmaps](/images/manual/use-cases/deerflow-configmap.png#bordered)
3. On the resource details page, click <span class="material-symbols-outlined">edit_square</span> in the top-right to open the YAML editor.
4. Add the following key/value pairs under the `data` section:
   ```yaml
   SEARCH_API: tavily
   TAVILY_API_KEY: tvly-xxx # Your Tavily API Key
   ```
   ![Configure Tavily](/images/manual/use-cases/deerflow-configure-tavily.png#bordered)
5. Click **Confirm** to save the changes.
6. Return to **Deployments** in the resource list, locate **deerflow**, and then click **Restart**.

   ![Restart DeerFlow](/images/manual/use-cases/deerflow-restart.png#bordered)

7. In the confirmation dialog, type `deerflow`, and then click **Confirm**.
8. Wait for the status icon to turn green, which indicates the new configuration has been loaded.

#### Update a C/S app

**Example scenario**

Modify Ollama's timeout settings (Client) and model concurrency (Server).

**Procedure**

1. Modify the proxy in User namespace:

    a. Go to **Browse > laresprime > ollamav2-laresprime** > **Deployments** > **ollamav2**, and then click <span class="material-symbols-outlined">edit_square</span>.

   ![Locate Ollama instance in Control Hub](/images/manual/use-cases/locate-ollama-instance.png#bordered)

    b. In the YAML editor, find the `containers` section, locate and check the `env` field. The configuration here references `nginx.conf`.

   ![Edit YAML for Ollama deployment](/images/manual/use-cases/edit-yaml-ollama.png#bordered)

    c. Click **Cancel** to close the editor.

    d. Return to the resource group **ConfigMaps**, click the `nginx-config` instance, and then click <span class="material-symbols-outlined">edit_square</span> in the upper‑right corner. 

   ![Locate Nginx config instance](/images/manual/use-cases/locate-nginx-config.png#bordered)
    
    e. In the YAML editor, find the `data` section, locate the `nginx-config` field, and then modify the `proxy_read_timeout` value from `300s` to `600s`.

   ![Edit Nginx configuration](/images/manual/use-cases/edit-nginx-conf.png#bordered)
 
    f. Click **Confirm**.

    g. Return to **Deployments** > **ollamav2**, and then click **Restart** to apply the change.

2. Modify core in System namespace: 

    a. Go to **Browse** > **System** > **ollamaserver-shared** > **Deployments** > **ollama**, and then click <span class="material-symbols-outlined">edit_square</span>.

   ![Locate Ollama in System namespace](/images/manual/use-cases/locate-ollama-sys-namespace.png#bordered)    

    b. In the YAML editor, find the `containers` section, locate the `envFrom` field. The configuration here references `ollama-env`.

   ![Edit YAML](/images/manual/use-cases/edit-yaml-envfrom.png#bordered)  
    
    c. Click **Cancel** to close the editor.

    d. Return to the resource group **ConfigMaps**, click the `ollama-env` instance, and then click <span class="material-symbols-outlined">edit_square</span> in the upper‑right corner.

   ![Edit Ollama env](/images/manual/use-cases/edit-ollama-env.png#bordered)
    
    e. In the YAML editor, find the `data` section, and then change the value of `OLLAMA_MAX_LOADED_MODELS` from `3` to `5`.

   ![Edit Ollama variable](/images/manual/use-cases/modify-var-ollama.png#bordered)
    
    f. Click **Confirm**.

    g. Return to **Deployments** > **ollama**, and then click **Restart** to apply the change.

### Modify variables in a Secret

Secrets store encrypted data like passwords and tokens.

Because they operate identically to Configmaps, you can modify them by following the same workflow as [standard applications](#update-a-standard-app).

:::info
When you open the YAML editor for a Secret, all values entered under the `data` field must be Base64 encoded.
:::
