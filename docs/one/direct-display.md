---
outline: [2, 3]
description: Use Olares One locally with an HDMI display, keyboard, and mouse while keeping your apps and data accessible from a browser.
head:
  - - meta
    - name: keywords
      content: Olares One, HDMI display, direct display, Olares Desktop, Node Display
---

# Use Olares One with a direct display <Badge type="warning" text="Alpha"/>

Connect an HDMI display, keyboard, and mouse to Olares One to access your apps and data directly, without another computer. You can use Olares Desktop on the connected display and in a browser at the same time without either session interrupting the other.

:::warning Alpha feature
This feature is currently in the **Alpha** stage and is not recommended for production environments. It is disabled by default and requires you to install the preview `olares-desktop` package. It may contain performance issues and require additional manual configurations. If you encounter any issues, please report them to the [Olares GitHub repository](https://github.com/beclab/Olares/issues).
:::

## Prerequisites

**Hardware**
- Your Olares One is set up and powered on.
- An HDMI display, keyboard, and mouse connected to Olares One.
- Audio output connected to Olares One if you want to play games locally.
- If Olares OS is activated, a mobile device with the LarePass app installed is required to retrieve the login password from Vault.

**System**
- Olares OS v1.12.7 or later running on the Olares One.

## Step 1: Access Olares One directly

1. Prepare the Olares One host login password.

   - If Olares OS is not activated, use the default password `olares`.
   - If Olares OS is activated, open the LarePass app on your mobile device. In Vault, find the item marked with the <span class="material-symbols-outlined">terminal</span> icon, and tap it to reveal the password.

   :::info Not the same as your Olares Desktop password
   This password logs you in to the Olares One host system. It is different from the password you use to sign in to Olares Desktop in your browser.
   :::

2. On the text-based login prompt displayed on the connected monitor, enter the username `olares`, and then press **Enter**.

   ```text
   olares login:
   ```

3. Enter the host login password you prepared above, and then press **Enter**. For security, characters will not appear on the screen as you type.

:::tip
If you prefer to prepare Olares One remotely, you can use SSH or Control Hub to download and install the package before connecting the display. See [Access Olares One terminal](./access-overview.md) for the available methods.
:::

## Step 2: Download and install the Olares Desktop preview package

1. Download the package:

   ```bash
   cd /tmp && wget https://cdn.olares.com/olares-one/desktop/olares-desktop_0.1.0_20260928_amd64.deb
   ```

2. After the download finishes, install the package:

   ```bash
   sudo bash -c 'apt update && DEBIAN_FRONTEND=noninteractive apt-get install -y -f ./olares-desktop_0.1.0_20260928_amd64.deb'
   ```

## Step 3: Start the local desktop

1. Start the local desktop:

   ```bash
   sudo start-desktop
   ```

2. Wait for the Olares sign-in screen to appear, and then enter your Olares Desktop password to log in.

## Step 4: Switch between top-level windows

Press and hold **Ctrl**, then press **Tab** repeatedly to cycle through the available top-level windows. Release **Ctrl** to open the highlighted window.

The default windows are:
- **Olares Desktop**: The same Olares Desktop interface that you access in a browser. You can install apps from Market and access your apps and data in the same way.
- **Node Display**: A local interface for managing the Olares One device.
- **`olares@olares:~`**: The local terminal.

![Switch between Olares Desktop, Node Display, and the local terminal](/images/one/direct-display-window-switcher.png#bordered)

## Step 5: Use the local desktop

Use the local Olares Desktop just as you would in a browser. For example, you can:

- Open Files to browse and manage your files.
- Open Market to find and install apps.
- Open installed apps from Olares Desktop.

## Step 6: Stop the direct display interface

1. Save your work in any open app.
2. Stop the direct display interface using one of the following methods:

   - To stop it manually, press **Ctrl+Tab** to switch to **`olares@olares:~`**, and run:

     ```bash
     sudo stop-desktop
     ```

   - To stop it automatically, disconnect the HDMI cable.
