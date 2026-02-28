---
outline: [2, 3]
description: Learn how to access the Olares One host terminal for command-line usage via Control Hub or SSH.
head:
  - - meta
    - name: keywords
      content: SSH, Olares Terminal, Control Hub
---

# Connect to Olares One via SSH  <Badge type="tip" text="10 min" />

Certain development and operational tasks, such as connecting Olares One with other devices to form a cluster, require direct command-line access. 

Olares One is primarily used as a headless device, so you can access the terminal remotely without a monitor or keyboard connected to it.

You can connect to the host shell using one of the following methods:
- **Control Hub Terminal** is a web-based interface for direct root access. It is recommended for quick tasks.
- **Secure Shell (SSH)** is the standard protocol for remote management and complex operations.

## Prerequisites
**Hardware**
- Your Olares One is set up and connected to a network.
- A client device, such as a computer, is required to access the terminal.

**Experience**
- Basic familiarity with terminal commands and the command-line interface (CLI).

## Method 1: Access via Control Hub
For quick access without configuring SSH clients, you can use the web-based terminal built directly into Control Hub.

1. Open the Control Hub app. 
2. In the left sidebar, under the **Terminal** section, click **Olares**.
   ![Terminal](/images/manual/olares/controlhub-terminal.png#bordered)

You can now execute system commands directly in the embedded terminal.

:::tip Run as `root`
The Control Hub terminal runs as `root` by default. You do not need to use `sudo` before commands.
:::

## Method 2: Access via SSH

SSH establishes a secure session over the network, allowing you to use command-line operations for Olares One on your current device.

### Step 1: Get IP address of Olares One
To connect via SSH, you first need to find the internal IP address of your Olares One.

1. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.
   ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)

2. Tap on the device card.
3. Scroll down to the **Network** section. You can find the **Intranet IP** there.

:::tip Check via Control Hub
You can check the IP using the `ifconfig` command in the Control Hub terminal.

Look for your active connection, typically named `enp3s0` (wired) or `wlo1` (wireless). The IP address follows `inet`.
:::

### Step 2: Check SSH password in Vault

<!--@include: ./reusables-reset-ssh.md{7,16}-->

### Step 3: Connect via SSH

1. Open a terminal on your computer. 
2. Run the `ssh` command using Olares One's local IP address:
    ```bash
    # The default username for Olares One is olares.
    ssh olares@<host_ip_address>
    ```
3. Enter the password when prompted.

## Reset SSH password
<!--@include: ./reusables-reset-ssh.md{19,}-->
