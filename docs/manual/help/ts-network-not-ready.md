---
outline: [2, 3]
description: Troubleshoot the issue where your Olares One is connected to the network but remains unreachable through standard access methods.
head:
  - - meta
    - name: keywords
      content: Olares, network not ready, connection error, Olares One, SSH, pod status
---

# Network not ready or olares connection error

Use this guide to troubleshoot an Olares One device that appears powered on and connected to the network but has unexpectedly stopped responding.

## Condition

- The LarePass app displays **Network not ready** and accessing your Olares desktop shows an **olares connection error**, but your router shows the device is connected and the device responds to a network `ping`.
- Restarting the device and your router does not fix the issue.

## Cause

These symptoms do not identify a single cause. A successful `ping` only confirms that the device responds to basic network traffic. Olares might still be unavailable because system services have not started, one or more pods are unhealthy, the device is short on resources, or the access path is failing.

## Solution

Use the shortest available path to access the host, then check whether the problem is inside Olares or limited to the network access path.

### Step 1: Attempt SSH connection

Try this method first because it is the most convenient way to access your device and collect diagnostic information.

1. Get the local IP address of Olares One.

    a. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.
    ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)
    
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

    d. If the connection is successful, skip to [Step 3](#step-3-check-system-pod-status).

### Step 2: Log in locally

When the SSH access is unavailable, log in to the device locally using a monitor and keyboard.

1. Connect a monitor and keyboard to your Olares One. A text-based login prompt is displayed on your screen automatically.

    ```text
    olares login:
    ```

2. Type the username `olares` and press **Enter**.
3. Type the same SSH password obtained in **Step 1** and press **Enter**.

### Step 3: Check system pod status

1. Once you log in successfully, run the following command to get the status of all pods across all namespaces:

    ```bash
    kubectl get pods -A
    ```
    
2. Check the **STATUS** column and continue based on the result:

    - If a pod shows an error state such as `CrashLoopBackOff`, `Error`, `ImagePullBackOff`, or remains `Pending`, record only its **NAMESPACE**, **NAME**, **STATUS**, and **RESTARTS** values. A job in `Completed` state is not an error by itself.
    - If no pod shows an error and restart counts are not increasing, the symptom is more likely related to the access or network path. Record whether local access, the Olares domain, and LarePass VPN each work.

3. Record the time of the check, your time zone, and the installed Olares version.
4. Follow [Collect diagnostic information](../collect-diagnostic-information.md) to create a log archive and share it through a private channel.

:::warning Do not post complete logs publicly
Pod output and system logs can contain Olares IDs, hostnames, IP addresses, domains, and application metadata. A public GitHub Issue can include the symptom and the limited fields listed above, but not the complete command output or log archive.
:::
