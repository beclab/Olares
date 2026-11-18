---
description: Use USB drives with Olares. Olares automatically mounts inserted USB storage devices, and you can access or eject them from Files.
head:
  - - meta
    - name: keywords
      content: Olares, USB drive, external storage, auto-mount, eject USB
---
# Use a USB drive

Olares automatically detects and mounts inserted USB storage devices.

- Once you insert a USB device, it will be mounted automatically. No command-line operations are required.
- Open **Files** > **External** in Olares or LarePass to access the drive. In Olares 1.12.8, members have no access by default. The Super Admin or an admin with management access to the drive must grant access through [Access settings](olares/files/manage-common-external-access.md).
- Before unplugging the drive, wait for file operations to finish. Log in to Olares as the Super Admin or an admin with management access to the drive. In **Files**, right-click the USB drive and select **Unmount**. Unplug it only after unmounting succeeds. Members must ask an admin to unmount the drive.
