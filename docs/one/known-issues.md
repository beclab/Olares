---
description: This page documents the known issues and unexpected behaviors you might encounter when using Olares One, along with their corresponding solutions or workarounds.
---

# Known issues

Use this page to identify and troubleshoot currently known issues with your Olares One device. We regularly update this list with temporary workarounds and permanent fixes as they become available.

## Olares One initial setup fails at 9%

Olares One fails during the initial setup process with the installation stopping at around 9% and prompting you to uninstall or reinstall.

During startup, your Olares One automatically synchronizes its clock with the internet to generate security certificates. This background update usually happens instantly. However, if the time check is delayed, the device might use its factory default time (UTC+8, Beijing Time). This creates a "future" timestamp on its security certificates, which forces the activation to fail. 

### Solution

Uninstall the incomplete installation and reactivate the device.

#### Step 1: Attempt SSH connection

Try this method first if you do not already have a monitor and keyboard connected to your Olares device.

1. Get the local IP address of Olares One.

    a. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.
    ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)
    
    b. Tap the Olares One device card.

    c. Scroll down to the **Network** section and note the **Intranet IP**.
2. Check SSH password in Vault.

    a. Tap **Vault** in the LarePass app. When prompted, enter your local password to unlock.

    b. In the top-left corner, tap **Authenticator** to open the side navigation, and then tap **All vaults** to display all saved items.
        ![Switch Vault filter](/images/one/ssh-switch-filter.png#bordered)

    c. Find the item with the <span class="material-symbols-outlined">terminal</span> icon and tap it to reveal the password.
        ![Check saved SSH password in Vault](/images/one/ssh-check-password-in-vault.png#bordered)

3. Connect via SSH.
    
    a. Open a terminal on your computer.

    b. Type the following command, replace `<local_ip_address>` with the Intranet IP, and then press **Enter**:
    
    ```bash
    ssh olares@<host_ip_address>
    ```
    c. When prompted, type the SSH password, and then press **Enter**.

    e. If the connection is successful, skip to [Step 3](#step-3-run-the-uninstall-command).

#### Step 2: Log in locally

When the SSH access is unavailable, log in to the device locally using a monitor and keyboard.

1. Connect a monitor and keyboard to your Olares One. A text-based login prompt is displayed on your screen automatically:

    ```text
    olares login:
    ```

3. Type the username `olares` and press **Enter**.
4. Type the same SSH password obtained in **Step 1** and press **Enter**.

#### Step 3: Run the uninstall command

1. Once logged in, type the following command and press **Enter**. This command removes all installed components and data, restoring the device to the unactivated state.

    ```bash
    sudo olares-cli uninstall
    ```
2. Wait until the uninstallation is completed and the device automatically reboots.

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

4. Find your Olares One from the list of available devices, and then tap **Install now** on it. The installation should now proceed and complete successfully.
5. When installation finishes, try to install and activate Olares OS again.
