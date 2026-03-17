---
outline: [2, 3]
description: Diagnose and collect information when the System section in LarePass shows "System error".
---

# "System error" in LarePass

Use this guide when the **System** section in LarePass displays "System error".

This guide uses Olares One as the example device. If you are using another Olares device, you can follow the same general process where applicable.

![System error in LarePass](/images/manual/help/ts-sys-err.png#bordered){width=90%}

## Condition

- The **System** section in LarePass shows "System error".
- LarePass cannot retrieve the system status of your Olares device.
- Olares desktop might be inaccessible.

## Cause

The "System error" message can be triggered by different underlying issues. A common cause is that one or more system pods on the Olares device are not running normally. When this happens, LarePass cannot retrieve system status and displays "System error".

## Solution

Follow the steps below to access the device terminal, identify any pod that is not running normally, inspect its error details, and share the results with the Olares team. This helps narrow down possible causes and speed up troubleshooting.

### Step 1: Try to access Olares desktop

If you can still access the Olares desktop, open Control Hub and use its built-in terminal.

1. Open a browser and access your Olares desktop:

    ```text
    https://desktop.<your-olares-id>.olares.com
    ```

2. Open Control Hub.
3. In the left sidebar, under the **Terminal** section, click **Olares**.
    ![Open terminal](/images/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

If you can access the terminal successfully, go to [Step 4](#step-4-check-system-pod-status).

### Step 2: Attempt SSH connection

If you cannot access the Olares desktop, try SSH first.

:::info Same network required
Your computer and the Olares device should be on the same local network.
:::

1. Get the local IP address of your Olares device. 
If you cannot find the local IP address, continue to get the SSH password below, and then go to **Step 3**.

    a. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.

    b. Tap the Olares One device card.

    c. Scroll down to the **Network** section and note the **Intranet IP**.

2. Check SSH password in Vault.

    a. Tap **Vault** in the LarePass app. When prompted, enter your local password to unlock.

    b. In the top-left corner, tap **Vault** to open the side navigation, and then tap **All vaults** to display all saved items.

    c. Find the item with the <span class="material-symbols-outlined">terminal</span> icon and tap it to reveal the password.
        ![Check saved SSH password in Vault](/images/one/ssh-check-password-in-vault.png#bordered)

3. Connect via SSH.
    
    a. Open a terminal on your computer.

    b. Type the following command, replace `<local_ip_address>` with the Intranet IP, and then press **Enter**:
    
    ```bash
    ssh olares@<local_ip_address>
    ```
        
    c. When prompted, type the SSH password, and then press **Enter**.

If the connection is successful, go to [Step 4](#step-4-check-system-pod-status).

If you cannot connect through SSH, go to [Step 3](#step-3-log-in-locally).

### Step 3: Log in locally

Use a monitor and keyboard to log in to the device locally.

1. Connect a monitor and keyboard to your Olares device. A text-based login prompt is displayed on your screen automatically.

    ```text
    olares login:
    ```

2. Type the username `olares` and press **Enter**.
3. Type the same SSH password obtained in **Step 2** and press **Enter**.

### Step 4: Check system pod status

1. Run the following command to get the status of all pods across all namespaces:

    ```bash
    kubectl get pods -A
    ```

2. Check the **STATUS** column for any pods that are not in the `Running` state.
3. Note down the exact **NAMESPACE** and **NAME** of each problematic pod.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

### Step 5: Inspect the pod error

1. Run the following command, replacing `<namespace>` and `<pod-name>` with the values you noted in the previous step:

    ```bash
    kubectl describe pod <pod-name> -n <namespace>
    ```

    For example:

    ```bash
    kubectl describe pod backup-66f8c76996-d7vnq -n os-framework
    ```

2. Scroll down to the **Events** section to find the detailed error message.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-event-detail.png#bordered){width=90%}

### Step 6: Contact support

Create an issue in the [Olares GitHub repository](https://github.com/beclab/Olares/issues) and provide the following information:

- The output of `kubectl describe pod <pod-name> -n <namespace>`
- A screenshot of the error message, if available

This information helps our team investigate and resolve the issue faster.