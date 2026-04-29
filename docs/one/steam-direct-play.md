---
outline: [2, 3]
description: Learn how to play Steam games directly on Olares One by connecting a monitor, keyboard, and mouse.
head:
  - - meta
    - name: keywords
      content: Steam, Local Gaming, Direct Play, Linux Gaming
---

# Play Steam games locally on Olares One <Badge type="tip" text="15 min" />

Connect a monitor, keyboard, and mouse to your Olares One to play Steam games directly on the device.

## Prerequisites

**Hardware** <br>
- Olares One connected to a stable network (Ethernet recommended).
- Monitor, keyboard, and mouse connected to the Olares One.
- Sufficient disk space to download games.

**Software** <br>
- A valid Steam account.

## Step 1: Install Steam Headless

1. Open Market, and search for "Steam".
2. Click **Get**, then **Install**.
   ![Install Steam Headless](/images/manual/use-cases/steam-install-steam-headless1.png#bordered)

3. A prompt will appear asking you to set environment variables:
   - `SUNSHINE_USER`: Set the username for Sunshine access.
   - `SUNSHINE_PASS`: Set a secure password.
4. Wait for the installation to complete.

## Step 2: Install the Steam client

1. Open Steam Headless and click **Connect**.
   ![Connect to Steam](/images/manual/use-cases/steam-connect-to-steam.png#bordered)

2. The Steam client will automatically begin downloading and installing.
   ![Install Steam](/images/manual/use-cases/steam-install-steam.png#bordered)
   ![Update Steam](/images/manual/use-cases/steam-update-steam.png#bordered)

3. When installation completes, sign in with your Steam account.
   ![Sign in to Steam](/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

## Step 3: Download and play games

1. In Steam, go to **Library** to view your purchased games.
2. Select a game you want to play and click **Install**.
3. Wait for the download to complete, then click **Play**.

## FAQs

### Keyboard or mouse not responding on first connection or after Steam restarts
If your keyboard or mouse does not respond when first connected, or after Steam restarts, try unplugging the device and plugging it back in.

This issue may occur when the device is already connected before Steam starts but is not detected during startup. Reconnecting the device may help the keyboard or mouse become available again.

### Mouse not responding on first connection
If your mouse doesn't respond when first connected, try unplugging it and plugging it back in. This is a common occurrence with USB peripherals on initial connection.

### Why does my monitor show the Steam interface even when I'm not playing?
Olares One usually displays a terminal prompt when connected to a monitor. However, running the Steam application activates a graphical interface that takes over the display.

To return the monitor to the standard terminal view, stop the Steam application via **Market** or **Settings**.

## Resources
- [Stream Steam games to any device](steam-stream.md)
