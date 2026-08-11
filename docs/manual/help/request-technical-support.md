---
outline: [2, 3]
description: Learn how to get technical support for Olares through the Ticket app, Olares Space, or GitHub.
head:
  - - meta
    - name: keywords
      content: Olares, technical support, system logs, GitHub issue, Ticket, help
---
# Get technical support

If you cannot resolve an issue using the troubleshooting guides, contact the Olares Support team for assistance through one of the following channels.

## Submit tickets in the Ticket app

Install and use the Ticket app to create and manage support tickets, track their progress, and communicate with the Olares Support team directly from your Olares device. It can automatically collect system information and logs, so you don't need to export them manually.

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
   You can also export system logs manually from Olares **Settings > Advanced > Export system logs**, and then attach them in **Attachments**. For detailed steps, see [Export system logs](/manual/olares/settings/developer.md#export-system-logs).
   :::

5. Click **Create ticket**.

### Manage tickets

After submitting a ticket, you can track its progress, view replies from the Olares team, and communicate by replying to resolve the issue in a timely manner.

#### View your tickets

View your ticket status, and open any ticket to check replies and details.

1. Select a status from the left sidebar to filter tickets. **All tickets** shows all your tickets, and the other sections show tickets by status.
2. Click a ticket to see its details and view replies from the Olares Support team.

#### Reply to a ticket

Use replies to communicate with the Olares Support team, provide updates, or ask follow-up questions.

1. Open the ticket details page.
2. Click **Add message**.
3. In the **New message** panel, enter your message. You can also attach files or click **Collect logs** to include system logs.
4. Click **Send message**.

#### Close or resolve a ticket

When the issue is fixed or no longer needs follow-up, mark the ticket as **Resolved** or **Closed**:

1. Open the ticket details page.
2. Click **Close** or **Resolved** at the bottom of the page.

:::info
You cannot delete a submitted ticket. If you no longer need it, close it.
:::

#### Reopen a ticket

You can reopen a **Closed** or **Resolved** ticket if the issue returns or was not fully fixed.

1. Open the ticket details page.
2. Click **Reopen ticket**.
3. Add a message explaining why you reopened it.

The ticket status changes back to **Open**, and the Olares team will continue to handle it.

## Submit tickets in Olares Space

Create and manage support tickets directly in Olares Space through a web browser.

In Olares Space, you can create a ticket in one of the following ways:

- **Manually**: Fill in the web form and optionally attach exported system logs.
- **Automatically with Olares CLI**: Upload logs directly from the Olares terminal. A ticket is created automatically for follow-up.

For detailed steps, see [Create and manage support tickets in Olares Space](../space/tickets.md).

## Report issues on GitHub

Use the Olares GitHub repository if you prefer to report the issue publicly.

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
