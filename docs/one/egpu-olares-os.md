---
outline: [2, 3]
description: Connect an NVIDIA eGPU to Olares One, check it in Olares OS, and install the temporary Gen1 workaround when needed.
---

# Set up an eGPU on Olares OS

Use this guide to connect the eGPU. Install the Gen1 workaround only if startup stalls or the GPU is not detected.

:::danger Always power off before changing the connection
Olares OS does not support eGPU hot-plugging. Shut down Olares One before connecting or disconnecting the eGPU.
:::

## Connect the eGPU

1. Open **Settings** > **My hardware** > **Shutdown**.

   ![Shut down Olares One](/images/one/shut-down-olares-one.png#bordered)

2. Scan the QR code with LarePass, then tap **Confirm**.
3. Wait until Olares One is completely off.
4. Prepare the eGPU:

   - If the GPU is already installed, connect the device's power adapter.
   - If you use an eGPU dock or enclosure, install the desktop GPU and connect all required GPU power cables.

5. Turn on the eGPU.
6. Disconnect other high-bandwidth Thunderbolt devices. Using a certified Thunderbolt 5 cable, connect the eGPU directly to a Thunderbolt 5 (USB-C) port on Olares One.
7. Press the power button on Olares One.

## Check the connection

1. Log in to Olares and open **Dashboard**.
2. Select the **GPU** card.
3. Check that the built-in GPU and eGPU both appear.

   ![Verify the eGPU in Dashboard](/images/one/egpu-verify.png#bordered)

You can also check from the terminal:

```bash
nvidia-smi
```

If the eGPU appears in Dashboard and `nvidia-smi`, do not install the workaround.

If startup stalls or the eGPU does not appear in Dashboard or `nvidia-smi`, shut down Olares One, disconnect the eGPU, and start Olares One again. Install the workaround below before reconnecting the eGPU.

## Install the workaround if needed

The workaround sets the external GPU's PCIe link to Gen1 before the NVIDIA driver loads. It targets only Thunderbolt-connected NVIDIA GPUs and does not change the built-in GPU.

:::info Performance impact
The workaround reduces bandwidth between system memory and GPU memory. Model loading, CPU offload, gaming, and real-time rendering may be slower. It does not change the GPU's compute resources.
:::

:::warning Temporary workaround
This workaround is provided by Olares and is not an official NVIDIA fix. Install only the files linked from this page. A future Olares release will include the workaround, so manual setup will no longer be required.
:::

1. Start Olares One without the eGPU connected.
2. Download all three files to the same directory:

   - <a href="/downloads/one/egpu/fix-egpu-link.sh" download>`fix-egpu-link.sh`</a>
   - <a href="/downloads/one/egpu/egpu-gen1-fix.service" download>`egpu-gen1-fix.service`</a>
   - <a href="/downloads/one/egpu/99-egpu-gen1-fix.rules" download>`99-egpu-gen1-fix.rules`</a>

3. Open a terminal in that directory and check that `setpci` is available:

   ```bash
   command -v setpci
   ```

   The command should return a path such as `/usr/sbin/setpci`. If it returns nothing, stop and contact Olares support.

4. Install the files and enable the service:

   ```bash
   sudo install -m 0755 fix-egpu-link.sh       /usr/local/sbin/fix-egpu-link.sh
   sudo install -m 0644 egpu-gen1-fix.service  /etc/systemd/system/egpu-gen1-fix.service
   sudo install -m 0644 99-egpu-gen1-fix.rules /etc/udev/rules.d/99-egpu-gen1-fix.rules
   sudo systemctl daemon-reload
   sudo udevadm control --reload-rules
   sudo systemctl enable egpu-gen1-fix.service
   ```

5. Check the result:

   ```bash
   sudo systemctl is-enabled egpu-gen1-fix.service
   ```

   The command should return `enabled`.

6. Follow [Connect the eGPU](#connect-the-egpu) again.

## Check the workaround

After Olares One starts, check that the eGPU appears in Dashboard or `nvidia-smi`. Then check the workaround log:

```bash
sudo tail -n 20 /var/log/egpu-gen1-fix.log
```

Look for a line similar to this one:

```plain
after rescan: 0000:0a:00.0 speed=2.5 GT/s PCIe driver=nvidia
```

`2.5 GT/s PCIe` means the external GPU link is running at Gen1. The PCI address differs by system.

If the eGPU is still missing or unstable, follow [Troubleshoot eGPU issues](./ts-egpu.md).

## Disconnect the eGPU

1. Open **Settings** > **My hardware** > **Shutdown**.
2. Approve the shutdown in LarePass and wait until Olares One is completely off.
3. Turn off the eGPU.
4. Disconnect the Thunderbolt cable from Olares One.
5. Start Olares One again.

## Remove the workaround

:::warning Disconnect the eGPU first
Shut down Olares One and disconnect the eGPU before removing the workaround.
:::

1. Start Olares One without the eGPU connected.
2. Remove the workaround:

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

## Related resources

- [Connect an eGPU to Olares One](./egpu.md)
- [Troubleshoot eGPU issues](./ts-egpu.md)
