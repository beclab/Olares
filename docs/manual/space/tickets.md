---
outline: [2, 3]
description: Submit support tickets, upload logs with olares-cli, and manage your requests in Olares Space.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, tickets, support, Olares CLI, logs
---
# Create and manage support tickets in Olares Space

Use the Tickets page in Olares Space to report issues, ask for help, and track the status of your support requests.

## Create tickets

Use one of the following methods to create a support ticket in Olares Space.

### Create a ticket manually

Submit a new support request through the web form. Describe your issue, and attach screenshots, exported system logs, or other files to help the Olares Support team diagnose and resolve your issue faster.

:::tip Prepare system logs in advance
If you want to attach system logs, export them first from Olares **Settings > Advanced > Export system logs**. The export might take some time depending on the amount of logs collected. For detailed instructions, see [Export system logs](../olares/settings/developer.md#export-system-logs).
:::

1. Open [Olares Space](https://space.olares.com/) in your browser, and scan the QR code using LarePass to log in.
2. On the left sidebar, select **Tickets**.
3. Click **+ Create ticket**, and then fill in the ticket details:

   ![Create a ticket in Olares Space](/images/how-to/space/create_ticket.png#bordered)

   - **Request type**: Select the type of request that best matches your issue.
   - **Product or area**: Select the product or area related to your issue.
   - **Topic**: Optional. Select a more specific topic.
   - **Title**: Enter a short summary of the issue.
   - **Description**: Describe the issue in detail, including what triggered it, how to reproduce it, and what you expected to happen.
   - **Attachments**: Optional. Click the upload area or drag files into it.

4. At the bottom of the form, choose an action:
   - **Create ticket**: Send the ticket to the support team.
   - **Save draft**: Save your progress without submitting.
   - **Delete draft**: Remove the current draft.

### Create a ticket automatically via Olares CLI

Upload system logs directly from your Olares device. This creates a ticket automatically with the collected logs attached.

1. On the **Tickets** page, click <i class="material-symbols-outlined">terminal</i> next to **+ Create ticket**.

   ![Upload logs with Olares CLI](/images/how-to/space/upload-log-cli.png#bordered)

2. In the **Upload logs with Olares CLI** window, copy the **Command**.

   :::info Verification code
   The verification code is valid for a limited time and refreshes automatically when the countdown ends. Copy the command before the code expires.
   :::

   ![Upload logs dialog](/images/how-to/space/upload_logs_window.png#bordered){width=60%}

3. Access the terminal on your Olares device. Choose the method that fits your setup:
   - [Access Terminal in Control Hub](/one/access-terminal-control-hub.md): Use the web-based terminal in your browser.
   - [Access via SSH](/one/access-terminal-ssh.md): Connect from your computer over the network.
   - [Access physically](/one/access-physical-console.md): Log in directly on the device.
4. Run the command copied from the window:

   ```bash
   sudo olares-cli logs upload --code <verification-code> --olares-id <your-olares-id>
   ```

   Example:

   ```bash
   sudo olares-cli logs upload --code 90341739 --olares-id laresprime@olares.com
   ```

5. When the command finishes, it creates a ticket with your logs attached. Note the ticket number for follow-up, such as `TKT-10220`.

   The following example shows running the command in the terminal inside Control Hub:

   ![Upload logs using olares-cli](/images/how-to/space/cli-upload-log.png#bordered)

6. Return to the **Tickets** page, and find the ticket with the number you noted in Step 5. It is named **Olares CLI logs {creation-date}** and appears in the **Open** status.

   ![CLI logs ticket](/images/how-to/space/cli-logs-ticket.png#bordered)

## Manage tickets

After submitting a ticket, you can track its progress, view replies from the Olares Support team, and communicate by replying to resolve the issue in a timely manner.

### View your tickets

View your ticket status, and open any ticket to check replies and details.

1. To view all your tickets, open the **Tickets** page. The list shows the title, status, ticket number, request type, and creation time for each ticket.
2. Click a ticket to open its details page and view replies from the Olares Support team.
3. To filter by status, select a state from the drop-down list:
   - **All**: Every ticket.
   - **Draft**: Tickets you saved but have not submitted yet.
   - **Open**: Tickets waiting for the first response.
   - **In progress**: Tickets actively being handled.
   - **Resolved**: Tickets that have been resolved.
   - **Closed**: Tickets that have been closed.

### Reply to a ticket

Use replies to communicate with the Olares Support team, provide updates, or ask follow-up questions.

1. On the **Tickets** page, click the ticket you want to update.
2. Click **Add message** at the bottom of the ticket details page.
3. In the **New message** panel, enter your message and optionally attach files.
4. Click **Send message**.

### Close or resolve a ticket

When the issue is fixed or no longer needs follow-up, mark the ticket as **Resolved** or **Closed**:

1. Open the ticket details page.
2. Click **Mark as resolved** or **Close ticket** at the bottom of the page.

### Reopen a ticket

You can reopen a **Closed** or **Resolved** ticket if the issue returns or was not fully fixed.

1. Open the ticket details page.
2. Click **Reopen ticket** at the bottom of the page.
3. Add a message explaining why you reopened it.

The ticket status changes back to **Open**, and the Olares Support team will continue to handle it.

## FAQ

### How are tickets associated with my Olares ID?

The **Tickets** page in Olares Space only shows tickets created with the Olares ID you are currently logged in with. To view a specific ticket, log in to Olares Space using the same Olares ID that was used to create it.

## Learn more

- [Get support](../help/request-technical-support.md): Learn more ways to get help through the Ticket app or GitHub.
