---
outline: [2, 3]
description: Manage cookies for Wise so it can access protected websites and feeds. Learn why cookies matter, supported formats, how to obtain them, and ways to upload and maintain cookies.
---

# Manage cookies for Wise

Some websites require cookies for accessing content. Wise relies on cookies to access this protected content.

If the required cookies are missing or expired, actions like **Save to library** or **Subscribe to RSS feed** may not work as expected.

Wise supports the following cookie formats:

- **Netscape**
- **JSON**
- **Header String**

This guide explains how to manage cookies in Olares for Wise.

## Get cookies for batch upload

If you want to upload cookies in batch to Wise or Integration, you need cookie strings that you can paste manually. The simplest way to obtain them is to use a cookie management extension (for example, `cookie-editor`) and export cookies for the current site in formats Wise understands.

Typical steps:

1. Install a cookie management extension from the Chrome Web Store.
2. Log in to the target website in your browser.
3. Open the cookie extension on that page.
4. Export cookies for the site in one of the following formats:
   - **Netscape**
   - **JSON**
   - **Header String**

You can then paste the exported cookie text into Wise (batch upload) or Integration using the methods below.

## Upload cookies

Olares provides three ways to upload cookies, depending on whether you are working with the current page or managing a batch of credentials.

### Upload cookies for current page

Use LarePass extension when you are currently on the site and want Wise to access that specific page. LarePass will both read and upload cookies in one flow.

:::tip Install LarePass browser extension
LarePass extension is currently available for Chrome browsers only. You can download it from the [official page](https://www.olares.com/larepass).
:::

While browsing the website:

1. Open the LarePass browser extension and select **<i class="material-symbols-outlined">inventory_2</i>**.
2. Click **Upload** to upload the cookies for the current page.
![Upload cookies via LarePass extension](/images/manual/olares/wise-upload-cookie-larepass.png#bordered){width=70%}
After upload, Wise will use these cookies when you **Save to library** or **Subscribe to RSS feed** for that page.

### Upload cookies to Wise in batch

Use this method when you already have cookie strings (Netscape, JSON, or Header String) and want to upload them directly to Wise for one or more sites.

1. In Wise, click <i class="material-symbols-outlined">settings</i> **Settings** in the bottom left corner.
2. Select **Preferences**.
3. Enable **Batch cookie upload**.
4. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar.  
5. Select **Upload Cookie**.
6. Paste the cookie text (Netscape, JSON, or header string) into the input box.
7. Click **Confirm**.
![Upload cookies to Wise in batch](/images/manual/olares/wise-upload-cookie-in-pre.png#bordered)

Cookies uploaded this way are stored for Wise and will be used whenever Wise accesses matching domains.

### Upload cookies to Integration

Use this method when you want to manage cookies centrally so that multiple apps (including Wise) can reuse them.

1. Open **Settings** from the Launchpad.
2. Go to **Integration** > **Manage Your Cookies**.
3. Click **Import Cookie** (for first time use) or <i class="material-symbols-outlined">add_circle</i> icon in the top-right corner (when existing cookies exist).
4. Paste your cookie text (Netscape, JSON, or header string) and click **Confirm**.

Cookies stored here are available at the system level and can be used by Wise when accessing the corresponding sites.

## Manage cookies

You usually do not need to manage cookies frequently, but it can be helpful to review or clean them up from time to time.

### Check cookie status

You can review added cookies as follows:

1. Open **Settings** from the Launchpad.
2. Go to **Integration** > **Manage Your Cookies**.

Here you can view existing cookies and their details. When needed, you can update them by re-uploading new cookie strings using the same steps as above.

### Remove cookies

If you no longer need access to a particular site, or want to clean up old credentials:

1. Open **Settings** from the Launchpad.
2. Go to **Integration** > **Manage Your Cookies**.
3. Find your target cookie item and click <i class="material-symbols-outlined">delete</i> or **Delete All**.

After removal, Wise will no longer use those cookies and may not be able to access protected content for that site until new cookies are provided.