---
outline: [2, 3]
description: Check an Olares One that does not appear to power on or shows no display, and collect the right details before requesting support.
head:
  - - meta
    - name: keywords
      content: Olares One, no power, black screen, no display, status LED, startup troubleshooting
---

# Olares One has no power or display

Use this guide when Olares One does not appear to power on, or its status LED turns on but a connected display remains blank.

## Identify the symptom

Check the status LED before changing the system:

| What you see | What it indicates |
|---|---|
| No status LED and no fan or startup activity | The device may not be receiving power. |
| The status LED is solid white | Olares One is powered on. Continue with the display and startup checks below. |
| Olares is reachable from LarePass or a browser, but the display is blank | The system is running; focus on the monitor, cable, and display input. |

## If there is no sign of power

1. Confirm that the power cable is fully connected to Olares One and its power adapter.
2. Connect the adapter directly to a working wall outlet. If possible, verify the outlet with another device.
3. Disconnect optional USB devices and external storage, then press the power button once.
4. Observe the status LED and listen for fan or startup activity.

If there is still no sign of power, stop here and contact Olares support. Do not open the chassis or replace internal components as a generic troubleshooting step.

## If the device has power but no display

1. Make sure the monitor is powered on and set to the input used by Olares One.
2. Reseat the display cable at both ends. If available, try another cable or monitor.
3. Disconnect optional peripherals, leaving only power, the display, and a wired keyboard connected.
4. Restart Olares One. When the logo appears, or immediately after powering it on, press **Delete** repeatedly to try to enter the BIOS.

If the BIOS appears, Olares One is producing a display signal. The remaining problem may be the operating system startup rather than power or display hardware. If the device is connected to your router but Olares is not reachable, follow [Network not ready or Olares connection error](/manual/help/ts-network-not-ready.md).

If the BIOS does not appear, check whether the keyboard's **Caps Lock** indicator responds. Record the result, but do not perform a blind BIOS reset, firmware flash, operating-system reinstall, or storage removal unless Olares support confirms that the procedure applies to your exact case.

## Information to provide to support

Record:

- The device serial number.
- Status LED color and behavior.
- Whether you hear the fan or other startup activity.
- The monitor, input, and cable tested.
- Whether the BIOS appears and whether the keyboard responds.
- What happened immediately before the issue, such as a restart, power interruption, update, or display change.

Photos or a short video of the LED and display behavior can help distinguish a power problem from a startup or display problem.

