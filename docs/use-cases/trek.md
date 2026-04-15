---
outline: [2, 3]
description: Plan trips collaboratively with TREK on Olares. Create itineraries, manage budgets, share with friends, and export travel plans as PDFs.
head:
  - - meta
    - name: keywords
      content: Olares, TREK, trip planner, travel planning, collaborative, itinerary, budget, packing list, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-16"
---

# Plan trips collaboratively with TREK

TREK is a self-hosted, real-time collaborative trip planner. It combines interactive maps, detailed itineraries, budgeting, packing lists, and team features into a single app. Running TREK on Olares keeps all your travel data private while letting you plan trips together with friends and family.

## Learning objectives

In this guide, you will learn how to:
- Install and set up TREK on Olares.
- Build trip plans, including daily schedules, budgets, and packing lists.
- Invite friends and collaborate on travel plans in real time.
- Secure your account and back up your travel data.
- Configure advanced settings, such as third-party single sign-on (SSO) and map API keys.

## Install TREK

1. Open Market and search for "TREK".

   ![TREK](/images/manual/use-cases/trek.png#bordered)

2. Click **Get**, and then click **Install**. 
3. When prompted, set the environment variables:
   - **ADMIN_EMAIL**: Your admin email address.
   - **ADMIN_PASSWORD**: Your admin password.
   
   :::info Password requirements
   The password must be at least 8 characters and include uppercase letters, lowercase letters, and numbers.
   :::

4. Click **Confirm** and wait for the installation to finish.

## Set up TREK

1. Open TREK from then Launchpad, and then sign in with the email and password you set during installation.
2. On the first signin, TREK requires you to reset your password. Enter a new password, and then click **Update password**.

## Use TREK

### Create a trip plan

1. On the home page, click **Create First Trip**.

   ![Create first trip](/images/manual/use-cases/trek-create-trip.png#bordered)

2. Specify the trip details.
   - **Cover Image**: Upload a cover image for your trip.
   - **Title**: Specify the name of the trip, such as `Paris Summer 2026`.
   - **Description**: Enter a description for the trip, such as the overal theme or goal.
   - **Dates**: Select the start and end dates for your trip.
   - **Number of Days**: Select the duration of the trip.

3. Click **Create New Trip**. The trip appears on the **My Trips** page.

   ![First trip created](/images/manual/use-cases/trek-trip-created.png#bordered)

### Plan your daily itinerary

Build a day-by-day plan by adding places and organizing them into each day's schedule.

1. Click the newly created trip to open the trip planner where you start adding places and activities.

   ![Trip planner](/images/manual/use-cases/trek-trip-planner.png#bordered)

2. Click **Add Place/Activity**.
3. Enter the location to search such as `Eiffel Tower`, click <i class="material-symbols-outlined">search</i>, select the target one from the results list, and then click **Add**. 

   The place appears on the right panel in the trip planner.

   ![Add a place](/images/manual/use-cases/trek-place-added.png#bordered)   

4. Drag the place into a specific day on your itinerary. 

   For example:
      - **Day 1**: Eiffel Tower, Trocadero Gardens
      - **Day 2**: Louvre Museum, Tuileries Garden
      - **Day 3**: Notre-Dame Cathedral, Sainte-Chapelle, Latin Quarter

5. Reorder places within a day by dragging and dropping them. 
6. Drag an activity across days to shift it to a new date.
7. Click a place to add notes, enter a visit time, or view it on the interactive map.
8. Click a place to add notes, set the visit time, or view it on Google Maps.

   ![Itinerary view](/images/manual/use-cases/trek-itinerary.png#bordered)

:::tip Route optimization
Select **Optimize** to automatically reorder places within a day for the most efficient path. You can also export the route to Google Maps for navigation.

   ![Optimize route](/images/manual/use-cases/trek-optimize-route.png#bordered){width=40%}
:::

### Check weather forecasts

Click a date in your itinerary to view the weather forecast for that destination. TREK provides up to 16-day forecasts through Open-Meteo (no API key needed), with historical climate averages as a fallback for dates further out.

![Weather forecast](/images/manual/use-cases/trek-weather.png#bordered)

### Export your itinerary as PDF

After your plan is ready, export it as a PDF to share with travel companions or print for offline reference.

1. Open the trip you want to export.
2. Click **PDF** at the top of your itinerary.

   ![Export plan as a PDF](/images/manual/use-cases/trek-export-pdf.png#bordered){width=40%}

3. In the popup window, click **Save as PDF**.   

   TREK generates a PDF with a cover page, your day-by-day itinerary, images, and notes.

### Attach travel documents

Keep booking confirmations, e-tickets, and travel insurance documents organized by attaching them directly to your itinerary items, places, or reservations. Each file supports a maximum size of 50 MB.

1. In your trip planner, click the **Files** tab.
2. Upload the files to attach.
3. In the **Assign File** window, add a note for file, and then select where to link the document, such as a specific day or place.

   ![Assign file](/images/manual/use-cases/trek-documents.png#bordered)

4. Close the window.

### Track trip expenses

Keep track of trip expenses with category-based budgeting and multi-currency support.

1. In your trip planner, click the **Budget** tab.
2. Enter a category name for your expenses, such as `Food`, `Transport`, `Accommodation`, or `Activities`.

   ![Create budget category](/images/manual/use-cases/trek-budget-category.png#bordered)

3. Click <i class="material-symbols-outlined">add_2</i>. The budget planner is displayed.

   ![Budget planner](/images/manual/use-cases/trek-budget-table.png#bordered)

4. Specify your preferred currency from the drop-down menu in the top-right corner.
5. Specify the details for the expense:

   - **NAME**: Enter the item name, such as `Dinner cruise on the Seine`.
   - **TOTAL**: Enter the total cost.
   - **PERSONS**: Enter the number of people sharing the cost.
   - **DAYS**: Enter the duration of the expense.
   - **DATE**: Enter the date of the expense.
   - **NOTE**: Enter additional context.

6. Select <i class="material-symbols-outlined">add</i> at the end of the row to add the entry. 

   TREK automatically calculates the **PER PERSON**, **PER DAY**, and **P. P / DAY** amounts, and updates your total budget on the right.

7. To add more expense catetory, enter the category name on the right panel, and then click <i class="material-symbols-outlined">add</i> next to it.

   TREK displays a pie chart breakdown of your spending by category.

   ![Budget management](/images/manual/use-cases/trek-budget.png#bordered)

### Log reservations

Keep track of your flights, accommodations, restaurants, and tour bookings in one place.

1. In your trip planner, click the **Book** tab.
2. Click **Manual Booking** to open the **New Reservation** window.
3. Select a **BOOKING TYPE**, such as **Flight**.
4. Specify the reservation details. For example, for a hotel stay:

   - **TITLE**: Enter the name of the reservation, such as Hotel Le Meurice.
   - **LINK TO DAY ASSIGNMENT**: Select a specific day in your itinerary to link this booking.
   - **DATE and END DATE**: Specify your check-in and check-out dates.
   - **STATUS**: Select the current state of the booking, such as Pending or Confirmed.
   - **LOCATION / ADDRESS**: Enter the hotel's address.
   - **BOOKING CODE**: Enter your confirmation number.
   - **FILES**: Select **Attach file** to upload your booking confirmation or e-ticket.
   - **PRICE** and **BUDGET CATEGORY**: enter the total cost to automatically sync this reservation with your trip budget.

   <!--![Reservations](/images/manual/use-cases/trek-reservations.png#bordered)-->

5. Click **Add**.

### Build packing lists

Create categorized packing lists, assign responsibilities, and track your packing progress.

1. In your trip planner, click the **Lists** tab.
2. Click **Add category**, enter a catetory name such as `Clothing`, `Electronics`, or `Toiletries`, and then click <i class="material-symbols-outlined">check</i> at the end of the row.
3. Under your new category, enter the items to pack such as `Walking shoes` and specify the quantity for each item.
4. To assign the category to a specific travel companion, click <i class="material-symbols-outlined">person_add</i>.
5. Select the checkbox next to an item as you pack it. TREK updates your overall packing progress at the top of the page.

   ![Packing list](/images/manual/use-cases/trek-packing-list.png#bordered)

6. To save time on future trips, select **Save as template** in the top-right corner to save your current list. When planning your next trip, click **Apply template** to load a saved template, so you can start with a pre-populated checklist.

### Add trip notes

Jot down daily reminders, travel ideas, or specific plans on your itinerary.

1. In your trip planner, click the **Plan** tab.
2. Locate the specific day where you want to add a note, and then click <i class="material-symbols-outlined">docs</i>.
3. Select an icon that matches the theme of your note.
4. In the **Note** field, enter a short title or summary, such as `Buy Metro tickets`.
5. In the **Daily Note** field, enter additional details, such as `Get a carnet of 10 tickets at the station before heading to the Louvre`.
6. Click **Add**.

   ![Add notes to days](/images/manual/use-cases/trek-add-note.png#bordered){width=40%}

## Collaborate with others

### Invite members to a trip

:::info External access
To invite people outside your Olares network, first set the **Authentication level** of the app to **Public** in **Settings** > **Applications** > **TREK**.

![Authentication level of TREK](/images/manual/use-cases/trek-auth-level.png#bordered){width=70%}
:::

<Tabs>
<template #Option-1:-Share-an-invite-link>
Generate a read-only link so friends or family can view your itinerary without logging in.

1. Open a trip, and then click **Share** in the upper-right corner.
2. Under **Public Link**, select the trip modules you want to make visible, such as **Map & Plan**, **Bookings**, or **Packing**.
3. Click **Create link**.
4. Copy the generated link and send it to your travel companions.

![Invite link](/images/manual/use-cases/trek-invite-link.png#bordered)
</template>
<template #Option-2:-Create-a-user-account>
1. Click your user avatar in the upper-right corner, and then click **Admin**.
2. On the **Users** tab, click **Create User**.

   ![Create user](/images/manual/use-cases/trek-create-user.png#bordered)

3. In the **Create Users** window:

   a. Enter the new member's name, email, and password.

   b. Select the role to assign.

   c. Click **Create User**.

4. Click My Trips in the upper-left corner, and then open the trip you want to share.
5. Click **Share** in the upper-right corner.
6. In the **Share Trip** window, select the user from the **Invite User** list, and then click **Invite**.

   ![Invite user](/images/manual/use-cases/trek-invite-user.png#bordered)

   The invited member logs in and views the shared trip immediately.

   <!-- ![Share trip](/images/manual/use-cases/trek-share-trip.png#bordered) -->

   <!-- ![Synced trip](/images/manual/use-cases/trek-synced-trip.png#bordered) -->
</template>
</Tabs>

### Collaborate in real time

When members join a trip, all changes sync instantly through WebSocket. The **Collab** tab provides additional team features:
- **Chat**: Discuss plans with your travel group in real time.
- **Shared notes**: Post notes visible to all trip members.
- **Polls**: Create polls to vote on destinations, restaurants, or activities (for example, "Day 3: Versailles or Montmartre?").
- **Activity sign-ups**: Track who is joining each day's activities.

![Team collaboration](/images/manual/use-cases/trek-collaboration.png#bordered)

## Next steps

- [Trek advanced settings](trek-advanced-settings.md)

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

- [TREK on GitHub](https://github.com/mauriceboe/NOMAD)
