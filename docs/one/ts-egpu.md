---
outline: [2, 3]
description: Fix common Olares One eGPU startup and detection problems, then collect the right diagnostics for Olares OS or Windows.
---

# Troubleshoot eGPU issues on Olares One

Find the symptom that matches your problem. If the quick checks do not resolve it, collect diagnostics before asking for help.

:::danger Do not hot-plug on Olares OS
Shut down Olares One before connecting or disconnecting an eGPU. For a cold start, turn on the eGPU, connect it to Olares One, and then start Olares One.
:::

## Olares OS startup stalls at the logo

1. Power off Olares One.
2. Disconnect the eGPU.
3. Start Olares One again.
4. Check that the Gen1 workaround is enabled:

   ```bash
   sudo systemctl is-enabled egpu-gen1-fix.service
   ```

   The command must return `enabled`. If it does not, reinstall the workaround by following [Set up an eGPU on Olares OS](./egpu-olares-os.md).

5. Before trying again, [collect the Olares OS diagnostics](#collect-olares-os-diagnostics). The report may include logs from the stalled startup.

## Olares OS starts but does not detect the eGPU

Check each item in order:

1. Check that the eGPU is on and its power supply is connected.
2. If you use an eGPU dock or enclosure with a desktop GPU, check the power supply rating and all GPU power connectors.
3. Connect the eGPU directly to a Thunderbolt 5 (USB-C) port on Olares One. Remove docks and other Thunderbolt devices from the connection path.
4. Use a certified Thunderbolt cable, preferably the one supplied with the external graphics device.
5. Repeat the cold-start order. Turn on the eGPU, connect the cable, and then start Olares One.
6. Check whether the operating system and NVIDIA driver detect the GPU:

   ```bash
   lspci -nn | grep -i nvidia
   nvidia-smi
   ```

Both commands should list the external GPU. If `lspci` lists it but `nvidia-smi` does not, collect diagnostics before changing the driver.

## Olares OS loses the eGPU under load

1. Check the workaround log:

   ```bash
   sudo tail -n 20 /var/log/egpu-gen1-fix.log
   ```

   Look for `speed=2.5 GT/s PCIe`. This shows that the eGPU link is running at Gen1.

2. Check the external graphics device's power supply. If you use an eGPU dock or enclosure with a desktop GPU, also check the power supply rating and all GPU power connectors.
3. Remove other high-bandwidth Thunderbolt devices and try again with the eGPU connected directly to Olares One.
4. Collect diagnostics immediately after the disconnect.

## Windows reports a GPU error

1. In **Device Manager** > **Display adapters**, open the affected device.
2. Record the complete message and error code under **General** > **Device status**.
3. Follow [Recover the built-in GPU](./egpu-windows.md#recover-the-built-in-gpu).
4. Restart Windows, then check that both GPUs appear without warning icons.

If either GPU is still missing, collect the Windows information below.

## Collect Olares OS diagnostics

1. Download <a href="/downloads/one/egpu/collect-egpu-info.sh" download>`collect-egpu-info.sh`</a>.
2. Open a terminal in the download directory and run:

   ```bash
   chmod +x collect-egpu-info.sh
   sudo ./collect-egpu-info.sh
   ```

3. Open the generated `egpu-report-*.txt` file and read the `Summary` section.

   - If the eGPU was connected when you ran the script and the report says `External GPU detected : NO`, check the power, cable, port, and connection path again.
   - If you disconnected the eGPU after a stalled startup, `External GPU detected : NO` is expected. When the previous boot log is available, the report still includes it.

The script reads system state and writes one report in the current directory. It does not change settings, connect to the internet, or upload the report.

:::warning Review the report before sharing
The report may contain the hostname, kernel command line, hardware topology, and system logs. Remove any information you do not want to post publicly.
:::

## Collect Windows diagnostics

Prepare:

1. A screenshot of **Device Manager** > **Display adapters**, including any warning icons.
2. The NVIDIA driver version shown in NVIDIA App.
3. The complete Device Manager error message and code for each affected GPU.

Remove personal information from screenshots before sharing them.

## Ask for help in the Olares forum

If the problem continues, create a post in the [Olares forum](https://www.olares.com/forum/). Attach the diagnostic report or screenshots and include:

```plain
System and version:
External graphics device, or eGPU dock or enclosure and GPU:
Connection path, including any dock or hub:
Power-on order:
What happened:
Steps already tried:
Attachments:
```

## Related resources

- [Connect an eGPU to Olares One](./egpu.md)
- [Set up an eGPU on Olares OS](./egpu-olares-os.md)
- [Set up an eGPU on Windows](./egpu-windows.md)
