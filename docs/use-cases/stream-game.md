---
outline: [2, 3]
description: Comprehensive tutorial on streaming Steam games with Olares. Learn to install Steam Headless, configure the streaming service, and stream games on Moonlight from both local and remote networks.
---

# Stream your favorite games with Steam Headless

Want to enjoy some gaming with the power of Olares? You are all set. With the Steam Headless app, Olares easily transforms into a steam streaming server. You can now play your favorite games on any compatible device via Moonlight or Steam Link.

This guide walks you through installing Steam Headless on Olares, configuring the streaming service, and using the Moonlight client to stream your games.


## Objectives

By the end of this tutorial, you will learn how to:

- Install Steam Headless on Olares and configure compatibility for Windows games on the Steam client.
- Configure the streaming service, which involves pairing the Sunshine streaming server with the host on the Moonlight client.
- Stream games on the Moonlight client from both local and remote networks.

## Before you begin 

Before you begin, ensure the following:

- Olares is up and running on a machine equipped with an NVIDIA GPU.
- Your streaming device has Moonlight installed. Visit the [Moonlight website](https://moonlight-stream.org/) to download and install the appropriate version.
- [Enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your streaming device.
- You have a valid Steam account to access your games.

## Install Steam Headless

1. In Olares Market, find Steam Headless under the **Entertainment** category, and click **Get**.  
2. Open Steam Headless, and then click **Connect** to access the Steam Headless console.
3. In the Steam Headless console, click the **Install** button to install and update the Steam client. Once completed, the Steam login page should appear.
   ![Install Steam client](/images/manual/tutorials/install-steam-client.png#bordered)

4. Log into your Steam account and complete the initial setup.

   ![Steam login interface](/images/manual/tutorials/steam-login.png#bordered)

::: tip Retry installation upon failures
Due to network issues, Steam installation or update may fail. If this happens, go to the top-left menu in Steam console and navigate to **Applications** > **Internet** > **Steam** to reinstall.
:::

## Configure game compatibility

Olares runs on a Linux environment. You'll need to enable [Proton](https://github.com/ValveSoftware/Proton), a compatibility layer, to support Windows games.

1. In the Steam client page, navigate to **Steam** > **Settings** in the top left corner.
2. Go to the **Compatibility** tab and check **Enable Steam Play for all other titles**.  
   ![Steam Settings](/images/manual/tutorials/steam-setting.png#bordered)
3. Restart the Steam client to view your complete game library.  

## Configure the streaming service

Steam Headless also integrates Sunshine, the streaming server. To stream games using Moonlight, you need to pair the host PC on moonlight with Sunshine. 

### Prepare for paring

1. From your Steam client in the browser, get the URL and append `:47990`, for example, `https://139ebc4f0.<your Olares ID>.olares.com:47990`. Open this URL to access the Sunshine console page.

   ![Sunshine console page](/images/manual/tutorials/access-sunshine.png#bordered)

2. On your first visit, log in using the default credentials:  
   - Username: `sam`  
   - Password: `password` 

3. Click the **Pin** tab to open the pairing page, where you will be prompted to enter the pairing code.
   
   ![Sunshine's paring page](/images/manual/tutorials/pin-sunshine.png#bordered)


### Add host in Moonlight

1. Open the Moonlight client on your streaming device and click the <i class="material-symbols-outlined">add_to_queue</i> button in the top-right corner.

2. When promoted to enter the IP address of your host PC, enter the URL of Steam, for example, `139ebc4f0.<your Olares ID>.olares.com`.

   ::: tip Note
   Do not include `https://`.
   ::: 

3. Click **OK**, and a host icon appears in locked status.
4. Click the host icon to receive a pairing code.

   ![Get pin code](/images/manual/tutorials/get-pin-code.png#bordered)

### Complete paring

1. In Sunshine's paring page, enter the pairing code.
2. Enter a name for the device and click **Send** to finish pairing. If successful, you will see the message: "Success! Please check Moonlight to continue".
3. Check the host status in Moonlight. The icon should be active now. 
   ![Paring successfully](/images/manual/tutorials/active-host-moonlight.png#bordered)  

## Stream your game

Now you are done with configuring and ready for gaming. 

### Stream locally 

If you are streaming locally from the same network of Olares:

1. Open Moonlight on your streaming device.
2. Click your steam host, and then click the Steam icon to enter the Steam Big Screen mode and play. 

   ![Stream game](/images/manual/tutorials/stream-success.png#bordered)

### Stream remotely 

With Olares's VPN feature, you can enjoy a smooth streaming experience even across different networks.

To enable the VPN:

<!--@include: ./remote.reusables.md{4,22}-->

Once VPN is enabled on your device, simply follow the same instructions as local streaming.

## FAQs

### Why isn't the game displaying in full screen?

This could be due to your resolution settings. Try adjusting the resolution:
- In Moonlight: adjust in **Settings** > **Basic Settings** > **Resolution and FPS**.
- On the Steam console page: adjust in **Applications** > **Settings** > **Display**.  
  
  ![Set display](/images/manual/tutorials/set-steam-display.png#bordered)

### How do I exit full-screen streaming?

To exit streaming in the full-screen mode:
- **On Windows**: Use the shortcut combo **Ctrl + Alt + Shift + Q**.  
- **On Mac**: Use the shortcut combo **Control (^) + Option (⌥) + Shift + Q**. 
- **On mobile devices**: use the controller button combo **Start + Select + L1 + R1**.  

After finishing your streaming session, exit the Steam Big Picture mode to release system resources on Olares.

### Where are my downloaded games stored?

By default, games are stored at: 

`/Cache/olares/steam-headless/c0/.steam/debian-installation/steamapps/common`

We recommend not changing this default directory.






 



