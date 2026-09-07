---
search: false
head:
  - - meta
    - name: keywords
      content: Olares, local service domain, .local domain, LAN access, hosts file
---
<!-- Reusable local domain content. Include by named region. -->

<!-- #region local-domain-overview -->
When your device is on the same local network as Olares, you can connect directly over the LAN. You can either keep using your standard `olares.com` URLs or use `.local` URLs.

<!-- #region local-domain-url-format -->
A multi-level `.local` hostname mirrors the standard Olares URL and works with both system and community apps.

:::tip
Use `https://` with the standard URL and `http://` with the `.local` URL.
:::

**Standard URL**
```text
https://<entrance_id>.<username>.olares.com
```
**`.local` URL**
```text
http://<entrance_id>.<username>.olares.local
```
<!-- #endregion local-domain-url-format -->
<!-- #endregion local-domain-overview -->

<!-- #region larepass-local-domains-summary -->
On Windows and macOS, LarePass Desktop can configure direct LAN access for the current computer. It adds both `olares.com` and `olares.local` hostnames to the hosts file and maps them to the Olares LAN IP.

1. Make sure the computer and Olares are on the same LAN.
2. Turn off **VPN connection** in LarePass Desktop.
3. Start the update from either location:
   - Click **Enable local host access** in the lower-left corner.
   - Click your avatar, go to **Settings** > **Enable local service domain**, and click **Enable**.
4. Review the entries in **Edit local service domain**, then click **Update**.

   ![Edit local hosts](/images/manual/tutorials/larepass-edit-local-hosts.png#bordered)

5. Enter the administrator password for your computer and confirm the change.
6. Wait for the **Success** message.


You can then use the standard `https://<entrance_id>.<username>.olares.com` URL or its `http://<entrance_id>.<username>.olares.local` equivalent.

:::info Disable local host access
1. In LarePass Desktop, click your avatar and go to **Settings**.
2. Under **Enable local service domain**, click **Disable**.
3. In the **Disable local service domain** confirmation window, click **Disable** again.

LarePass removes all hosts entries it previously added to the computer.
:::
<!-- #endregion larepass-local-domains-summary -->

<!-- #region windows-local-domain -->
On Windows, use LarePass Desktop to add the required entries to the hosts file so multi-level `.local` hostnames resolve to the Olares LAN IP.

1. Make sure the computer and Olares are on the same LAN.
2. Turn off **VPN connection** in LarePass Desktop.
3. Start the update from either location:
   - Click **Enable local host access** in the lower-left corner.
   - Click your avatar, go to **Settings** > **Enable local service domain**, and click **Enable**.
4. Review the entries in **Edit local service domain**, then click **Update**.
5. Enter the administrator password for your computer and confirm the change.
<!-- #endregion windows-local-domain -->

<!-- #region larepass-local-domains -->
Use **Local service domain** in LarePass Desktop to manage the hosts entries for your Olares. LarePass maps both `olares.com` and `olares.local` hostnames to the Olares LAN IP, so requests from this computer stay on the LAN.

This mode applies only to the current computer and is available only while the computer and Olares are on the same LAN.

:::warning Turn off VPN first
LarePass VPN and local service domains are mutually exclusive. Turn off **VPN connection** before adding or updating the hosts entries.
:::

1. Make sure the computer and Olares are on the same LAN.
2. Turn off **VPN connection** in LarePass Desktop.
3. Start the update from either location:
   - Click **Enable local host access** in the lower-left corner.
   - Click your avatar, go to **Settings** > **Enable local service domain**, and click **Enable**.
4. Review the entries in **Edit local service domain**, then click **Update**.

   ![Edit local hosts](/images/manual/tutorials/larepass-edit-local-hosts.png#bordered)

5. When the password prompt appears, enter the administrator password for your computer and confirm the change.
6. Wait for the **Success** message.

You can now open either URL format shown above.

LarePass prompts you to update the hosts entries when the Olares LAN IP changes or when installing or uninstalling apps changes the required hostnames. Apply the update before using the affected URLs.

:::tip
You do not need to edit the hosts file manually. LarePass manages the entries on both Windows and macOS.
:::

:::info Disable local host access
1. Click your avatar and go to **Settings**.
2. Find **Enable local service domain** and click **Disable**.
3. In the **Disable local service domain** confirmation window, click **Disable** again.

LarePass removes all hosts entries it previously added to the computer. Standard `olares.com` URLs then use their normal network route. Multi-level `.local` URLs may no longer resolve on systems without native support.
:::
<!-- #endregion larepass-local-domains -->

<!-- #region larepass-local-domain-faq -->
### Local service domain

#### Why can't I find Enable local service domain in LarePass?

The option appears only when your computer and Olares are on the same LAN. Check the network connection on both devices, then reopen LarePass.

#### Why can't I enable LarePass VPN?

Local service domain and LarePass VPN cannot be enabled at the same time. Go to **Settings** > **Enable local service domain** and disable the local service domain first.

#### Why is LarePass asking me to update the hosts file again?

The required entries can change when the Olares LAN IP changes or when you install or uninstall apps. Review the entries and click **Update** so all managed hostnames point to the current LAN IP.

#### What happens when I disable the local service domain?

LarePass removes the hosts entries it manages. Standard `olares.com` URLs then use their normal network route, and multi-level `.local` URLs may no longer resolve on systems without native support.
<!-- #endregion larepass-local-domain-faq -->

<!-- #region local-domain-faq -->
### `.local` browser issues

#### Why doesn't the .local domain work in Chrome on macOS?

Chrome may block local URLs if macOS has not granted it local network access.

1. Open the Apple menu and go to **System Settings**.
2. Go to **Privacy & Security** > **Local Network**.
3. Find **Google Chrome** and **Google Chrome Helper** and turn their toggles on.
4. Restart Chrome and try the `.local` URL again.

![Enable local network](/images/manual/larepass/mac-chrome-local-access.png#bordered){width=400}

#### Why does the app show "connection not secure" or fail to load in Chrome?

Chrome sometimes forces HTTPS for `.local` hostnames, which is not supported.

Enter `http://` explicitly at the start of the URL, for example, `http://desktop.<username>.olares.local`.

![Incorrect local address](/images/manual/get-started/incorrect-local-address.png#bordered)

#### Why does the iframe flicker when I open a .local URL in Safari?

Safari applies stricter handling to `.local` and other non-HTTPS content in iframes, which can make an iframe flicker or reload.

1. Open **Safari** and go to **Settings**.
2. Open the **Privacy** tab.
3. Enable **Prevent cross-site tracking** and **Hide IP address from trackers**.

   ![Safari Privacy settings for .local](/images/manual/get-started/safari-privacy-settings.png#bordered){width=70%}
4. Reload the `.local` page.
<!-- #endregion local-domain-faq -->
