---
outline: [2, 3]
description: Learn how to locate and modify application environment variables in Control Hub for debugging, updates, or configuration changes.
---

# Configure environment variables

Control Hub allows you to view and modify application environment variables for debugging, feature configuration, or temporary adjustments.

## Before you begin

Check where the target variable should be configured before making changes.

| Metric | System-level variables | App-specific variables |
|:---|:---|:---|
| Where to configure | Settings | Control Hub |
| Scope | Shared across all apps<br> that reference them | Specific to a single app resource |
| What it covers | Common variables pre-configured<br> by Olares, such as API keys and<br> mail server settings | App-specific parameters not covered by system-level variables |
| Persist after app upgrades | Yes | No |

Check [Settings](../settings/developer.md#set-system-environment-variables) first to see if the variable you need is already available there. If not, configure it directly in Control Hub using the steps below.

## Identify where a variable is stored

Variables in an app can be stored in different Kubernetes resource types. Knowing which resource holds your target variable determines how you modify it and whether a restart is needed.

| Resource type | Typical content | Restart after editing |
|:---|:---|:---|
| Deployment | Direct name/value pairs | Automatic |
| Configmap | Configuration data, startup<br> parameters, config files | Manual restart required |
| Secret | Sensitive data such as <br>passwords, tokens, credentials | Manual restart required |

### Standard apps

To determine where a variable is stored, open the Deployment in the YAML editor and check the `spec` > `containers` section. The injection method tells you the source:

- `env`: The variable is defined directly in the Deployment as a name/value pair.
- `envFrom` with `configMapRef`: The variable is stored in the referenced Configmap.
- `valueFrom` with `secretKeyRef`: The variable is stored in the referenced Secret.

For example, in the following YAML, `envFrom` references the Configmap `lobechat-config`, while `env` directly defines `PGID`, `PUID`, and `TZ`.

```yaml{9-18}
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

### C/S apps

Some applications use a Client/Server (C/S) architecture, such as Ollama. Their variables are distributed across two namespaces: the User namespace for client-side resources and the System namespace for server-side resources.

For C/S apps, environment variables are typically stored in Configmaps across both namespaces. Navigate to the correct namespace based on what you want to change:

- To modify external access behavior, go to the User namespace and look for Configmaps containing `nginx` or `sidecar` in the name.
- To modify core application parameters, go to the System namespace and look for Configmaps containing `env`, `config`, or similar identifiers.

## Modify variables in a Deployment

Use this method for direct adjustments to a workload. The following example changes the time zone for Jellyfin so the media library displays local timestamps instead of UTC.

1. In Control Hub, select the Jellyfin project from the Browse panel.

2. Under **Deployments**, click **jellyfin**, then click <i class="material-symbols-outlined">edit_square</i>.

    ![Browse to Jellyfin's deployment](/images/manual/olares/jellyfin-env-var.png#bordered)

3. In the YAML editor, find the `containers` section, locate the `env` field for jellyfin, and then change the value of `TZ`:

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
4. Click **Confirm**. The pod restarts automatically to apply the changes.

## Modify variables in a Configmap

Use this method to add third-party API keys, modify startup parameters, or update configuration files.

### Standard apps {#modify-standard-apps}

The following example adds a Tavily API key to DeerFlow's configuration to enable web search.

1. In Control Hub, select the DeerFlow project from the Browse panel.
2. Under **Configmaps**, click `deerflow-config`.
    ![Browse to DeerFlow's configmaps](/images/manual/use-cases/deerflow-configmap.png#bordered)

3. On the resource details page, click <i class="material-symbols-outlined">edit_square</i> in the top-right to open the YAML editor.
4. Add the following key/value pairs under the `data` section:
   ```yaml
   SEARCH_API: tavily
   TAVILY_API_KEY: tvly-xxx # Your Tavily API Key
   ```
   ![Configure Tavily](/images/manual/use-cases/deerflow-configure-tavily.png#bordered)
5. Click **Confirm** to save the changes.
6. Return to **Deployments** > **deerflow**, and then click **Restart**.

   ![Restart DeerFlow](/images/manual/use-cases/deerflow-restart.png#bordered)

7. In the confirmation dialog, type `deerflow`, and then click **Confirm**.

Wait for the status icon to turn green, which indicates the new configuration has been loaded.

### C/S apps {#modify-cs-apps}

Depending on what you want to change, you may need to modify variables in the User namespace, the System namespace, or both.

#### Modify client settings in the User namespace

The following steps change Ollama's proxy read timeout from `300s` to `600s`.

1. In Control Hub, select the Ollama project from the Browse panel.

2. Under **Deployments**, click **ollamav2**, then click <i class="material-symbols-outlined">edit_square</i>.
   ![Locate Ollama instance in Control Hub](/images/manual/use-cases/locate-ollama-instance.png#bordered)

3. In the YAML editor, find the `containers` section, locate and check the `env` field. The configuration here references `nginx.conf`.

   ![Edit YAML for Ollama deployment](/images/manual/use-cases/edit-yaml-ollama.png#bordered)

4. Click **Cancel** to close the editor.

5. Click **Configmaps** to expand the resource group, click `nginx-config`, and then click <i class="material-symbols-outlined">edit_square</i> in the upper-right corner.

   ![Locate Nginx config instance](/images/manual/use-cases/locate-nginx-config.png#bordered)

6. In the YAML editor, find the `data` section, locate the `nginx.conf` key, and then modify the `proxy_read_timeout` value from `300s` to `600s`.

   ![Edit Nginx configuration](/images/manual/use-cases/edit-nginx-conf.png#bordered)

7. Click **Confirm**.

8. Return to **Deployments** > **ollamav2**, and then click **Restart** to apply the change.

Wait for the status icon to turn green, which indicates the new configuration has been loaded.

#### Modify server settings in the System namespace

The following steps change Ollama's maximum loaded models from `3` to `5`.

1. In Control Hub, scroll down in the Browse panel and click **System** to expand the system section.

2. Select **ollamaserver-shared**, and under **Deployments**, click **ollama**, and then click <i class="material-symbols-outlined">edit_square</i>.

   ![Locate Ollama in System namespace](/images/manual/use-cases/locate-ollama-sys-namespace.png#bordered)

3. In the YAML editor, find the `containers` section, and check the `envFrom` field. The configuration here references `ollama-env`.

   ![Edit YAML](/images/manual/use-cases/edit-yaml-envfrom.png#bordered)

4. Click **Cancel** to close the editor.

5. Return to the resource group **Configmaps**, click the `ollama-env` instance, and then click <i class="material-symbols-outlined">edit_square</i> in the upper-right corner.

   ![Edit Ollama env](/images/manual/use-cases/edit-ollama-env.png#bordered)

6. In the YAML editor, find the `data` section, and then change the value of `OLLAMA_MAX_LOADED_MODELS` from `3` to `5`.

   ![Edit Ollama variable](/images/manual/use-cases/modify-var-ollama.png#bordered)

7. Click **Confirm**.

8. Return to **Deployments** > **ollama**, and then click **Restart** to apply the change.

Wait for the status icon to turn green, which indicates the new configuration has been loaded.

## Modify variables in a Secret

The workflow for Secrets is the same as for Configmaps. Follow the steps for [standard apps](#modify-standard-apps) or [C/S apps](#modify-cs-apps), depending on your app type.

:::info
When you open the YAML editor for a Secret, all values under the `data` field must be Base64 encoded.
:::

## FAQ

### Changes to a ConfigMap or Secret are not applied

After you modify a ConfigMap or Secret, the associated workload (Deployment) does not automatically reload the configuration. You must restart the workload to pick up the new values.

Use one of the following methods to restart the workload:
- **In Control Hub**  
  
  Go to **Deployments** under the app's namespace, click the target workload, and then click **Restart**.

- **Via Market or Settings**  
  
  If you are not sure which Deployment to restart, stop and then resume the app:
   - Go to **Market** > **My Olares**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the app's operation button, click **Stop**, and then click **Resume**.
   - Go to **Settings** > **Applications**, click the app, click **Stop**, and then click **Resume**.

Both methods will apply and load the latest configuration from the ConfigMap or Secret.
