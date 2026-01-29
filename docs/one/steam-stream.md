---
outline: [2, 3]
description: Comprehensive tutorial on streaming Steam games with Olares. Learn to install Steam Headless, configure the streaming service, and stream games on Moonlight from both local and remote networks.
head:
  - - meta
    - name: keywords
      content: Steam, Steam Headless, Moonlight, Sunshine, Game Streaming, Remote Play, Cloud Gaming
---

# Stream Steam games to any device <Badge type="tip" text="30 min" />

Equipped by an NVDIA 5090 mobile GPU, with the Steam Headless app, Olares One transforms into a powerful game streaming server. You can play your favorite games on any compatible device via Moonlight.

## Learning objectives
- Set up Steam Headless on your Olares One.
- Configure the Sunshine streaming service.
- Pair your device via Moonlight and stream games locally or remotely.

## Prerequisites
**Hardware** <br>
- Olares One connected to a stable network (Ethernet recommended).
- Sufficient disk space to download games.

**Software** <br>
- [Moonlight](https://moonlight-stream.org/) installed on the client device.
- A valid Steam account.

**LarePass** (Required for remote access)
- LarePass required on your client device (desktop or mobile) if you plan to stream outside your home network.

## Step 1: Install Steam Headless

<<<<<<< HEAD
1. Open the Market, and search for "Steam".
=======
1. Open the Market, and search for "Steam."
>>>>>>> 24de23a9 (add steam)
2. Click **Get**, then **Install**.
   ![Install Steam Headless](/images/manual/use-cases/steam-install-steam-headless1.png#bordered)

3. A prompt will appear asking you to set environment variables. This creates your login credentials for the Sunshine Web UI:
    - `SUNSHINE_USER`: Set the username for Sunshine access.
    - `SUNSHINE_PASS`: Set a secure password.
      :::tip Remember your login credentials
      These are your initial login credentials for Sunshine. You must use them to log in to Sunshine the first time.
      :::
4. Wait for the installation to complete.

## Step 2: Install the Steam client
Once the headless app is running, you need to initialize the Steam client inside it.
1. Open Steam Headless and click **Connect**.
   ![Connect to Steam](/images/manual/use-cases/steam-connect-to-steam.png#bordered)

2. The Steam client will automatically begin downloading and installing.
   ![Install Steam](/images/manual/use-cases/steam-install-steam.png#bordered)
   ![Update Steam](/images/manual/use-cases/steam-update-steam.png#bordered)

3. When installation completes, the Steam login screen appears. Sign in with your Steam account.
   ![Sign in to Steam](/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

::: tip Retry installation upon failures
If Steam installation or update fails due to network issues, go to the top-left menu in the Steam Headless console and navigate to **Applications** > **Internet** > **Steam** to restart the installation.
:::

## Step 3: Pair Sunshine with Moonlight
Steam Headless uses Sunshine to stream video. You must pair it with the Moonlight app on your playing device.

### 1. Access the Sunshine console

You need to access the Sunshine Web UI running on your Olares One to enter a pairing PIN.

1. Copy the URL of your current Steam Headless browser tab.
2. Open a new browser tab and modify the URL to access port `47990`. The address varies depending on your network.
    - **Same network**: Use `http` and your `.local` address. For example:

   ```plain
   http://139ebc4f0.<your Olares ID>.olares.local:47990
   ```

    - **Different network (via VPN)**: Enable LarePass VPN on your device, then append `:47990`. For example:

   ```plain
   https://139ebc4f0.<your Olares ID>.olares.com:47990
   ```

3. Press **Enter** to open the Sunshine console page.
4. Sign in using the `SUNSHINE_USER` and `SUNSHINE_PASS` credentials you created earlier.
   ![Sign in to Sunshine](/images/manual/use-cases/steam-sign-in-to-sunshine.png#bordered)
5. Click the **PIN** tab. The page will now wait for a pairing PIN.
   ![PIN on Sunshine](/images/manual/use-cases/steam-pin-on-sunshine.png#bordered)

### Add the host in Moonlight

Moonlight usually detects your Olares host automatically when it is on the same local network as your Olares One.

If it doesn't appear, or if you are connecting over different networks, follow these steps to add the host manually. The process shown below uses the macOS Moonlight client.

1. Open Moonlight on your client device.
2. Click the **Add Host** button, which looks like a computer with a plus icon.
3. Enter your Olares domain without the `https://` prefix. For example:

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

## Step 4: Stream your games
:::tip Optimizing remote play
For the best experience when streaming remotely:
1. Connect your client device to 5GHz Wi-Fi or Ethernet.
2. Ensure **LarePass VPN** is active.
:::
You can stream your games through Moonlight either locally or remotely, depending on your network setup.

The following steps demonstrate local streaming.

1. Open Moonlight on your client device.
2. Select your Olares host and click the **Steam** icon. 
   ![Steam in Moonlight](/images/manual/use-cases/steam-in-moonlight.png#bordered)

3.  Steam **Big Picture Mode** will launch. Select a game from your library and start playing.

## FAQs

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
   ![Restart Steam Headless](/images/manual/use-cases/steam-restart.png#bordered)

3. In the confirmation prompt, enter `steamheadless` and click **Confirm**.
   ![Confirm restart](/images/manual/use-cases/steam-confirm-restart.png#bordered){width=80%}

4. Once restarted, pair with Sunshine again in Moonlight.

### How do I change my Sunshine username or password?

You can change your Sunshine credentials directly from the Sunshine web console.

1. Open Sunshine in your browser using your local address, for example: `http://139ebc4f0.<your Olares ID>.olares.local:47990`.
2. Log in with your current username and password.
3. Go to the **Change Password** tab.
   ![Change Sunshine password](/images/manual/use-cases/steam-change-sunshine-pw.png#bordered)
4. Enter a new password (and username if desired), then click **Save**.

## Resources
- [Play Steam games locally on Olares One](steam-direct-play.md)