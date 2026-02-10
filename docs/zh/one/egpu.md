---
outline: [2,3]
description: Learn how to expand your Olares One's graphical performance by connecting an external GPU (eGPU).
head:
  - - meta
    - name: keywords
      content: eGPU, external graphics card, Thunderbolt 5, hardware expansion
---

# Connect to an external graphics card (eGPU) <Badge type="tip" text="5 min" />
Olares One supports connecting external graphics cards (eGPU) to boost performance for gaming, AI model training, etc.

:::danger Power off before connecting
Do not hot-swap the external GPU.

Olares One does not support connecting or disconnecting the external GPU while the system is running. Doing so might cause system crashes, data loss, or hardware damage.

Always shut down Olares One completely before connecting or disconnecting the device.
:::

## Prerequisites
**Hardware**<br>
- The external GPU must support the Thunderbolt 5 protocol and be compatible with Ubuntu/Linux environments.
- Ensure the external GPU has its own independent power source connected.

:::warning Compatibility
The external GPU must be compatible with Linux (Ubuntu). Unsupported devices might not be recognized by the system.
:::

## Step 1: Connect the eGPU

1. Open Settings, and select **My hardware** > **Shutdown**.
   ![Shut down Olares One](/images/one/shut-down-olares-one.png#bordered)

2. Scan the displayed QR code using the LarePass app. When the prompt appears in the app, tap **Confirm** to shut down Olares One.
3. Once Olares One is fully powered off, plug the cable from your external GPU into the USB-C port on Olares One.
4. Press the power button to turn on Olares One.

## Step 2: Verify connection
To verify that the external graphics card is recognized:
1. Log in to Olares, and open Dashboard.
2. Select the **GPU** card. You should see your external card listed alongside the internal GPU.
   ![Verify eGPU connection](/images/one/egpu-verify.png#bordered)

## Disconnect the eGPU
To safely remove the eGPU:

1. Shut down Olares One completely using the steps above.
2. Turn off the power to the external GPU.
3. Unplug the USB-C cable from Olares One.
4. Press the power button to turn on Olares One.