---
outline: [2, 3]
description: Install PhotoPrism on Olares to organize, browse, and share your personal photo collection with AI-powered tagging and face recognition.
head:
  - - meta
    - name: keywords
      content: Olares, PhotoPrism, photo management, self-hosted, AI tagging, face recognition, photo album
app_version: "1.0.13"
doc_version: "1.0"
doc_updated: "2026-03-30"
---

# Manage your photo library with PhotoPrism

PhotoPrism is an AI-powered photo management app built for the decentralized web. It provides a clean, intuitive interface that automatically tags your photos and recognizes faces, making it easy to organize and browse your personal photo collection on Olares.

## Learning objectives

In this guide, you will learn how to:
- Install PhotoPrism and build a photo index.
- Organize photos into albums using people, labels, dates, locations, and moments.
- Correct face recognition results manually.

## Install PhotoPrism

1. Open Market and search for "PhotoPrism".

   <!-- ![Search for PhotoPrism in Market](/images/manual/use-cases/photoprism.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Index your photos

Before you can browse or organize photos in PhotoPrism, you need to build an index. PhotoPrism periodically scans the `Pictures` directory in Files for new photos and indexes them automatically. To make newly uploaded photos available right away, you can start an index manually.

1. Open PhotoPrism. Navigate to **Settings** > **General** to adjust the interface language and feature options as needed.

2. Navigate to **Library** > **Index**, and click **Start** to build the photo index.

   <!-- ![Build photo index](/images/manual/use-cases/photoprism-index.png#bordered) -->

:::tip When to index manually
Only indexed photos are visible in PhotoPrism. If you have just uploaded new photos and want to use them immediately, start a manual index from **Library** > **Index**.
:::

## Create albums

PhotoPrism automatically categorizes your photos by people, labels, dates, locations, and moments. You can use these categories to quickly add photos to albums.

### Add photos by people

1. Select **People** from the left menu.

2. Newly recognized faces appear under the **New** tab. Click the button below each face to assign a name.

   <!-- ![Name a recognized face](/images/manual/use-cases/photoprism-people-name.png#bordered) -->

3. Select a recognized person, click the number button in the bottom-right corner, and select **Add to album** to add all photos of that person to a specific album.

   <!-- ![Add person photos to album](/images/manual/use-cases/photoprism-people-album.png#bordered) -->

4. To review or correct face recognition results, click a person's avatar to view all recognized photos.

   a. If a photo is incorrectly identified, click the photo to open its details. Select the **People** tab and click the eject button to remove the photo from that person's group.

      <!-- ![Remove incorrect face match](/images/manual/use-cases/photoprism-people-remove.png#bordered) -->

   b. If a face is missed, click **New** > **Show all new faces**, find the missing photo, and open its details. On the **People** tab, manually enter the person's name.

      <!-- ![Manually add a person](/images/manual/use-cases/photoprism-people-add.png#bordered) -->

:::info Rebuild index after manual changes
Face recognition runs after the photo index is built. If you manually change the people associated with a photo, rebuild the index to update recognition results for other photos.
:::

### Add photos by labels

1. Select **Labels** from the left menu to view the tags PhotoPrism has automatically assigned to your photos.

   <!-- ![View labels](/images/manual/use-cases/photoprism-labels.png#bordered) -->

2. Select a label, click the number button in the bottom-right corner, and select **Add to album** to add all photos with that label to a specific album.

### Add photos by date and location

1. Select **Calendar** from the left menu to view photos organized by the date they were taken.

   <!-- ![View photos by calendar](/images/manual/use-cases/photoprism-calendar.png#bordered) -->

2. Select a specific month, click the number button in the bottom-right corner, and select **Add to album** to add all photos from that period to a specific album.

3. Select **Places** from the left menu and click **Expand** to view countries or regions. PhotoPrism organizes photos by location automatically.

   <!-- ![View photos by location](/images/manual/use-cases/photoprism-places.png#bordered) -->

4. Select a location, click the number button in the bottom-right corner, and select **Add to album** to add all photos from that location to a specific album.

### Add photos by moments

Moments are smart photo collections that PhotoPrism generates automatically based on combinations of time, location, and labels.

1. Select **Moments** from the left menu to view automatically generated collections.

   <!-- ![View moments](/images/manual/use-cases/photoprism-moments.png#bordered) -->

2. Select a moment, click the number button in the bottom-right corner, and select **Add to album** to add all photos from that moment to a specific album.

:::tip How moments work
Moments are generated after you upload a certain number of photos. As you add more photos over time, existing moments update automatically.
:::
