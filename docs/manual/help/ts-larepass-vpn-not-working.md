---
outline: [2, 3]
description: Troubleshoot LarePass VPN connection issues on macOS, Windows, and mobile devices.
---
# LarePass VPN not working

Use this guide when LarePass VPN does not connect on macOS, Windows, or mobile devices.

## Condition

- On any client, the VPN stays on **Connecting** for a while and then turns off automatically.
- On macOS:
  - The VPN toggle does not respond, or the VPN status stays on **Connecting**.
  - LarePass VPN used to work on this device but now fails to connect or drops immediately.
- On Windows, the VPN toggle does not respond or the VPN cannot be enabled.

## Cause

Depending on the symptom, the issue may be caused by one of the following:

- **Incorrect system time**: If the system time on your LarePass client device is incorrect, the VPN handshake may fail and the VPN may turn off after remaining on **Connecting** for a while.
- **macOS extension issue**: LarePass requires a system-level network extension and VPN configuration. If the setup was incomplete, or if the extension became stuck or corrupted, macOS may block the VPN tunnel.
- **Windows antivirus block**: Third-party antivirus or security software may mistakenly flag LarePass as suspicious, preventing the VPN service from starting.

## Solution

Follow the section that matches the symptom on your device.

### Sync device time

1. Open the date and time settings on the device where you are using LarePass.
2. Turn on automatic time synchronization.
   - **Mobile**: Check your phone's date and time settings.
   - **Desktop**: Check your computer's date and time settings.
3. Reopen LarePass and enable the VPN connection again.

### macOS: Reset the network extension

Reset the network extension and complete the full setup flow to restore the VPN.

:::info
Depending on your macOS version, the UI might look slightly different.
:::

1. Open **System Settings**, search for "Extension", and select **Extensions**.
2. Scroll to the **Network Extensions** section and click <span class="material-symbols-outlined">info</span> to view loaded extensions.
   
   ![Network Extensions section in System Settings](/images/manual/help/ts-vpn-network-extensions.png#bordered){width=60%}

3. Find **LarePass**, click the three dots (**...**), and select **Delete Extension**.
4. Confirm the uninstallation.
5. Restart your Mac.
6. Open the LarePass desktop client and re-enable the VPN.
7. Complete the system prompts to restore the extension and VPN configuration:

   a. When macOS prompts to add the LarePass network extension, click **Open System Settings**.
   
   ![Prompt to add LarePass network extension](/images/manual/help/ts-vpn-add-network-extension.png#bordered){width=30%}

   b. Toggle on **LarePass**.
   
   ![Toggle on LarePass network extension](/images/manual/help/ts-vpn-toggle-on-network-extension.png#bordered){width=60%}

   c. When prompted to add VPN configurations, click **Allow**.
   
   ![Prompt to add VPN configuration](/images/manual/help/ts-vpn-add-vpn-configuration.png#bordered){width=30%}

### Windows: Add LarePass to the allowlist

:::info LarePass blocked on first launch
If your antivirus blocked LarePass when you first opened it after installation, allow the app in your security software before following the steps below.
:::

1. In your antivirus or security software, open the **Allowlist**, **Exclusions**, or **Exceptions** settings.
2. Add the main LarePass executable or installation directory to the allowlist. Common locations include:
   - `C:\Users\<your-username>\AppData\Local\LarePass\`
   - `C:\Program Files\LarePass\`
3. Apply the changes and restart your antivirus or security software if required.
4. Quit and reopen the LarePass desktop client.
5. Try enabling **VPN connection** again from within LarePass.