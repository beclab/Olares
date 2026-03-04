---
outline: [2,3]
description: Diagnose the cause when the System section in LarePass shows "System error".
---
# `System error` in LarePass

Use this guide when the **System** section in LarePass displays `System error`.
    ![System error in LarePass](/images/manual/help/ts-sys-err.png#bordered){width=90%}

## Condition

- The **System** section in LarePass shows `System error`.
- Your Olares device is accessible, but LarePass cannot retrieve system status.

## Cause

One or more system pods on the Olares device are not running normally. When this happens, LarePass cannot retrieve system status and displays `System error`.

## Solution: Identify the failing pod

Check the status of system pods and identify any pods that are not running.

1. Open a browser and access your Olares system at `https://desktop.<your-olaresID>.olares.com`.
2. Open Control Hub. In the left sidebar, under the **Terminal** section, click **Olares**.
    ![Open terminal](/images/manual/help/ts-sys-err-terminal.png#bordered){width=90%}
3. Run the following command to get the status of all pods across all namespaces:
    ```bash
    kubectl get pods -A
    ```
    ![Check pod status](/images/manual/help/ts-sys-err-pod-status.png#bordered){width=90%}
4. Check the **STATUS** column and locate any pods that are not in the `Running` state. Note down the exact NAMESPACE and NAME of the problematic pod.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}
5. Inspect the problematic pod to view detailed error messages.

    Run the following command, replacing `<namespace>` and `<pod-name>` with the values you noted in the previous step:

    ```bash
    kubectl describe pod <pod-name> -n <namespace>
    ```

    In this example:

    ```bash
    kubectl describe pod backup-66f8c76996-d7vnq -n os-framework
    ```
6. Scroll down to the **Events** section in the output to identify the cause of the failure.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-event-detail.png#bordered){width=90%}
7. Collect the command output or take a screenshot, then contact technical support via WhatsApp or email at hi@olares.com.