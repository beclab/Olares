---
outline: [2, 3]
description: Learn how to get support for Olares through the Ticket app, Olares Space, or GitHub.
head:
  - - meta
    - name: keywords
      content: Olares, support, help, Ticket, Olares Space Tickets, system logs, GitHub issue
---

# Get support

If you cannot resolve an issue using our troubleshooting guides, contact the Olares Support team for assistance. 

## Choose the right support channel

To get the fastest resolution, choose the channel that best fits your situation:

| Channel | Best for |
| --- | --- |
| **[Ticket](#submit-a-ticket)** | Private, one-on-one technical support and securely sharing system logs with the Olares Support team. |
| **[GitHub](#report-issues-on-github)** | Reporting technical issues or requesting features that need ongoing, public technical discussion. |
| **[Forum](https://www.olares.com/forum/)** | Community discussions, questions, and sharing experiences. |
| **[Discord](https://discord.gg/olares)** | Real-time chat and quick questions. |

## Submit a ticket

You can submit a support ticket in one of two ways, depending on whether you can currently access your Olares device.

<Tabs>
<template #Ticket-app>

Use the Ticket app when you want a graphical interface that automatically collects system information and logs directly from your device.

:::tip Prerequisites
- **System version:** Ensure your Olares system is upgraded to v1.12.6 or later.
- **Account match:** The Olares ID used to log in to LarePass must match the account currently logged in on your Olares device.
- **Ticket access:** Tickets are linked to the Olares ID used at creation. The Ticket app only displays tickets associated with your current ID. If you switch to a new ID, you must log in to Olares Space with your original ID to view past tickets.
:::

1. Install the Ticket app from Market, and then open it.
2. Log in by scanning the QR code using LarePass.
3. Create a ticket and fill in the required issue details.
4. Review the auto-collected system information. Toggle off **Include in ticket** if you want to exclude it.
5. Expand **System logs** and click **Collect logs** to attach them to the ticket automatically.

Once submitted, you can track the ticket's progress, view replies, and communicate with the Olares Support team until your issue is resolved.

</template>
<template #Olares-Space>

If you cannot access your Olares device (e.g., the system has crashed or is unresponsive), use [Olares Space](https://www.olares.com/space/login?redirect=/). Any tickets you previously submitted through the Ticket app will also be visible here.

In Olares Space, you can create a ticket in two ways:

- **Manually:** Fill out the web form and upload exported system logs as attachments.
- **Automatically (CLI):** Upload logs directly from the Olares terminal using the `olares-cli`. This automatically generates a ticket.

For detailed instructions, see [Create and manage support tickets in Olares Space](../space/tickets.md).

</template>
</Tabs>

## Report issues on GitHub

For Olares OS system issues or feature requests that benefit from public technical discussion, consider opening an issue on GitHub.

:::warning Security reporting
Follow our [Security Policy](https://github.com/beclab/Olares/blob/main/SECURITY.md). Please do not report vulnerabilities through public issues, discussions, or community channels.
:::

1. [Export your system logs](/manual/olares/settings/developer.md#export-system-logs) from Olares Settings.
2. Open the appropriate GitHub repository, and create an **Issue** for specific technical problems or open a **Discussion** for general questions.

    - **Olares OS issues:** [beclab/Olares](https://github.com/beclab/Olares)
    - **App issues:** [beclab/apps](https://github.com/beclab/apps)

3. Describe the issue in detail and attach your exported system log file. Always include:

    - Exact steps to reproduce the issue
    - Any error messages or unexpected behaviors
    - Your environment details (Operating system, Olares version, etc.)