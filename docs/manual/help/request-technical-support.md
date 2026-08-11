---
outline: [2, 3]
description: Learn how to get technical support for Olares through the Ticket app, Olares Space, or GitHub.
head:
  - - meta
    - name: keywords
      content: Olares, technical support, system logs, GitHub issue, Ticket, help
---
# Get technical support

If you cannot resolve an issue using the troubleshooting guides, contact the Olares team for assistance through one of the following channels.

## Submit a ticket in the Ticket app

The Ticket app has a built-in log collection feature, so you can submit a ticket directly without manually exporting logs.

:::info Link Olares Space first
Before submitting a ticket, make sure your Olares Space account is linked to your Olares device. If not, you will be prompted to bind it in LarePass under **Settings** > **Integration**. See [Monitor Olares in Olares Space](../space/manage-olares.md) for linking instructions.
:::

### Install Ticket

Install the Ticket app and log in with LarePass to submit tickets from your Olares device.

1. Open Market, search for "Ticket", and then install the app.

   ![Ticket app in Market](/images/manual/help/ticket.png#bordered)

2. When the installation finishes, open Ticket. A login QR code appears on the screen.
3. Open the LarePass app on your mobile device and go to the **Settings** tab.
4. Tap the scan icon in the upper-right corner, scan the QR code on the screen, and then tap **Confirm** to log in.

### Submit a ticket

Create a ticket with details about your issue. The app can automatically collect system information and logs from your Olares device.

1. Click **+ Create ticket** on the main page, or click <i class="material-symbols-outlined">add_circle</i> next to **All tickets** in the left sidebar.

   ![Create ticket](/images/manual/help/create-ticket.png#bordered)

2. In **Ticket details**, fill in the form:
   - **Request type**: Select the type of request that best matches your issue.
   - **Product or area**: Select the product or area related to your issue.
   - **Topic**: Optional. Select a more specific topic.
   - **Title**: Enter a short summary of the issue.
   - **Description**: Describe the issue in detail, including what triggered it, how to reproduce it, and what you expected to happen.
   - **Attachments**: Optional. Click the upload area or drag files into it.
3. In **System information**, review the device information automatically collected from your Olares device. If you do not want to include it in the ticket, toggle off **Include in ticket**.
4. In **System logs**, expand the section, and then click **Collect logs** to automatically collect and attach system logs.

   When the collection finishes, the log file appears as an attachment.

   ![Collect logs](/images/manual/help/collect-logs.png#bordered)

   :::tip Prefer not to use automatic collection
   You can also export system logs manually from Olares **Settings > Advanced > Export system logs**, and then attach them in **Attachments**. For detailed instructions, see [Export system logs](/manual/olares/settings/developer.md#export-system-logs).
   :::

5. Click **Create ticket**.

### Manage tickets

After submitting a ticket, you can track its progress, view replies from the Olares team, and communicate by replying to resolve the issue in a timely manner.

#### View ticket status

Check the status of your tickets and open any ticket to see its details.

1. Select a status from the left sidebar to filter tickets. **All tickets** shows all your tickets, and the other sections show tickets by status.
2. Click a ticket to see its details.

#### Reply to a ticket

Use replies to add more details or answer questions from the Olares team.

1. Open the ticket details page.
2. Click **Add message**.
3. In the **New message** panel, enter your message. You can also attach files or click **Collect logs** to include system logs.
4. Click **Send message**.

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
2. Click **Reopen ticket**.
3. Add a message explaining why you are reopening it.

The ticket status changes back to **Open**, and the Olares team will continue to handle it.

## Submit a ticket in Olares Space

You can also create and manage support tickets directly in Olares Space through a web browser. This option does not require installing a separate Ticket app or having access to an Olares device, unless you want to upload logs using `olares-cli`.

For detailed steps, see [Manage support tickets in Olares Space](../space/tickets.md).

## Report on GitHub

Use the Olares GitHub repository if you prefer to report the issue publicly, or if you cannot access your Olares device.

1. [Export your system logs](/manual/olares/settings/developer.md#export-system-logs) from Olares Settings.
2. Visit the [Olares GitHub Repository](https://github.com/beclab/Olares) and choose one of the following options:
   - Open a new **[Discussion](https://github.com/beclab/Olares/discussions/new?category=q-a)** for general questions or assistance.
   - Create a new **[Issue](https://github.com/beclab/Olares/issues/new)** for bug reports or technical problems.
3. Describe the issue and attach the exported system log file. Include the following information when applicable:
   - Steps to reproduce the issue
   - Any error messages or unexpected behaviors
   - Your environment details (operating system, Olares version, etc.)

## FAQs

### How are tickets associated with my Olares ID?

Tickets are linked to the Olares ID used when they were created. The Ticket app only shows tickets created with your current Olares ID.

If you reinstall Olares and switch to a new Olares ID, you will not see tickets created with the previous ID in Ticket. To view those tickets, log in to Olares Space using the original Olares ID, and then find them on the **Tickets** page.

### Can I edit a ticket after submitting it?

No. You cannot edit the title or description after submission. You can add replies to provide additional information.

### Can I delete a ticket?

No. Submitted tickets cannot be deleted. Close them instead.
