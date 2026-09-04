---
outline: [2, 3]
description: Learn how to get support for Olares through the Ticket app, Olares Space, or GitHub.
head:
  - - meta
    - name: keywords
      content: Olares support, contact Olares support, Ticket app, Olares support ticket, Olares Space tickets, Olares system logs, olares-cli ticket
---

# Get support

If you cannot resolve an issue using our troubleshooting guides, contact the Olares Support team for assistance. 

## Where to get support

To get the fastest resolution, choose the option that best fits your situation:

| Option | Use when |
| :--- | :--- |
| <nobr>[Ticket](#submit-a-ticket)</nobr> | Private, one-on-one support and securely sharing system logs with our team. |
| <nobr>[GitHub Olares repo](https://github.com/beclab/Olares)</nobr> | Reporting Olares OS system issues or requesting OS features. |
| <nobr>[GitHub apps repo](https://github.com/beclab/apps)</nobr> | Reporting technical issues with apps, or requesting new apps and app updates. |
| <nobr>[Forum](https://www.olares.com/forum/)</nobr> | Sharing knowledge, finding or posting tutorials, discussing features, and getting community help. |
| <nobr>[Security reporting](https://github.com/beclab/Olares/security/advisories/new)</nobr> | Reporting system vulnerabilities. |

## Submit a ticket

You can submit a support ticket via the Ticket app or Olares Space, depending on whether you can currently access your Olares device.

:::tip Prerequisites
- **System version:** Ensure your Olares system is upgraded to v1.12.6 or later.
- **Account match:** The Olares ID used to log in to LarePass must match the account currently logged in on your Olares device.
:::

### Submit via the Ticket app

Use the Ticket app when your device is accessible.

1. Install the Ticket app from Market, and then open it.
2. Log in by scanning the QR code using LarePass.
3. Create a ticket and fill in the required issue details.
4. Review the auto-collected system information. Toggle off **Include in ticket** if you want to exclude it.
5. Expand **System logs** and click **Collect logs** to attach them to the ticket automatically.

Once submitted, you can track the ticket's progress, view replies, and communicate with the Olares Support team until your issue is resolved.

:::info
The Ticket app only displays tickets associated with the Olares ID currently logged in on your device. If you switch to a new ID, you can log in to [Olares Space](https://space.olares.com/) with your original ID to view the previous tickets.
:::

### Submit via Olares Space

Use [Olares Space](https://space.olares.com/) if your device is inaccessible or if you cannot use the Ticket app. Any tickets you previously submitted through the Ticket app will also be visible here.

In Olares Space, you can create a ticket in two ways:

- **Manually**: Fill out the web form and attach the manually [exported system logs](/manual/olares/settings/developer.md#export-system-logs).
- **Automatically (CLI)**: Select <i class="material-symbols-outlined">terminal</i>, and then run the command shown in Olares Space on your device terminal. This command uses Olares CLI to collect and upload system logs.

   ![Upload logs dialog](/images/how-to/space/upload_logs_window1.png#bordered)

  When the command finishes, it automatically creates a ticket with your logs attached. Note the ticket number, such as `TKT-10324`, and find it on the **Tickets** page to follow up.

   ![Upload logs using olares-cli](/images/how-to/space/cli-upload-log1.png#bordered)   

For more information, see [Create and manage support tickets in Olares Space](../space/tickets.md).
