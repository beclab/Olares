---
outline: [2, 3]
description: Control your Olares instances through system monitoring, worker node management, and shared GPU solutions. Track storage usage, traffic consumption and maintain instance health.
---

# Manage Olares 

This page covers Olares management tasks in Olares Space, including monitoring system data, adding worker nodes, and managing cloud services.

## View system status

You can monitor the system status of Olares through **Olares Space**:

1. In your LarePass app, go to **Settings** > **Integration**.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner and link your Olares Space account to the Olares device. This authorizes Olares Space's access to your system data.
3. Log into [**Olares Space**](https://space.olares.com/).
4. On the **Olares** page, view **Storage usage** and **Traffic consumption** in the system panel.

![System Panel](/images/how-to/space/my_olares.jpg#bordered)

:::info
For self-hosted Olares users, it's important to monitor **Traffic statistics** for intranet penetration services, and **Storage usage** for backup services. These services may incur charges based on usage.
:::

## Add worker nodes

For cloud Olares users, you can improve performance by adding worker nodes:

1. Click <i class="material-symbols-outlined">more_horiz</i> in the upper right corner, and select **Add Worker**.
2. On the guide page, choose your preferred hardware configuration.
3. Review the fees for storage and traffic.
4. Confirm your order and submit.

## Return Olares

If you no longer need your Olares service, you can return the instance by following these steps:

1. Click <i class="material-symbols-outlined">more_horiz</i> in the upper right corner.
2. Select **Destroy Olares**.
3. Confirm the action and settle your usage:
   - If you are eligible for a refund, the amount will be credited back to your account balance.
   - If additional payment is required, please confirm and settle the payment.

## Shared GPU solution

Currently, we do not offer cloud instances that include GPUs. However, for users who need GPU capabilities, we provide a shared GPU solution via rCuda. This solution is ideal for applications like Stable Diffusion, costing approximately $0.02 per image.

::: tip NOTE
For Large Language Models (LLMs), the shared GPU solution is still under development and may require further enhancements.
:::

If you need GPU support, please reach us on [Discord](https://discord.com/invite/BzfqrgQPDK).