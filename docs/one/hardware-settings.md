---
description: View Olares One hardware details in Settings, and manage power mode, CPU frequency limit, and automatic startup.
head:
  - - meta
    - name: keywords
      content: Olares One, hardware settings, power mode, CPU frequency, automatic startup, My hardware
---

# Manage hardware settings

On the **My Olares** > **My hardware** page in Settings, you can view hardware details and manage performance options for your Olares One.

![My hardware](/images/manual/olares/my-hardware-1.12.6.png#bordered)

## View hardware details

Check information such as **Model**, **Device status**, **Device Identifier**, **CPU**, and **GPU**.

## Switch power mode

Olares One supports two performance profiles:

- **Silent mode**: Limits CPU and GPU power for quiet operation, suitable for everyday workloads.
- **Performance mode**: Enables maximum CPU and GPU performance for demanding tasks such as AI inference or gaming.

## Limit CPU frequency

Turn on this switch to limit the CPU frequency from 5.4 GHz to 5.0 GHz. Turn it off to restore the original maximum frequency.

## Set automatic startup

Turn on this switch to start the device automatically when power is connected or restored after a power outage.

:::info
Requires Olares OS 1.12.6 or later and EC firmware 1.03 or later. If either prerequisite is not met, the toggle is visible but disabled.

To check your current EC firmware version or update EC firmware, see [Manage BIOS and EC](update-firmware.md).
:::
