---
outline: [2, 3]
description: Submit support tickets, upload logs with olares-cli, and manage your requests in Olares Space.
head:
  - - meta
    - name: keywords
      content: Olares support, Olares Space tickets, Olares system logs, Olares CLI logs
---
# Create and manage support tickets in Olares Space

Use the Tickets page in Olares Space to report issues, ask for help, and track the status of your support requests. Create a ticket automatically by uploading logs with the Olares CLI, or manually through the web form.

The Tickets page only shows tickets created with the Olares ID you are currently logged in with. To view a specific ticket, log in to Olares Space using the same Olares ID that was used to create it.

:::tip Prerequisites
Olares OS v1.12.6 or later is required.
:::

## Create a ticket automatically via Olares CLI

Run the command shown in Olares Space on your device terminal to upload system logs and create a ticket automatically.

1. Open [Olares Space](https://space.olares.com/) in your browser, and scan the QR code using LarePass to log in.
2. On the left sidebar, select **Tickets**, and then click <i class="material-symbols-outlined">terminal</i> .

   ![Upload logs with Olares CLI](/images/how-to/space/upload-log-cli.png#bordered)

3. In the **Upload logs with Olares CLI** window, copy the **Command**.

   :::info Verification code
   The verification code is valid for a limited time and refreshes automatically when the countdown ends. Copy the command before the code expires.
   :::

   ![Upload logs dialog](/images/how-to/space/upload_logs_window1.png#bordered)

4. Access the terminal on your Olares device. For example, [access via SSH](/one/access-terminal-ssh.md).
5. Run the copied command in the terminal. It uses Olares CLI to collect and upload system logs automatically.
6. When the command finishes, it creates a ticket with your logs attached. Note the ticket number for follow-up, such as `TKT-10324`.

   ![Upload logs using olares-cli](/images/how-to/space/cli-upload-log1.png#bordered)

7. Return to the **Tickets** page, and find the ticket with the number you noted in Step 6. It is named **Olares CLI logs {creation-date}** and appears in the **Open** status.

   ![CLI logs ticket](/images/how-to/space/cli-logs-ticket2.png#bordered)

Once submitted, you can track the ticket's progress, view replies, and communicate with the Olares Support team until your issue is resolved.

## Create a ticket manually

Submit a new support request through the web form.

1. Open [Olares Space](https://space.olares.com/) in your browser, and scan the QR code using LarePass to log in.
2. On the left sidebar, select **Tickets**, and then click **+ Create ticket**.
3. Fill in the required issue details, and then click **Create ticket**.

   ![Create a ticket in Olares Space](/images/how-to/space/create_ticket.png#bordered)

   :::tip Export system logs
   If you want to attach system logs, export them from Olares **Settings** > **Advanced** > **[Export system logs](../olares/settings/developer.md#export-system-logs)**.
   :::

## Learn more

- [Get support](../help/request-technical-support.md): Learn about other support options, such as the Ticket app and GitHub.
