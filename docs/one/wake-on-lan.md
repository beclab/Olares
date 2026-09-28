---
outline: [2, 3]
description: Configure Wake-on-LAN on Olares One running Ubuntu, then wake it from a phone, Linux, macOS, or Windows device on the same local network.
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu, Wake-on-LAN, WOL, Magic Packet
---

# Set up Wake-on-LAN for Olares One on Ubuntu

Wake-on-LAN (WOL) lets you wake Olares One by sending a Magic Packet from another device on the same local network. This guide explains how to prepare Olares One and send the packet from a phone, Linux, macOS, or Windows.

:::info Local network only
This guide covers waking Olares One from the same local network. Waking it over the internet requires additional router and security configuration and is not covered here.
:::

## Before you begin

- Ubuntu is installed on Olares One. If needed, follow [Install Ubuntu Server on Olares One](install-ubuntu-server.md) or [Install Ubuntu Desktop on Olares One](install-ubuntu-desktop.md).
- Olares One is connected to your router through its Ethernet port. Wake-on-LAN does not work over its Wi-Fi connection.
- You can access the Ubuntu terminal with an account that has `sudo` permission.
- A phone or computer is connected to the same local network as Olares One.

## Configure Wake-on-LAN on Olares One

### Check the EC firmware

Wake-on-LAN requires EC firmware version 1.01 or later. For the best compatibility, update to the latest EC firmware. To check or update the version, see [Manage BIOS and EC](update-firmware.md).

### Find the network interface and MAC address

1. Make sure the Ethernet cable is connected to Olares One and your router.
2. Open the Ubuntu terminal.
3. Run the following command:

   ```bash
   ip address
   ```

4. Find the wired interface. Its name usually starts with `en`, such as `enp129s0`.
5. Record these values for the wired interface:

   - **Interface name**: The name shown at the beginning of the interface entry.
   - **MAC address**: The value after `link/ether`.
   - **IPv4 address**: The value after `inet`, without the subnet suffix. For example, record `192.168.0.92` from `192.168.0.92/23`.

### Check the Wake-on-LAN status

1. Install `ethtool`:

   ```bash
   sudo apt update
   sudo apt install ethtool -y
   ```

2. Check the current Wake-on-LAN status. Replace `<network-interface>` with the wired interface name you recorded earlier.

   ```bash
   sudo ethtool <network-interface> | grep Wake-on
   ```

3. Check that the output contains `g` in **Supports Wake-on** and shows `Wake-on: g`:

   ```text
   Supports Wake-on: pumbg
   Wake-on: g
   ```

   The `g` value means the interface can wake the device when it receives a Magic Packet.

4. If **Supports Wake-on** contains `g` but **Wake-on** is set to another value, enable Magic Packet wake-up:

   ```bash
   sudo ethtool --change <network-interface> wol g
   ```

5. Run the status command again and check that it now shows `Wake-on: g`.

:::tip Setting resets after a restart
Some Ubuntu network configurations reset the Wake-on-LAN setting during startup. If Olares One stops responding to Magic Packets after a restart, check the status again and re-enable it before suspending or shutting down the device.
:::

### Suspend or shut down Olares One

Keep Olares One connected to AC power and Ethernet. Then use one of the following commands.

- Suspend Olares One:

  ```bash
  sudo systemctl suspend
  ```

- Shut down Olares One:

  ```bash
  sudo shutdown -h now
  ```

Start with suspend if this is your first time using Wake-on-LAN. Some network environments or power settings might not support waking from a full shutdown.

## Send a Magic Packet to wake Olares One

Use the MAC address you recorded earlier. The sending device must be connected to the same local network as Olares One.

:::info Choose one sending device
You only need to configure the phone or computer that you will use to send the Magic Packet. Each method below sends the same type of packet, so you do not need to complete every subsection.
:::

### From a phone

The following steps use Easy WOL as an example. You can use another app that sends Wake-on-LAN Magic Packets.

1. Connect your phone to the same local network as Olares One.
2. Install and open a Wake-on-LAN app.
3. Add Olares One and enter the following information:

   - **Device Name**: Enter a name that helps you identify Olares One.
   - **Address**: Enter the IPv4 address you recorded earlier.
   - **MAC**: Enter the wired interface MAC address.
   - **Port**: Enter `9`.

   ![Wake-on-LAN device settings on a phone](/images/one/wol-add-device.png#bordered){width=50%}

4. Save the device. To wake Olares One, tap the saved device. This sends the Magic Packet immediately.

   ![Wake-on-LAN packet sent confirmation](/images/one/wol-packet-sent.png#bordered){width=35%}

5. Wait for Olares One to start.

### From Linux

1. Install `wakeonlan`:

   ```bash
   sudo apt update
   sudo apt install wakeonlan -y
   ```

2. Send a Magic Packet. Replace `<mac-address>` with the wired interface MAC address of Olares One.

   ```bash
   wakeonlan <mac-address>
   ```

   The command sends the Magic Packet immediately. Wait for Olares One to start.

### From macOS

1. If Homebrew is not installed, follow the installation instructions on the [Homebrew website](https://brew.sh/).
2. Install `wakeonlan`:

   ```bash
   brew install wakeonlan
   ```

3. Send a Magic Packet. Replace `<mac-address>` with the wired interface MAC address of Olares One.

   ```bash
   wakeonlan <mac-address>
   ```

   The command sends the Magic Packet immediately. Wait for Olares One to start.

### From Windows

:::info No download required
PowerShell generates and sends the Magic Packet directly. You do not need to download `magic_pkt.zip` or install a separate Wake-on-LAN utility.
:::

1. Open PowerShell.
2. Replace `<mac-address>` in the following script with the wired interface MAC address of Olares One, and then run the script:

   ```powershell
   $mac = "<mac-address>"
   $macBytes = [byte[]]($mac -split '[:-]' | ForEach-Object {
     [Convert]::ToByte($_, 16)
   })
   $packet = [byte[]](,0xFF * 6 + ($macBytes * 16))
   $udp = [System.Net.Sockets.UdpClient]::new()
   $udp.EnableBroadcast = $true
   [void]$udp.Send($packet, $packet.Length, "255.255.255.255", 9)
   $udp.Close()
   ```

3. Running the script sends the Magic Packet immediately. Wait for Olares One to start.

## Troubleshooting

If Olares One does not wake:

- Make sure it remains connected to AC power and Ethernet.
- Make sure the sending device is on the same local network.
- Check that you entered the MAC address of the wired interface, not the Wi-Fi interface.
- Check that the EC firmware is version 1.01 or later.
- Before suspending or shutting down Olares One, confirm that `ethtool` shows `Wake-on: g`.
- If waking from a shutdown does not work, suspend Olares One and try again.
- Check whether your router isolates Wi-Fi devices from wired devices. If client isolation is enabled, the Magic Packet might not reach Olares One.
- On Linux or macOS, try sending the packet to your subnet broadcast address:

  ```bash
  wakeonlan -i <broadcast-address> <mac-address>
  ```
