---
outline: [2, 3]
description: Technical guide for installing and activating an Olares device using Olares CLI.
---

# Activate an Olares device using the Olares CLI

This tutorial walks you through activating an Olares device (e.g., Olares One) using the Olares CLI tool. The process assumes the device is freshly unboxed and has not been installed or activated.

## Learning bjectives

In this tutorial, you will learn how to:

- Download and extract the Olares CLI tool.
- Install the Olares OS on a new device.
- Retrieve a Fast Reverse Proxy (FRP) host for remote access.
- Run the activation command to configure your device.

## Prerequisites

- **Connection to device**: You have logged into the Olares device physically or via SSH.
- **Network connectivity**: The device must have access to the internet to download packages, query FRP servers, and complete activation.
- **Root privileges**: Ensure you have root privileges. All installation and activation commands require `sudo` or root access on the device.
- **LarePass preparation**: You have registered an Olares ID using the LarePass app.

    ![Fast creation](/images/manual/get-started/create-olares-id.png)

## Step 1: Download and extract the CLI tool

1. Download the Olares CLI package.

    ```bash
    curl -sSOL https://cdn.olares.com/common/olares-cli-amd64.8cbdc32.tar.gz
    ```

2. Extract the downloaded file.

    ```bash
    tar xzf olares-cli-amd64.8cbdc32.tar.gz
    ```

## Step 2: Retrieve the FRP list

Find an available FRP host to enable remote access to your device.

1. Run the following command. Replace `{olares-id}` with your registered Olares ID.

    ```bash
    olares-cli wizard frp {olares-id}
    ```

    **Example:**

    ```bash
    olares-cli wizard frp alice123@olares.com
    ```

2. Select a host address from the output list and save it for the activation step.

## Step 3: Install Olares OS

:::info For Olares One hardware
A fresh Olares One device is shipped in an uninstalled state. You must run the installation command to set up Olares first before you attempt activation.
:::

1. Run the install command.

    ```bash
    sudo olares-cli install
    ```

2. Wait for the installation process to finish. The terminal outputs a local gateway address and a default password. Save these details for the activation step.

    **Example:**
    - **Wizard URL**: The local gateway address, such as `http://192.168.50.123:30180`.
    - **Password**: The default Olares login password.

    ![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

## Step 4: Activate Olares

Run the activation command to configure and secure your device. This process connects your Olares ID to the device and configures network tunneling and credentials.

1. Prepare your activation command based on the following required parameters:

    | Parameter | Description |
    |:----------|:------------|
    | `olares-id` | The unique identifier within the Olares ecosystem.<br>Find it in the LarePass app after registration. |
    | `mnemonic` | Your user backup phrase from the LarePass app. |
    | `password` | The default Olares login password generated in Step 3. |
    | `reset-password` | Specify a new login password for Olares. |
    | `authurl` | Enter the `wizard-url` generated in Step 3. |
    | `vault` | Enter the `wizard-url` and append `/server`.  |
    | `bfl` | Enter the `wizard-url` generated in Step 3. |
    | `host` | Specify the FRP host address selected in Step 2. |
    | `enable-tunnel` | Enter `true` to activate using tunnel mode. |

2. Replace the placeholders in the following command with your specific values, and then run it.

    ```bash
    sudo olares-cli wizard activate {olares-id} \
    --mnemonic "{mnemonic}" \
    --password="{password}" \
    --reset-password="{reset-password}" \
    --authurl={authurl} \
    --vault={vault} \
    --bfl={bfl} \
    --host={host} \
    --enable-tunnel=true
    ```

    **Example:**
    
    If the Olares ID is `alice123@olares.com`, the Wizard URL is `http://192.168.50.123:30180`, and the selected FRP host is `bb.hongkong.frp.olares.com`, run:

    ```bash
    sudo olares-cli wizard activate alice123@olares.com \
    --mnemonic "abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef" \
    --password="v0kSmyVN" \
    --reset-password="Ab1234@" \
    --authurl=http://192.168.50.123:30180 \
    --vault=http://192.168.50.123:30180/server \
    --bfl=http://192.168.50.123:30180 \
    --host=bb.hongkong.frp.olares.com \
    --enable-tunnel=true
    ```

## Next step

You can now log into Olares using your Olares ID and the login password you specified in `reset-password`.

## Learn more

- [Create an Olares ID](../get-started/create-olares-id.md)
- [Olares CLI](../../developer/install/cli/olares-cli.md)
