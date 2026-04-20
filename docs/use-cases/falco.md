---
outline: deep
description: Deploy Falco on Olares to monitor Linux kernel events in real time and detect runtime security threats across hosts, containers, and Kubernetes workloads.
head:
  - - meta
    - name: keywords
      content: Olares, Falco, runtime security, eBPF, Kubernetes, container security, threat detection, Falcosidekick
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-20"
---

# Monitor runtime security with Falco

Falco is an open-source cloud-native runtime security tool built on eBPF. It watches Linux kernel events in real time and fires alerts when it spots suspicious behavior on hosts, in containers, or across Kubernetes workloads.

On Olares, Falco runs as a shared application. Agents collect events on each node, and a central web UI brings everything into one place.

## Learning objectives

In this guide, you will learn how to:
- Understand how Falco is deployed on Olares.
- View security alerts in the Falcosidekick Web UI.
- Tune event retention, detection rules, output channels, and plugins.
- Troubleshoot common plugin issues.

## Prerequisites

- Admin access to Olares. Falco runs in a client/server architecture, and only administrators can install or configure it. If you are a regular user, ask your administrator to install the Falco shared application first.

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
4. The Web UI reads from Redis and renders the dashboard.

## Install Falco

1. Open Market and search for "Falco".
   <!-- ![Falco in Market](/images/manual/use-cases/falco.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## View alerts in the Falcosidekick Web UI

The Falcosidekick Web UI is the recommended production setup. Alerts flow from each `falco-agent` into Falcosidekick and are rendered in the UI. You can also point Falcosidekick at external systems to forward alerts elsewhere. See [Configure output channels](#configure-output-channels) for details.

### Dashboard

The dashboard gives you a real-time snapshot of what's happening across nodes.

<!-- ![Falco dashboard](/images/manual/use-cases/falco-dashboard.png#bordered) -->

| Panel | What it shows |
|:------|:--------------|
| Global statistics | Aggregate alert counts for the selected time window. |
| Filter bar | Narrow results by source, priority, or tag. |
| Snapshot counters | Live totals under the current filter: `Total`, `Critical`, and `Notice`. |
| Pie chart | Alert distribution by source, priority, and tag. |
| Rule bar chart | Alerts grouped by rule. Useful for spotting noisy rules that need allowlists or threshold tuning. |
| Timeline by priority | Alert volume over time, split by priority. |
| Timeline by source | Alert volume over time, split by source. |

### Events

The **Event** page lists every alert with its full context.

<!-- ![Falco events](/images/manual/use-cases/falco-events.png#bordered) -->

| Column | Description |
|:-------|:------------|
| Timestamp | When the alert fired, for example `2026-04-14 20:35:37`. |
| Source | Where the event was collected from. |
| Hostname | The node or Pod that triggered the alert. |
| Priority | Severity, color-coded. |
| Rule | The rule name from the Falco rule library. |
| Output | The full alert message with context variables. |
| Tags | Classification tags. |

Click the details icon on any row to open the event view. From there, switch to the **JSON** tab to inspect or copy the raw payload.

## Configure Falco

:::warning Admin only
Configuration requires admin privileges. Regular users cannot change Falco settings.
:::

This section covers the four areas administrators typically tune:

| Area | What you control |
|:-----|:-----------------|
| [Event retention](#set-event-retention) | How long alerts stay in the Web UI before cleanup. |
| [Detection rules](#manage-detection-rules) | Which behaviors trigger alerts. |
| [Output channels](#configure-output-channels) | Where alerts are sent (Web UI, file, external systems). |
| [Plugins](#install-and-use-plugins) | Extra event sources such as Kubernetes audit logs. |

For anything not covered here, see the [official Falco documentation](https://falco.org/docs/).

### Set event retention

Falcosidekick keeps alerts for 72 hours by default. To change this:

1. Navigate to **Settings** > **Applications** > **Falco** > **Environment variables**.

2. Edit `FALCOSIDEKICK_UI_TTL`. Use a numeric value followed by a unit suffix (`s`, `m`, `h`, `d`, `w`, `M`, `y`), for example `7d` for seven days. Leave the value empty to keep events indefinitely.
   <!-- ![Edit FALCOSIDEKICK_UI_TTL](/images/manual/use-cases/falco-edit-ttl.png#bordered) -->

3. Click **Confirm** to save the change, then click **Apply** to apply it.

4. Go to `system/falcoserver-shared/falco-central` and restart the process.
   <!-- ![Restart falco-central](/images/manual/use-cases/falco-restart-central.png#bordered) -->

5. Confirm the new value in either the YAML view at `system/falcoserver-shared/falco-central` or the `webui` environment variables on the same page.

### Manage detection rules

Falco ships with a default rule set. You can layer custom rules on top, or disable rules that don't apply to your environment.

#### Understand the rule format

Every rule follows the same format:

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
| `condition` | When the rule fires. |
| `output` | Alert template. Supports fields like `%proc.cmdline`. |
| `priority` | Severity level. One of:<br>`EMERGENCY`, `ALERT`, `CRITICAL`, `ERROR`, `WARNING`, `NOTICE`, `INFORMATIONAL`, `DEBUG`. |
| `tags` | Tags for filtering and grouping. |

:::warning Restart required
Every rule change takes effect only after you restart the `falco-agent` DaemonSet. Rule names must be unique and match exactly. A mismatched name stops `falco-agent` from starting.
:::

#### View active rules

Go to `system/falcoserver-shared/falco-agent` and open the startup logs. Falco loads three rule files:

| File | Purpose |
|:-----|:--------|
| `falco_rules.yaml` | Upstream default rules. Read-only. View them at `/etc/falco/falco_rules.yaml` in the `falco-agent` image. |
| `custom_rules.yaml` | Your custom rules, managed through the `falco-custom-rules` ConfigMap. |
| `falco_disable_rules.yaml` | Rules you've disabled, managed through the `falco-disable-rules` ConfigMap. |

<!-- ![View active rules](/images/manual/use-cases/falco-view-rules.png#bordered) -->

#### Create a custom rule

1. Navigate to `system/falcoserver-shared/falco-custom-rules` and open the YAML editor. Add your rule under `data.custom_rules.yaml`. For example:

    ```yaml
    kind: ConfigMap
    apiVersion: v1
    metadata:
      name: falco-custom-rules
      namespace: falcoserver-shared
      labels:
        app.kubernetes.io/managed-by: Helm
      annotations:
        meta.helm.sh/release-name: falcoserver
        meta.helm.sh/release-namespace: falcoserver-shared
    data:
      custom_rules.yaml: |
        - rule: Test - Terminal Shell In Container
          desc: Test rule to validate Falco custom rules pipeline
          condition: >
            evt.type in (execve, execveat)
            and container and shell_procs and proc.name in (bash, sh, zsh)
            and k8s.ns.name exists
            and not (k8s.ns.name in ("kube-system", "falco", "falcoserver-shared"))
          output: >
            TEST custom rule matched (ns=%k8s.ns.name user=%user.name command=%proc.cmdline container=%container.id image=%container.image.repository)
          priority: WARNING
          tags: [container, test]
    ```

2. Restart the `falco-agent` DaemonSet from `system/falcoserver-shared/falco-agent`.

3. Verify the rule is active:
    - Check `/etc/falco/rules.d/managed/custom_rules.yaml` in the `falco-agent` image to confirm the rule was written.
    - Open the Falcosidekick Web UI dashboard and look for alerts from the new rule.

#### Disable a rule

1. Navigate to `system/falcoserver-shared/falco-custom-rules` and edit the `falco-disable-rules` ConfigMap. For example, to disable the `Fileless execution via memfd_create` rule:

    ```yaml
    kind: ConfigMap
    apiVersion: v1
    metadata:
      name: falco-disable-rules
      namespace: falcoserver-shared
      labels:
        app.kubernetes.io/managed-by: Helm
      annotations:
        meta.helm.sh/release-name: falcoserver
        meta.helm.sh/release-namespace: falcoserver-shared
    data:
      falco_disable_rules.yaml: |
        - rule: "Fileless execution via memfd_create"
          override:
            enabled: replace
          enabled: false
    ```

2. Restart the `falco-agent` DaemonSet.

3. Open the Falcosidekick Web UI dashboard and confirm the rule no longer fires.

### Configure output channels

Falco supports several output channels. The Web UI is enabled by default. You can also log alerts to a local file or forward them to external systems.

#### Send alerts to the Web UI (default)

Out of the box, `falco-agent` forwards alerts to Falcosidekick over HTTP, which the Web UI then displays. Verify the configuration at `system/falcoserver-shared/falco-agent`:

```plain
- '-o'
- http_output.enabled=true
- '-o'
- http_output.url=http://falco-sidekick.{{ .Release.Namespace }}:2801/
```

#### Write alerts to a file

To write alerts to a local log file:

1. Navigate to **Settings** > **Applications** > **Falco** > **Environment variables**, set `File_output` to `true`, then click **Apply**.
   <!-- ![Enable file output](/images/manual/use-cases/falco-enable-file-output.png#bordered) -->

2. Restart the `falco-agent` DaemonSet from `system/falcoserver-shared/falco-agent`.

3. Check the configuration in the YAML editor at `system/falcoserver-shared/falco-agent` to confirm the output is applied.

4. Read the log file at **Files** > **Applications** > **Data** > **falco** > **logs** > `events.log`.

    :::info
    The log directory is mounted in the admin environment. Only administrators can read it.
    :::

#### Forward alerts to external systems

For Slack, Elasticsearch, webhooks, and other destinations, configure Falcosidekick directly. See the [Falcosidekick documentation](https://github.com/falcosecurity/falcosidekick) for the full list.

### Install and use plugins

Falco plugins add new event sources. The example below installs the `k8saudit` plugin for Kubernetes audit logging.

1. Go to `System/falcoserver-shared/falco-plugin-installer` and open the toolbox terminal. Install the plugin artifacts:

    ```bash
    falcoctl artifact install k8saudit
    falcoctl artifact install k8saudit-rules
    falcoctl artifact install json
    ```

2. Enable the plugins. Edit `/etc/falco/config.d/plugins.local.yaml` in the `falco-agent` container (or directly from **Files** > **Applications** > **Data** > **falco**):

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

3. Restart the `falco-agent` DaemonSet.

4. Check the `falco-agent` logs. A successful install shows `k8s_audit` loaded at startup.

## Troubleshooting

### falco-agent fails to start after installing k8saudit

**Cause**

The `k8saudit` rules were installed, but the plugins were not enabled in `plugins.local.yaml`. When `falco-agent` restarts, it tries to load rules for a plugin that isn't active and crashes.

**Solution**

1. Open the `falco-plugin-installer` toolbox.
2. Remove the orphaned rule file:
    ```bash
    rm /etc/falco/rules.d/managed/k8s_audit_rules.yaml
    ```
3. Restart the `falco-agent` DaemonSet.

To bring `k8saudit` back, redo [Install and use plugins](#install-and-use-plugins) from the beginning, making sure to enable the plugins in Step 2 before restarting.

## Learn more

- [Falco official documentation](https://falco.org/docs/): Full reference for rules, conditions, and plugins.
- [Falcosidekick documentation](https://github.com/falcosecurity/falcosidekick): Supported output destinations and configuration options.
