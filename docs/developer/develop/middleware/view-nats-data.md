---
outline: [2, 3]
description: Learn how to subscribe to and publish messages in Olares using NATS CLI, and understand the NATS Subject naming rules and permission model.
---
# Subscribe and publish messages with NATS

This guide explains how to use the `nats-box` CLI tool to test NATS message subscription and publication within the Olares cluster, and provides an overview of the NATS Subject naming rules and permission model.

## Get connection information

Before connecting, obtain NATS connection details from the Control Hub.

1. Open Control Hub from Launchpad.
2. In the left navigation pane, go to Middleware and select **Nats**.
3. On the Subject panel, select a target Subject and record the corresponding information from the same row:
    - **Subject**: The target message subject.
    - **User**: The connection username.
    - **Password**: The connection password.

    ![Nats details](/public/images/developer/develop/middleware/mw-nats-details.png#bordered){width=60% style="margin-left:0"}

## Access via CLI

`nats-box` provides a convenient way to test NATS subscriptions and publications from within the cluster.

### Deploy `nats-box`

1. Download the example [`nats-box.yaml`](http://cdn.olares.com/common/nats-box.yaml) file, then upload it to the Olares machine.
2. Navigate to the directory containing the YAML file and deploy `nats-box`:
    ```bash
    kubectl apply -f nats-box.yaml
    ```
3. Retrieve the name of the `nats-box` Pod:
    ```bash
    kubectl get pods -n os-platform | grep nats-box
    ```
4. Enter the `nats-box` container:
    ```bash
    kubectl exec -it -n os-platform <nats-box-pod> -- sh
    ```

### Subscribe to messages

Use the Subject, User, and Password obtained from Control Hub to subscribe:
```bash
nats sub <subject-from-controlhub> --user=<user-from-controlhub> --password=<password-from-controlhub> --all
```

### Publish messages

Publish a message to the specified Subject:
```bash
nats pub <subject-from-controlhub> '{"hello":"world"}' --user=<user-from-controlhub> --password=<password-from-controlhub>
```

## Subject naming and permission reference

This section describes the Subject naming convention and permission model used in Olares.

### Subject structure

NATS Subjects use a three-level structure separated by dots (.): `<prefix>.<event>.<olaresId>`.

| Level | Name | Description |
|--|--|--|
| 1st |`<prefix>` | Source Identifier.<br>- **System services**: Fixed as `os`.<br> - **Third-party apps**: Uses the corresponding `appId`. |
| 2nd | `<event>` | Event type or Domain. <br>Examples: `users`, `groups`, `files`, `notification`. |
| 3rd |`<olaresId>` | Represents the Olares ID of the user space. | 

### Permission model
Read and write permissions for Subjects vary depending on the application type.

| App type | Permission scope | Description |
|--|--|--|
| User space app| Read-only | Can only subscribe to Subjects with a three-level structure containing its own `<olaresId>`. |
| System/Cluster app| System-level access | **Subscribe**: Can subscribe to system-level Subjects (e.g., `os.users`, `os.groups`).<br>**Write**: Can write to second-level Subjects within its own space. <br>**Global Read**: Requires separate approval to subscribe to all second-level Subjects. |