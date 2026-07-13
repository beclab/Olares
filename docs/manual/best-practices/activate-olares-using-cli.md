---
outline: [2, 3]
description: Technical guide for installing and activating an Olares device using Olares CLI.
---

# Activate an Olares device using the Olares CLI

Activate a new or uninitialized Olares device, such as Olares One, using the Olares CLI tool.

## Learning objectives

In this tutorial, you will learn how to:
- Install Olares on a new device.
- Obtain and run the correct Olares CLI tool for your system version.
- Retrieve a Fast Reverse Proxy (FRP) host for remote access.
- Run the activation command to configure your device.

## Prerequisites

Before you begin, ensure the following requirements are met:

- You can access the Olares device directly with a keyboard and monitor, or via SSH.
- The device has internet access to download packages, query FRP servers, and complete activation.
- You can run commands as the root user, or prepend commands with `sudo`.
- You have created an Olares ID using the LarePass app, and have [backed up your 12-word mnemonic phrase](../larepass/back-up-mnemonics.md).

    ![Fast creation](/images/manual/get-started/create-olares-id-1.12.6.png)

## Step 1: Install Olares

:::info For Olares One hardware
A fresh Olares One device is shipped in an uninstalled state. You must run the installation command to set up Olares first before you attempt activation.
:::

1. Run the install command as root.

    ```bash
    sudo olares-cli install
    ```
2. When prompted to enter the domain name, enter `olares.com`.
3. When prompted for the Olares ID, enter the one you registered in the LarePass app. For example, `alice2026`.
4. Wait for the installation process to finish. The terminal outputs a wizard URL and a default password. Note down these details for the activation step.

    **Example:**
    - **Wizard URL**: The local gateway address, such as `http://192.168.31.127:30180`.
    - **Password**: The default Olares login password.

    ![Wizard URL](/images/manual/get-started/wizard-url-and-login-password1.png)

## Step 2: Prepare the CLI tool

Determine your CLI preparation steps based on your Olares version.

1. Run the following command to check the current Olares version.

    ```bash
    sudo olares-check
    ```

2. Check the version in the output, and select the method based on your version.

    <Tabs>
    <template #v1.12.6-and-later>

    The v1.12.6 and later systems include a default CLI equipped with the activation feature. No additional downloads are necessary. Proceed directly to Step 3.
    </template>
    <template #v1.12.5-and-earlier>

    :::warning Use the standalone CLI
    The built-in `olares-cli` included in v1.12.5 and earlier systems do not have the activation feature. You must download a standalone daily build CLI to perform the activation. Review the following requirements for using the tool:
    - **Do not overwrite system files**: A strict version correspondence exists between the system's built-in `olares-cli`, `olaresd`, and the cluster version. Therefore, never move or copy the downloaded standalone CLI to overwrite the system `/usr/bin/olares-cli` file. Doing so breaks this version chain and impacts future system upgrades.
    - **Execution path differences**: Run `./olares-cli` to execute the standalone version downloaded to the current directory. Do not run `olares-cli` directly, because it executes the built-in system version which lacks the activation feature.
    :::

    1. Download the standalone Olares CLI package.

        ```bash
        curl -sSOL https://cdn.olares.com/common/olares-cli-amd64.8cbdc32.tar.gz
        ```

    2. Extract the downloaded file.

        ```bash
        tar xzf olares-cli-amd64.8cbdc32.tar.gz
        ```

    3. Grant executable permissions to the extracted binary file.

        ```bash
        chmod +x olares-cli
        ```
    </template>
    </Tabs>

## Step 3: Retrieve the FRP list

Find an available FRP host to enable remote access to your device.

1. Run the following command based on your Olares version. Replace `{olares-id}` with your registered Olares ID.

    <Tabs>
    <template #v1.12.6-and-later>

    ```bash
    olares-cli wizard frp {olares-id}
    ```

    **Example:**

     ```bash
    olares-cli wizard frp alice2026@olares.com
    ```
    </template>
    <template #v1.12.5-and-earlier>

    ```bash
    ./olares-cli wizard frp {olares-id}
    ```

    **Example:**

    ```bash
    ./olares-cli wizard frp alice2026@olares.com
    ```
    </template>
    </Tabs>

2. Select a host address from the output list and note it down for the activation step. For example, `bb.hongkong.frp.olares.com`.


## Step 4: Activate Olares

Run the activation command to configure and secure your device. This process connects your Olares ID to the device and configures network tunneling and credentials.

1. Prepare your activation command based on the following required parameters:

    | Parameter | Description |
    |:----------|:------------|
    | `olares-id` | The Olares ID you created in LarePass, for example `alice2026@olares.com`. |
    | `mnemonic` | The 12-word mnemonic phrase of your Olares ID. |
    | `password` | The default Olares login password from Step 1. |
    | `reset-password` | A new login password to replace the default one. |
    | `authurl` | The Wizard URL from Step 1. |
    | `vault` | The Wizard URL from Step 1, followed by `/server`.  |
    | `bfl` | The Wizard URL from Step 1. |
    | `host` | The FRP host address from Step 3. |
    | `enable-tunnel` | Set to `true` to enable tunnel mode. |

2. Based on your Olares version, replace the placeholders in the following command with your specific values, and then run it.

    <Tabs>
    <template #v1.12.6-and-later>

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
    
    If the Olares ID is `alice2026@olares.com`, the Wizard URL is `http://192.168.31.127:30180`, and the selected FRP host is `bb.hongkong.frp.olares.com`, run:

    ```bash
    sudo olares-cli wizard activate alice2026@olares.com \
    --mnemonic "abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef" \
    --password="b8Ln6qbz" \
    --reset-password="Abw1234@" \
    --authurl=http://192.168.31.127:30180 \
    --vault=http://192.168.31.127:30180/server \
    --bfl=http://192.168.31.127:30180 \
    --host=bb.hongkong.frp.olares.com \
    --enable-tunnel=true
    ```
    </template>
    <template #v1.12.5-and-earlier>

    ```bash
    sudo ./olares-cli wizard activate {olares-id} \
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
    
    If the Olares ID is `alice2026@olares.com`, the Wizard URL is `http://192.168.31.127:30180`, and the selected FRP host is `bb.hongkong.frp.olares.com`, run:

    ```bash
    sudo ./olares-cli wizard activate alice2026@olares.com \
    --mnemonic "abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef" \
    --password="b8Ln6qbz" \
    --reset-password="Abw1234@" \
    --authurl=http://192.168.31.127:30180 \
    --vault=http://192.168.31.127:30180/server \
    --bfl=http://192.168.31.127:30180 \
    --host=bb.hongkong.frp.olares.com \
    --enable-tunnel=true
    ```
    </template>
    </Tabs>

3. Wait until the terminal displays a message indicating that the activation finished successfully.

## Next step

You can now log into Olares using your Olares ID and the login password you specified in `reset-password`.

## Learn more

- [Create an Olares ID](../get-started/create-olares-id.md)
- [Olares CLI](../../developer/install/cli/olares-cli.md)
