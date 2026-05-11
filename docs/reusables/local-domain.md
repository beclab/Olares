---
search: false
---
<!-- Reusable .local domain content. Include by named region. -->

<!-- #region local-domain-overview -->
When your device is on the same local network as Olares, you can use a `.local` domain to reach your services so traffic stays on your LAN.

<!-- #region local-domain-url-format -->
Use a multi-level `.local` hostname that mirrors your standard URL. This format works with Olares system apps and community apps.

:::tip
Use `http://`, not `https://`, with the `.local` URL.
:::

**Standard URL**
```text
https://<entrance_id>.<username>.olares.com
```
**Local URL**
```text
http://<entrance_id>.<username>.olares.local
```
<!-- #endregion local-domain-url-format -->
<!-- #endregion local-domain-overview -->

### Windows

<!-- #region windows-local-domain -->
On Windows, `.local` hostnames are not resolved by default. Use the LarePass desktop app to add the necessary entries to your hosts file so multi-level `.local` URLs resolve to your Olares device.

1. Open the LarePass app, click your avatar, then **Settings**.
2. Scroll to **Enable local service domain** and click **Add**. LarePass will update your hosts file automatically.

   ![Enable local service domain](/images/one/larepass-win-update-hosts.png#bordered)
3. When the update completes, a success message appears. If a command line window opens, you can close it.
4. (Optional) To verify the changes to the hosts file:

   a. Go to `C:\Windows\System32\drivers\etc\`.

   b. Open the `hosts` file in a text editor. You should see the `.local` entries that LarePass added.

   ![Hosts file updated by LarePass](/images/one/larepass-updated-hosts.png#bordered)
<!-- #endregion windows-local-domain -->

<!-- #region local-domain-faq -->
### Why doesn't the .local domain work in Chrome on macOS?

Chrome may block local URLs if macOS has not granted it local network access.

1. Open the Apple menu and go to **System Settings**.
2. Go to **Privacy & Security** > **Local Network**.
3. Find **Google Chrome** and **Google Chrome Helper** and turn their toggles on.
4. Restart Chrome and try the `.local` URL again.

![Enable local network](/images/manual/larepass/mac-chrome-local-access.png#bordered){width=400}

### Why does the app show "connection not secure" or fail to load in Chrome?

Chrome sometimes forces HTTPS for `.local` hostnames, which is not supported.

Use `http://` explicitly at the start of the URL (e.g. `http://desktop.<username>.olares.local`). On your home network, this unencrypted local connection is expected and keeps the `.local` domain working.

![Incorrect local address](/images/manual/get-started/incorrect-local-address.png#bordered)

### Why does the iframe flicker when I open a .local URL in Safari?

Safari applies stricter handling to `.local` (and other non-HTTPS) content in iframes, which can make the iframe flicker or reload. Enabling two options in **Privacy** settings fixes it.

To fix it:

1. Open **Safari** and go to **Settings**.
2. Open the **Privacy** tab.
3. Enable the two options:
   - Prevent cross-site tracking
   - Hide IP address from trackers

   ![Safari Privacy settings for .local](/images/manual/get-started/safari-privacy-settings.png#bordered){width=70%}
4. Reload the `.local` page.
<!-- #endregion local-domain-faq -->
