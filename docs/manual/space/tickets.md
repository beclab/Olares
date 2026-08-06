---
outline: [2, 3]
description: Submit support tickets, upload logs with olares-cli, and manage your requests in Olares Space.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, tickets, support, olares-cli, logs
---
# Manage support tickets in Olares Space

Use the Tickets page in Olares Space to report issues, ask for help, and track the status of your support requests.

## Create tickets

Use one of the following methods to create a support ticket in Olares Space.

### Create a ticket manually

Submit a new support request through the web form. Describe your issue, and attach screenshots, exported system logs, or other files to help the support team diagnose and resolve your issue faster.

:::tip Prepare system logs in advance
If you want to attach system logs, export them first from Olares **Settings > Advanced > Export system logs**. The export might take some time depending on the amount of logs collected. For detailed instructions, see [Export system logs](../olares/settings/developer.md#export-system-logs).
:::

1. [Log in to Olares Space](manage-accounts.md#log-in-to-olares-space).
2. On the left sidebar, click **Tickets**.
3. Click **+ New ticket**.
   ![Create a ticket in Olares Space](/images/how-to/space/create_ticket.png#bordered)
4. Fill in the form:
   - **Issue Type**: Select the category that best matches your issue.
   - **Subcategory**: Optional. Select a more specific subcategory.
   - **Issue Title**: Enter a short summary of the issue.
   - **Description**: Describe the issue, what triggered it, and how to reproduce it. Use the toolbar to format text, and paste or drop images directly into the editor.
   - **Attachments**: Optional. Click the upload area or drag files into it. You can attach up to 50 files, and each file can be up to 2 GB.
5. At the bottom of the form, choose an action:
   - **Submit**: Send the ticket to the support team.
   - **Save draft**: Save your progress without submitting.
   - **Delete draft**: Remove the current draft.

### Create a ticket automatically via olares-cli

Upload system logs directly from your Olares device. This creates a ticket automatically with the collected logs attached.

1. On the **Tickets** page, click <i class="material-symbols-outlined">terminal</i> next to **+ New ticket**.
2. In the **Upload logs via olares-cli** window, copy the **Full command line**.

   :::info Verification code
   The verification code is valid for a limited time and refreshes automatically when the countdown ends. Copy the full command line before the code expires.
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
   sudo olares-cli logs upload --code 75753956 --olares-id laresprime@olares.com
   ```

5. When the command finishes, it creates a ticket with your logs attached. Note the ticket number for follow-up.

   The following example shows running the command in the terminal inside Control Hub:

   ![Upload logs using olares-cli](/images/how-to/space/cli-upload-log.png#bordered)

6. Return to Olares Space and open the **Tickets** page.
7. Find the ticket with the number you noted in Step 5. It is named **Olares CLI logs {creation-date}** and appears in the **Open** status.

   ![CLI logs ticket](/images/how-to/space/cli-logs-ticket.png#bordered)

## Manage tickets

Use the ticket list to track the status and details of your existing support requests, and update tickets as needed.

### View your tickets

1. To view all your tickets, open the **Tickets** page. The list shows the title, status, ticket number, issue type, and creation time for each ticket.
2. To filter tickets by status, click the status drop-down list and select a state:
   - **All**: Every ticket.
   - **Draft**: Tickets you saved but have not submitted yet.
   - **Open**: Tickets waiting for the first response.
   - **In progress**: Tickets the support team is actively handling.
   - **Resolved**: Tickets that have been resolved.
   - **Closed**: Tickets that have been closed.

### Reply to a ticket

Use replies to add more details, answer questions, or provide updates.

1. On the **Tickets** page, click the ticket you want to update.
2. Click **Add a Reply** at the bottom of the ticket details page.
3. In the **Add a Reply** dialog, enter your message in the **Description** field. You can also attach files.
4. Click **Send reply**.

### Close or resolve a ticket

When the issue is fixed or no longer needs follow-up, mark the ticket as **Resolved** or **Closed**:

1. Open the ticket details page.
2. Click **Resolved** or **Close** at the bottom of the page.

### Reopen a ticket

You can reopen a **Closed** or **Resolved** ticket if the issue returns or was not fully fixed.

1. Open the ticket details page.
2. Click **Reopen** at the bottom of the page.
3. Add a reply explaining why you are reopening it.

The ticket status changes back to **In progress**, and the support team continues handling it.

## FAQs

### How are tickets associated with my Olares ID?

The **Tickets** page in Olares Space only shows tickets created with the Olares ID you are currently logged in with. To view a specific ticket, log in to Olares Space using the same Olares ID that was used to create it.

### Can I edit a ticket after submitting it?

No. You cannot edit the title or description after submission. You can only add replies to provide additional information.

### Can I delete a ticket?

No. Submitted tickets cannot be deleted. You can close them instead.
