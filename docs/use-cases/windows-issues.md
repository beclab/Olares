---
outline: [2,3]
description: This page documents the known issues and unexpected behaviors you might encounter when using Windows on Olares, along with their corresponding solutions or workarounds.
---

# Known issues

Use this page to identify and troubleshoot currently known issues with Windows on Olares.

## Windows cannot be remotely accessed after upgrading from 1.12.4 to 1.12.5

After upgrading Olares from version 1.12.4 to 1.12.5, Windows may become unavailable for remote access after the device restarts.

This issue may occur because some Tailscale ACL-related annotations are not refreshed correctly during the upgrade, which can prevent remote access to Windows.

### Workaround

Refresh the related annotations and restart the required service.

1. Open the built-in terminal in Control Hub:

   a. Open Control Hub.

   b. In the left sidebar, under **Terminal**, click **Olares**.
    ![Olares terminal](/images/developer/develop/controlhub-terminal.png#bordered){width=90%}
2. Run the following commands in order. Replace `<olaresid>` with your actual Olares ID.
    ```bash
    kubectl annotate configmap tailscale-acl -n user-space-<olaresid> tailscale-acl-md5-

    kubectl patch deployment headscale -n user-space-<olaresid> --type=json -p='[{"op":"remove","path":"/spec/template/metadata/annotations/tailscale-acl-md5"}]'
    ```
3. Restart `app-service`.
    ```bash
    kubectl rollout restart sts app-service -n os-framework
    ```

4. After the restart is complete, wait a few minutes, and then try remotely connecting to Windows again.