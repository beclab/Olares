---
outline: [2, 3]
description: Set up your cloud based Olares instance in Olares Space. Configure system resources, monitor installation progress, and activate your instance with proper credentials.
---
# Create a cloud-based Olares

Cloud-based Olares offer a convenient deployment option, enabling you to set up an Olares environment without managing hardware directly.
This section provides a step-by-step guide to creating and managing your cloud-based Olares. 

::: tip NOTE
- The **Host Service** is currently in an invite-only beta phase. To access the cloud-based Olares, contact us for an invitation.
- If you are looking for information about setting up a self-hosted Olares, refer to the [Getting Started Guide](../manual/get-started/index.md).
:::

## Prerequisites

Ensure you get an Olares ID to log in to Olares Space and activate the Olares instance.

## Create an Olares

1. Log in to [Olares Space](https://space.olares.com/) by scanning the QR code using LarePass.
2. Navigate to the **Cluster** page and select the second option to start creating.

    ![Basic Configuration](/images/how-to/space/basic_configuration.jpg#bordered)
3. Configure the environment for installation as below:
   - **Select Cloud Provider**: Choose a cloud service provider and the data center location closest to your users or workloads.
   - **Hardware Configuration**: Select the instance's CPU, RAM, and storage resources.
   - **Olares Version & Kubernetes Setup**: Choose the appropriate version of Olares and the Kubernetes/K3S solution to be installed.

    :::tip
    If you intend to host large language models (LLMs), select the **Alibaba Cloud Hong Kong** region. Currently, it is the only region that offers instances with shared GPU services.
    :::

4. Review the fees for storage and traffic. 

    ![Storage and Network Fees](/images/how-to/space/storage_and_network.jpg#bordered)

    ::: tip NOTE
    Each instance comes with a set amount of free storage and bandwidth. If your usage exceeds these quotas, additional fees will apply based on your cloud provider's pricing. 
    :::
5. Review your order details, including instance configuration, selected options, and fees.
6. Complete the payment to initiate the installation process.

## Monitor installation

The creation and installation of your cloud-based Olares typically take around 10 minutes. During this time, you can monitor the progress and logs in real-time.

### System statuses

The installation follows several key stages, represented by different statuses:

| Status                 | Description                                                                   |
|------------------------|-------------------------------------------------------------------------------|
| **Unpaid**             | 	Instance created but pending payment. Can be canceled.                       |
| **Fetching**           | 	Payment confirmed. System is creating resources.                             |
| **Queuing**            | Resource creation request has been submitted.                                 |
| **Pending**            | Resources created. Waiting for OS installation.                               |
| **Installing**         | OS installation is in progress.                                               |
| **Restoring**          | OS restoration is in progress from backup.                                    |
| **Restore_error**      | Restoration failed because of incorrect snapshot password.                    |
| **Restarting**         | 	System is restarting.                                                        |
| **Stopping**           | 	System is shutting down.                                                     |
| **Starting**           | System is starting up.                                                        |
| **Running**            | System is operating normally. You can restart, stop, or destroy the instance. |
| **Stopped**            | System is not running. You can restart or destroy the instance.               |
| **Errored**            | System encountered an error during resource creation or installation.         |
| **Destroying**         | Instance is being destroyed.                                                  |
| **Destroyed**          | Instance has been destroyed.                                                  |
| **Canceled**           | 	Instance is terminated due to cancellation or payment issues.                |
| **Pending Activation** | 	System is waiting for activation. Will start after activations.              |

### Real-time logs

Click **Log** to view detailed logs and monitor the installation process in real time.

## Activate Olares  

When the installation enters the **Pending Activation** state, activate Olares:

1. Click **Activation**. A pop-up window will display Olares ID, one-time password, and a wizard URL.

    ![One Time Password](/images/how-to/space/one_time_password.jpg#bordered)

2. Access the wizard URL in your browser, and use the one-time password to log into Olares for the first time. 
3. Change the Olares password via LarePass when prompted. 
4. Follow the on-screen instructions to finish the rest of activation process. 