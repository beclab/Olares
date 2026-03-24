---
outline: [2, 3]
description: 
---

# Configure environment variables in Control Hub

Control Hub allows you to modify application environment variables to support debugging, feature updates, or temporary adjustments.


## Before you begin

- Before making changes, go to **Settings** > **Advanced** > **System environment variables**. Global variables (like API keys or mail server details) configured there persist after application updates. For more information, see [Set system environment variables](../settings/developer.md#set-system-environment-variables).
- Changes made directly in Control Hub are overwritten and lost when an application is upgraded.

## 1. Locate environment variables

Identify which Kubernetes resource contains the variable you want to change. Variables typically reside in one of following locations:
- **Deployments**: Within the workload YAML.
- **Secrets**: For sensitive data like passwords or tokens.
- **Configmaps**: For general configuration data or files.

### Find the target resource

1. Open Control Hub and click **Browse** from the left sidebar.
2. Select the target namespace in the first column.
3. View its associated resources grouped under the categories like **Deployments**, **Secrets**, and **Configmaps**.
4. Select the specific resource instance in **Deployments** to open its details page.
5. Select **Edit YAML** next to the resource name and locate the `spec > containers` section:
    - **`env`**: If the variable appears as a name/value pair, modify it directly here.
    - **`envFrom`**: If the variable refers to a `configMapRef` or `secretRef`, the variable is actually stored in the referenced Configmap or Secret. You must locate that referenced resource to make modifications.

    Example: In the following YAML, `envFrom` references the ConfigMap `lobechat-config`, and `env` directly defines `PGID` (group ID), `PUID` (user ID), and `TZ` (time zone).

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
6. (Optional) If the variable comes from a referenced Configmap or Secret, locate that resource in the corresponding resource group, and edit it following the same steps above.

### Find variables for apps in C/S architecture

Some applications adopt a Client/Server (C/S) architecture, such as Ollama. Their environment variables are distributed across two different namespace categories, each serving a distinct configuration role:

| Namespace category | Role | Common content | Function |
|:---|:---|:---|:---|
| User | Client/Proxy | Nginx config, Envoy sidebar config | Controls external access, routing, and load balancing |
| System | Server/Application core | Application parameters (e.g., model settings, concurrency, debug switches) | Controls core behavior |

The steps to locate variables are the same as above, but you operate in two different namespaces:
- To modify external access behavior, go to the **User** namespace, and look for Configmaps typically containing `nginx` or `sidecar` in the name.
- To modify core application parameters, go to the **System** namespace, and look for Configmaps typically containing `env`, `config`, or similar identifiers.

## 2. Modify environment variables

Use the following methods based on where the variables are stored. This section provides some examples to show how to configure 

### Modify variables in a Deployment

Use this for direct, temporary adjustments to a workload, such as changing log levels or time zones.

The following steps use the example of modifying time zone for Jellyfin so that the media library can display timestamps in the specified time zone.

1. Go to **Browse** > **demo0002** > **jellyfin-demo0002** > **Deployments** > **jellyfin**, and then click <span class="material-symbols-outlined">edit_square</span>.

    ![Browse to Jellyfin's deployment](/images/manual/olares/jellyfin-env-var.png#bordered)

2. In the YAML editor, find the `containers` section, locate the `env` field for jellyfin, and then change the value of the time zone variable `TZ`:

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

### Modify application configurations (Configmap)

Configmaps store environment variables, startup parameters, and configuration files. Use this to add third-party API keys or modify service addresses.

#### Update a standard application

The following steps uses the example of adding a Tavily API key to DeerFlow configuration to enable web search.

1. In Control Hub, go to **Browse** > **laresprime** > **deerflow-laresprime**.
2. From the resource list, expand **Configmaps**, and then click `deerflow-config`.
    ![Browse to DeerFlow's configmaps](/images/manual/use-cases/deerflow-configmap.png#bordered)
3. On the resource details page, click <span class="material-symbols-outlined">edit_square</span> in the top-right to open the YAML editor.
4. Add the following key-value pairs under the `data` section:
   ```yaml
   SEARCH_API: tavily
   TAVILY_API_KEY: tvly-xxx # Your Tavily API Key
   ```
   ![Configure Tavily](/images/manual/use-cases/deerflow-configure-tavily.png#bordered)
5. Click **Confirm** to save the changes.
6. Restart the service to apply the search configurations.

    a. Return to **Deployments** in the resource list, locate **deerflow** and click **Restart**.

   ![Restart DeerFlow](/images/manual/use-cases/deerflow-restart.png#bordered)

    b. In the confirmation dialog, type `deerflow` and click **Confirm**.

    c. Wait for the status icon to turn green, which indicates the service has successfully restarted and the new configuration has been loaded.
7. (Optional) If you are not sure which container to restart for applying the changes, stop and then resume the deerflow app via Market or Settings. This will also apply the changes.

#### Update a C/S application

The following steps use the example of modifying Ollama's timeout and model concurrency.

1. Modify proxy in User namespace:

    a. Go to **Browse > laresprime > ollamav2-laresprime** > **Deployments** > **ollamav2**, and then click <span class="material-symbols-outlined">edit_square</span>.

    b. In the YAML editor, find the `containers` section, locate and check the `env` field. The configuration here references `nginx.conf`.

    c. Click **Cancel** to close the editor.

    d. Return to the resource group **ConfigMaps**, click the `nginx-config` instance, and then click <span class="material-symbols-outlined">edit_square</span> in the upper‑right corner. 

    e. In the YAML editor, find the `data` section, locate the `nginx-config` field, and then modify the `proxy_read_timeout` value from `300s` to `600s`.
 
    f. Click **Confirm**.

2. Modify core in System namespace: 

    a. Go to **Browse** > **System** > **ollamaserver-shared** > **Deployments** > **ollama**, and then lick <span class="material-symbols-outlined">edit_square</span>.

    b. In the YAML editor, find the `containers` section, locate the `envFrom` field. The configuration here references `ollama-env`.

    c. Click **Cancel** to close the editor.

    d. Return to the resource group **ConfigMaps**, click the `ollama-env` instance, and then click <span class="material-symbols-outlined">edit_square</span> in the upper‑right corner.
    
    e. In the YAML editor, find the `data` section, and then change the value of `OLLAMA_MAX_LOADED_MODELS` from `3` to `5`.
 
    f. Click **Confirm**.

3. Apply changes using one of the following methods:

    - Go to **Deployments** and **Restart** the workload. 
    - If you are not sure which workload to restart, stop the application and then resume it via the **Market** or **Settings**.

    :::info
    Because `ollama-env` is referenced via environment variables (`envFrom`), you must restart the associated pod for the changes to take effect.
    ::: 

### Modify sensitive data

Secrets store encrypted data like tokens and passwords.

1. Go to **Browse** and select the target **Namespace**.
2. Expand **Secrets** and select the target secret.
3. Select **Edit** and update the value under the `data` field. Values in Secrets must be **Base64 encoded**.
4. Select **Confirm**.
5. Restart the application via the **Market** or **Settings** to apply the new credentials.
