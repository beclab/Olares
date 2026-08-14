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

## Where to get support

To get the fastest resolution, choose the option that best fits your situation:

| Option | Use when |
| :--- | :--- |
| [Ticket](#submit-a-ticket) | Private, one-on-one technical support and securely sharing system logs with our team. |
| [GitHub Olares repo](https://github.com/beclab/Olares) | Reporting Olares OS system issues or requesting OS features. |
| [GitHub apps repo](https://github.com/beclab/apps) | Reporting technical issues or requesting features related to apps in Olares Market. |
| Security reporting | Reporting system vulnerabilities. Please follow our [Security Policy](https://github.com/beclab/Olares/blob/main/SECURITY.md). |
| [Forum](https://www.olares.com/forum/)| Sharing knowledge, finding or posting tutorials, discussing features, and getting community help. |

## Submit a ticket

You can submit a support ticket via the Ticket app or Olares Space, depending on whether you can currently access your Olares device.

:::tip Prerequisites
- **System version:** Ensure your Olares system is upgraded to v1.12.6 or later.
- **Account match:** The Olares ID used to log in to LarePass must match the account currently logged in on your Olares device.
:::

### Submit via the Ticket app

Use the Ticket app when your device is accessible. It provides a graphical interface and automatically collects necessary logs directly from your device.

Tickets are linked to the Olares ID used at creation. The Ticket app only displays tickets associated with your current ID. If you switch to a new ID, you must log in to Olares Space with your original ID to view past tickets.

1. Install the Ticket app from Market, and then open it.
2. Log in by scanning the QR code using LarePass.
3. Create a ticket and fill in the required issue details.
4. Review the auto-collected system information. Toggle off **Include in ticket** if you want to exclude it.
5. Expand **System logs** and click **Collect logs** to attach them to the ticket automatically.

Once submitted, you can track the ticket's progress, view replies, and communicate with the Olares Support team until your issue is resolved.

### Submit via Olares Space

Use [Olares Space](https://www.olares.com/space/login?redirect=/) if your device is inaccessible or if you cannot use the Ticket app. Any tickets you previously submitted through the Ticket app will also be visible here.

In Olares Space, you can create a ticket in two ways:

- **Manually**: Fill out the web form and upload exported system logs as attachments.
- **Automatically (CLI)**: Run the command shown in Olares Space on your device terminal. This command uses Olares CLI to collect and upload system logs, and then creates a ticket automatically.

   ![Upload logs dialog](/images/how-to/space/upload_logs_window1.png#bordered)

For more information, see [Create and manage support tickets in Olares Space](../space/tickets.md).
