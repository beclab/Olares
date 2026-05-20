---
outline: [2, 3]
description: Run Karakeep on Olares to save links, notes, images, and PDFs in one self-hosted workspace. Access your collection from a phone and add AI auto-tagging with a local model.
head:
  - - meta
    - name: keywords
      content: Olares, Karakeep, Hoarder, bookmark manager, self-hosted, AI auto-tag, mobile app, Ollama, video download
app_version: "1.0.4"
doc_version: "1.0"
doc_updated: "2026-05-15"
---

# Save and organize bookmarks with Karakeep

Karakeep (formerly Hoarder) is a self-hosted bookmark and content management app that stores links, notes, images, and PDFs in one place. It automatically fetches page metadata, indexes content for full-text search, supports shared lists, and can auto-tag entries with a local AI model.

Running Karakeep on Olares keeps your saved content and AI processing entirely on your device.

## Learning objectives

In this guide, you will learn how to:

- Install Karakeep on Olares.
- Save a bookmark from the web interface.
- Sign in to Karakeep from your phone.
- Enable AI auto-tagging with a local model.
- Enable video downloads for bookmarked links.
- Configure SMTP for email invitations.

## Install Karakeep

1. Open Market and search for "Karakeep".

   ![Karakeep in Market](/images/manual/use-cases/karakeep.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, open Karakeep from the Launchpad and create the first account. The first user to register becomes the Karakeep administrator.

## Save your first bookmark

Karakeep automatically fetches the page title, content, and metadata when you paste a URL.

1. Open Karakeep from the Launchpad.
2. Paste any URL into the input box at the top of the dashboard, and then click **Save**. Karakeep saves the link and starts fetching the page in the background.

   ![Save a URL in Karakeep](/images/manual/use-cases/karakeep-save-url.png#bordered)

For the full feature set, including notes, lists, and collaboration, see the [Karakeep documentation](https://docs.karakeep.app/).

## Access Karakeep from mobile devices

The Karakeep mobile app lets you save images and links from your phone directly into your Olares-hosted instance. To connect, you need a reachable endpoint and an API key.

### Install the Karakeep mobile app

Install the Karakeep app from the App Store (iOS) or Google Play (Android). You can find direct download links on the Karakeep page in Market.

![Karakeep mobile app download](/images/manual/use-cases/karakeep-mobile-stores.png#bordered)

### Allow mobile app access

By default, Karakeep is private to your Olares account. To let the mobile app sign in, change the authentication level.

1. Open Settings, and then go to **Applications** > **Karakeep** > **Entrances** > **Karakeep**.
2. Set **Authentication level** to one of the following, and then click **Submit**:

   - **Internal**: Requires LarePass VPN. Recommended for personal use.
   - **Public**: Accessible without LarePass VPN, but exposes Karakeep to the internet and uses FRP traffic.

   ![Set Karakeep authentication level](/images/manual/use-cases/karakeep-auth-level.png#bordered)

:::warning Public access
Use a strong account password and keep your API key secure.
:::

### Get your Karakeep endpoint URL

On the same **Entrances** page, copy the endpoint URL shown for Karakeep. For example:

```text
https://abc123.{username}.olares.com
```

Save this URL for later.

### Generate an API key

1. Return to Karakeep in your browser.
2. Click your profile icon, and then select **User Settings**.
3. Go to the **API Keys** tab, and then click **New API Key**.
4. Enter a name for the key (for example, `mobile`), and then click **Create**. Copy the generated key. You only see it once.

   ![Generate API key in Karakeep](/images/manual/use-cases/karakeep-api-key.png#bordered)

### Sign in from the mobile app

1. If you selected **Internal** in the authentication step, open LarePass on your phone and turn on the VPN.

   ![Enable LarePass VPN on mobile](/images/manual/get-started/larepass-vpn-mobile.png#bordered)

2. Open the Karakeep app, and then select **Use API key instead**.
3. Enter your details:

   - **Server Address**: Paste the endpoint URL you copied earlier.
   - **API Key**: Paste the API key you generated.

4. Tap **Sign in**. You can now save images and links from your phone.

   ![Karakeep mobile sign-in](/images/manual/use-cases/karakeep-mobile-signin.png#bordered)

Karakeep also supports browser extensions and other clients. For the full list, see [karakeep.app/apps](https://karakeep.app/apps/).

## Auto-tag bookmarks with a local model

Karakeep can use a local model hosted on Olares to generate tags for your saved content. This guide uses Qwen3.5 27B Q4_K_M (Ollama) as the text model.

### Prerequisites

- A local model app installed from Market with the model fully downloaded. For image tagging, also install [Ollama](ollama.md) and pull a vision model such as `llava`.

### Get the model endpoint and name

1. Open your model app from the Launchpad. The model name appears on the page (for example, `qwen3.5:27b-q4_K_M`). Note it for later.

   ![Get model name](/images/manual/use-cases/deerflow2-get-model-name.png#bordered)

2. Open Settings, and then go to **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)**.
3. Under **Shared entrances**, select the model app to view the endpoint URL.

   ![Get shared endpoint](/images/manual/use-cases/ollama-shared.png#bordered){width=70%}

4. Copy the shared endpoint. For example:

   ```text
   http://94a553e00.shared.olares.com
   ```

### Connect Karakeep to the local model

1. Open Settings, and then go to **Applications** > **Karakeep** > **Manage environment variables**.
2. Click <i class="material-symbols-outlined">edit_square</i> next to each variable, enter the value, and then click **Confirm**:

   - **OLLAMA_BASE_URL**: The shared endpoint URL from the previous step. For example:
     ```text
     http://94a553e00.shared.olares.com
     ```
   - **INFERENCE_TEXT_MODEL**: The model name from the previous step. For example:
     ```text
     qwen3.5:27b-q4_K_M
     ```
   - **INFERENCE_IMAGE_MODEL** (optional): The name of a vision model pulled in Ollama. Set this only if you want Karakeep to tag images.

3. Click **Apply**, and wait for Karakeep to restart.

   ![Manage Karakeep environment variables](/images/manual/use-cases/karakeep-manage-env-vars.png#bordered)

:::info Text-only tagging
If you only need text tagging, leave `INFERENCE_IMAGE_MODEL` empty.
:::

### Generate tags for existing bookmarks

1. Open Karakeep, click your profile icon, and then select **User Settings**.
2. Go to **AI Settings**. The default prompts work for most cases.
3. Click your profile icon, select **Admin Settings**, and then go to **Background Jobs**.
4. In **Inference Jobs**, click **Regenerate AI Tags for All Bookmarks**. The queue size increases.

   ![Regenerate tags in Karakeep](/images/manual/use-cases/karakeep-regenerate-tags.png#bordered)

5. After the queue empties, refresh the dashboard. Tags appear on your bookmarks.

   ![Generated tags](/images/manual/use-cases/karakeep-generated-tags.png#bordered)

## Enable video downloads

Karakeep can use `yt-dlp` to automatically download videos from saved links to Olares. After a video is downloaded, you can watch it offline in Karakeep or download the video file to your local computer from the bookmark attachments.

1. Open Settings, and then go to **Applications** > **Karakeep** > **Manage environment variables**.
2. Click <i class="material-symbols-outlined">edit_square</i> next to `CRAWLER_VIDEO_DOWNLOAD`, set the value to `true`, and then click **Confirm**.
3. Click **Apply**, and wait for Karakeep to restart.
4. After Karakeep restarts, add a video link as you would any other bookmark. 

   Karakeep downloads the video to Olares in the background. This may take some time depending on the video size and network conditions.
   
5. On the bookmark card, click <i class="material-symbols-outlined">pan_zoom</i> to open the bookmark details.
   
   After the video is downloaded to Olares, **Video** appears in the content type drop-down at the top of the detail page.

6. Optional: To download the video file to your local computer, find **Video** under **Attachments** in the right sidebar, and then click <i class="material-symbols-outlined">download</i> next to it.

   ![Video download](/images/manual/use-cases/karakeep-video-download.png#bordered)

:::warning Download failures
Video downloads can fail due to network issues or anti-bot protections on the source site. If a download fails, see [Why does my video download fail](#why-does-my-video-download-fail) for troubleshooting.
:::

## Configure SMTP for email invitations

To invite users by email, configure the system SMTP settings in Olares. Karakeep uses the system SMTP settings by default.

1. Open Settings, and then go to **Advanced** > **System environment variables**.
2. Set the following system SMTP variables according to the SMTP settings provided by your mail provider:

   - `OLARES_USER_SMTP_ENABLED`: Set to `true`.
   - `OLARES_USER_SMTP_SERVER`: SMTP server host, such as `smtp.gmail.com`.
   - `OLARES_USER_SMTP_PORT`: SMTP port, such as `587` for TLS.
   - `OLARES_USER_SMTP_USERNAME`: SMTP username. For Gmail, use your full Gmail address.
   - `OLARES_USER_SMTP_PASSWORD`: SMTP password or app password.
   - `OLARES_USER_SMTP_FROM_ADDRESS`: Sender email address.

   For most mail providers, these variables are enough. If your provider requires specific SSL/TLS settings, also configure the related SMTP security variables as instructed by the provider.

3. Click **Apply**. Karakeep restarts for the new SMTP settings to take effect.

After Karakeep restarts, you can send email invitations from the admin user management page.

## FAQs

### Why does my video download fail?

#### Cause

Video sites apply anti-bot measures such as CAPTCHAs, IP blocking, and headless browser detection. These measures can block `yt-dlp`, the tool Karakeep uses for video downloads. Network instability can also cause partial downloads.

#### Solution

1. Open Control Hub, click **Browse**, select your Karakeep project, expand **Deployments**, and select the running pod.
2. In the right panel, scroll down to the **Containers** section, locate the `karakeep` container, then click <i class="material-symbols-outlined">article</i> next to it. 

   ![Check container logs](/images/manual/use-cases/karakeep-container-logs.png#bordered)


3. Search for `[VideoCrawler]` to find the specific error.

   For example, if you see an error like the following, the source site has blocked the download request:

   ```text
   ERROR: [youtube] mZp8yCueuKU: Sign in to confirm you're not a bot
   ```

4. If your network uses a dynamic IP address, wait for the IP address to change and try again.

Some failures are caused by restrictions on the source website. In these cases, Karakeep may not be able to download the video until the restriction is lifted or `yt-dlp` supports the updated site behavior.

## Learn more

- [Karakeep documentation](https://docs.karakeep.app/): Official feature reference, API documentation, and third-party client integrations.
- [Download and run local AI models via Ollama](ollama.md): Install Ollama to host a vision model for image tagging.
- [Set up Open WebUI for local AI chat](openwebui.md): Reference workflow for shared model endpoints on Olares.