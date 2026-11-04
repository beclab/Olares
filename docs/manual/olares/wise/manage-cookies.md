---
outline: [2, 3]
description: Upload, update, or remove cookies that may help Wise access sites and feeds requiring sign-in.
head:
  - - meta
    - name: keywords
      content: Olares, Wise, cookies, protected sites, batch upload, LarePass extension
---

# Manage cookies for Wise <Badge type="tip" text="Olares 1.12.8+" />

If Wise cannot save, subscribe to, or download content from a site that requires sign-in, upload cookies for that site to let Wise access it. Cookies may expire, and access is not guaranteed.

This guide covers how to upload, update, and remove cookies for Wise.

## Upload cookies

You can upload cookies while browsing with LarePass or paste exported cookies manually.

### Upload via LarePass browser extension

This is the easiest way to handle cookies. When you are browsing a site that requires a login, [LarePass browser extension](https://www.olares.com/larepass) can capture and upload the cookies for the current page in one flow.

1. Navigate to the target website and make sure you are logged in.
2. Open the LarePass browser extension and select the "Collect" icon.
3. Click **Upload cookies** to upload the cookies for the current page.

![Upload cookies via LarePass extension](/images/manual/olares/wise-cookie-required.png#bordered){width=40%}

After upload, try saving or subscribing again. If access still fails, check that the cookies are current and match the site.

### Upload manually in batch

You can manually paste cookies in Netscape, JSON, or Cookie header format. This is useful if you use browser extensions like [Cookie-Editor](https://chromewebstore.google.com/detail/cookie-editor/iphcomljdfghbkdcfndaijbokpgddeno?utm_source=ext_app_menu) to export credentials from specific websites.

You can upload cookie strings via the Wise app interface or directly through the system Settings.

<Tabs>
<template #Upload-to-Wise>

This method allows you to upload cookies directly without leaving the Wise application.

:::info
Wise supports uploading only. To view, manage, or delete these cookies later, open **Olares Settings** > **Integrations** > **Manage cookies**.
:::

1. In Wise, click the settings icon in the bottom-left corner and select **Preferences**.
2. Under **Upload settings**, select the **Enable batch cookie upload** checkbox.
3. Click the plus icon in the bottom-left corner and select **Upload cookies**.
4. In the **Upload cookies** dialog, select **Netscape**, **JSON**, or **Cookie header** and paste the exported cookies.
5. Click **Confirm**.

    ![Upload cookies to Wise in batch](/images/manual/olares/wise-upload-cookie-in-pre1.png#bordered)

Wise can use the uploaded cookies for requests to matching domains.

</template>

<template #Upload-to-Integration>

This method allows you to upload, view, and manage your cookies in the same interface.

1. Open **Olares Settings**.
2. Go to **Integrations** > **Manage cookies** to open **Cookie management**.
3. Click **Import cookies** to open the **Upload cookies** dialog.
4. Select **Netscape**, **JSON**, or **Cookie header**, paste the exported cookies into the input box, and click **Confirm**.

Wise can use these cookies when accessing matching domains.
</template>
</Tabs>

## Update or remove cookies

1. Open **Olares Settings** and go to **Integrations** > **Manage cookies** to review the cookies you have uploaded.
2. To update an expired cookie, upload a fresh value using either method above.
3. To remove a cookie you no longer need, find its item and click <i class="material-symbols-outlined">delete</i>.

After removal, Wise may lose access to content requiring those cookies until you upload new ones.
