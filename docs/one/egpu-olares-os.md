---
outline: [2, 3]
description: Connect an NVIDIA eGPU to Olares One, apply the temporary Gen1 workaround, verify the GPU, and disconnect it safely.
---

# Set up an eGPU on Olares OS

Use this guide to connect an NVIDIA eGPU to Olares One and apply a temporary Gen1 workaround that improved initialization reliability in the tested Olares OS environment.

:::warning Power off before connecting
Do not connect or disconnect the eGPU while Olares OS is running. Power off Olares One completely before changing the connection.
:::

In the tested Olares OS environment, the eGPU sometimes failed to initialize when its PCIe link operated at Gen4. The workaround in this guide limits the enclosure's PCIe link to Gen1 before initialization.

## Before you start

This workaround has been tested with:

- Olares OS 1.12.6 based on Ubuntu 24.04
- NVIDIA driver 595.84
- AOOSTAR EG02
- NVIDIA GeForce RTX 4060 Ti

It also allowed the built-in RTX 5090M and external RTX 4060 Ti to run together in this configuration.

Other enclosures and GPUs have not been tested. Check the [eGPU support overview](./egpu.md) before continuing.

You also need:

- `sudo` access to Olares OS
- Terminal access
- A Thunderbolt eGPU enclosure with its own power supply
- A certified Thunderbolt 5 cable, preferably the cable supplied with the enclosure

<!-- TODO(tech-review): Confirm that the Olares OS, Ubuntu base, and NVIDIA driver versions above are accurate and approved for publication. -->

## Performance impact

This workaround limits the eGPU link to Gen1. It does not reduce the GPU's compute resources, but it may reduce performance when data moves between system memory and GPU memory.

- Workloads that keep the model and working data in VRAM may be affected mainly during model loading.
- CPU offload, limited VRAM, frequent PCIe transfers, gaming, and real-time rendering may experience a larger impact.

Actual performance depends on the workload.

<!-- TODO(tech-review): Confirm whether the tested link width was Gen1 x4 before publishing an approximate bandwidth figure. -->

## Install the workaround

:::warning System-level workaround
This workaround changes the eGPU's PCIe link settings and temporarily removes and rescans the device during startup. Use only the files provided on this page. If your configuration differs from the tested configuration, check the [eGPU support overview](./egpu.md) before continuing.
:::

1. Download these files into the same directory:

   - <a href="/downloads/one/egpu/fix-egpu-link.sh" download>`fix-egpu-link.sh`</a>
   - <a href="/downloads/one/egpu/egpu-gen1-fix.service" download>`egpu-gen1-fix.service`</a>
   - <a href="/downloads/one/egpu/99-egpu-gen1-fix.rules" download>`99-egpu-gen1-fix.rules`</a>

2. Open a terminal in the download directory and check that `setpci` is available:

   ```bash
   command -v setpci
   ```

   The command should return a path such as `/usr/sbin/setpci`. The workaround requires `setpci`, which is provided by the `pciutils` package. If the command returns nothing, stop here and confirm the supported installation method for your Olares OS version.

3. Install the files:

   ```bash
   sudo install -m 0755 fix-egpu-link.sh       /usr/local/sbin/fix-egpu-link.sh
   sudo install -m 0644 egpu-gen1-fix.service  /etc/systemd/system/egpu-gen1-fix.service
   sudo install -m 0644 99-egpu-gen1-fix.rules /etc/udev/rules.d/99-egpu-gen1-fix.rules
   sudo systemctl enable egpu-gen1-fix.service
   ```

4. Confirm that the service is enabled:

   ```bash
   sudo systemctl is-enabled egpu-gen1-fix.service
   ```

   The command should return `enabled`.

The workaround is designed to target removable NVIDIA display controllers and exclude the built-in GPU.

<!-- TODO(tech-review):
Confirm the installation procedure, including:
- whether daemon-reload and udev rule reload are required before shutdown
- whether removable reliably distinguishes Thunderbolt eGPUs from the built-in GPU
- whether the systemd unit dependencies are valid on supported Olares OS versions
-->

## Shut down and connect the eGPU

1. Open **Settings**, then select **My hardware** > **Shutdown**.

   ![Shut down Olares One](/images/one/shut-down-olares-one.png#bordered)

2. Scan the QR code with LarePass. When prompted, tap **Confirm** to shut down Olares One.
3. Wait until Olares One is completely off.
4. Install the GPU in its enclosure and connect the enclosure to its dedicated power supply.
5. Power on the enclosure, then connect it to the Thunderbolt 5 (USB-C) port on Olares One with a certified Thunderbolt 5 cable.
6. Press the power button to start Olares One.

The workaround runs automatically during startup.

## Verify the connection

Complete both checks below. Dashboard confirms that Olares detects the eGPU. The link speed and log confirm that the workaround is active.

### Confirm eGPU detection in Dashboard

1. Log in to Olares and open **Dashboard**.
2. Select the **GPU** card. Confirm that both the built-in GPU and external GPU appear.

   ![Verify the eGPU in Dashboard](/images/one/egpu-verify.png#bordered)

### Confirm that the workaround is active

1. Confirm that both GPUs appear and find the PCI address of the external NVIDIA display controller:

   ```bash
   nvidia-smi
   lspci -nn | grep -i nvidia
   ```

   The output may also include the built-in GPU and NVIDIA audio functions. Use the address of the external display controller, not its audio device.

2. Replace `0000:0a:00.0` below with the external display controller address:

   ```bash
   cat /sys/bus/pci/devices/0000:0a:00.0/current_link_speed
   sudo tail -n 20 /var/log/egpu-gen1-fix.log
   ```

`2.5 GT/s PCIe` means Gen1 is active. A successful log includes:

```plain
after rescan: 0000:0a:00.0 speed=2.5 GT/s PCIe driver=nvidia
```

If startup stalls, power off Olares One, disconnect the eGPU, and start it again. Then see [Troubleshoot eGPU issues](./ts-egpu.md).

<!-- TODO(tech-review): Provide a reliable command for identifying the external GPU PCI address without requiring users to infer it from lspci output. -->

## Disconnect the eGPU

1. Open **Settings** > **My hardware** > **Shutdown**.
2. Confirm the shutdown in LarePass and wait until Olares One is completely off.
3. Turn off the enclosure.
4. Disconnect the Thunderbolt cable from Olares One.
5. Press the power button to start Olares One again.

## Remove the workaround

:::warning Disconnect the eGPU first
Before removing the workaround, shut down Olares One and disconnect the eGPU. Starting Olares One with the eGPU connected after removal may cause the initialization issue to return.
:::

1. Follow [Disconnect the eGPU](#disconnect-the-egpu) and start Olares One without the eGPU connected.

2. Open a terminal and remove the workaround:

   ```bash
   sudo systemctl disable egpu-gen1-fix.service
   sudo rm /usr/local/sbin/fix-egpu-link.sh
   sudo rm /etc/systemd/system/egpu-gen1-fix.service
   sudo rm /etc/udev/rules.d/99-egpu-gen1-fix.rules
   sudo systemctl daemon-reload
   sudo udevadm control --reload-rules
   ```

3. Restart Olares One:

   ```bash
   sudo reboot
   ```

The next time you connect the eGPU, it uses the default PCIe link behavior. The initialization issue may return.

## Technical background

An eGPU carries PCIe traffic through a Thunderbolt tunnel. In the tested configuration, forcing the enclosure's PCIe link to Gen1 allowed the NVIDIA driver to initialize more reliably.

In repeated cold-start testing with Olares OS 1.12.6, startup frequently stalled without the workaround. Whether this workaround is still required on later Olares OS releases has not been verified.

The underlying cause has not been confirmed. It may involve the interaction between the NVIDIA driver, PCIe link behavior, and the Thunderbolt path. In testing, Windows used a different driver stack and remained stable at Gen4 with the same enclosure and GPU.

<!-- TODO(tech-review): Confirm the cold-start sample size and results before restoring a numeric failure rate. -->
<!-- TODO(tech-review): Confirm whether current Olares OS releases still require this workaround. -->
<!-- TODO(tech-review): Confirm whether Windows Gen4 testing used the same enclosure, GPU, cable, and host. -->

## Related pages

- [Olares One eGPU support overview](./egpu.md)
- [Troubleshoot eGPU issues](./ts-egpu.md)
