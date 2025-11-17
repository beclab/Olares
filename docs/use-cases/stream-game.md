---
outline: [2, 3]
description: Comprehensive tutorial on streaming Steam games with Olares. Learn to install Steam Headless, configure the streaming service, and stream games on Moonlight from both local and remote networks.
---

# Stream your games with Steam Headless

Want to enjoy gaming powered by Olares? You're all set. With the Steam Headless app, Olares easily transforms into a powerful game streaming server. You can now play your favorite games on any compatible device via Moonlight.

This guide walks you through installing Steam Headless on Olares, configuring the Steam client, pairing the streaming service, and connecting with the Moonlight client to play.


## Learning objectives

By the end of this tutorial, you will learn how to:
- Install and set up Steam Headless on your Olares device.
- Set up the Sunshine streaming service and pair it with the Moonlight client.
- Stream your Steam games from both local and remote networks.

## Prerequisites

Before you begin, make sure:
- Olares running on a machine equipped with an NVIDIA GPU.
- Moonlight installed on your streaming device. Visit the [Moonlight website](https://moonlight-stream.org/) to download and install the appropriate version.
- A valid Steam account to access your games.
- [LarePass VPN enabled](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your client devices (desktop or mobile) if you plan to stream remotely (outside your home network).
> For local streaming, LarePass VPN is not required.

## Install and configure Steam Headless

Install the app from the Olares Market and then complete the initial setup within the Steam client itself.

### Install Steam Headless

Follow these steps to install and configure Steam Headless:

1. Open the **Market** app in your Olares web interface.
2. Find **Steam Headless** in the **Fun** category or use the search bar.
3. Click **Get**, and then click **Install**.
![Install Steam Headless](../public/images/manual/use-cases/steam-install-steam-headless.png#bordered)
4. A prompt will appear asking you to configure environment variables. This creates your login credentials for the Sunshine streaming dashboard:
  - `SUNSHINE_USER`: Create a username for Sunshine access.
  - `SUNSHINE_PASS`: Set the corresponding password.
:::tip Remember your login credentials 
These are your initial login credentials for Sunshine. You must use them to log in the first time. 
:::
5. Wait a few minutes for the installation and initialization to complete.

### Complete the initial Steam setup

Sign in to Steam to complete setup and access your game library.
1. Once installed, open Steam Headless and click **Connect**.
   ![Connect to Steam](../public/images/manual/use-cases/steam-connect-to-steam.png#bordered)
2. The Steam client will automatically install and update. This process may take several minutes.
   ![Install Steam](../public/images/manual/use-cases/steam-install-steam.png#bordered)
   ![Update Steam](../public/images/manual/use-cases/steam-update-steam.png#bordered)
3. When the installation completes, the Steam login screen appears. Sign in with your Steam account and complete the initial setup.
   ![Sign in to Steam](../public/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

::: tip Retry installation upon failures 
If Steam installation or update fails due to network issues, go to the top-left menu in the Steam Headless console and navigate to Applications > Internet > Steam to restart the installation. 
:::
Once Steam is ready, you can connect it to Moonlight through Sunshine.

## Pair Sunshine with Moonlight

Steam Headless integrates Sunshine, an open-source streaming server. To stream games, you must pair it with your Moonlight client.

### Access the Sunshine console

Access the Sunshine console to pair your Moonlight client with Olares. The connection method varies depending on your network.

1. From your Steam Headless browser tab, copy its URL.
2. In your browser, enter the address that matches your connection:
   - **Same network**: Replace the domain suffix with `.local`, and append `:47990`. For example:
    
   ```
   https://139ebc4f0.<your Olares ID>.olares.local:47990
   ```

   - **Different network (via VPN)**: [Enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device, then use the `.com` address and append `:47990`. For example: 
   
   ```
   https://139ebc4f0.<your Olares ID>.olares.com:47990
   ```

3. Press **Enter** to open the Sunshine console page. 
4. Sign in using the `SUNSHINE_USER` and `SUNSHINE_PASS` credentials you created earlier.
   ![Sign in to Sunshine](../public/images/manual/use-cases/steam-sign-in-to-sunshine.png#bordered)
5. Click the **PIN** tab. The page will now wait for a pairing PIN.
   ![PIN on Sunshine](../public/images/manual/use-cases/steam-pin-on-sunshine.png#bordered)

### Add the host in Moonlight

Moonlight usually detects your Olares host automatically when it is on the same local network as your Olares device.

If it doesn't appear, or if you are connecting over different networks, follow these steps to add the host manually. The process shown below uses the macOS Moonlight client.

:::info Connecting from a different network? 
If your Moonlight client and Olares device are on different networks, you must first [enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) on the device running Moonlight. 
:::

1. Open Moonlight on your streaming device.
2. Click the **Add Host** button (looks like a computer with a plus icon).
3. Enter your Olares domain without the `https://` prefix. For example: 

   ```
   139ebc4f0.<your Olares ID>.olares.com
   ```
4. Click **OK**. A new locked host icon appears.
5. Click the locked icon. Moonlight will display a 4-digit pairing PIN.
   ![Get pairing PIN](../public/images/manual/use-cases/steam-get-pairing-pin.png#bordered)

### Complete pairing

1. Return to the **Sunshine PIN** page in your browser.
2. Enter the PIN displayed in Moonlight and give your device a name.
   ![Enter pairing PIN](../public/images/manual/use-cases/steam-enter-pairing-pin.png#bordered)
3. Click **Send**. When pairing succeeds, you'll see:
   ```text
   Success! Please check Moonlight to continue
   ```
   ![Pairing succeeds](../public/images/manual/use-cases/steam-pairing-succeeds.png#bordered)
4. Back in Moonlight, your host icon should now appear unlocked and active.
   ![Host in Moonlight](../public/images/manual/use-cases/steam-host-in-moonlight.png#bordered)

Once paired, you're ready to start streaming.

## Stream your games

You can stream locally or remotely depending on your usage scenario.

### Stream locally 

If your streaming device and Olares are on the same network:

1. Open Moonlight.
2. Select your unlocked host and click the Steam icon.
  ![Steam in Moonlight](../public/images/manual/use-cases/steam-in-moonlight.png#bordered)

Steam Big Picture mode opens. You can now start playing your games.

### Stream remotely 

To play from outside your home network, you can connect securely to Olares using LarePass VPN.

1. Follow the instructions to [enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass).
2. Once connected, open Moonlight.
3. Select your unlocked host and click the Steam icon.
Steam Big Picture mode opens. 

You can now start playing your games.

## FAQs

### Why isn't the game displaying in full screen?

This may be caused by resolution settings. Try adjusting the resolution:

- **In Moonlight**: Go to **Settings** > **Basic Settings** > **Resolution and FPS**.
![Display in Moonlight](../public/images/manual/use-cases/steam-display-in-moonlight.png#bordered)
- **On the Steam console page**: Go to **Applications** > **Settings** > **Display**.  
![Display in Steam Headless](../public/images/manual/use-cases/steam-display-in-steam-hd.png#bordered)

### How do I exit full-screen streaming?

Use the following shortcuts:
- **Windows**: `Ctrl + Alt + Shift + Q`  
- **Mac**: `Control (^) + Option (⌥) + Shift + Q` 
- **Mobile**: `Start + Select + L1 + R1`

After finishing, exit Steam Big Picture mode to release system resources on Olares.
   ![Exit Steam Big Picture Mode](../public/images/manual/use-cases/steam-exit-big-pic.png#bordered)

### Where are my downloaded games stored?

By default, games are saved in: 

```
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
   ![Restart Steam Headless](../public/images/manual/use-cases/steam-restart.png#bordered)
3. In the confirmation prompt, enter `steamheadless` and click **Confirm**.
   ![Confirm restart](../public/images/manual/use-cases/steam-confirm-restart.png#bordered)
4. Once restarted, pair with Sunshine again in Moonlight.

### How do I change my Sunshine username or password?

You can change your Sunshine credentials directly from the Sunshine web console.

1. Open Sunshine in your browser using your local address, for example: `http://139ebc4f0.<your Olares ID>.olares.local:47990`. 
2. Log in with your current username and password.
3. Go to the **Change Password** tab.
   ![Change Sunshine password](../public/images/manual/use-cases/steam-change-sunshine-pw.png#bordered)
4. Enter a new password (and username if desired), then click **Save**.