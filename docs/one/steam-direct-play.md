---
outline: [2, 3]
description: Learn how to turn your Olares device into a gaming console by connecting a monitor and peripherals directly.
head:
  - - meta
    - name: keywords
      content: Steam, Local Gaming, Direct Play, Linux Gaming, Audio Configuration
---

# Play Steam games locally on Olares One <Badge type="tip" text="30 min" />

While Olares One is typically used as a headless system, you can connect a monitor, keyboard, and mouse to play games directly.

However, simply plugging in a monitor is not enough. Because Steam Headless is optimized for remote streaming by default, you must manually configure the audio settings to output audio locally.

## Learning objectives
- Set up Steam Headless on your Olares One.
- Grant the application permission to access the host sound card.
- Modify display and audio settings for local gameplay.

## Prerequisites
**Hardware** <br>
- Olares One connected to a stable network (Ethernet recommended).
- Monitor, keyboard, and mouse connected to the Olares One.
- Sufficient disk space to download games.

**Software** <br>
- [Moonlight](https://moonlight-stream.org/) installed on the client device.
- A valid Steam account.

**Experience**
- Basic familiarity with terminal commands and the command-line interface (CLI).

## Step 1: Install Steam Headless

1. Open the Market, and search for "Steam".
2. Click **Get**, then **Install**.
   ![Install Steam Headless](/images/manual/use-cases/steam-install-steam-headless1.png#bordered)

3. A prompt will appear asking you to set environment variables. This creates your login credentials for the Sunshine Web UI:
   - `SUNSHINE_USER`: Set the username for Sunshine access.
   - `SUNSHINE_PASS`: Set a secure password.
     :::tip Remember your login credentials
     These are the initial login credentials for Sunshine. You will need them if you access the Sunshine dashboard later.
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

## Step 3: Grant sound card permissions
To output audio to your monitor, the Steam container needs direct access to your sound card. You must edit the deployment configuration to "pass through" these hardware components.

:::warning Edit with caution
Incorrect YAML syntax or indentation may prevent the application from starting.
- **Verify changes:** Double-check that your indentation matches the existing lines.
- **Revert when necessary:** If the app fails to start after applying changes, check **Revision records** and simply delete the lines you added to restore the previous configuration.
  ![Check revision records](../public/images/manual/use-cases/steam-yaml-revision-records.png#bordered)
  :::

1. Open Control Hub, then select the Steam project from the sidebar.
2. Navigate to **Deployments** > **steamheadless**, and click <span class="material-symbols-outlined">edit_square</span> to open the editor.
   ![Open Yaml editor](../public/images/manual/use-cases/steam-open-yaml-editor.png#bordered)

3. Locate the `volumes` section and append the following entry to the list.
   ```yaml
    spec:
      volumes:    
    # Keep existing volumes like steam-headless-claim0
    # ...
    # Add the following entry at the end:
        - name: snd
          hostPath:
            path: /dev/snd
            type: ''
   ```
4. Locate the `volumeMounts` section and append the following mounting point:
   ```yaml
       volumeMounts:
       # Keep existing mounts like steam-headless-claim0
       # ... 
       # Add the following entry at the end:
         - name: snd
           mountPath: /dev/snd
   ```
5. Click **Confirm** to apply the changes. The container will restart with the new permissions.

## Step 4: Configure display and input
By default, the system is configured for streaming. You need to modify the X11 configuration to use your physical monitor and peripherals.

1. In Control Hub, click on the running Pod to view its details.
   ![Click Pod](../public/images/manual/use-cases/steam-click-pod.png#bordered)

2. Click the Terminal icon next to **steam-headless** to access the container's shell.
   ![Access container's shell](../public/images/manual/use-cases/steam-access-container-shell.png#bordered)

3. Open the X configuration file:
   ```bash
   nano /etc/X11/xorg.conf
   ```
4. Modify the file to match the configuration below.

   a. Delete the `Section "InputDevice"` blocks to stop using virtual inputs.

   b. Update the `Display` subsection to match your monitor's resolution. For example:
   ```bash
   SubSection     "Display"
   Virtual     1920 1080
   Depth       24
   Modes      "1920x1080R" "1920x1080" "1280x800" "1024x768" "1920x1080" "1600x900" "1440x900"
   ```

   :::details Reference `xorg.conf` configuration
   ```bash
   Section "ServerLayout"
   Identifier     "Default Layout"
   Screen      0  "Default Screen"
   EndSection
   
   Section "Monitor"
   Modeline "1920x1080R"  138.50  1920 1968 2000 2080  1080 1083 1088 1111 +hsync -vsync
   Identifier     "Configured Monitor"
   HorizSync       5.0 - 1000.0
   VertRefresh     5.0 - 200.0
   ModeLine       "1920x1080" 148.50 1920 2448 2492 2640 1080 1084 1089 1125 +hsync +vsync
   EndSection
   
   Section "Device"
   Identifier     "Configured Video Device"
   Driver         "nvidia"
   Option         "ModeValidation" "NoMaxPClkCheck, NoEdidMaxPClkCheck, NoMaxSizeCheck, NoHorizSyncCheck, NoVertRefreshCheck, NoVirtualSizeCheck, NoTotalSizeCheck, NoDualLinkDVICheck, NoDisplayPortBandwidthCheck, AllowNon3DVisionModes, AllowNonHDMI3DModes, AllowNonEdidModes, NoEdidHDMI2Check, AllowDpInterlaced"
   Option         "AllowEmptyInitialConfiguration"
   Option         "PrimaryGPU" "yes"
   Option         "AllowExternalGpus" "True"
   BusID          "PCI:1:0:0"
   EndSection
   
   Section "Screen"
   Identifier     "Default Screen"
   Device         "Configured Video Device"
   Monitor        "Configured Monitor"
   DefaultDepth    24
   Option         "ProbeAllGpus" "False"
   Option         "BaseMosaic" "False"
   Option         "AllowEmptyInitialConfiguration" "True"
   Option         "ConnectedMonitor" "DFP"
   SubSection     "Display"
   Virtual     1920 1080
   Depth       24
   Modes      "1920x1080R" "1920x1080" "1280x800" "1024x768" "1920x1080" "1600x900" "1440x900"
   EndSubSection
   EndSection
   
   Section "ServerFlags"
   Option "AutoAddGPU" "false"
   EndSection
   ```
   :::
5. Press `Ctrl+O` to save, then `Ctrl+X` to exit.
6. Restart the display server to apply changes:
   ```bash
   supervisorctl stop xorg
   supervisorctl start xorg
   ```

## Step 5: Enable audio output
Finally, you must manually select the correct audio output channel.

1. Identify your sound card. Run the following command inside the container terminal:
   ```bash
   aplay -l
   ```
   Look for your monitor name in the output. For example:
   ```bash
   **** List of PLAYBACK Hardware Devices ****
   card 0: NVidia [HDA NVidia], device 3: HDMI 0 [U2790B]
   Subdevices: 0/1
   Subdevice #0: subdevice #0
   card 0: NVidia [HDA NVidia], device 7: HDMI 1 [HDMI 1]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 0: NVidia [HDA NVidia], device 8: HDMI 2 [HDMI 2]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 3: HDMI 0 [HDMI 0]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 7: HDMI 1 [HDMI 1]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 8: HDMI 2 [HDMI 2]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 9: HDMI 3 [HDMI 3]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   ```

   Look for your monitor name in the output. In this example, the output corresponds to `card 0`, `device 3`.

2. Load the audio module. Replace `0,3` with your specific card and device numbers:
   ```bash
   pactl load-module module-alsa-sink device=plughw:0,3 sink_name=nvhdmi
   ```

3. Check the audio channel status (use the same card number you selected above):
   ```bash
   amixer -c 0 scontents
   ```
   Example output:
   ```bash
   Simple mixer control 'IEC958',0
   Capabilities: pswitch pswitch-joined
   Playback channels: Mono
   Mono: Playback [off]
   Simple mixer control 'IEC958',1
   Capabilities: pswitch pswitch-joined
   Playback channels: Mono
   Mono: Playback [off]
   Simple mixer control 'IEC958',2
   Capabilities: pswitch pswitch-joined
   Playback channels: Mono
   Mono: Playback [off]
   ```
4. If the output contains `Mono: Playback [off]`, it means the channel is muted. Unmute it by running:
   ```bash
   amixer -c 0 sset 'IEC958' on
   ```
Audio now plays directly from your monitor or speakers. You are ready to play.

## Optional: Enable audio for remote streaming

:::info
For detailed streaming and pairing instructions, refer to [Stream your games with Steam Headless](stream-game.md).
:::

Local play and remote streaming use exclusive audio modes.
- **Local play** requires the `module-alsa-sink` module loaded.

- **Remote streaming** requires this module unloaded so Sunshine can capture the audio.

If you decide to stream games to another device later, you must unload the local audio module to restore sound to the stream.

To enable audio for streaming:
1. Open the container terminal via Control Hub.
2. Unload the local audio module:
   ```bash
   pactl unload-module module-alsa-sink
   ```

## FAQs
### Why does my monitor show the Steam interface even when I'm not playing?
Olares One usually displays a terminal prompt when connected to a monitor. However, running the Steam application activates a graphical interface that takes over the display.

To return the monitor to the standard terminal view, you must stop the Steam application via **Market** or **Settings**.

## Resources
- [Stream Steam games to any device](steam-stream.md)