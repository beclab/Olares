---
outline: [2, 3]
description: Troubleshoot the issue where your Olares One is connected to the network but cannot be accessed via the LarePass app, web browser desktop, or SSH.
---

# Olares One is online but unreachable

Use this guide when your Olares One suddenly stops responding after working normally for a few days, even though it is still powered on and connected to your network.

## Condition

- Your router shows the device is connected and the device responds to a network `ping`, but:
    - The LarePass app displays **Network not ready** or **Olares not found**.
    - Accessing your Olares desktop `https://desktop.{your-olares-id}.olares.com` in a web browser shows an olares connection error.
- Restarting the device and your router does not fix the issue.

## Cause

Your device’s underlying operating system is running normally, which is why it successfully connects to your router and appears online. However, the core Olares software services (the Kubernetes cluster) have unexpectedly frozen or crashed. 

Because these core services are down, the specific network ports required for secure access like port `22` for SSH and port `443` for HTTPS stop working. As a result, the device cannot accept incoming connections from the LarePass app, your web browser, or your terminal. 

This freeze is typically caused by an abnormal software container or a system service that has become stuck.

## Solution

Follow these steps in order to gather diagnostic information so Olares Support can help you get back online.

### Step 1: Attempt a remote SSH connection

While this specific issue usually blocks SSH connections, it is always best to try this method first because it is the most convenient way to access your device and collect diagnostic information.

1. Follow instructions [on this page](https://docs.olares.com/one/access-terminal-ssh.html#method-2-access-via-ssh) to access Olares One via SSH.
2. If the connection is successful, skip to [Step 3](#step-3-check-system-status).
3. If the connection times out or is refused, proceed to **Step 2**.

### Step 2: Log in locally

When the remote SSH access is blocked, you must log in to the device locally using a monitor and keyboard.

1. Connect a monitor and keyboard to your Olares One.
2. Power on Olares One and wait for the system to boot. A text-based login prompt is displayed on your screen automatically:

    ```text
    olares login:
    ```

3. Log in with the username `olares` and the same SSH password obtained in **Step 1**.
4. If the login is successful, proceed to **Step 3**.
5. If the login fails, skip the next step and contact Olares Support directly for advanced recovery options.

### Step 3: Check system status

1. Once you log in successfully, type the following command and then press **Enter** to view the list of internal software components:

    ```bash
    kubectl get pods -A
    ```
    
2. Check the **STATUS** column for any components that are not showing `Running` or `Completed`.
3. Take a clear photo of the full command output, or manually note down the abnormal components.
4. Send this photo or your notes with descriptions to Olares Support by [submitting a GitHub Issue](https://github.com/beclab/Olares/issues/new).

