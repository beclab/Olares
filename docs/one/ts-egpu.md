---
outline: [2, 3]
description: Diagnose Olares One eGPU startup, detection, disconnect, and Windows driver issues, then collect information for support.
---

# Troubleshoot eGPU issues on Olares One

Use this guide when startup stalls, the eGPU is not detected, the GPU disconnects, or Windows reports a GPU driver error.

## Try the quick fixes

| Platform | Symptom | What to check |
|---|---|---|
| Olares OS | Startup stalls at the Olares logo | Power off Olares One, disconnect the eGPU, and start it again. Then confirm that the [Gen1 workaround](./egpu-olares-os.md) is installed and enabled. |
| Olares OS | The system starts but does not detect the eGPU | Check the enclosure power, connection order, and certified Thunderbolt 5 cable. Confirm that the enclosure is connected to the Thunderbolt 5 (USB-C) port on Olares One. |
| Olares OS | The eGPU disappears while in use | Confirm that the Gen1 workaround is active. Check that the enclosure power supply is sufficient and that all GPU power connectors are secure. |
| Windows | The built-in GPU reports an error | Perform a [clean driver installation](./egpu-windows.md#recover-the-built-in-gpu), then restart Windows. |

For Olares OS cold starts, power the enclosure, connect the Thunderbolt cable, then start Olares One. 

For the initial Windows driver setup, follow the connection order in the [Windows setup guide](./egpu-windows.md).

## Verify the eGPU on Olares OS

```bash
lspci -nn | grep -i nvidia
nvidia-smi
```

Both the built-in GPU and eGPU should appear. To check whether the Gen1 workaround is active, run the diagnostic script below and review its `Summary` section. A link speed of `2.5 GT/s PCIe` means Gen1 is active.

## Collect Olares OS diagnostics

1. Download <a href="/downloads/one/egpu/collect-egpu-info.sh" download>`collect-egpu-info.sh`</a>.
2. Run it after the affected startup. If startup stalled, disconnect the eGPU, start normally, and run the script. When available, the report also includes logs from the previous startup.

   ```bash
   chmod +x collect-egpu-info.sh
   sudo ./collect-egpu-info.sh
   ```

3. Open `egpu-report-*.txt` and review `Summary`.

   - If the eGPU was connected when you ran the script and `External GPU detected : NO` appears, check the enclosure power, Thunderbolt 5 cable, and Thunderbolt 5 (USB-C) port.
   - If you disconnected the eGPU to recover from a stalled startup, `External GPU detected : NO` is expected. When available, the report may still include logs from the previous startup.

The script does not change system settings or upload data. It reads system information and writes one report file to the current directory. The report includes system and driver versions, Thunderbolt and PCIe topology, link speeds, BAR allocation, workaround status, and relevant logs.

:::warning Review the report before sharing
The report may contain the device hostname, kernel command line, hardware topology, and system logs. Review the file and remove any information you do not want to share publicly.
:::

## Collect Windows diagnostics

Collect:

1. **Device Manager** > **Display adapters** screenshot, including warning icons.
2. NVIDIA driver version.
3. The complete message and error code under **Device properties** > **General** > **Device status**.

Review screenshots and remove any personal or device information you do not want to share publicly.

Also record the enclosure and GPU models, operating system, connection path, startup order, symptom, frequency, and attempted fixes. For intermittent startup failures, perform several cold starts with the same connection order and record each result. A cold start means starting Olares One after it has been completely powered off.

## Ask for help in the Olares forum

If the issue continues, or if you want to discuss an unlisted hardware configuration, create a post in the [Olares forum](https://www.olares.com/forum/). Provide as much of the following information as you can:

- **System**: Windows or Olares OS, including the version
- **Hardware**: The eGPU enclosure and GPU model
- **Setup**: How the eGPU is connected and, if relevant, the startup order
- **Issue**: What happened and whether it occurs consistently
- **Attempted fixes**: Any setup steps, cable changes, or other fixes you have tried
- **Attachments**: The Olares OS diagnostic report, or relevant Windows screenshots and error details

Review diagnostic files and screenshots before posting, and remove any information you do not want to share publicly.

## Resources

- [Olares One eGPU support overview](./egpu.md)
- [Set up an eGPU on Olares OS](./egpu-olares-os.md)
- [Set up an eGPU on Windows](./egpu-windows.md)
