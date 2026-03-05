---
outline: [2, 3]
description: Troubleshoot the issue where an Olares One device is pingable but cannot be accessed via SSH or web browser.
---

# Olares One reachable on LAN but gateway not running (“Network Not Ready”)

Use this guide when your Olares One device stops responding after working normally for a few days, while still being reachable on the network.

## Condition

- Your router shows the device is connected and has an IP address (it responds to a network "ping").
- The LarePass app displays **Network not ready** or **Olares not found**.
- Accessing `https://<device-IP>` in a web browser shows an "olares connection error".
- Restarting the device and your router does not fix the issue.

## Cause

Your device’s underlying operating system is running normally, which is why it successfully connects to your router and appears online. However, the core Olares software services (the Kubernetes cluster) have unexpectedly frozen or crashed. 

Because these core services are down, the specific network ports required for secure access like port `22` for SSH and port `443` for HTTPS stop working. As a result, the device cannot accept incoming connections from the LarePass app, your web browser, or your terminal. 

This freeze is typically caused by an abnormal container or a system service that has become stuck.

## Solution

Follow these steps in order to gather diagnostic information so our support team can help you get back online.

### Step 1: SSH into Olares One

While this specific issue usually blocks SSH connections, it is always best to try this method first because it is the most convenient way to access your device and collect diagnostic information.

1. Follow instructions [on this page](https://docs.olares.com/one/access-terminal-ssh.html#method-2-access-via-ssh) to access Olares One via SSH.
2. If the connection is successful, skip to **Step 3**.
3. If the connection times out or is refused, proceed to **Step 2**.

### Step 2: Log in to Olares One locally

When SSH access is blocked, you must log in locally using a monitor and keyboard.

1. Connect a monitor, a keyboard, and a mouse to your Olares One.
2. Power on Olares One and wait for the system to boot to the login prompt.
3. Log in with username `olares` and the same SSH password.

If local login also fails, skip the next step and contact Olares Support directly for advanced recovery options.

### Step 3: Check system status

1. Once you log in successfully, run the following command to view the list of internal software components:

    ```bash
    kubectl get pods -A
    ```
    
2. Look for components with the status that is not `Running` or `Completed`.
3. Take a clear screen shot of the full command output, or manually note down the abnormal components.
4. Send this screen shot with descriptions to Olares Support by submitting a GitHub Issue.
