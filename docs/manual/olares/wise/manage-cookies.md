---
outline: [2, 3]
description: Manage cookies for Wise so it can access protected websites and feeds. Learn why cookies matter, supported formats, how to obtain them, and ways to upload and maintain cookies.
---

# Manage cookies for Wise

Some websites require a login to access their content. To allow Wise to fetch content from these protected sources, you may need to upload cookies so it can authenticate the connection.

This grants Wise permission to pass login checks, ensuring that features like **Save to library** or **Subscribe to RSS feed** work securely on restricted sites.

Wise supports the following cookie formats:

- **Netscape**
- **JSON**
- **Header String**

This guide explains how to manage cookies in Olares for Wise.

## Upload cookies

Olares provides flexible options for uploading cookies, depending on whether you want to capture them instantly while browsing or manage them manually in batch.

### Upload via LarePass browser extension

This is the easiest way to handle cookies. When you are browsing a site that requires a login, [LarePass browser extension](https://www.olares.com/larepass) can capture and upload the cookies for the current page in one flow.

1. Navigate to the target website and make sure you are logged in.
2. Open the **LarePass** browser extension and select **<i class="material-symbols-outlined">hard_drive</i>**.
3. Click **Upload** to upload the cookies for the current page.

![Upload cookies via LarePass extension](/images/manual/olares/wise-upload-cookie-larepass.png#bordered){width=50%}

After upload, Wise will use these cookies when you **Save to library** or **Subscribe to RSS feed** for that site.

### Upload manually in batch

Use this method when you already have cookie strings (Netscape, JSON, or Header String) for one or more sites and want to upload them manually.

#### Step 1: Get cookies for batch upload

To upload cookies manually, you first need cookie strings that you can paste. The simplest way to obtain them is to use a cookie management extension (for example, `cookie-editor`) and export cookies for the current site in formats Wise understands.

Typical steps:

1. Install a cookie management extension from the Chrome Web Store.
2. Log in to the target website in your browser.
3. Open the cookie extension on that page.
4. Export cookies for the site in one of the following formats:
   - **Netscape**
   - **JSON**
   - **Header String**

You can use the exported cookie text in the next step, or reuse any existing cookie strings you already have in one of the supported formats.

#### Step 2: Choose an upload location

You can upload cookie strings either directly to Wise or to Integration, depending on whether you want them to be app-specific or system-wide.

<Tabs>
<template #Upload-to-Wise>

**Best for cookies that are only needed within the Wise app.**

1. In Wise, click <i class="material-symbols-outlined">settings</i> **Settings** in the bottom left corner.
2. Select **Preferences**.
3. Enable **Batch cookie upload**.
4. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar.
5. Select **Upload Cookie**.
6. Paste the cookie text (Netscape, JSON, or Header String) into the input box.
7. Click **Confirm**.

![Upload cookies to Wise in batch](/images/manual/olares/wise-upload-cookie-in-pre.png#bordered)

Cookies uploaded this way are stored for Wise and will be used whenever Wise accesses matching domains.

</template>

<template #Upload-to-Integration>

**Best for managing cookies centrally so multiple Olares apps can reuse them.**

1. Open **Settings** from the Launchpad.
2. Go to **Integration** > **Manage Your Cookies**.
3. Click **Import Cookie** (for first-time use) or the <i class="material-symbols-outlined">add_circle</i> icon in the top-right corner (if cookies already exist).
4. Paste your cookie text (Netscape, JSON, or Header String) and click **Confirm**.

Cookies stored here are available at the system level, and Wise will automatically use them when accessing matching domains.
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