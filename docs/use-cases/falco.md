---
outline: deep
description: Deploy Falco on Olares to monitor Linux kernel events in real time and detect runtime security threats across hosts, containers, and Kubernetes workloads.
head:
  - - meta
    - name: keywords
      content: Olares, Falco, runtime security, eBPF, Kubernetes, container security, threat detection, Falcosidekick
app_version: "1.0.11"
doc_version: "1.0"
doc_updated: "2026-04-23"
---

# Monitor runtime security with Falco

Falco is an open-source cloud-native runtime security tool built on eBPF. It watches Linux kernel events in real time and fires alerts when it spots suspicious behavior on hosts, in containers, or across Kubernetes workloads.

On Olares, Falco runs as a shared application. Agents collect events on each node, and a central Falcosidekick UI brings everything into one place.

Use this guide when you want to install Falco on Olares and review runtime security alerts from hosts, containers, and Kubernetes workloads.

## Learning objectives

In this guide, you will learn how to:
- View security alerts in the Falcosidekick UI.
- Configure event retention, detection rules, output channels, and plugins.
- Troubleshoot common plugin issues.

## Prerequisites

- **Admin access required**: Falco runs in a client/server architecture, and only administrators can install or configure it. If you are a regular user, ask your administrator to install the Falco shared application first.

## How Falco works on Olares

Falco uses distributed collection with centralized display, so you can monitor every node from a single UI.

### Components

| Component | Kind | Role |
|:----------|:-----|:-----|
| `falco-agent` | DaemonSet | Runs on every node and captures kernel events that match Falco rules. |
| `falco-plugin-installer` | DaemonSet | Toolbox for installing plugins and rules. Ships with `falcoctl`. |
| `falco-sidekick` | Service | Receives HTTP output from all `falco-agent` instances. |
| `webui` | Service | Serves the dashboard and event views. |

### Event flow

1. `falco-agent` on each node captures kernel events locally.
2. When an event matches a rule, `http_output` forwards it to `falco-sidekick`.
3. `falco-sidekick` writes the event to Redis.
4. The Falcosidekick UI reads from Redis and renders the dashboard.

## Install Falco

1. Open Market and search for "Falco".

  ![Falco in Market](/images/manual/use-cases/falco.png#bordered){width=90%}

2. Click **Get**, then **Install**, and wait for installation to complete.

## View alerts in Falcosidekick UI

After Falco is installed, open the Falco application to access the Falcosidekick UI and review security alerts.

The Falcosidekick UI is the default place to review alerts on Olares. If needed, administrators can also forward alerts to external systems. See [Configure output channels](#configure-output-channels).

### Dashboard

The **Dashboard** page gives you a real-time overview of alert activity across nodes.

  ![Falco dashboard](/images/manual/use-cases/falco-dashboard.png#bordered){width=90%}

| Panel | What it shows |
|:------|:--------------|
| Global statistics | Aggregate alert counts for the selected time window. |
| Filter bar | Narrow results by source, priority, or tag. |
| Snapshot counters | Live totals under the current filter: `Total`, `Critical`, and `Notice`. |
| Pie chart | Alert distribution by source, priority, and tag. |
| Rule bar chart | Alerts grouped by rule. Useful for spotting noisy rules that need <br>allowlists or threshold tuning. |
| Timeline by priority | Alert volume over time, split by priority. |
| Timeline by source | Alert volume over time, split by source. |

### Events

The **Events** page lists every alert with its full context.

  ![Falco events](/images/manual/use-cases/falco-events.png#bordered){width=90%}

| Column | Description |
|:-------|:------------|
| Timestamp | When the alert was generated, for example `2026-04-14 20:35:37`. |
| Source | Where the event came from. |
| Hostname | The host associated with the alert. |
| Priority | Alert severity, color-coded. |
| Rule | The rule name from the Falco rule library. |
| Output | The full alert message with context variables. |
| Tags | Classification tags. |

To inspect an alert in detail:
1. On the **Events** page, find the alert you want to inspect.
2. Click **{…}** on the right side of the row.
3. Review the detail panel. Switch to the **JSON** tab if you need the raw payload.

## Configure Falco

Use this section when you need to change how Falco stores, detects, or forwards alerts.

:::warning Admin only
Configuration requires admin privileges. Regular users cannot change Falco settings.
:::


| Area | What you control |
|:-----|:-----------------|
| [Event retention](#set-event-retention) | How long alerts stay in the Falcosidekick UI before cleanup. |
| [Detection rules](#manage-detection-rules) | Which behaviors trigger alerts. |
| [Output channels](#configure-output-channels) | Where alerts are sent (Falcosidekick UI, file, external systems). |
| [Plugins](#install-and-use-plugins) | Extra event sources such as Kubernetes audit logs. |

### Set event retention

Falco keeps alerts for 72 hours by default. To change how long alerts are kept:

1. Go to **Settings** > **Applications** > **Falco** > **Manage environment variables**.
2. Click <i class="material-symbols-outlined">edit_square</i> next to `FALCOSIDEKICK_UI_TTL`. 
3. Enter a duration with a unit suffix, such as `7d` for seven days. Supported suffixes include `s`, `m`, `h`, `d`, `w`, `M` and `y`. Leave the value empty to keep events indefinitely.
4. Click **Confirm**, then click **Apply**.

  ![Edit FALCOSIDEKICK_UI_TTL](/images/manual/use-cases/falco-edit-ttl.png#bordered){width=90%}

5. Optional: To verify that the new value is applied, open Control Hub, and go to **Browse** > **System** > **falcoserver-shared** > **Deployments** > **falco-central**.
   - Click <i class="material-symbols-outlined">edit_square</i> to open the YAML file, locate `FALCOSIDEKICK_UI_TTL`, and check its value.
   - In the right panel, under **Environment variables**, click **webui** and check the value of `FALCOSIDEKICK_UI_TTL`.

### Manage detection rules

Falco uses rules to decide which behaviors should generate alerts.

Use this section when you want to:
- Check which rule files are currently loaded.
- Add a custom rule.
- Disable a rule that is not relevant in your environment.

:::warning Restart required
Rule changes take effect only after you restart the `falco-agent` DaemonSet.

Rule names must be unique and match exactly. A mismatched rule name can prevent `falco-agent` from starting.
:::

#### Understand the rule format

A Falco rule usually includes:

```yaml
- rule: Test - Terminal Shell In Container
  desc: Test rule to validate Falco custom rules pipeline
  condition: container and shell_procs and proc.name in (bash, sh, zsh)
  output: >
    TEST custom rule matched (user=%user.name command=%proc.cmdline container=%container.id image=%container.image.repository)
  priority: WARNING
  tags: [container, test]
```

| Field | Description |
|:------|:------------|
| `rule` | Unique rule name. |
| `desc` | Short description. |
| `condition` | The condition that triggers the alert. |
| `output` | The alert message template. Supports fields like `%proc.cmdline`. |
| `priority` | Alert severity. One of:<br>`EMERGENCY`, `ALERT`, `CRITICAL`, `ERROR`, `WARNING`, `NOTICE`,<br> `INFORMATIONAL`, `DEBUG`. |
| `tags` | Tags for filtering and grouping. |

#### Check loaded rule files

Falco loads rule files at startup. The exact set of loaded files depends on your current configuration.

For example, Falco may load:

| File | Purpose |
|:-----|:--------|
| `falco_rules.yaml` | Upstream default rules provided by Falco.  |
| `custom_rules.yaml` | Custom rules that you add for your own environment. |
| `falco_disable_rules.yaml` | Rules that you explicitly disable. |

To check which rule files are currently loaded:

1. Open Control Hub.
2. Go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**. 
3. Click your pod to open the details panel.
4. Under **Containers**, click <i class="material-symbols-outlined">article</i> next to **falco** to open the logs. 

  ![View active rules](/images/manual/use-cases/falco-view-rules.png#bordered){width=90%}

5. Look for the `Loading rules from:` section in the logs.

#### View default rules 

To view the default Falco rules:

1. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**.
2. Click your pod to open the details panel.
3. Under **Containers**, click <i class="material-symbols-outlined">terminal</i> next to **falco** to open the terminal. 
4. Run the following command:

    ```bash
    cat /etc/falco/falco_rules.yaml
    ```

#### Create a custom rule

1. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Configmaps** > **falco-custom-rules**. 
2. In the right panel, click <i class="material-symbols-outlined">edit_square</i> next to **falco-custom-rules**.
3. Change `custom_rules.yaml:` to `custom_rules.yaml: |`, then add your rule on the next line.

   Example:

   ```yaml
   data:
     custom_rules.yaml: |
       - rule: Test - Terminal Shell In Container
         desc: Test rule to validate the custom rule pipeline
         condition: >
           evt.type in (execve, execveat)
           and container
           and shell_procs
           and proc.name in (bash, sh, zsh)
           and k8s.ns.name exists
           and not (k8s.ns.name in ("kube-system", "falco", "falcoserver-shared"))
         output: >
           TEST custom rule matched (ns=%k8s.ns.name user=%user.name command=%proc.cmdline container=%container.id image=%container.image.repository)
         priority: WARNING
         tags: [container, test]
    ```
4. Click **Confirm** to save the changes.
5. Restart **falco-agent**.
  
    a. Go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**.

    b. In the right panel, click <i class="material-symbols-outlined">more_vert</i>, then select **Restart**.
6. Optional: Verify that the rule is active.
    - On the same page, click your pod to open the details panel. Under **Containers**, click <i class="material-symbols-outlined">terminal</i> next to **falco**, then run:
      ```bash
      cat /etc/falco/rules.d/managed/custom_rules.yaml
      ```
    - On the Falcosidekick UI dashboard, check the **Rules** dropdown list for the new rule. The rule only appears after it is triggered.

#### Disable a rule

1. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Configmaps** > **falco-disable-rules**. 
2. In the right panel, click <i class="material-symbols-outlined">edit_square</i> next to **falco-disable-rules** to open the YAML editor.
3. Add your rule on the line below `falco_disable_rules.yaml: |`.

    For example, to disable the `Terminal shell in container` rule:

    ```yaml
    data:
      falco_disable_rules.yaml: |
        - rule: Terminal shell in container
          override:
            enabled: replace
          enabled: false
    ```

4. Click **Confirm**.
5. Restart **falco-agent**.
  
    a. Go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**.

    b. In the right panel, click <i class="material-symbols-outlined">more_vert</i>, then select **Restart**.

6. Optional: Verify that the rule is disabled.
    - On the same page, click your pod to open the details panel. Under **Containers**, click <i class="material-symbols-outlined">terminal</i> next to **falco**, then run:
      ```bash
      cat /etc/falco/rules.d/managed/falco_disable_rules.yaml
      ```
    - On the Falcosidekick UI dashboard, check the **Rules** dropdown list. The disabled rule may still appear from past events, but new events will no longer trigger it, and it will disappear once those historical records expire.

### Configure output channels

By default, Falco sends alerts to the Falcosidekick UI. You can also write alerts to a local file or forward them to external systems.

#### Send alerts to the Falcosidekick UI

By default, `falco-agent` forwards alerts to Falcosidekick over HTTP, and the alerts are then displayed in the Falcosidekick UI.

1. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**. 
2. In the right panel, click <i class="material-symbols-outlined">edit_square</i> next to **falco-agent** to open the YAML editor.
3. Check for the following output configuration:
    
    Example:
    ```plain
    - '-o'
    - http_output.enabled=true
    - '-o'
    - http_output.url=http://falco-sidekick.falcoserver-shared:2801/
    ```

#### Write alerts to a file

To write alerts to a local log file:

1. Go to **Settings** > **Applications** > **Falco** > **Manage environment variables**.
2. Set `File_OUTPUT` to `true`.
   
   ![Enable file output](/images/manual/use-cases/falco-enable-file-output.png#bordered){width=90%}

3. Click **Confirm**, then click **Apply**.
4. Optional: Verify that file output is enabled.

    a. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**. 
    
    b. In the right panel, click <i class="material-symbols-outlined">edit_square</i> next to **falco-agent** to open the YAML editor. 
    
    c. Check whether the configuration includes `file_output.enabled=true`.

5. New alerts are written to `events.log` in Files at `/Data/falco/logs/`.

    :::info
    The log directory is mounted in the admin environment. Only administrators can read it.
    :::

#### Forward alerts to external systems

To forward alerts to Slack, Elasticsearch, Webhook, or other external destinations, configure Falcosidekick directly.

See the [Falcosidekick documentation](https://github.com/falcosecurity/falcosidekick) for the full list of supported outputs.

### Set up plugins

Falco plugins add additional event sources. The example below installs the `k8saudit` plugin for Kubernetes audit logging.

#### Install plugins

1. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-plugin-installer**.
2. Click your pod to open the details panel.
3. Under **Containers**, click <i class="material-symbols-outlined">terminal</i> next to **toolbox**. 
4. Run the following commands one by one to install the plugin artifacts:

    ```bash
    falcoctl artifact install k8saudit
    falcoctl artifact install k8saudit-rules
    falcoctl artifact install json
    ```
5. Optional: Verify that the plugin is installed.
    
    a. Go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**.

    b. Click your pod to open the details panel. 

    c. Under **Containers**, click <i class="material-symbols-outlined">article</i> next to **falco**.

    d. Check whether `k8s_audit_rules.yaml` appears in the `Loading rules from:` section.

#### Enable plugins

<Tabs>
<template #Edit-in-terminal>

1. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-plugin-installer**.
2. Click your pod to open the details panel.
3. Under **Containers**, click <i class="material-symbols-outlined">terminal</i> next to **toolbox**.
4. Run the following commands:
  ```bash
  cd /etc/falco/config.d/
  vi plugins.local.yaml
  ```
5. Update the file with the following example configuration:
  
    ```yaml
    plugins:
      - name: k8saudit
        library_path: /var/lib/falco/plugins/libk8saudit.so
        init_config: ""
        open_params: "http://:9765/k8s-audit"
      - name: json
        library_path: /var/lib/falco/plugins/libjson.so
        init_config: ""
    load_plugins: [k8saudit, json]
    ```
6. Save the file and exit the editor.

</template>
    
<template #Edit-in-Files>

1. In Files, open `/Data/falco/plugins.local.yaml`.
2. Update the file with the following example configuration:
  
    ```yaml
    plugins:
      - name: k8saudit
        library_path: /var/lib/falco/plugins/libk8saudit.so
        init_config: ""
        open_params: "http://:9765/k8s-audit"
      - name: json
        library_path: /var/lib/falco/plugins/libjson.so
        init_config: ""
    load_plugins: [k8saudit, json]
    ```
3. Save the file.

</template>
</Tabs>

**After updating `plugins.local.yaml`:**

1. Restart **falco-agent**.
  
    a. Go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**.

    b. In the right panel, click <i class="material-symbols-outlined">more_vert</i>, then select **Restart**.

2. Optional: Verify that the plugin is enabled.
    
    a. Go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**.

    b. Click your pod to open the details panel.
    
    c. Under **Containers**, click <i class="material-symbols-outlined">article</i> next to **falco**.

    d. Check whether the following info appears in the log:
      - `Enabled event sources: k8s_audit`
      - `Opening 'k8s_audit' source with plugin 'k8saudit'`


## Troubleshooting

### falco-agent fails to start after installing k8saudit rules

You may see an error like this in the logs:

```plain
LOAD_UNUSED_LIST (Unused list): List not referred to by any other rule/macro
Error: Plugin requirement not satisfied, must load one of: k8saudit (>= 0.7.0), k8saudit-aks (>= 0.1.0), k8saudit-eks (>= 0.4.0), k8saudit-gke (>= 0.1.0), k8saudit-ovh (>= 0.1.0)
```

**Cause**

If the `k8saudit` rules are installed but the plugin is not enabled successfully in `plugins.local.yaml`, **falco-agent** fails to start after restart.

**Solution**

1. Open Control Hub, then go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-plugin-installer**.
2. Click your pod to open the details panel.
3. Under **Containers**, click <i class="material-symbols-outlined">terminal</i> next to **toolbox**.
4. Remove the `k8s_audit_rules.yaml` file:
    ```bash
    rm /etc/falco/rules.d/managed/k8s_audit_rules.yaml
    ```
5. Restart **falco-agent**.
  
    a. Go to **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**.

    b. In the right panel, click <i class="material-symbols-outlined">more_vert</i>, then select **Restart**.

If you want to continue using `k8saudit`, repeat [Install and use plugins](#install-and-use-plugins) from the beginning.

## Learn more

- [Falco official documentation](https://falco.org/docs/): Full reference for rules, conditions, and plugins.
- [Falcosidekick documentation](https://github.com/falcosecurity/falcosidekick): Supported output destinations and configuration options.