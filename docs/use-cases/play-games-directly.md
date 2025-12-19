---
outline: [2, 3]
description: Learn how to turn your Olares device into a gaming console by connecting a monitor and peripherals directly.
---

# Play Steam games locally on Olares

While Olares is typically used as a headless system, you can connect a monitor, keyboard, and mouse to play games directly on the device.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Configure system permissions to allow the container to access your physical hardware.
- Modify display and audio settings for local gameplay.

## Prerequisites
Before you begin, ensure you have:
- Olares version 1.12.3 or later.
- Olares running on a machine equipped with an NVIDIA GPU.
- A valid Steam account to access your games.
- A monitor, keyboard, and mouse connected to the Olares device.

## 1. Install and configure Steam Headless
Install the app from the Olares Market and then complete the initial setup within the Steam client itself.

### Install and Steam Headless

Follow these steps to install and configure Steam Headless:

1. Open the Market, and search for "Steam".
2. Click **Get**, then **Install**.
   ![Install Steam Headless](../public/images/manual/use-cases/steam-install-steam-headless.png#bordered)
3. A prompt will appear asking you to configure environment variables. This creates your login credentials for the Sunshine Web UI:
    - `SUNSHINE_USER`: Set the username for Sunshine access.
    - `SUNSHINE_PASS`: Set the corresponding password.
      :::tip Remember your login credentials
      These are the initial login credentials for Sunshine. You will need them if you access the Sunshine dashboard later.
      :::
4. Wait for the installation to complete.

### Complete the initial Steam setup
Sign in to Steam to complete setup and access your game library.
:::tip Retry installation upon failures
If Steam installation or update fails due to network issues, go to the top-left menu in the Steam Headless console and navigate to **Applications** > **Internet** > **Steam** to restart the installation.
:::
1. Once installed, open Steam Headless and click **Connect**.
   ![Connect to Steam](../public/images/manual/use-cases/steam-connect-to-steam.png#bordered)
2. The Steam client will automatically install and update. This process may take several minutes.
   ![Install Steam](../public/images/manual/use-cases/steam-install-steam.png#bordered)
   ![Update Steam](../public/images/manual/use-cases/steam-update-steam.png#bordered)
3. When the installation completes, the Steam login screen appears. Sign in with your Steam account.
   ![Sign in to Steam](../public/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

## 2. Grant sound card permissions
To output audio to your monitor, the Steam container needs direct access to your sound card. You must edit the deployment configuration to "pass through" these hardware components.

1. Open Control Hub, then select the Steam project from the sidebar.
2. Navigate to **Deployments** > **steamheadless**, and click <span class="material-symbols-outlined">box_edit</span> to open the editor.
   ![Open Yaml editor](../public/images/manual/use-cases/steam-open-yaml-editor.png#bordered)
   
3. Locate the `volumes` section and append the following entry to the list.
   ```yaml
    spec:
      volumes:
        - name: steam-headless-claim0
          hostPath:
            path: >-
              /olares/userdata/Cache/pvc-appcache-onetest02-x2kjhiuu1cgftrzb/steamheadless/steamheadless/c0
            type: DirectoryOrCreate
        - name: steam-headless-claim1
          hostPath:
            path: >-
              /olares/userdata/Cache/pvc-appcache-onetest02-x2kjhiuu1cgftrzb/steamheadless/steamheadless/c1
            type: DirectoryOrCreate
        - name: input-devices
          hostPath:
            path: /dev/input/
            type: ''
        - name: dshm
          emptyDir:
            medium: Memory
        # Add the following
        - name: snd
          hostPath:
            path: /dev/snd
            type: ''
   ```
4. Locate the `volumeMounts` section and append the following mounting point:
   ```yaml
       volumeMounts:
         - name: steam-headless-claim0
           mountPath: /home/default
         - name: steam-headless-claim1
           mountPath: /mnt/games
         - name: input-devices
           mountPath: /dev/input/
         - name: dshm
           mountPath: /dev/shm
         # Add the following
         - name: snd
           mountPath: /dev/snd
   ```
5. Click **Confirm** to apply the changes. The container will restart with the new permissions.

## 3. Configure display and input
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

   b. Update the `Display` subsection to match your monitor's resolution.

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

## 4. Enable audio output
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

3. Check the audio channel status.
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
For detailed streaming and paring instructions, refer to [Stream your games with Steam Headless](stream-game.md).
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
The Olares device usually displays a terminal prompt. However, running the Steam application activates a graphical interface that takes over the display.

To return the monitor to the standard terminal view, you must stop the Steam application via **Market** or **Settings**.


