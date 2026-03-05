---
outline: [2, 3]
description: Troubleshoot LarePass VPN not working on macOS or Windows.
---

# LarePass VPN not working

Use this guide when the LarePass VPN toggle does nothing, the VPN stays stuck in "connecting", or a previously working VPN connection suddenly stops on macOS or Windows.

## Condition

**macOS** 
- Clicking the VPN toggle in the LarePass desktop client does nothing, or the VPN status stays stuck in "connecting".
- LarePass VPN used to work on this device but now fails to connect or drops immediately.
**Windows** 
- Clicking the VPN toggle in the LarePass desktop client does nothing, or the VPN cannot be enabled.

## Cause

-  **macOS**: LarePass VPN requires both a system-level network extension and a VPN configuration to be fully set up. If you skipped or did not complete either step during the initial setup prompt, or if the network extension has become stuck or corrupted, macOS will block LarePass from creating the VPN tunnel.

- **Windows**: Third-party antivirus or security software may mistakenly flag the LarePass desktop client as suspicious, preventing the VPN service from starting.

## Solution

### macOS

Reset the network extension and complete the full setup flow to restore the VPN.

:::info
Depending on your macOS version, the UI might look slightly different.
:::

1. Open **System Settings**, search for "Extension", and select **Extensions**.
2. Scroll to the **Network Extensions** section and click <span class="material-symbols-outlined">info</span> to view loaded extensions.
   ![Network Extensions section in System Settings](/images/manual/help/ts-vpn-network-extensions.png#bordered){width=70%}

3. Find **LarePass**, click the three dots (**...**), and select **Delete Extension**.
4. Confirm the uninstallation.
5. Restart your Mac.
6. Open the LarePass desktop client and re-enable the VPN.
7. Complete the system prompts to restore the extension and VPN configuration:

   a. When macOS prompts to add the LarePass network extension, click **Open System Settings**.
   ![Prompt to add LarePass network extension](/images/manual/help/ts-vpn-add-network-extension.png#bordered){width=40%}

   b. Toggle on **LarePass**.
   ![Toggle on LarePass network extension](/images/manual/help/ts-vpn-toggle-on-network-extension.png#bordered){width=70%}

   c. When prompted to add VPN configurations, click **Allow**.
   ![Prompt to add VPN configuration](/images/manual/help/ts-vpn-add-vpn-configuration.png#bordered){width=40%}

### Windows

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
