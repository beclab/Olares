---
outline: [2,3]
description: Troubleshoot known Windows VM issues on Olares, including Remote Desktop connection failures after an upgrade, with causes and workarounds.
head:
  - - meta
    - name: keywords
      content: Olares, Windows VM, Windows on Olares issues, Windows VM troubleshooting, RDP connection failed, Windows VM known issues
---

# Windows VM known issues

Use this page to identify and troubleshoot currently known issues with Windows on Olares.

## Remote desktop connection fails after upgrading to Olares 1.12.5

After upgrading Olares from version 1.12.4 to 1.12.5, you might be unable to access Windows remotely after the device restarts.

This happens because of an outdated Tailscale Access Control List (ACL) setting, which controls which devices can access your Windows VM. Because this setting does not update automatically during the upgrade, it blocks remote desktop connections.

### Workaround

To restore remote access, update the access control settings and restart the required service.

1. [Connect to the Olares host via SSH](../manual/access-olares-terminal.md) with an account that has `sudo` permission.

2. Run the following commands one at a time. Replace `<username>` with the part before `@` in your Olares ID. For example, if your Olares ID is `alice123@olares.com`, then `<username>` is `alice123`.

    ```bash
    sudo kubectl annotate configmap tailscale-acl -n user-space-<username> tailscale-acl-md5-

    sudo kubectl patch deployment headscale -n user-space-<username> --type=json -p='[{"op":"remove","path":"/spec/template/metadata/annotations/tailscale-acl-md5"}]'

    sudo kubectl rollout restart sts app-service -n os-framework
    ```

3. After the restart is complete, wait a few minutes, and then try remotely connecting to Windows again.
