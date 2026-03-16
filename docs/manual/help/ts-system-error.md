---
outline: [2,3]
description: Diagnose and collect information when the System section in LarePass shows "System error".
---
# "System error" in LarePass

Use this guide when the **System** section in LarePass displays "System error". 

There can be multiple underlying causes for this message, so follow the steps below to collect diagnostic information first, and then contact the Olares team with the results.
![System error in LarePass](/images/manual/help/ts-sys-err.png#bordered){width=90%}

## Condition

- The **System** section in LarePass shows "System error".
- LarePass cannot retrieve the system status of your Olares device.
- Olares desktop might be inaccessible.

## Cause

The "System error" message can be triggered by different underlying issues. A common cause is that one or more system pods on the Olares device are not running normally. When this happens, LarePass cannot retrieve system status and displays "System error".

## Solution

Access the device terminal to identify any pod that is not running normally, inspect its error details, and then share this information with the Olares team. This helps narrow down the possible causes and speeds up troubleshooting.

### Step 1: Access the terminal

- If you can access Control Hub in Olares desktop, follow [Option A](#option-a-access-the-command-line-from-the-desktop).
-  If you cannot access Control Hub, follow [Option B](#option-b-access-the-command-line-through-ssh).

#### Option A: Access via Control Hub

1. Open a browser and access your Olares desktop: 
    ```text
    https://desktop.<your-olaresID>.olares.com
    ```
2. Open Control Hub. In the left sidebar, under the **Terminal** section, click **Olares**.
    ![Open terminal](/images/manual/help/ts-sys-err-terminal.png#bordered){width=90%}
    
#### Option B: Access via SSH

:::warning
To connect through SSH, make sure your computer and the Olares device are on the same local network. Otherwise, the SSH connection will fail.
:::

1. (Optional) Obtain the local IP address using one of the following methods.

    <Tabs>
    <template #From-LarePass-mobile-client>

    a. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.

    b. Tap the Olares device card.

    c. Scroll down to the **Network** section and note the **Intranet IP**.

    </template>

    <template #Via-monitor>
    a. Connect your Olares device to a monitor and a keyboard.

    b. Open a terminal, and run `ifconfig`.
    
    c. Look for your active interface, typically `enp3s0` (wired) or `wlo1` (wireless). The IP address appears after `inet`.

    </template>

    </Tabs>

2. Run the following command, replacing `<local_ip_address>` with the Intranet IP you get from the previous step.
    ```bash
    ssh olares@<local_ip_address>
    ```
3. If prompted to confirm the connection, type `yes` and press Enter.
4. When prompted, enter the SSH password.
    :::tip
    If you have not changed it, the default SSH password is `olares`.
    :::

### Step 2: Identify the problematic pod

1. Run the following command to get the status of all pods across all namespaces:
    ```bash
    kubectl get pods -A
    ```
2. Check the **STATUS** column and locate any pods that are not in the `Running` state. 
3. Note down the exact **NAMESPACE** (the first column) and **NAME** (the second column) of each problematic pod.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

### Step 3: Inspect the pod error

1. Run the following command, replacing `<namespace>` and `<pod-name>` with the values you noted in the previous step:

    ```bash
    kubectl describe pod <pod-name> -n <namespace>
    ```

    In this example:

    ```bash
    kubectl describe pod backup-66f8c76996-d7vnq -n os-framework
    ```

2. Scroll down to the **Events** section to find the detailed error message.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-event-detail.png#bordered){width=90%}

### Step 4: Contact support

Create an issue in the [Olares GitHub repository](https://github.com/beclab/Olares/issues) and provide the following information:

- The output of `kubectl describe pod <pod-name> -n <namespace>`.
- A screenshot of the error message, if available.

This information helps our team investigate and resolve the issue faster.