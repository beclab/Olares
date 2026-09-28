---
outline: [2, 3]
description: Get help when LarePass shows "System error", collect logs through Ticket or Settings, and optionally check system pods.
head:
  - - meta
    - name: keywords
      content: Olares, system error, LarePass, pod status, kubectl, troubleshoot
---

# "System error" in LarePass

Use this guide when the **System** section in LarePass displays "System error". The message can have several causes. Start by collecting logs for the Olares team, or use the optional terminal checks below to narrow down the cause.

:::warning Do not uninstall LarePass or Olares OS
Do not uninstall LarePass, or open the **System error** page to uninstall Olares OS or restore factory settings. This error alone does not mean you need to reinstall. Keep LarePass available for account access and troubleshooting; uninstalling Olares OS or restoring factory settings can erase your data.
:::

![System error in LarePass](/images/manual/help/ts-sys-err.png#bordered){width=90%}

## Condition

- The **System** section in LarePass shows **System error** instead of **Running**.
- The Olares desktop might be inaccessible.

## Cause

The message means LarePass could not obtain a healthy system state. One or more system pods might be unhealthy, but the message alone does not identify the failing component or root cause.

## Solution

### Collect logs and contact support

You do not need to run terminal commands before asking for help. Choose an available option:

- **Ticket app (Olares 1.12.7 or later)**: If the app is accessible, use **Collect logs** under **System logs** to collect and attach logs to your support ticket. See [Submit via the Ticket app](request-technical-support.md#submit-via-the-ticket-app) for prerequisites and steps.
- **Settings**: If Olares Settings is accessible, export logs from **Advanced** > **Export system logs** and attach the archive to a support ticket. See [Export system logs](../olares/settings/developer.md#export-system-logs).

If neither is accessible, [submit a ticket through Olares Space](request-technical-support.md#submit-via-olares-space) and describe the issue, even if you cannot collect logs yet. Include a screenshot of **System error**, when it appeared and your time zone, your Olares version if known, and whether it followed an update or restart. Share full logs only through a private support channel, not a public GitHub issue.

### Advanced diagnostics (optional)

If you are comfortable using a terminal, or the Olares team asks for more details, follow these steps to identify unhealthy system pods and inspect their error events.

:::info
This guide uses Olares One as an example. If you installed Olares on your own hardware, the diagnostic steps are the same, but the way you access the terminal might differ.
:::

#### Step 1: Try to access Olares desktop

If you can still access the Olares desktop, open Control Hub and use its built-in terminal.

1. Open a browser and access your Olares Desktop:

    ```text
    https://desktop.<username>.olares.com
    ```

2. Open Control Hub.
3. In the left sidebar, under the **Terminal** section, click **Olares**.
    ![Open terminal](/images/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

If you can access the terminal successfully, skip to [Step 4](#step-4-check-system-pod-status).

#### Step 2: Connect via SSH

If you cannot access the Olares desktop, try connecting via SSH.

:::info Same network required
Your computer and Olares One should be on the same local network.
:::

1. Get the local IP address of your Olares One.

    a. Open the LarePass app and go to **Settings** > **System** to open the **Olares management** page.

    b. Tap the Olares One device card.

    c. Scroll down to **Network** and note the **Intranet IP**.

2. Find your SSH password in Vault.

    a. Tap **Vault** in the LarePass app. When prompted, enter your local password to unlock.

    b. In the top-left corner, tap **Vault** to open the side navigation, then tap **All vaults**.

    c. Find the item with the <span class="material-symbols-outlined">terminal</span> icon and tap it to reveal the password.
        ![Check saved SSH password in Vault](/images/one/ssh-check-password-in-vault.png#bordered)

3. Open a terminal on your computer and connect via SSH.

    a. Run the following command, replacing `<local_ip_address>` with the Intranet IP you noted earlier:

    ```bash
    ssh olares@<local_ip_address>
    ```

    b. When prompted, enter the SSH password.

If the connection is successful, skip to [Step 4](#step-4-check-system-pod-status).

#### Step 3: Log in locally

If SSH is also unavailable, log in directly on the device using a monitor and keyboard.

1. Connect a monitor and keyboard to your Olares One. A text-based login prompt appears automatically:

    ```text
    olares login:
    ```

2. Type the username `olares` and press **Enter**.
3. Enter the SSH password from [Step 2](#step-2-connect-via-ssh) and press **Enter**.

#### Step 4: Check system pod status

1. Run the following command to get the status of all pods across all namespaces:

    ```bash
    kubectl get pods -A
    ```

2. Check **STATUS** and **RESTARTS**. Look for states such as `CrashLoopBackOff`, `Error`, `ImagePullBackOff`, or a pod that remains `Pending`. A job in `Completed` state is not an error by itself.
3. Note the **NAMESPACE** and **NAME** of each pod that shows an error or a restart count that keeps increasing. If none do, skip to [Step 6](#step-6-record-the-result-and-collect-logs).
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

#### Step 5: Inspect the pod error

:::warning Review output before sharing it
`kubectl describe` can include IP addresses, node names, Olares IDs, domains, and configuration values. Do not paste its complete output into a public issue.
:::

1. Run the following command, replacing `<namespace>` and `<pod-name>` with the values you noted in the previous step:

    ```bash
    kubectl describe pod <pod-name> -n <namespace>
    ```

    For example:

    ```bash
    kubectl describe pod backup-66f8c76996-d7vnq -n os-framework
    ```

2. Scroll down to the **Events** section to find the detailed error message.
    ![Pod event details](/images/manual/help/ts-sys-err-pod-event-detail.png#bordered){width=90%}

#### Step 6: Record the result and collect logs

Record the following minimum information:

- The affected pod's **NAMESPACE**, **NAME**, **STATUS**, and **RESTARTS** values
- The error lines in the **Events** section
- The time the error appeared and your time zone
- Your installed Olares version and whether the error followed an update or restart

If no pod shows an error and restart counts are not increasing, record that result instead. It means the message cannot be explained by pod status alone and needs a different diagnostic branch.

Add the results to your support ticket. If you still need a log archive and cannot use Ticket or Settings, follow [Collect diagnostic information](../collect-diagnostic-information.md) to collect it from the terminal.

For a reproducible software bug, you can open a [GitHub Issue](https://github.com/beclab/Olares/issues/new) with the symptom and the limited fields above after redacting IDs, hostnames, IP addresses, and domains. Keep the full log archive in the private support ticket.
