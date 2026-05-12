---
outline: [2, 3]
description: Technical guide for installing and activating an Olares device using Olares CLI.
---

# Activate an Olares device using the Olares CLI

This tutorial walks you through activating an Olares device (e.g., Olares One) using the Olares CLI tool. The process assumes the device is freshly unboxed and has not been installed or activated.

## Learning objectives

In this tutorial, you will learn how to:

- Download and extract the Olares CLI tool.
- Install the Olares OS on a new device.
- Retrieve a Fast Reverse Proxy (FRP) host for remote access.
- Run the activation command to configure your device.

## Prerequisites

  Before you begin, make sure the following requirements are met:

  - You can access the Olares device directly with a keyboard and monitor, or via
  SSH.
  - The device has internet access to download packages, query FRP servers, and
  complete activation.
  - You can run commands as the root user, or prepend commands with `sudo`.
  - You have created an Olares ID using the LarePass app, and have [backed up your
   12-word mnemonic phrase](../larepass/back-up-mnemonics.md).

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

## Step 3: Install Olares

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
    | `olares-id` | The Olares ID you created in LarePass, <br>for example
  `alice123@olares.com`. |
    | `mnemonic` | The 12-word mnemonic phrase of your Olares ID. |
    | `password` | The default Olares login password from Step 3. |
    | `reset-password` | A new login password to replace the default one. |
    | `authurl` | The Wizard URL from Step 3. |
    | `vault` | The Wizard URL from Step 3, followed by `/server`.  |
    | `bfl` | The Wizard URL from Step 3. |
    | `host` | The FRP host address from Step 2. |
    | `enable-tunnel` | Set to `true` to enable tunnel mode. |

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
