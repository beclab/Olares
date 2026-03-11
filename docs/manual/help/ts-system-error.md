---
outline: [2,3]
description: Diagnose and collect information when the System section in LarePass shows "System error".
---
# "System error" in LarePass

Use this guide when the **System** section in LarePass displays "System error". There can be multiple underlying causes for this message, so follow the steps below to collect diagnostic information first, and then contact the Olares team with the results.
![System error in LarePass](/images/manual/help/ts-sys-err.png#bordered){width=90%}
## Condition

- The **System** section in LarePass shows "System error".
- Your Olares device is accessible, but LarePass cannot retrieve system status.

## Cause

The "System error" message can be triggered by different underlying issues. A common cause is that one or more system pods on the Olares device are not running normally. When this happens, LarePass cannot retrieve system status and displays "System error".

## Solution

Use the built-in terminal to locate any failing pods, retrieve their error messages, and then share this information with the Olares team. This helps narrow down the possible causes and speeds up troubleshooting.

### Step 1: Identify the failing pod

Check the status of system pods and identify any pods that are not running.

1. Open a browser and access your Olares system at `https://desktop.<your-olaresID>.olares.com`.
2. Open Control Hub. In the left sidebar, under the **Terminal** section, click **Olares**.
    ![Open terminal](/images/manual/help/ts-sys-err-terminal.png#bordered){width=90%}
    
3. Run the following command to get the status of all pods across all namespaces:
    ```bash
    kubectl get pods -A
    ```
4. Check the **STATUS** column and locate any pods that are not in the `Running` state. Note down the exact **NAMESPACE** (the first column) and **NAME** (the second column) of the problematic pod.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

### Step 2: Inspect the pod error

View the detailed error message for the problematic pod.

1. Run the following command, replacing `<namespace>` and `<pod-name>` with the values you noted in the previous step:

    ```bash
    kubectl describe pod <pod-name> -n <namespace>
    ```

    In this example:

    ```bash
    kubectl describe pod backup-66f8c76996-d7vnq -n os-framework
    ```
2. Scroll down to the **Events** section in the output to identify the error message related to the failure.
    ![Locate problematic pod](/images/manual/help/ts-sys-err-pod-event-detail.png#bordered){width=90%}

### Step 3: Contact support

Create an issue in the [Olares GitHub repository](https://github.com/beclab/Olares/issues) and provide the following information:

- The output of `kubectl describe pod <pod-name> -n <namespace>`.
- A screenshot of the error message, if available.

This information helps our team investigate and resolve the issue faster.