---
outline: [2, 3]
description: Stream Steam games from Olares with Steam Headless and Moonlight. Configure local and remote game streaming to phones, computers, and other devices.
head:
  - - meta
    - name: keywords
      content: Olares, Steam Headless, game streaming, Moonlight, self-hosted game streaming, stream Steam games, Steam Headless on Olares
---

# Stream your games with Steam Headless

Want to enjoy gaming powered by Olares? You're all set. With the Steam Headless app, Olares easily transforms into a powerful game streaming server. You can now play your favorite games on any compatible device via Moonlight.

This guide walks you through installing Steam Headless on Olares, configuring the Steam client, pairing the streaming service, and connecting with the Moonlight client to play.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install and set up Steam Headless on your Olares device.
- Configure the Sunshine streaming service.
- Pair your device via Moonlight and stream games locally or remotely.

## Prerequisites

Before you begin, make sure:
- Olares running on a machine equipped with an NVIDIA GPU.
- Moonlight installed on your streaming device. Visit the [Moonlight website](https://moonlight-stream.org/) to download and install the appropriate version.
- A valid Steam account to access your games.
- [LarePass VPN enabled](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your client devices (desktop or mobile) if you plan to stream outside your home network.
:::tip
For local streaming, LarePass VPN is not required.
:::

## Install and configure Steam Headless

Install the app from the Olares Market and then complete the initial setup within the Steam client itself.

### Install Steam Headless

Follow these steps to install and configure Steam Headless:

1. Open the Market, and search for "Steam".
2. Click **Get**, then **Install**.
   ![Install Steam Headless](/images/manual/use-cases/steam-install-steam-headless1.png#bordered)

3. A prompt will appear asking you to configure environment variables. This creates your login credentials for the Sunshine Web UI:
   - `SUNSHINE_USER`: Set the username for Sunshine access.
   - `SUNSHINE_PASS`: Set a secure password.
     :::tip Remember your login credentials
     These are your initial login credentials for Sunshine. You must use them to log in to Sunshine the first time. 
     :::
4. Wait for the installation to complete.

### Complete the initial Steam setup

Once the headless app is running, you need to initialize the Steam client inside it.
1. Open Steam Headless and click **Connect**.
   ![Connect to Steam](/images/manual/use-cases/steam-connect-to-steam.png#bordered)

2. The Steam client will automatically begin downloading and installing.
   ![Install Steam](/images/manual/use-cases/steam-install-steam.png#bordered)
   ![Update Steam](/images/manual/use-cases/steam-update-steam.png#bordered)
3. When installation completes, the Steam login screen appears. Sign in with your Steam account.
   ![Sign in to Steam](/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

:::tip Retry installation upon failures 
If Steam installation or update fails due to network issues, go to the top-left menu in the Steam Headless console and navigate to **Applications** > **Internet** > **Steam** to restart the installation. 
:::
Once Steam is ready, you can connect it to Moonlight through Sunshine.

## Pair Sunshine with Moonlight

Steam Headless uses Sunshine to stream video. You must pair it with the Moonlight app on your playing device.

### Access the Sunshine console

First, open the Sunshine Web UI to prepare for pairing. The address depends on your network setup.

   :::info Sunshine requires HTTPS
   Always add `https://` when opening the Sunshine Web UI. Do not use `http://`, even when connecting through a local IP or `.local` domain.
   :::

1. Open the Sunshine Web UI using one of these methods:

   - **Overlay gateway (recommended for local networks)**

     :::info Overlay gateway requirements
     Use this method only when Olares runs on a native Linux host connected through wired Ethernet. If your setup does not meet these requirements, use the `.local` method below.

     If you cannot turn on the system-level service, ask the Super admin to enable it. For details, see [Manage overlay gateway for applications](../manual/olares/settings/overlay-gateway.md).
     :::

     Overlay gateway gives Steam Headless a local IP, allowing video, audio, and control streams to travel directly over the LAN instead of through `.local` resolution, internal proxies, and cluster network forwarding.

     a. Open **Settings** > **Network** > **Overlay gateway**.

     b. Ensure the system-level service is on, then enable overlay gateway for Steam Headless.

     c. Wait for Steam Headless to return to **Running**. The local addresses assigned to the app appears below it.

        ![View Steam Headless overlay gateway addresses](/images/manual/use-cases/steam-enable-overlay-gateway.png#bordered)

     d. Copy the **Sunshine Web UI** address, add `https://`, and open it in your browser. For example:

        ```plain
        https://192.168.50.117:47990
        ```

   - **Same network (`.local` domain)**

     Copy the URL of your current Steam Headless browser tab, change the hostname to its `.local` form, append port `47990`, and open it in a browser. Either local hostname format works:

     ```plain
     https://139ebc4f0.<your Olares ID>.olares.local:47990
     https://139ebc4f0-<your Olares ID>-olares.local:47990
     ```

     On Windows, import Olares hosts with the LarePass desktop app or use the single-level hostname with hyphens. For details, see [Access Olares services locally](../manual/best-practices/local-access.md).

   - **Different network (LarePass VPN)**: [Enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass), then copy your Steam Headless `.com` URL, append port `47990`, and open it in a browser. For example:

     ```plain
     https://139ebc4f0.<your Olares ID>.olares.com:47990
     ```

2. Once the page loads, sign in using the `SUNSHINE_USER` and `SUNSHINE_PASS` credentials you created earlier.
   ![Sign in to Sunshine](/images/manual/use-cases/steam-sign-in-to-sunshine.png#bordered)
3. Click the **PIN** tab. The page will now wait for a pairing PIN.
   ![PIN on Sunshine](/images/manual/use-cases/steam-pin-on-sunshine.png#bordered)

### Add the host in Moonlight

Next, add Olares as a host in Moonlight. The steps below use the macOS client as an example.

On the same local network, Moonlight might detect Steam Headless automatically. If it does not appear, or if you are connecting remotely, add it manually:

1. Open Moonlight on your streaming device.
2. Click **Add Host**, which looks like a computer with a plus icon.
3. Enter the host address for your network setup. Do not add `http://` or `https://`:

   - **Overlay gateway**: Enter the address shown under **Moonlight HTTP** in the screenshot above, including its port. Do not add a protocol. For example:

     ```plain
     192.168.50.117:47989
     ```

   - **Same network (`.local` domain)**: Enter your `.local` hostname without a port. For example:

     ```plain
     139ebc4f0.<your Olares ID>.olares.local
     ```

   - **Different network (LarePass VPN)**: Enter your `.com` hostname without a port. For example:

     ```plain
     139ebc4f0.<your Olares ID>.olares.com
     ```

4. Click **OK**. A new locked host icon appears.
5. Click the locked icon. Moonlight will display a 4-digit pairing PIN.
   ![Get pairing PIN](/images/manual/use-cases/steam-get-pairing-pin.png#bordered)

### Complete pairing

1. Return to the **Sunshine PIN** page in your browser.
2. Enter the PIN displayed in Moonlight and give your device a name.
   ![Enter pairing PIN](/images/manual/use-cases/steam-enter-pairing-pin.png#bordered)

3. Click **Send**.
4. Upon success, you will see a confirmation message, and the lock icon in Moonlight will disappear.
   ![Host in Moonlight](/images/manual/use-cases/steam-host-in-moonlight.png#bordered)

Once paired, you're ready to start streaming.

## Stream your games
:::tip Optimizing remote play
For the best experience when streaming remotely:
1. Connect your client device to 5GHz Wi-Fi or Ethernet.
2. Ensure **LarePass VPN** is active.
:::
You can stream your games through Moonlight either locally or remotely, depending on your network setup.

The following steps demonstrate local streaming.

1. Open Moonlight on your client device.
2. Select your unlocked host and click the **Steam** icon.  
   ![Steam in Moonlight](/images/manual/use-cases/steam-in-moonlight.png#bordered)
3.  Steam **Big Picture Mode** will launch. Select a game from your library and start playing.

## FAQs

### Why can't I access the Sunshine Web UI using the `.local` address?

Olares supports `.local` addresses with the HTTP protocol for most services. The Sunshine Web UI is different because it requires HTTPS to secure local communication. If you use `http://` with your `.local` URL, the Sunshine page will not load.

To fix this, use `https://` instead of `http://` in your browser's address bar (for example, `https://139ebc4f0.<your Olares ID>.olares.local:47990`).
### Why isn't the game displaying in full screen?

This may be caused by resolution settings. Try adjusting the resolution:

- **In Moonlight**: Go to **Settings** > **Basic Settings** > **Resolution and FPS**.
  ![Display in Moonlight](/images/manual/use-cases/steam-display-in-moonlight.png#bordered)
- **On the Steam console page**: Go to **Applications** > **Settings** > **Display**.  
  ![Display in Steam Headless](/images/manual/use-cases/steam-display-in-steam-hd.png#bordered)

### How do I exit full-screen streaming?

Use the following shortcuts:
- **Windows**: `Ctrl + Alt + Shift + Q`
- **Mac**: `Control (^) + Option (⌥) + Shift + Q`
- **Mobile**: `Start + Select + L1 + R1`

After finishing, exit Steam Big Picture mode to release system resources on Olares.
![Exit Steam Big Picture Mode](/images/manual/use-cases/steam-exit-big-pic.png#bordered)

### Where are my downloaded games stored?

You can check the downloaded games in the Files app. By default, games are saved in:

```plain
/Cache/olares/steam-headless/c0/.steam/steam/steamapps/common
```

We recommend not changing this default directory.

### Why do I get an error when re-pairing the host in Moonlight?

If you delete your Olares host in Moonlight and try to pair again, you may encounter the following errors:

- **The PIN from the PC didn't match. Please try again.**
- **Request timed out (Error 4)**
- **Connection closed (Error 2)**

This usually happens when the Sunshine service is not responding.
To fix it, simply restart Steam Headless in Olares and try pairing again:

1. Open Control Hub from Launchpad.
2. Navigate to **Browser** > **steamheadless** > **Deployments** > **steamheadless** > **Restart**.
   ![Restart Steam Headless](/images/manual/use-cases/steam-restart.png#bordered)

3. In the confirmation prompt, enter `steamheadless` and click **Confirm**.
   ![Confirm restart](/images/manual/use-cases/steam-confirm-restart.png#bordered){width=80%}

4. Once restarted, pair with Sunshine again in Moonlight.

### How do I change my Sunshine username or password?

You can change your Sunshine credentials directly from the Sunshine web console.

1. Open Sunshine in your browser using your local address, for example: `https://139ebc4f0.<your Olares ID>.olares.local:47990`.
2. Log in with your current username and password.
3. Go to the **Change Password** tab.
   ![Change Sunshine password](/images/manual/use-cases/steam-change-sunshine-pw.png#bordered)

4. Enter a new password (and username if desired), then click **Save**.
