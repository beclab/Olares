---
outline: [2, 3]
description: Plan trips collaboratively with TREK on Olares. Create itineraries, manage budgets, share with friends, and export travel plans as PDFs.
head:
  - - meta
    - name: keywords
      content: Olares, TREK, trip planner, travel planning, collaborative, itinerary, budget, packing list, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-13"
---

# Plan trips collaboratively with TREK

TREK is a self-hosted, real-time collaborative trip planner. It combines interactive maps, detailed itineraries, budgeting, packing lists, and team features into a single app. Running TREK on Olares keeps all your travel data private while letting you plan trips together with friends and family.

## Learning objectives

In this guide, you will learn how to:
- Install TREK and set up your account
- Create and manage trip plans with itineraries, budgets, and packing lists
- Share trips and collaborate with others in real time

## Install TREK

1. Open Market and search for "TREK".

   <!-- ![TREK](/images/manual/use-cases/trek.png#bordered) -->

2. Click **Get**, then **Install**. When prompted, set the following environment variables:
   - **ADMIN_EMAIL**: Your admin email address.
   - **ADMIN_PASSWORD**: Your admin password.
   :::info Password requirements
   The password must be at least 8 characters and include uppercase letters, lowercase letters, and numbers.
   :::
3. Wait for installation to complete.

## Set up TREK

1. Open TREK from Launchpad. Log in with the email and password you set during installation.

   <!-- ![Log in to TREK](/images/manual/use-cases/trek-login.png#bordered) -->

2. On first login, TREK requires you to reset your password. Enter a new password and click **Update Password**.

   <!-- ![Reset password](/images/manual/use-cases/trek-reset-password.png#bordered) -->

## Use TREK

### Create a trip

1. On the home page, click **Create First Trip**.

   <!-- ![Create first trip](/images/manual/use-cases/trek-create-trip.png#bordered) -->

2. Enter the trip details:
   - **Trip name**: For example, `Paris Summer 2026`.
   - **Destination**: Enter your destination, such as `Paris, France`.
   - **Dates**: Select the start and end dates for your trip.

3. Click **Create**. TREK opens the trip planner where you can start adding places and activities.

### Plan your itinerary

Build a day-by-day plan by adding places and organizing them into each day's schedule.

1. In your trip, click **Add Place** to search for a location. For example, search for `Eiffel Tower` and select it from the results.

2. Drag the place into a specific day on your itinerary. For a week in Paris, you might organize it like this:
   - **Day 1**: Eiffel Tower, Trocadero Gardens
   - **Day 2**: Louvre Museum, Tuileries Garden
   - **Day 3**: Notre-Dame Cathedral, Sainte-Chapelle, Latin Quarter

3. Reorder places within a day by dragging and dropping them. TREK also supports cross-day moves, so you can easily shift an activity to another day.

4. Click on a place to add notes, set a visit time, or view it on the interactive map.

   <!-- ![Itinerary view](/images/manual/use-cases/trek-itinerary.png#bordered) -->

:::tip Route optimization
Click **Optimize Route** to automatically reorder places within a day for the most efficient route. You can also export the route to Google Maps for navigation.
:::

### Check weather forecasts

Click on a date in your itinerary to view the weather forecast for that destination. TREK provides up to 16-day forecasts through Open-Meteo (no API key needed), with historical climate averages as a fallback for dates further out.

<!-- ![Weather forecast](/images/manual/use-cases/trek-weather.png#bordered) -->

### Export your itinerary as PDF

Once your plan is ready, export it as a PDF to share with travel companions or print for offline reference.

1. Open the trip you want to export.
2. Click **PDF**.
3. TREK generates a PDF with a cover page, your day-by-day itinerary, images, and notes.

<!-- ![Export as PDF](/images/manual/use-cases/trek-export-pdf.png#bordered) -->

### Manage travel documents

Attach booking confirmations, e-tickets, travel insurance documents, and other PDFs to specific itinerary items, places, or reservations. Each file can be up to 50 MB.

1. Navigate to the trip's **Files** tab.
2. Click **Upload** and select the file to attach.
3. Choose where to link the document (a specific day, place, or reservation).

<!-- ![Document management](/images/manual/use-cases/trek-documents.png#bordered) -->

### Track your budget

Keep track of trip expenses with category-based budgeting and multi-currency support.

1. Navigate to the trip's **Budget** tab.
2. Click **Add Expense** and fill in the details:
   - **Category**: Select a category such as `Food`, `Transport`, `Accommodation`, or `Activities`.
   - **Amount**: Enter the expense amount and currency (for example, `45 EUR` for a Seine river cruise).
   - **Description**: Add a brief note, such as `Seine dinner cruise`.

3. TREK displays a pie chart breakdown of your spending by category and calculates per-person and per-day costs.

<!-- ![Budget management](/images/manual/use-cases/trek-budget.png#bordered) -->

### Manage reservations

Track flights, hotels, restaurants, and activity bookings in one place.

1. Navigate to the trip's **Bookings** tab.
2. Click **Manual Booking** and enter the details:
   - **Type**: Select `Flight`, `Hotel`, `Restaurant`, or `Activity`.
   - **Details**: For example, for a hotel reservation: hotel name, check-in/check-out dates, confirmation number.
   - **Status**: Mark it as `Confirmed`, `Pending`, or `Cancelled`.

3. Optionally attach a confirmation document to the reservation.

<!-- ![Reservations](/images/manual/use-cases/trek-reservations.png#bordered) -->

### Create packing lists

Build categorized packing lists with item assignments and progress tracking.

1. Navigate to the trip's **Packing List** tab.
2. Click **Add Item** and enter what to pack. Organize items by category such as `Clothing`, `Electronics`, or `Toiletries`.
3. Assign items to specific travelers if you are planning with others.
4. Check off items as you pack them. TREK shows your overall packing progress.

<!-- ![Packing list](/images/manual/use-cases/trek-packing-list.png#bordered) -->

To create reusable packing templates for future trips:

1. Navigate to **Admin** > **Configuration** > **Packing Templates**.
2. Create a new template with pre-defined categories and items.
3. When creating a new trip's packing list, apply the template to start with a pre-populated checklist.

<!-- ![Packing templates](/images/manual/use-cases/trek-packing-templates.png#bordered) -->

### Take notes

Add notes to individual days with timestamps and icon tags. You can create custom note categories to organize different types of travel notes.

1. Navigate to the trip's **Notes** tab.
2. To add a note category, click **Setting**, enter a category name (for example, `Restaurant Tips` or `Local Phrases`), click **+**, then click **Save**.

   <!-- ![Notes categories](/images/manual/use-cases/trek-notes-categories.png#bordered) -->

3. Click **Add Note**, select a category, and write your note.
4. Drag and drop notes to reorder them.

<!-- ![Notes drag and drop](/images/manual/use-cases/trek-notes-reorder.png#bordered) -->

## Collaborate with others

### Invite members to a trip

:::info External access
To let people outside your Olares network collaborate, first set the app's entrance to **Public** in Olares **Settings** > **Applications** > **TREK**.
:::

You can invite people to a trip in two ways:

#### Option 1: Share an invite link

1. Open a trip and click **Share** in the top-right corner.
2. Click **Create Invite Link**.
3. Configure the maximum number of uses and an expiration time.
4. Copy the link and send it to your travel companions. They can register and join the trip through this link.

<!-- ![Invite link](/images/manual/use-cases/trek-invite-link.png#bordered) -->

#### Option 2: Create a user account

1. Navigate to **Admin** > **Users** > **Create User**.
2. Enter the new member's name, email, and password, then click **Create**.

   <!-- ![Create user](/images/manual/use-cases/trek-create-user.png#bordered) -->

3. Open the trip you want to share and click **Share** > **Invite User**.
4. Select the user from the list and click **Invite**.

   <!-- ![Share trip](/images/manual/use-cases/trek-share-trip.png#bordered) -->

   <!-- ![Invite user](/images/manual/use-cases/trek-invite-user.png#bordered) -->

The invited member can log in and see the shared trip immediately.

<!-- ![Synced trip](/images/manual/use-cases/trek-synced-trip.png#bordered) -->

### Collaborate in real time

Once members join a trip, all changes sync instantly through WebSocket. The **Collab** tab provides additional team features:

- **Chat**: Discuss plans with your travel group in real time.
- **Shared notes**: Post notes visible to all trip members.
- **Polls**: Create polls to vote on destinations, restaurants, or activities (for example, "Day 3: Versailles or Montmartre?").
- **Activity sign-ups**: Track who is joining each day's activities.

<!-- ![Team collaboration](/images/manual/use-cases/trek-collaboration.png#bordered) -->

## Set up OIDC single sign-on

TREK supports third-party login through Google, Apple, Authentik, Keycloak, or any OIDC provider. The following example uses Google.

1. Go to the [Google Cloud Console](https://console.cloud.google.com/auth/clients) and create an OAuth client. Set the following:
   - **Authorized JavaScript origins**: `https://<your-trek-domain>`
   - **Authorized redirect URIs**: `https://<your-trek-domain>/oidc/callback`

   <!-- ![Google OAuth client](/images/manual/use-cases/trek-google-oauth.png#bordered) -->

2. After creating the client, copy the **Client ID** and **Client Secret**.

   <!-- ![Client credentials](/images/manual/use-cases/trek-client-credentials.png#bordered) -->

3. In TREK, navigate to **Admin** > **Configuration** > **OIDC** and paste the Client ID and Client Secret. Click **Save**.

   <!-- ![Paste OIDC credentials](/images/manual/use-cases/trek-oidc-config.png#bordered) -->

4. Log out. On the login page, you can now sign in with your Google account.

   <!-- ![Google login](/images/manual/use-cases/trek-google-login.png#bordered) -->

## Add Google API keys

Adding a Google API key enables place photos, ratings, and opening hours when you search for locations in your itinerary.

1. Go to the [Google Cloud Console](https://console.cloud.google.com/) and create an API key.
2. In TREK, navigate to **Admin** > **Settings** > **API Keys**, paste your Google API key, and click **Save**.

   <!-- ![API key settings](/images/manual/use-cases/trek-api-key.png#bordered) -->

3. Places you add now display photos, ratings, and opening hours.

   <!-- ![Place details](/images/manual/use-cases/trek-place-details.png#bordered) -->

## Enable two-factor authentication

TREK supports TOTP-based two-factor authentication (2FA) with apps like Google Authenticator or Authy.

### Enable 2FA

1. Navigate to **Settings** > **Two-factor authentication (2FA)** > **Set up authentication**.
2. Scan the QR code with your authenticator app and enter the generated code.
3. Click **Enable 2FA**.

   <!-- ![2FA setup](/images/manual/use-cases/trek-2fa-setup.png#bordered) -->

4. Save the backup codes displayed on screen. Store them in a safe place, then click **OK**.

   <!-- ![Backup codes](/images/manual/use-cases/trek-2fa-backup-codes.png#bordered) -->

### Disable 2FA

1. Navigate to **Settings** > **Two-factor authentication (2FA)**.
2. Enter your current password and a 2FA code from your authenticator app.
3. Click **Disable 2FA**.

   <!-- ![Disable 2FA](/images/manual/use-cases/trek-2fa-disable.png#bordered) -->

## Back up your data

TREK supports both manual and automatic backups.

- **Manual backup**: Navigate to **Admin** > **Backups** to create a new backup or upload an existing backup file.

  <!-- ![Manual backup](/images/manual/use-cases/trek-backup-manual.png#bordered) -->

- **Automatic backup**: Configure scheduled backups under **Admin** > **Backups** > **Auto Backup** settings.

  <!-- ![Auto backup](/images/manual/use-cases/trek-backup-auto.png#bordered) -->

## Set up email notifications

TREK can send email notifications for trip updates and invitations. You can either use the global SMTP configuration from Olares Settings, or configure SMTP manually within TREK.

### Option 1: Use global Olares SMTP settings

If you have already configured SMTP in Olares system environment variables, TREK automatically inherits the global configuration. No additional setup is needed.

To configure global SMTP settings:

1. Go to **Settings** > **Advanced** > **System environment variables**.

2. Add or edit the following SMTP variables:

   | Variable | Description |
   |:---------|:------------|
   | `OLARES_USER_SMTP_ENABLED` | Set to `true` to enable SMTP. |
   | `OLARES_USER_SMTP_SERVER` | SMTP server domain (for example, `smtp.gmail.com`). |
   | `OLARES_USER_SMTP_PORT` | SMTP server port. Typically `465` (SSL) or `587` (TLS). |
   | `OLARES_USER_SMTP_USERNAME` | Your email address or SMTP username. |
   | `OLARES_USER_SMTP_PASSWORD` | Your email password or app-specific password. |
   | `OLARES_USER_SMTP_FROM_ADDRESS` | Sender email address. |
   | `OLARES_USER_SMTP_SECURE` | Set to `true` to use a secure connection. |

   For each variable, click **Add environment variables**, enter the key and value, and click **Save**.

3. Click **Apply** at the bottom of the page.

<!-- ![Global SMTP settings](/images/manual/use-cases/trek-smtp-global.png#bordered) -->

### Option 2: Configure SMTP manually for TREK

If you haven't set up global SMTP, or want to use a different SMTP server for TREK:

1. In Olares, go to **Settings** > **Applications** > **TREK** and disable the global SMTP configuration.
2. In TREK, navigate to **Admin** > **Notifications**, enter your SMTP server details, and click **Save**.

<!-- ![SMTP manual configuration](/images/manual/use-cases/trek-smtp-manual.png#bordered) -->

## FAQs

### I forgot my TREK password. How do I reset it?

Ask the TREK admin to reset your password from **Admin** > **Users**. Alternatively, you can view the initial credentials set during installation:

1. In Control Hub, select the TREK project from the Browse panel.
2. Under **Deployments**, click **trek**, then click <i class="material-symbols-outlined">edit_square</i>.
3. In the YAML editor, find the `containers` section and locate the `ADMIN_EMAIL` and `ADMIN_PASSWORD` environment variables.

### Map search returns no results

TREK uses OpenStreetMap by default. For more comprehensive search results, add a Google Places API key under **Admin** > **Settings** > **API Keys**. See [Add Google API keys](#add-google-api-keys) for details.

### What is the file upload size limit?

Each file can be up to 50 MB. Supported formats include jpg, png, gif, webp, heic, pdf, doc, xls, txt, and csv.

## Learn more

- [TREK on GitHub](https://github.com/mauriceboe/NOMAD): Source code and release notes.
