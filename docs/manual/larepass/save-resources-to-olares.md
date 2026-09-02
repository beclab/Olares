---
outline: [2, 3]
description: Use Fetch in LarePass to save files, torrents, Hugging Face models, and supported online resources directly to Olares.
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, Fetch, torrent, magnet link, Hugging Face, remote download
---

# Save resources to Olares

Fetch in LarePass saves resources from the internet directly to Olares. Add a link or torrent file, choose a destination, and let Olares handle the transfer in the background. This feature is available only in the LarePass desktop client.

Use this guide when you want to save a remote resource to Olares. To download a file already stored on Olares to your computer, see [Download files](manage-files.md#download-files).

:::warning Check content permissions
Make sure you have the right to download the requested content, such as content you own, content in the public domain, or content you have permission to download. You are responsible for complying with applicable copyright laws and the source website’s terms of service.
:::

## Create a Fetch task

:::info Fetch audio or video resources
To download media from supported websites, you need to install the `yt-dlp` component first. In LarePass, go to **Settings** > **Downloader** and click **Install in Olares Market** under **yt-dlp**. 

Torrents, Hugging Face, and general links do not require this.
:::

1. Open the LarePass desktop client.
2. Go to **Transfer** > **Fetch**, then select **Save to Olares**.
3. In the **Add task** window, do one of the following:
   - Paste an HTTP, HTTPS, magnet, or Hugging Face address.
   - Select the upload area to choose a `.torrent` file, or drag the file into the area.
4. After you paste a link or upload a torrent file, LarePass parses it and automatically opens the task details. Check the detected resource and selected files.
5. Under **Save to**, choose where to save the resource on Olares.
6. Select **Download** to create the task.

The task appears under **Transfer** > **Fetch**, where you can follow its progress.

## Supported sources

| Source | What to add |
|--------|-------------|
| Torrent | A `.torrent` file or magnet link |
| Web link | A direct HTTP or HTTPS file link |
| Hugging Face | A public model or repository address |
| Audio or video | A URL from a supported media website; requires `yt-dlp` |

## Manage Fetch tasks

Go to **Transfer** > **Fetch** and select a tab:

- **Fetching**: Pause, resume, or cancel a task.
- **Completed**: Check task details, open the containing folder, or delete a task. For a torrent task, you can also stop or resume seeding.

## Things to know

- Links must be complete, valid, and publicly accessible. Fetch saves unsupported website URLs as HTML files. Use a direct file link when available.
- Torrent downloads depend on available peers. If a task has no progress, check again later or use another source.
