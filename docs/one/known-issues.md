---
description: This page documents the known issues and unexpected behaviors you might encounter when using Olares One, along with their corresponding solutions or workarounds.
---

# Known issues

Use this page to identify and troubleshoot currently known issues with your Olares One device. We regularly update this list with temporary workarounds and permanent fixes as they become available.

## Olares One initial setup fails at 9%

Olares One fails during the initial setup process with the installation stopping at around 9% and prompting you to uninstall or reinstall.

During startup, the system performs an asynchronous NTP time synchronization before issuing security certificates. While this usually completes instantly, occasional delays can cause a certificate to be issued with a future timestamp. This is especially common if the device has not yet updated from its default shipped timezone of UTC+8, finally causing the activation to fail.

### Workaround

Uninstall the incomplete installation and reactivate the device.

#### Step 1: Attempt SSH connection

Try this method first if you do not already have a monitor and keyboard connected to your Olares device.

1. Get the local **IP** address of Olares One from the **Activate Olares** page on the LarePass app.
    ![IP address displayed on Activate Olares](/images/one/obtain-ip-from-install.png#bordered)
    
2. Open a terminal on your computer.
3. Type the following command, replace `<local_ip_address>` with the above local IP address, and then press **Enter**:
    
    ```bash
    ssh olares@<local_ip_address>
    ```
4. When prompted, type the default SSH password `olares`, and then press **Enter**.
5. If the connection is successful, skip to [Step 3](#step-3-run-the-uninstall-command).

#### Step 2: Log in locally

When the SSH access is unavailable, log in to the device locally using a monitor and keyboard.

1. Connect a monitor and keyboard to your Olares One. A text-based login prompt is displayed on your screen automatically:

    ```text
    olares login:
    ```

2. Type the username `olares` and press **Enter**.
3. When prompted, type the default SSH password `olares`, and then press **Enter**.

#### Step 3: Run the uninstall command

1. Once logged in, type the following command and press **Enter**. This command removes all installed components and data, restoring the device to the unactivated state.

    ```bash
    sudo olares-cli uninstall
    ```
2. Wait until the uninstallation is completed.

#### Step 4: Reinstall and activate using LarePass

:::tip Before reinstallation
To ensure accurate time synchronization, let the device remain powered on for a few minutes before reinstalling, allowing it to automatically calibrate its internal time.
:::

1. Discover and link your Olares One based on your network setup.

    <tabs>
    <template #Set-up-via-wired-LAN>

    a. Ensure your Olares One is connected to your router via Ethernet.

    b. In the LarePass app, tap **Discover nearby Olares**.
    ![Discover nearby Olares](/images/one/discover-nearby-olares.png#bordered){width=90%}  

    </template>

    <template #Set-up-via-Wi-Fi-(Bluetooth)>
    If wired access is not available, use Bluetooth to configure Wi-Fi credentials.

    a. In the LarePass app, tap **Discover nearby Olares**.

    b. Tap **Bluetooth network setup** at the bottom.

    c. Select your device from the Bluetooth list and tap **Network setup**. 

    d. Follow the prompts to connect Olares One to the Wi-Fi network your phone is currently using.

    e. Once connected, return to the main screen and tap **Discover nearby Olares** again.

    </template>
    </tabs>

2. Find your Olares One from the list of available devices, and then tap **Install now** on it. The installation should now proceed and complete successfully.