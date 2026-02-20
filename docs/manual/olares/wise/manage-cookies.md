---
outline: [2, 3]
description: Manage cookies for Wise so it can access protected websites and feeds. Learn why cookies matter, supported formats, how to obtain them, and ways to upload and maintain cookies.
---

# Manage cookies for Wise

Some websites require you to log in to access their content. To allow Wise to fetch content from these protected sources, you may need to upload cookies so Wise can authenticate its requests.

This allows Wise to authenticate access, ensuring that features like **Save to library**, **Subscribe to RSS feed** and **Download file** work securely on restricted sites.

This guide explains how to manage cookies in Olares for Wise.

## Upload cookies

Olares provides flexible options for uploading cookies, depending on whether you want to capture them instantly while browsing or manage them manually in batch.

### Upload via LarePass browser extension

This is the easiest way to handle cookies. When you are browsing a site that requires a login, [LarePass browser extension](https://www.olares.com/larepass) can capture and upload the cookies for the current page in one flow.

1. Navigate to the target website and make sure you are logged in.
2. Open the LarePass browser extension and select the "Collect" icon.
3. Click **Upload** to upload the cookies for the current page.

![Upload cookies via LarePass extension](/images/manual/olares/wise-upload-cookie-larepass.png#bordered){width=40%}

After upload, Wise will use these cookies when you **Save to library** or **Subscribe to RSS feed** for that site.

### Upload manually in batch

You can manually paste cookie strings in Netscape, JSON, or Header String formats. This is useful if you use browser extensions like [Cookie-Editor](https://chromewebstore.google.com/detail/cookie-editor/iphcomljdfghbkdcfndaijbokpgddeno?utm_source=ext_app_menu) to export credentials from specific websites.

You can upload cookie strings via the Wise app interface or directly through the system Settings.

<Tabs>
<template #Upload-to-Wise>

This method allows you to upload cookies directly without leaving the Wise application.

:::info
Wise supports uploading only. To view, manage, or delete these cookies later, you must navigate to the **Settings** > **Integration** > **Manage Your Cookies** page.
:::

1. In Wise, click <i class="material-symbols-outlined">settings</i> **Settings** in the bottom left corner.
2. Select **Preferences**.
3. Enable **Batch cookie upload**.
4. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar.
5. Select **Upload Cookie**.
6. Select the format for your cookie file, and paste the corresponding raw text into the input box.
7. Click **Confirm**.

![Upload cookies to Wise in batch](/images/manual/olares/wise-upload-cookie-in-pre.png#bordered)

Cookies uploaded this way are stored for Wise and will be used whenever Wise accesses matching domains.

</template>

<template #Upload-to-Integration>

This method allows you to upload, view, and manage your cookies in the same interface.

1. Open **Settings** from the Launchpad.
2. Go to **Integration** > **Manage Your Cookies**.
3. Click **Import Cookie** or the <i class="material-symbols-outlined">add_circle</i> icon in the top-right corner to open the **Upload Cookie** dialog.
4. Select the format for your cookie file, paste the corresponding raw text into the input box, and click **Confirm**.

Wise automatically detects and uses these cookies when accessing matching domains.
</template>
</Tabs>

## Manage cookies

You usually do not need to manage cookies frequently, but it can be helpful to review or clean them up from time to time.

### Check cookie status

To view existing cookies and their details:

1. Open **Settings** from the Launchpad.
2. Go to **Integration** > **Manage Your Cookies**.

When needed, you can update them by re-uploading new cookie strings using the same steps as above.

### Remove cookies

If you no longer need access to a particular site, or want to clean up old credentials:

1. Open **Settings** from the Launchpad.
2. Go to **Integration** > **Manage Your Cookies**.
3. Find your target cookie item and click <i class="material-symbols-outlined">delete</i> or **Delete All**.

After removal, Wise will no longer use those cookies and may not be able to access protected content for that site until new cookies are provided.