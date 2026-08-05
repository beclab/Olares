---
outline: [2, 3]
description: Learn how to get technical support for Olares through the Assist Hub app, Olares Space, or GitHub.
head:
  - - meta
    - name: keywords
      content: Olares, technical support, system logs, GitHub issue, Assist Hub, help
---
# Get technical support

If you cannot resolve an issue using the troubleshooting guides, contact the Olares team for assistance through one of the following channels.

## Submit a ticket in the Assist Hub app

The Assist Hub app is the recommended way to get help. It has a built-in log collection feature, so you can submit a ticket directly without manually exporting logs first.

:::info Link Olares Space first
Before submitting a ticket, make sure your Olares Space account is linked to your Olares device. If not, you will be prompted to bind it in LarePass under **Settings** > **Integration**. See [Monitor Olares in Olares Space](../space/manage-olares.md) for linking instructions.
:::

### Install Assist Hub

Install the Assist Hub app from Market and log in with LarePass to submit tickets from your Olares device.

1. Open Market, search for "Assist Hub", and then install the app.
2. Open the Assist Hub app. A login QR code appears on the screen.
3. Open the LarePass app on your mobile device and go to the **Settings** tab.
4. Tap the scan icon in the upper-right corner, scan the QR code on the screen, and then tap **Confirm** to log in.

### Submit a ticket

Create a ticket with details about your issue. Assist Hub can automatically collect system information and logs from your Olares device.

1. Click **Create Ticket** in the left sidebar or at the bottom of the page.
2. Fill in the form:
   - **Issue Type**: Select the category that best matches your issue.
   - **Subcategory**: Optional. Select a more specific subcategory.
   - **Issue Title**: Enter a short summary of the issue.
   - **Description**: Describe the issue, what triggered it, and how to reproduce it. You can paste or drop images directly into the editor.
   - **Attachments**: Optional. Click the upload area or drag files into it.
3. Review **System Information**. The information is automatically collected from your Olares instance. If you do not want to include the info in the ticket, toggle off **Attach in ticket**.
4. To collect and attach system logs, expand **Collect Logs**, and then click **Collect**. Wait until the collection completes.
5. Click **Submit** to send the ticket.

### Manage tickets

After submitting a ticket, you can view its status, add replies, and update its status.

#### View ticket status

Check the status of your tickets and open any ticket to see its details.

1. Select a status from the left sidebar to filter tickets. **Home** shows all tickets, and the other sections show tickets by status.
2. Click a ticket to see its details.

#### Reply to a ticket

Use replies to add more details or answer questions from the support team.

1. Open the ticket details page.
2. Click **Add a Reply**.
3. Enter your message and click **Send reply**.

#### Close or resolve a ticket

When your issue is resolved or no longer needs follow-up, mark the ticket as closed or resolved.

1. Open the ticket details page.
2. Click **Close** or **Resolved** at the bottom of the page.

:::info
You cannot delete a submitted ticket. If you no longer need it, close it.
:::

#### Reopen a ticket

If the issue returns or was not fully fixed, you can reopen a closed or resolved ticket.

1. Open the ticket details page.
2. Click **Reopen**.
3. Add a reply explaining why you are reopening it.

The ticket status changes back to **In progress**, and the support team continues handling it.

## Submit a ticket in Olares Space

You can also create and manage support tickets directly in Olares Space through a web browser. This option does not require installing a separate Assist Hub app or having access to an Olares device, unless you want to upload logs using `olares-cli`.

For detailed steps, see [Manage support tickets in Olares Space](../space/tickets.md).

## Report on GitHub

Use the Olares GitHub repository if you prefer to report the issue publicly, or if you cannot access your Olares device.

1. Export and download your system logs from Olares Settings:

   <!--@include: ../../reusables/export-system-logs.md#export-system-logs-steps-->

2. Visit the [Olares GitHub Repository](https://github.com/beclab/Olares) and choose one of the following options:
   - Open a new **[Discussion](https://github.com/beclab/Olares/discussions/new?category=q-a)** for general questions or assistance.
   - Create a new **[Issue](https://github.com/beclab/Olares/issues/new)** for bug reports or technical problems.
3. Describe the issue and attach the exported system log file. Include the following information when applicable:
   - Steps to reproduce the issue
   - Any error messages or unexpected behaviors
   - Your environment details (operating system, Olares version, etc.)
