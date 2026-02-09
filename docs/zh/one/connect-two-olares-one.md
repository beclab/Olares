---
outline: [2, 3]
description: A technical guide for deploying a scalable Olares cluster. Learn how to configure a master node, resolve hostname conflicts, and join worker nodes to the cluster.
head:
  - - meta
    - name: keywords
      content: Cluster, Kubernetes, Multi-node, Worker node, Master node
---
# Set up a multi-node Olares cluster <Badge type="tip" text="1.5 h" />
For advanced use cases requiring higher availability and distributed storage, you can physically connect two Olares One units to form a single, unified cluster.

## Learning objectives
- Configure a master node with distributed storage support.
- Resolve hostname conflicts between nodes.
- Join a worker node to the cluster using the Olares CLI.

## Before you begin
- The default username and password for Olares One are both `olares`.
  :::warning Reset default SSH password
  Even if you primarily use the Control Hub terminal, you must reset this password immediately in **Settings** > **My hardware** to secure your device against unauthorized access.
  :::
- SSH access grants powerful control over the system. Ensure you keep your credentials secure.

## Prerequisites
**Hardware**<br>
- Two Olares One units connected to the same local area network.
- Both units running the same version of Olares OS.
- You know the local IP addresses of both units.

**Experience**<br>
- Familiarity with terminal commands.
- Basic understanding of Kubernetes node management.

**Software**<br>
- LarePass installed on your mobile phone.

## Step 1: Set up the master node
:::danger Clean installation required
Setting up a cluster requires a clean environment. If Olares OS is already installed on this device, you must uninstall it first:
```bash
sudo olares-cli uninstall --all
```
:::

The first Olares One unit will serve as the master node.

1. Access the master node via SSH using its local IP address.
    ```bash
    ssh olares@<Master-IP-Address>
    ```
2. Initialize the local storage service MinIO  which backs the distributed file system.
    ```bash
    sudo olares-cli install storage
    ```
3. Install Olares with JuiceFS enabled. This allows data to be shared across multiple nodes.
    ```bash
    sudo olares-cli install --with-juicefs=true
    ```
4. The installation script will prompt you for your Olares ID details.

   For example, if your full Olares ID is `alice123@olares.com`:

   - **Domain name**: Press `Enter` to use the default domain name or type `olares.com`.
   - **Olares ID**: Enter the prefix of your Olares ID. In this example, enter `alice123`.

   Upon completion of the installation, the initial system information, including the Wizard URL and the initial login password, will appear on the screen. You will need them later in the activation stage.

   ![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

5. Use the Wizard URL and initial one-time password to activate. This process connects the Olares device with your Olares ID using LarePass.

   a. Enter the Wizard URL in your browser. You will be directed to the welcome page. Press any key to continue.
   ![Open wizard](/images/manual/get-started/open-wizard.png#bordered)

   b. Follow the on-screen instructions to continue the activation.

After setup is complete, the LarePass app returns to the home screen, and the Wizard redirects you to the Olares login page.

## Step 2: Set up the worker node

1. Access the worker node via SSH.
    ```bash
    ssh olares@<Worker-IP-Address>
    ```
2. Update the hostname:
    ```bash
    sudo hostnamectl set-hostname olares-worker
    ```
    :::info
    By default, all Olares One units have the hostname `olares`. Kubernetes requires unique hostnames for every node in a cluster. Before joining it to the cluster, you must ensure it has a unique hostname.
    :::

3. Verify that the worker node can communicate with the master and that versions are compatible.
    ```bash
    sudo olares-cli node masterinfo \
    --master-host=<Master-IP-Address> \
    --master-ssh-user=olares \
    --master-ssh-password=<Password>
    ```

   Example output:
    ```bash
    current: root
    2026-01-13T06:10:41.874Z        [Job] [Get Master Info] start ...
    2026-01-13T06:10:41.874Z        [Module] GetMasterInfo
    2026-01-13T06:10:42.528Z        Got master info:
    OlaresVersion: 1.12.3
    JuiceFSEnabled: true
    KubernetesType: k3s
    MasterNodeName: olares
    AllNodes: olares
    
    2026-01-13T06:10:42.528Z        [A] olares: GetMasterInfo success (654.711582ms)
    2026-01-13T06:10:42.529Z        [A] olares-worker: AddNodePrecheck success (37.084µs)
    2026-01-13T06:10:42.529Z        [Job] Get Master Info execute successfully!!! (654.871193ms)
    ```

4. Run the add node command. Ensure the `--version` matches the version found in the pre-check output (e.g., `1.12.3`):

    ```bash
    sudo olares-cli node add \
    --master-host=<Master-IP-Address> \
    --master-ssh-user=olares \
    --master-ssh-password=<Password> \
    --version=<Olares-Version>
    ```

5. Perform the same installation and activation steps for the worker node.

## Step 3: Verify the cluster

Once the join command completes, verify that the nodes are communicating correctly.

Check the status of all nodes in the Kubernetes cluster with the following command:
```bash
kubectl get nodes
```

## Resources
- [Olares CLI](../developer/install/cli/node.md): Explore the Olares CLI.
- [Olares environment variables](../developer/install/environment-variables.md): Learn about the environment variables that enable advanced configurations of Olares.