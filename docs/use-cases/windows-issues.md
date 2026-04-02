---
outline: [2,3]
description: This page documents the known issues and unexpected behaviors you might encounter when using Windows on Olares, along with their corresponding solutions or workarounds.
---

# Known issues

Use this page to identify and troubleshoot currently known issues with Windows on Olares.

## Remote desktop connection fails after upgrading to Olares 1.12.5

After upgrading Olares from version 1.12.4 to 1.12.5, you might be unable to access Windows remotely after the device restarts.

This happens because of an outdated Tailscale Access Control List (ACL) setting, which controls which devices can access your Windows VM. Because this setting does not update automatically during the upgrade, it blocks remote desktop connections.

### Workaround

To restore remote access, update the access control settings and restart the required service.

1. Open Control Hub.
2. In the left sidebar, under **Terminal**, click **Olares**.
    ![Olares terminal](/images/developer/develop/controlhub-terminal.png#bordered){width=90%}

3. Run the following commands one at a time. Replace `<olaresid>` with your Olares ID, which is the part before `@` in your Olares address. For example, in `alice123@olares.com`, the Olares ID is `alice123`.

    ```bash
    kubectl annotate configmap tailscale-acl -n user-space-<olaresid> tailscale-acl-md5-

    kubectl patch deployment headscale -n user-space-<olaresid> --type=json -p='[{"op":"remove","path":"/spec/template/metadata/annotations/tailscale-acl-md5"}]'

    kubectl rollout restart sts app-service -n os-framework
    ```

4. After the restart is complete, wait a few minutes, and then try remotely connecting to Windows again.