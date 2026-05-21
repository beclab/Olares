---
outline: [2, 3]
description: Common issues and solutions for dual-booting Olares OS with Windows on Olares One.
head:
  - - meta
    - name: keywords
      content: Olares One, dual boot, Windows, troubleshooting, display flickering, HDMI, external monitor
---

# Troubleshooting

Use this page to identify and resolve common issues when dual-booting Olares OS with Windows on Olares One.

## Monitors flicker when starting Windows with two external displays

### Condition

This issue applies if you experience all of the following:

- Olares One is set up to dual-boot Olares OS and Windows.
- Windows starts with two external monitors connected.
- One monitor is connected through HDMI, and the other is connected through a USB-C hub or adapter.
- One or both monitors flicker during Windows startup or shortly after Windows reaches the desktop.

You can confirm this issue if disconnecting and reconnecting the affected monitor restores normal display output.

### Cause

When Windows starts, it first loads a basic display driver to turn on the screens. This driver is enough to start Windows, but it may not handle high-resolution HDMI output and multi-monitor detection reliably in this dual-boot setup.

After Windows reaches the desktop, the NVIDIA driver takes over display output. However, if a monitor connection has already entered an unstable state during startup, the NVIDIA driver may not recover it automatically. Reconnecting the monitor forces Windows to detect the display again.

This is usually a temporary display initialization issue, not a sign that your monitors or GPU are damaged.

### Solution

If the monitor is currently flickering:

1. Disconnect the affected monitor from Olares One or from the USB-C hub.
2. Wait a few seconds.
3. Reconnect the monitor and wait for Windows to detect it again.

To prevent the issue from happening again:

1. Before booting into Windows, disconnect the USB-C display or hub and keep only the HDMI monitor connected.
2. Start Windows and wait until the desktop is fully loaded.
3. Connect the second monitor through the USB-C hub or adapter.